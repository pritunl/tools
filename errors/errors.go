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

// This module implements functions which manipulate errors and provide stack
// trace information.
//
// NOTE: This package intentionally mirrors the standard "errors" module.
// All dropbox code should use this.
package errors

import (
	"bytes"
	"fmt"
	"reflect"
	"runtime"
	"sync"
)

// This interface exposes additional information about the error.
type StackError interface {
	// This returns the error message without the stack trace.
	GetMessage() string

	// This returns the wrapped error or nil if this error does not wrap another error per the
	// Go 2 error introspection proposal:
	// https://go.googlesource.com/proposal/+/master/design/29934-error-values.md
	Unwrap() error

	// Implements the built-in error interface.
	Error() string

	// Returns stack addresses as a string that can be supplied to
	// a helper tool to get the actual stack trace. This function doesn't result
	// in resolving full stack frames thus is a lot more efficient.
	StackAddrs() string

	// Returns stack frames.
	StackFrames() []runtime.Frame

	// Returns string representation of stack frames.
	// Stack frame formatting looks generally something like this:
	// dropbox/legacy_rpc.(*clientV4).Do
	//   /srv/server/go/src/dropbox/legacy_rpc/client.go:87 +0xbf9
	// dropbox/exclog.Report
	//   /srv/server/go/src/dropbox/exclog/client.go:129 +0x9e5
	// main.main
	//   /home/cdo/tmp/report_exception.go:13 +0x84
	// It is discouraged to parse stack frames using string parsing since it can change at any time.
	// Use StackFrames() function instead to get actual stack frame metadata.
	GetStack() string
}

// Standard struct for general types of errors.
//
// For an example of custom error type, look at databaseError/newDatabaseError
// in errors_test.go.
type baseError struct {
	msg   string
	inner error

	stack       []uintptr
	framesOnce  sync.Once
	stackFrames []runtime.Frame
}

// This returns the error string without stack trace information.
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

// This returns a string with all available error information, including inner
// errors that are wrapped by this errors.
func (e *baseError) Error() string {
	return extractFullErrorMessage(e, true)
}

// Implements DropboxError interface.
func (e *baseError) GetMessage() string {
	return e.msg
}

// Implements DropboxError interface.
func (e *baseError) Unwrap() error {
	return e.inner
}

// Implements DropboxError interface.
func (e *baseError) StackAddrs() string {
	buf := bytes.NewBuffer(make([]byte, 0, len(e.stack)*8))
	for _, pc := range e.stack {
		fmt.Fprintf(buf, "0x%x ", pc)
	}
	bufBytes := buf.Bytes()
	return string(bufBytes[:len(bufBytes)-1])
}

// Implements DropboxError interface.
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

// Implements DropboxError interface.
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

// This returns a new baseError initialized with the given message and
// the current stack trace.
func New(msg string) StackError {
	return newBaseError(nil, msg)
}

// Same as New, but with fmt.Printf-style parameters.
func Newf(format string, args ...interface{}) StackError {
	return newBaseError(nil, fmt.Sprintf(format, args...))
}

// Wraps another error in a new baseError.
func Wrap(err error, msg string) StackError {
	return newBaseError(err, msg)
}

// Same as Wrap, but with fmt.Printf-style parameters.
func Wrapf(err error, format string, args ...interface{}) StackError {
	return newBaseError(err, fmt.Sprintf(format, args...))
}

// Internal helper function to create new baseError objects,
// note that if there is more than one level of redirection to call this function,
// stack frame information will include that level too.
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

// Constructs full error message for a given DropboxError by traversing
// all of its inner errors. If includeStack is True it will also include
// stack trace from deepest DropboxError in the chain.
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

// Return a wrapped error or nil if there is none.
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

// Keep peeling away layers or context until a primitive error is revealed.
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

// Return the lowest-level DropboxError. This can be used when
// reporting the stack of the original exception to try and get the most
// relevant stack instead of the highest level stack.
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

// Perform a deep check, unwrapping errors as much as possilbe and
// comparing the string version of the error.
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

// Performs a deep check of wrapped errors to find one which is selected by the given
// classifier func.  The classifer is called on all non-nil errors found, starting with topErr,
// then on each inner wrapped error in turn until it returns non-nil which ends the scan.
// If the classifier ever returns a non-nil error, it will be returned from this function along
// with `true` to indicate something was found.  Otherwise this function will return
// `topErr, false`.
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
