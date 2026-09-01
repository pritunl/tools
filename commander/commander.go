// Package commander runs external commands with a timeout, captured output,
// stdin input and extra environment variables, returning errors wrapped as
// errortypes values with stack traces.
//
//	ret, err := commander.Exec(&commander.Opt{
//		Name:    "ls",
//		Args:    []string{"-la"},
//		Timeout: 5 * time.Second,
//		PipeOut: true,
//	})
//	if err != nil {
//		return err
//	}
//	fmt.Print(string(ret.Output))
package commander

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"time"

	"github.com/pritunl/tools/errors"
	"github.com/pritunl/tools/errortypes"
)

var (
	envKeyReg = regexp.MustCompile(`[^a-zA-Z0-9_]|^[0-9]`)
	envValReg = regexp.MustCompile(`[^ -~]`)
)

// Opt describes a command to run with Exec.
type Opt struct {
	// Name is the program to run, resolved through PATH as by exec.Command.
	Name string
	// Args are the arguments passed to the program, excluding the program
	// name.
	Args []string
	// Dir is the working directory for the command. If empty the calling
	// process's working directory is used.
	Dir string
	// Env holds additional environment variables appended to the calling
	// process's environment. Keys must consist of letters, digits and
	// underscores and must not begin with a digit; values must contain only
	// printable ASCII. Exec returns a *errortypes.ParseError if either
	// constraint is violated.
	Env map[string]string
	// Timeout kills the command if it has not exited within this duration.
	// Zero means no timeout.
	Timeout time.Duration
	// Input, if non-empty, is written to the command's standard input,
	// which is then closed.
	Input string
	// PipeOut captures the command's standard output into Return.Output.
	// When false, standard output is discarded unless Ignore is set.
	PipeOut bool
	// PipeErr captures the command's standard error into Return.Output,
	// interleaved with standard output. When false, standard error is
	// discarded unless Ignore is set.
	PipeErr bool
	// Ignore lists substrings which, if present in the combined output of a
	// failed command, cause the failure to be treated as success with empty
	// output. Setting Ignore forces both standard output and standard
	// error to be captured.
	Ignore []string
}

// Return holds the result of a command run by Exec. It is returned even
// when Exec fails (except when Opt is nil) so that callers can log the
// command details alongside the error.
type Return struct {
	// Name is the program that was run, copied from Opt.
	Name string
	// Args are the arguments that were passed, copied from Opt.
	Args []string
	// Dir is the working directory that was used, copied from Opt.
	Dir string
	// Timeout is the timeout that was applied, copied from Opt.
	Timeout time.Duration
	// Output is the captured standard output and/or standard error, in the
	// order it was produced. It is empty if neither stream was captured or
	// if a failure was suppressed by Opt.Ignore.
	Output []byte
	// ExitCode is the command's exit status. It is zero on success and
	// also zero when the command could not be started or was killed by the
	// timeout.
	ExitCode int
	// Error is the error returned by Exec, if any, for convenience when
	// passing the Return to a logger.
	Error error
}

// Map returns the command details and result as a map suitable for use as
// structured log fields. The "error" key is only present when Error is
// non-nil.
func (r *Return) Map() map[string]interface{} {
	m := map[string]interface{}{
		"output":    string(r.Output),
		"cmd":       r.Name,
		"dir":       r.Dir,
		"args":      r.Args,
		"timeout":   r.Timeout.String(),
		"exit_code": r.ExitCode,
	}

	if r.Error != nil {
		m["error"] = r.Error
	}

	return m
}

