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

func (e ErrNotFound) Error() string {
	return fmt.Sprintf("%s not found", string(e))
}

// ErrInvalid represents an invalid error
type ErrInvalid string

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

// InvalidFieldError creates an ErrInvalidField for a specific field and reason
func InvalidFieldError(field, reason string) error {
	return ErrValidation{field, reason}
}

// EmptyFieldError creates an ErrInvalidField for an empty field
func EmptyFieldError(field string) error {
	return InvalidFieldError(field, "must not be empty")
}

// ErrForbidding is returned when an operation is attempted which is forbidden
// for the identified caller.
type ErrForbidden string

// Error() returns the underlying string of the error.
func (e ErrForbidden) Error() string {
	return string(e)
}

// StringError is any error that also happens to have an underlying type of string.
type StringError interface {
	error
	~string
}
