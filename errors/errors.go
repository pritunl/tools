// Copyright (c) 2014 Dropbox, Inc.
// All rights reserved.

// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are met:

// 1. Redistributions of source code must retain the above copyright notice, this
// list of conditions and the following disclaimer.

// 2. Redistributions in binary form must reproduce the above copyright notice,
// this list of conditions and the following disclaimer in the documentation
// and/or other materials provided with the distribution.

// 3. Neither the name of the copyright holder nor the names of its contributors
// may be used to endorse or promote products derived from this software without
// specific prior written permission.

// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
// ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
// WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
// DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE
// FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
// DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR
// SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER
// CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY,
// OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
// OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

// Package errors creates and wraps errors that carry a stack trace captured
// at the point of creation.
//
// Errors created with New, Newf, Wrap and Wrapf implement StackError, which
// extends the built-in error interface with access to the message, the
// wrapped inner error and the captured stack. Error strings include the
// messages of every wrapped error followed by the stack trace of the
// innermost StackError.
//
//	err = errors.Wrapf(err, "config: Failed to open '%s'", path)
//
// The package name intentionally mirrors the standard library "errors"
// package. StackError.Unwrap is compatible with the standard library's
// errors.Is and errors.As.
//
// This package is derived from the Dropbox godropbox errors package.
package errors

import (
	"bytes"
	"fmt"
	"reflect"
	"runtime"
	"sync"
)

// StackError is an error that records the call stack at the point it was
// created and may wrap an inner error. Values returned by New, Newf, Wrap
// and Wrapf implement StackError.
type StackError interface {
	// GetMessage returns this error's own message, without the messages of
	// wrapped errors and without the stack trace.
	GetMessage() string

	// Unwrap returns the wrapped error, or nil if this error does not wrap
	// another error. It follows the standard library convention so that
	// errors.Is and errors.As work with StackError values.
	Unwrap() error

	// Error implements the built-in error interface. It returns the
	// messages of this error and all wrapped errors, one per line, followed
	// by the stack trace of the innermost StackError.
	Error() string

	// StackAddrs returns the program counters of the captured stack as a
	// space-separated list of hexadecimal addresses. It does not resolve
	// stack frames and is therefore much cheaper than GetStack.
	StackAddrs() string

	// StackFrames returns the resolved frames of the captured stack. The
	// frames are resolved once and cached.
	StackFrames() []runtime.Frame

	// GetStack returns the captured stack as text, with each frame
	// formatted as the function name on one line followed by an indented
	// file:line and program counter:
	//
	//	main.main
	//		/home/user/main.go:13 +0x84
	//
	// The format is not stable and should not be parsed; use StackFrames
	// to inspect frames programmatically.
	GetStack() string
}

// baseError is the StackError implementation returned by New, Newf, Wrap
// and Wrapf.
type baseError struct {
	msg   string
	inner error

	stack       []uintptr
	framesOnce  sync.Once
	stackFrames []runtime.Frame
}

// GetMessage returns the message of err without stack trace information.
// For a StackError the messages of all wrapped errors are included, one per
// line. For any other error the result of its Error method is returned. If
// err is not an error a fixed placeholder string is returned.
func GetMessage(err interface{}) string {
	switch e := err.(type) {
	case StackError:
		return extractFullErrorMessage(e, false)
	case runtime.Error:
		return runtime.Error(e).Error()
	case error:
		return e.Error()
	default:
		return "Passed a non-error to GetMessage"
	}
}

// Error returns the messages of e and every wrapped error, followed by the
// stack trace of the innermost StackError in the chain.
func (e *baseError) Error() string {
	return extractFullErrorMessage(e, true)
}

// GetMessage returns e's own message.
func (e *baseError) GetMessage() string {
	return e.msg
}

// Unwrap returns the wrapped error, if any.
func (e *baseError) Unwrap() error {
	return e.inner
}

// StackAddrs returns the captured program counters as hexadecimal
// addresses separated by spaces.
func (e *baseError) StackAddrs() string {
	buf := bytes.NewBuffer(make([]byte, 0, len(e.stack)*8))
	for _, pc := range e.stack {
		fmt.Fprintf(buf, "0x%x ", pc)
	}
	bufBytes := buf.Bytes()
	return string(bufBytes[:len(bufBytes)-1])
}

// StackFrames resolves and returns the captured stack frames, caching the
// result for subsequent calls.
func (e *baseError) StackFrames() []runtime.Frame {
	e.framesOnce.Do(func() {
		e.stackFrames = make([]runtime.Frame, 0, len(e.stack))
		frames := runtime.CallersFrames(e.stack)
		for more := true; more; {
			var f runtime.Frame
			f, more = frames.Next()
			e.stackFrames = append(e.stackFrames, f)
		}
	})
	return e.stackFrames
}

// GetStack returns the captured stack formatted as text.
func (e *baseError) GetStack() string {
	stackFrames := e.StackFrames()
	buf := bytes.NewBuffer(make([]byte, 0, 256))
	for _, frame := range stackFrames {
		_, _ = buf.WriteString(frame.Function)
		_, _ = buf.WriteString("\n")
		fmt.Fprintf(buf, "\t%s:%d +0x%x\n",
			frame.File, frame.Line, frame.PC)
	}
	return buf.String()
}

// New returns a StackError with the given message and the stack trace of
// the caller.
func New(msg string) StackError {
	return newBaseError(nil, msg)
}

// Newf is like New but formats the message according to a format specifier
// in the manner of fmt.Sprintf.
func Newf(format string, args ...interface{}) StackError {
	return newBaseError(nil, fmt.Sprintf(format, args...))
}

