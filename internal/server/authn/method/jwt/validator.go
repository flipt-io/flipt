package jwt

import (
	"context"

	"github.com/hashicorp/cap/jwt"
)

// jwtKeySetValidator is a JWT validator that uses a key set and expected claims.
// it exists to provide a common interface for validating JWT tokens and simply delegates
// to the cap/jwt library for the actual validation.
type jwtKeySetValidator struct {
	validator *jwt.Validator
	expected  jwt.Expected
}

func NewValidator(validator *jwt.Validator, expected jwt.Expected) *jwtKeySetValidator {
	return &jwtKeySetValidator{validator: validator, expected: expected}
}

func (v *jwtKeySetValidator) Validate(ctx context.Context, token string) (map[string]any, error) {
	return v.validator.Validate(ctx, token, v.expected)
}
