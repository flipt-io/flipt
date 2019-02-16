package flipt

import "fmt"

type ErrNotFound string

func ErrNotFoundf(format string, args ...interface{}) error {
	return ErrNotFound(fmt.Sprintf(format, args...))
}

func (e ErrNotFound) Error() string {
	return fmt.Sprintf("%s not found", string(e))
}

type ErrInvalid string

func ErrInvalidf(format string, args ...interface{}) error {
	return ErrInvalid(fmt.Sprintf(format, args...))
}

func (e ErrInvalid) Error() string {
	return string(e)
}

type ErrInvalidField struct {
	field  string
	reason string
}

func (e ErrInvalidField) Error() string {
	return fmt.Sprintf("invalid field %s: %s", e.field, e.reason)
}

func InvalidFieldError(field, reason string) error {
	return ErrInvalidField{field, reason}
}

func EmptyFieldError(field string) error {
	return InvalidFieldError(field, "must not be empty")
}
