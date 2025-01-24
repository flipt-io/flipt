package config

import (
	"errors"
	"fmt"
)

const fieldErrFmt = "%s: %s %w"

var (
	// errValidationRequired is returned when a required value is
	// either not supplied or supplied with empty value.
	errValidationRequired = errors.New("non-empty value is required")
	// errPositiveNonZeroDuration is returned when a negative or zero time.Duration is provided.
	errPositiveNonZeroDuration = errors.New("positive non-zero duration required")
)

func errFieldWrap(typ, field string, err error) error {
	return fmt.Errorf(fieldErrFmt, typ, field, err)
}

func errFieldRequired(typ, field string) error {
	return errFieldWrap(typ, field, errValidationRequired)
}

func err(typ, msg string, args ...any) error {
	return fmt.Errorf("%s: %s", typ, fmt.Sprintf(msg, args...))
}
