package jwt

import (
	"context"
	"fmt"
	"slices"
	"strconv"

	"github.com/hashicorp/cap/jwt"
)

// CustomClaim is an additional requirement placed on the claims of a JWT.
// The claim named Name must be present in the token and its value must match
// at least one of Values. When the claim is an array, at least one of its
// elements must match at least one of Values.
type CustomClaim struct {
	Name   string
	Values []string
}

// ValidatorOption configures optional behaviour of the validator.
type ValidatorOption func(*jwtKeySetValidator)

// WithCustomClaims requires every one of the supplied claims to be satisfied
// by a token for it to be considered valid.
func WithCustomClaims(claims []CustomClaim) ValidatorOption {
	return func(v *jwtKeySetValidator) {
		v.customClaims = claims
	}
}

// jwtKeySetValidator is a JWT validator that uses a key set and expected claims.
// it exists to provide a common interface for validating JWT tokens and simply delegates
// to the cap/jwt library for the actual validation. Custom claims, which the
// cap/jwt library has no notion of, are checked afterwards.
type jwtKeySetValidator struct {
	validator    *jwt.Validator
	expected     jwt.Expected
	customClaims []CustomClaim
}

func NewValidator(validator *jwt.Validator, expected jwt.Expected, opts ...ValidatorOption) *jwtKeySetValidator {
	v := &jwtKeySetValidator{validator: validator, expected: expected}
	for _, opt := range opts {
		opt(v)
	}

	return v
}

func (v *jwtKeySetValidator) Validate(ctx context.Context, token string) (map[string]any, error) {
	claims, err := v.validator.Validate(ctx, token, v.expected)
	if err != nil {
		return nil, err
	}

	for _, custom := range v.customClaims {
		if err := custom.validate(claims); err != nil {
			return nil, err
		}
	}

	return claims, nil
}

func (c CustomClaim) validate(claims map[string]any) error {
	value, ok := claims[c.Name]
	if !ok {
		return fmt.Errorf("invalid %q claim: claim is missing", c.Name)
	}

	if elements, ok := value.([]any); ok {
		for _, element := range elements {
			if c.matches(element) {
				return nil
			}
		}
	} else if c.matches(value) {
		return nil
	}

	return fmt.Errorf("invalid %q claim: value is not one of the expected values", c.Name)
}

// matches reports whether a single, non-array, claim value is one of the
// expected values. Values of other types than string are compared using their
// JSON representation (e.g. true, 42), anything else never matches.
func (c CustomClaim) matches(value any) bool {
	switch v := value.(type) {
	case string:
		return slices.Contains(c.Values, v)
	case bool:
		return slices.Contains(c.Values, strconv.FormatBool(v))
	case float64:
		return slices.Contains(c.Values, strconv.FormatFloat(v, 'f', -1, 64))
	}

	return false
}
