package config

import (
	"errors"
	"fmt"
)

// errFieldWrap wraps an error with a field name and type.
func errFieldWrap(typ, field string, err error) error {
	if typ == "" {
		return fmt.Errorf("%s: %w", field, err)
	}
	return fmt.Errorf("%s: %s %w", typ, field, err)
}

// errFieldRequired returns a formatted error for required fields
func errFieldRequired(typ, field string) error {
	if typ == "" {
		return fmt.Errorf("%s non-empty value is required", field)
	}
	return fmt.Errorf("%s: %s non-empty value is required", typ, field)
}

// errFieldPositiveDuration returns a formatted error for positive non-zero duration fields
func errFieldPositiveDuration(typ, field string) error {
	if typ == "" {
		return fmt.Errorf("%s must be a positive duration", field)
	}
	return fmt.Errorf("%s: %s must be a positive duration", typ, field)
}

// errString creates a new error with type and message context
func errString(typ, msg string) error {
	if typ == "" {
		return errors.New(msg)
	}
	return fmt.Errorf("%s: %s", typ, msg)
}
