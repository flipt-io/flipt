package errors

import (
	"errors"
	"fmt"
)

// As is a utility for one-lining errors.As statements.
// e.g. cerr, match := errors.As[MyCustomError](err).
func As[E error](err error) (e E, _ bool) {
	return e, errors.As(err, &e)
}

// AsMatch is the same as As but it returns just a boolean to represent
// whether or not the wrapped type matches the type parameter.
func AsMatch[E error](err error) (match bool) {
	_, match = As[E](err)
	return
}

// New creates a new error with errors.New
func New(s string) error {
	return errors.New(s)
}

// NewErrorf is a generic utility for formatting a string into a target error type E.
func NewErrorf[E StringError](format string, args ...any) error {
	return E(fmt.Sprintf(format, args...))
}

// ErrNotFound represents a not found error
type ErrNotFound string

// ErrNotFoundf is a convience function for producing ErrNotFound.
var ErrNotFoundf = NewErrorf[ErrNotFound]

func (e ErrNotFound) Error() string {
	return fmt.Sprintf("%s not found", string(e))
}

// ErrInvalid represents an invalid error
type ErrInvalid string

// ErrInvalidf is a convience function for producing ErrInvalid.
var ErrInvalidf = NewErrorf[ErrInvalid]

func (e ErrInvalid) Error() string {
	return string(e)
}

// ErrValidation is a validation error for a specific field and reason
type ErrValidation struct {
	field  string
	reason string
}

func (e ErrValidation) Error() string {
	return fmt.Sprintf("invalid field %s: %s", e.field, e.reason)
}

// ErrCanceled is returned when an operation has been prematurely canceled by the requester.
type ErrCanceled string

// ErrCanceledf is a convience function for producing ErrCanceled.
var ErrCanceledf = NewErrorf[ErrCanceled]

func (e ErrCanceled) Error() string {
	return string(e)
}

// InvalidFieldError creates an ErrInvalidField for a specific field and reason
func InvalidFieldError(field, reason string) error {
	return ErrValidation{field, reason}
}

// EmptyFieldError creates an ErrInvalidField for an empty field
func EmptyFieldError(field string) error {
	return InvalidFieldError(field, "must not be empty")
}

// ErrUnauthenticated is returned when an operation is attempted by an unauthenticated
// client in an authenticated context.
type ErrUnauthenticated string

// ErrUnauthenticatedf is a convience function for producing ErrUnauthenticated.
var ErrUnauthenticatedf = NewErrorf[ErrUnauthenticated]

// Error() returns the underlying string of the error.
func (e ErrUnauthenticated) Error() string {
	return string(e)
}

type ErrUnauthorized string

var ErrUnauthorizedf = NewErrorf[ErrUnauthorized]

func (e ErrUnauthorized) Error() string {
	return string(e)
}

// StringError is any error that also happens to have an underlying type of string.
type StringError interface {
	error
	~string
}
