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

// ErrAlreadyExists is returned when attempting to create a resource that
// already exists
type ErrAlreadyExists string

// ErrAlreadyExistsf is a convience function for producing ErrAlreadyExists
var ErrAlreadyExistsf = NewErrorf[ErrAlreadyExists]

// Error returns the underlying string of the error
func (e ErrAlreadyExists) Error() string {
	return string(e)
}

// ErrConflict is returned when an operation fails to be applied because the
// underlying storage revision has advanced since the operations was first attempted
type ErrConflict string

// ErrConflictf is a convience function for producing ErrConflict
var ErrConflictf = NewErrorf[ErrConflict]

// Error returns the underlying string of the error
func (e ErrConflict) Error() string {
	return string(e)
}

// ErrNotImplemented is returned when an operation is attempted which is not supported
// either because it is not yet implemented or because it has been disabled in Flipt
type ErrNotImplemented string

// ErrNotImplementedf is a convience function for producing ErrNotImplemented
var ErrNotImplementedf = NewErrorf[ErrNotImplemented]

// Error returns the underlying string of the error
func (e ErrNotImplemented) Error() string {
	return string(e)
}

// ErrNotModified is returned when a request is made containing an If-None-Match header
// which matches the current digest for the given requested resource
// It us up to the client to reuse a previous response in this situation
type ErrNotModified string

var ErrNotModifiedf = NewErrorf[ErrNotModified]

// Error returns a formatted message of the underlying error string
func (e ErrNotModified) Error() string {
	return fmt.Sprintf("not modified: %q", string(e))
}

// StringError is any error that also happens to have an underlying type of string.
type StringError interface {
	error
	~string
}
