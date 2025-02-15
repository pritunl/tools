package errortypes

import (
	"github.com/pritunl/tools/errors"
)

type UnknownError struct {
	errors.StackError
}

type NotFoundError struct {
	errors.StackError
}

type ReadError struct {
	errors.StackError
}

type WriteError struct {
	errors.StackError
}

type ParseError struct {
	errors.StackError
}

type AuthenticationError struct {
	errors.StackError
}

type VerificationError struct {
	errors.StackError
}

type ApiError struct {
	errors.StackError
}

type DatabaseError struct {
	errors.StackError
}

type RequestError struct {
	errors.StackError
}

type ConnectionError struct {
	errors.StackError
}

type TimeoutError struct {
	errors.StackError
}

type ExecError struct {
	errors.StackError
}

type NetworkError struct {
	errors.StackError
}

type TypeError struct {
	errors.StackError
}

type ErrorData struct {
	Error   string `json:"error"`
	Message string `json:"error_msg"`
}

func (e *ErrorData) GetError() (err error) {
	err = &ParseError{
		errors.Newf("error: Parse error %s - %s", e.Error, e.Message),
	}
	return
}
