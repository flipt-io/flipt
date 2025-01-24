package config

import (
	"errors"
	"fmt"
)

var (
	// errValidationRequired is returned when a required value is
	// either not supplied or supplied with empty value.
	errValidationRequired = errors.New("non-empty value is required")
	// errPositiveNonZeroDuration is returned when a negative or zero time.Duration is provided.
	errPositiveNonZeroDuration = errors.New("positive non-zero duration required")
)

// errFieldWrap wraps an error with a field name and type.
func errFieldWrap(typ, field string, err error) error {
	return fmt.Errorf("%s: %s %w", typ, field, err)
}

// errFieldRequired returns a required field error with a field name and type.
func errFieldRequired(typ, field string) error {
	return errFieldWrap(typ, field, errValidationRequired)
}

// errString returns an error with a type and message.
func errString(typ, msg string, args ...any) error {
	return fmt.Errorf("%s: %s", typ, fmt.Sprintf(msg, args...))
}
