package config

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithForwardPrefix(t *testing.T) {
	ctx := context.Background()
	assert.Empty(t, getForwardPrefix(ctx))
	ctx = WithForwardPrefix(ctx, "/some/prefix")
	assert.Equal(t, "/some/prefix", getForwardPrefix(ctx))
}

func TestAuthenticationMethodJWTConfig_validate_ClaimsMapping(t *testing.T) {
	tests := []struct {
		name          string
		config        AuthenticationMethodJWTConfig
		expectError   bool
		errorContains string
	}{
		{
			name: "valid claims mapping with subset of allowed keys",
			config: AuthenticationMethodJWTConfig{
				JWKSURL: "https://example.com/.well-known/jwks.json",
				ClaimsMapping: map[string]string{
					"email": "/user/email",
					"name":  "/user/name",
				},
			},
			expectError: false,
		},
		{
			name: "nil claims mapping should be valid, use the defaults",
			config: AuthenticationMethodJWTConfig{
				JWKSURL:       "https://example.com/.well-known/jwks.json",
				ClaimsMapping: nil,
			},
			expectError: false,
		},
		{
			name: "invalid claim key should fail",
			config: AuthenticationMethodJWTConfig{
				JWKSURL: "https://example.com/.well-known/jwks.json",
				ClaimsMapping: map[string]string{
					"email":       "/user/email",
					"invalid_key": "/user/something",
				},
			},
			expectError:   true,
			errorContains: "invalid claim key 'invalid_key'",
		},
		{
			name: "multiple invalid claim keys should fail",
			config: AuthenticationMethodJWTConfig{
				JWKSURL: "https://example.com/.well-known/jwks.json",
				ClaimsMapping: map[string]string{
					"email":    "/user/email",
					"bad_key1": "/user/something",
					"bad_key2": "/user/other",
				},
			},
			expectError:   true,
			errorContains: "invalid claim key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.validate()

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
