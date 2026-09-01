// Package errortypes defines a set of error categories that wrap an
// errors.StackError. Each type embeds errors.StackError so it satisfies both
// the error and errors.StackError interfaces while allowing callers to
// classify an error by its concrete type.
//
// Errors are constructed by wrapping an errors value in a struct literal:
//
//	err = &errortypes.ReadError{
//		errors.Wrap(err, "config: Failed to read file"),
//	}
//
// and classified with a type assertion or errors.As:
//
//	var readErr *errortypes.ReadError
//	if errors.As(err, &readErr) { ... }
package errortypes

import (
	"github.com/pritunl/tools/errors"
)

// UnknownError is an error of no more specific category.
type UnknownError struct {
	errors.StackError
}

// NotFoundError reports that a requested resource does not exist.
type NotFoundError struct {
	errors.StackError
}

// ReadError reports a failure to read data from a file, stream or other
// source.
type ReadError struct {
	errors.StackError
}

// WriteError reports a failure to write data to a file, stream or other
// destination.
type WriteError struct {
	errors.StackError
}

// ParseError reports malformed or invalid input.
type ParseError struct {
	errors.StackError
}

// AuthenticationError reports failed authentication of a user or client.
type AuthenticationError struct {
	errors.StackError
}

// VerificationError reports that a signature, checksum or other integrity
// check failed.
type VerificationError struct {
	errors.StackError
}

// ApiError reports a failure returned by a remote API.
type ApiError struct {
	errors.StackError
}

// DatabaseError reports a failure of a database operation.
type DatabaseError struct {
	errors.StackError
}

// RequestError reports a failure to build or send a request.
type RequestError struct {
	errors.StackError
}

// ConnectionError reports a failure to establish or maintain a connection.
type ConnectionError struct {
	errors.StackError
}

// TimeoutError reports that an operation did not complete in time.
type TimeoutError struct {
	errors.StackError
}

// ExecError reports a failure to run an external command or a non-zero
// exit status from one.
type ExecError struct {
	errors.StackError
}

// NetworkError reports a network-level failure.
type NetworkError struct {
	errors.StackError
}

// TypeError reports a value of an unexpected type.
type TypeError struct {
	errors.StackError
}

// ErrorData is a machine-readable error payload, typically decoded from an
// API response, consisting of an error key and a human-readable message.
// When placed in logger.Fields under the key "error_data" it is rendered as
// the fields "error_key" and "error_msg".
type ErrorData struct {
	// Error is a short identifier for the error.
	Error string `json:"error"`
	// Message is a human-readable description of the error.
	Message string `json:"error_msg"`
}

// GetError converts the payload into a *ParseError whose message includes
// the error key and message.
func (e *ErrorData) GetError() (err error) {
	err = &ParseError{
		errors.Newf("error: Parse error %s - %s", e.Error, e.Message),
	}
	return
}
