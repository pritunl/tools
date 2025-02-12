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

type Opt struct {
	Name    string
	Args    []string
	Dir     string
	Env     map[string]string
	Timeout time.Duration
	Input   string
	PipeOut bool
	PipeErr bool
	Ignore  []string
}

type Return struct {
	Name     string
	Args     []string
	Dir      string
	Timeout  time.Duration
	Output   []byte
	ExitCode int
	Error    error
}

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
