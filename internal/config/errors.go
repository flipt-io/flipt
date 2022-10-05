package config

import (
	"errors"
	"fmt"
)

const fieldErrFmt = "field %q: %w"

var (
	// ErrValidationRequired is returned when a required value is
	// either not supplied or supplied with empty value.
	ErrValidationRequired = errors.New("non-empty value is required")
)

func errFieldWrap(field string, err error) error {
	return fmt.Errorf(fieldErrFmt, field, err)
}

func errFieldRequired(field string) error {
	return errFieldWrap(field, ErrValidationRequired)
}