// Exec runs the command described by opt and waits for it to exit.
//
// Output is captured only when opt.PipeOut, opt.PipeErr or opt.Ignore is
// set. If opt.Timeout elapses the command is killed and a
// *errortypes.ExecError is returned. A non-zero exit status is reported as
// a *errortypes.ExecError with ret.ExitCode set, unless the output contains
// one of the opt.Ignore substrings, in which case the command is treated as
// successful and ret.Output is cleared. Invalid opt.Env entries produce a
// *errortypes.ParseError. A nil opt produces a *errortypes.ParseError and a
// nil ret.
//
// When err is non-nil and ret is non-nil, ret.Error is set to err.
func Exec(opt *Opt) (ret *Return, err error) {
	var wrErr error
	var buffer bytes.Buffer
	ctx := context.Background()

	if opt == nil {
		err = &errortypes.ParseError{
			errors.New("utils: Missing exec options"),
		}
		return
	}

	ret = &Return{
		Name:    opt.Name,
		Args:    opt.Args,
		Dir:     opt.Dir,
		Timeout: opt.Timeout,
	}

	if opt.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, opt.Timeout)
		defer cancel()
	}

	cmd := exec.CommandContext(ctx, opt.Name, opt.Args...)

	if opt.Dir != "" {
		cmd.Dir = opt.Dir
	}
	if len(opt.Env) > 0 {
		env := os.Environ()
		for key, val := range opt.Env {
			if envKeyReg.MatchString(key) {
				err = &errortypes.ParseError{
					errors.Newf(
						"utils: Invalid environment variable name '%s'",
						key,
					),
				}
				return
			}

			if envValReg.MatchString(val) {
				err = &errortypes.ParseError{
					errors.Newf(
						"utils: Invalid environment variable value '%s'",
						val,
					),
				}
				return
			}

			env = append(env, fmt.Sprintf("%s=%s", key, val))
		}
		cmd.Env = env
	}

	hasIgnore := len(opt.Ignore) > 0
	if opt.PipeOut || hasIgnore {
		cmd.Stdout = &buffer
	}
	if opt.PipeErr || hasIgnore {
		cmd.Stderr = &buffer
	}

	if opt.Input != "" {
		var stdin io.WriteCloser

		stdin, err = cmd.StdinPipe()
		if err != nil {
			err = &errortypes.ExecError{
				errors.Wrapf(
					err,
					"utils: Failed to get stdin in exec '%s'", opt.Name,
				),
			}
			ret.Error = err
			return
		}

		err = cmd.Start()
		if err != nil {
			_ = stdin.Close()
			err = &errortypes.ExecError{
				errors.Wrapf(err, "utils: Failed to exec '%s'", opt.Name),
			}
			ret.Error = err
			return
		}

		go func() {
			defer func() {
				wrErr = stdin.Close()
				if wrErr != nil {
					wrErr = &errortypes.ExecError{
						errors.Wrapf(
							wrErr,
							"utils: Failed to close stdin in exec '%s'",
							opt.Name,
						),
					}
				}
			}()

			_, wrErr = io.WriteString(stdin, opt.Input)
			if wrErr != nil {
				wrErr = &errortypes.ExecError{
					errors.Wrapf(
						wrErr,
						"utils: Failed to write stdin in exec '%s'",
						opt.Name,
					),
				}
				return
			}
		}()

		err = cmd.Wait()
	} else {
		err = cmd.Run()
	}

	ret.Output = buffer.Bytes()

	if ctx.Err() == context.DeadlineExceeded {
		err = &errortypes.ExecError{
			errors.Wrapf(ctx.Err(), "utils: Command '%s' timed out", opt.Name),
		}
		ret.Error = err
		return
	}

	if exitErr, ok := err.(*exec.ExitError); ok {
		ret.ExitCode = exitErr.ExitCode()
	}

	if err != nil {
		for _, ignore := range opt.Ignore {
			if bytes.Contains(ret.Output, []byte(ignore)) {
				err = nil
				ret.Output = []byte{}
				break
			}
		}
	}

	if err == nil && wrErr != nil {
		err = wrErr
	}

	if err != nil {
		err = &errortypes.ExecError{
			errors.Wrapf(err, "utils: Failed to exec '%s'", opt.Name),
		}
		ret.Error = err
		return
	}

	return
}