// Wrap returns a StackError with the given message and the stack trace of
// the caller that wraps err. err may be nil, in which case the result is
// equivalent to New(msg).
func Wrap(err error, msg string) StackError {
	return newBaseError(err, msg)
}

// Wrapf is like Wrap but formats the message according to a format
// specifier in the manner of fmt.Sprintf.
func Wrapf(err error, format string, args ...interface{}) StackError {
	return newBaseError(err, fmt.Sprintf(format, args...))
}

// newBaseError creates a baseError wrapping err with the given message and
// the stack of the caller's caller. It must be called directly from an
// exported constructor so that the constructor's frame is skipped; any
// additional level of indirection will appear in the captured stack.
func newBaseError(err error, msg string) *baseError {
	var stackBuf [200]uintptr
	stackLength := runtime.Callers(3, stackBuf[:])
	stack := make([]uintptr, stackLength)
	copy(stack, stackBuf[:stackLength])
	return &baseError{
		msg:   msg,
		stack: stack,
		inner: err,
	}
}

// extractFullErrorMessage builds the full message for e by joining the
// messages of e and every wrapped error with newlines. If includeStack is
// true the stack trace of the innermost StackError in the chain is
// appended.
func extractFullErrorMessage(e StackError, includeStack bool) string {
	var ok bool
	var lastDbxErr StackError
	errMsg := bytes.NewBuffer(make([]byte, 0, 1024))

	dbxErr := e
	for {
		lastDbxErr = dbxErr
		errMsg.WriteString(dbxErr.GetMessage())

		innerErr := dbxErr.Unwrap()
		if innerErr == nil {
			break
		}
		dbxErr, ok = innerErr.(StackError)
		if !ok {
			// We have reached the end and traveresed all inner errors.
			// Add last message and exit loop.
			errMsg.WriteString("\n")
			errMsg.WriteString(innerErr.Error())
			break
		}
		errMsg.WriteString("\n")
	}
	if includeStack {
		errMsg.WriteString("\nORIGINAL STACK TRACE:\n")
		errMsg.WriteString(lastDbxErr.GetStack())
	}
	return errMsg.String()
}

// unwrapError returns the error wrapped by ierr, or nil if there is none.
// For a StackError it uses Unwrap; for other errors it reflectively reads a
// field named Err, which is the convention used by standard library error
// types such as *os.PathError and *net.OpError.
func unwrapError(ierr error) (nerr error) {
	// Internal errors have a well defined bit of context.
	if dbxErr, ok := ierr.(StackError); ok {
		return dbxErr.Unwrap()
	}

	// At this point, if anything goes wrong, just return nil.
	defer func() {
		if x := recover(); x != nil {
			nerr = nil
		}
	}()

	// Go system errors have a convention but paradoxically no
	// interface.  All of these panic on error.
	errV := reflect.ValueOf(ierr).Elem()
	errV = errV.FieldByName("Err")
	return errV.Interface().(error)
}

// RootError repeatedly unwraps ierr until an error that wraps nothing is
// reached and returns it. Both StackError wrapping and the standard library
// Err-field convention are followed. To guard against cycles, unwrapping
// stops after 20 levels and an error describing the failure is returned.
func RootError(ierr error) (nerr error) {
	nerr = ierr
	for i := 0; i < 20; i++ {
		terr := unwrapError(nerr)
		if terr == nil {
			return nerr
		}
		nerr = terr
	}
	return fmt.Errorf("too many iterations: %T", nerr)
}

// RootDropboxError returns the innermost StackError in the chain starting
// at dbxErr. Its stack trace is the one closest to where the failure
// originated and is usually the most useful one to report.
func RootDropboxError(dbxErr StackError) StackError {
	for {
		innerErr := dbxErr.Unwrap()
		if innerErr == nil {
			break
		}
		innerDBXErr, ok := innerErr.(StackError)
		if !ok {
			break
		}
		dbxErr = innerDBXErr
	}
	return dbxErr
}

// IsError reports whether err is, or wraps, errConst. It first compares the
// values directly and otherwise compares the string form of RootError(err)
// with that of errConst, so that sentinel errors can be matched through
// layers of wrapping regardless of whether they were stored by value or by
// pointer.
func IsError(err, errConst error) bool {
	if err == errConst {
		return true
	}
	// Must rely on string equivalence, otherwise a value is not equal
	// to its pointer value.
	rootErrStr := ""
	rootErr := RootError(err)
	if rootErr != nil {
		rootErrStr = rootErr.Error()
	}
	errConstStr := ""
	if errConst != nil {
		errConstStr = errConst.Error()
	}
	return rootErrStr == errConstStr
}

// FindWrappedError searches the chain of wrapped errors starting at topErr
// for one selected by classifier. classifier is called with each error in
// turn, outermost first, together with topErr, until it returns a non-nil
// error; that error is returned along with true. If classifier never
// selects an error, or topErr is nil, topErr and false are returned. Only
// StackError wrapping is followed.
func FindWrappedError(
	topErr error,
	classifier func(curErr, topErr error) error,
) (error, bool) {
	for curErr := topErr; curErr != nil; {
		classifiedErr := classifier(curErr, topErr)
		if classifiedErr != nil {
			return classifiedErr, true
		}

		dbxErr, ok := curErr.(StackError)
		if !ok || dbxErr == nil {
			break
		}
		curErr = dbxErr.Unwrap()
		if curErr == nil {
			break
		}
	}
	return topErr, false
}
