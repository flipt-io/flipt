package errors

import (
	"errors"
	"fmt"
)

// New creates a new error with errors.New
func New(s string) error {
	return errors.New(s)
}

// ErrNotFound represents a not found error
type ErrNotFound string

// ErrNotFoundf creates an ErrNotFound using a custom format
func ErrNotFoundf(format string, args ...interface{}) error {
	return ErrNotFound(fmt.Sprintf(format, args...))
}

func (e ErrNotFound) Error() string {
	return fmt.Sprintf("%s not found", string(e))
}

// ErrInvalid represents an invalid error
type ErrInvalid string

// ErrInvalidf creates an ErrInvalid using a custom format
func ErrInvalidf(format string, args ...interface{}) error {
	return ErrInvalid(fmt.Sprintf(format, args...))
}

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
