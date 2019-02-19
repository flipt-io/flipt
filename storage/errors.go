package storage

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
