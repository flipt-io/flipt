package server

import "fmt"

// errInvalidField represents a validation error
type errInvalidField struct {
	field  string
	reason string
}

func (e errInvalidField) Error() string {
	return fmt.Sprintf("invalid field %s: %s", e.field, e.reason)
}

// invalidFieldError creates an errInvalidField for a specific field and reason
func invalidFieldError(field, reason string) error {
	return errInvalidField{field, reason}
}

// emptyFieldError creates an errInvalidField for an empty field
func emptyFieldError(field string) error {
	return invalidFieldError(field, "must not be empty")
}
