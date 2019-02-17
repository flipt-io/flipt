package flipt

import "fmt"

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

// ErrInvalidField represents a validation error
type ErrInvalidField struct {
	field  string
	reason string
}

func (e ErrInvalidField) Error() string {
	return fmt.Sprintf("invalid field %s: %s", e.field, e.reason)
}

// InvalidFieldError creates an ErrInvalidField for a specific field and reason
func InvalidFieldError(field, reason string) error {
	return ErrInvalidField{field, reason}
}

// EmptyFieldError creates an ErrInvalidField for an empty field
func EmptyFieldError(field string) error {
	return InvalidFieldError(field, "must not be empty")
}
