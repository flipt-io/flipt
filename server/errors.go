package server

import "fmt"

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
