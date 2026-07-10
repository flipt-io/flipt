package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithForwardPrefix(t *testing.T) {
	ctx := t.Context()
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

func TestAuthenticationConfig_validate_FrontChannelLogoutRequiresSecureSession(t *testing.T) {
	tests := []struct {
		name          string
		config        AuthenticationConfig
		expectError   bool
		errorContains string
	}{
		{
			name: "allow_front_channel_logout with secure false should fail",
			config: AuthenticationConfig{
				Session: AuthenticationSessionConfig{
					Domain: "localhost",
					Secure: false,
				},
				Methods: AuthenticationMethodsConfig{
					OIDC: AuthenticationMethod[AuthenticationMethodOIDCConfig]{
						Enabled: true,
						Method: AuthenticationMethodOIDCConfig{
							Providers: map[string]AuthenticationMethodOIDCProvider{
								"google": {
									ClientID:                "client-id",
									ClientSecret:            "client-secret",
									RedirectAddress:         "http://localhost:8080",
									AllowFrontChannelLogout: true,
								},
							},
						},
					},
				},
			},
			expectError:   true,
			errorContains: "session secure must be true",
		},
		{
			name: "allow_front_channel_logout with secure true should pass",
			config: AuthenticationConfig{
				Session: AuthenticationSessionConfig{
					Domain: "localhost",
					Secure: true,
				},
				Methods: AuthenticationMethodsConfig{
					OIDC: AuthenticationMethod[AuthenticationMethodOIDCConfig]{
						Enabled: true,
						Method: AuthenticationMethodOIDCConfig{
							Providers: map[string]AuthenticationMethodOIDCProvider{
								"google": {
									ClientID:                "client-id",
									ClientSecret:            "client-secret",
									RedirectAddress:         "http://localhost:8080",
									AllowFrontChannelLogout: true,
								},
							},
						},
					},
				},
			},
			expectError: false,
		},
		{
			name: "allow_front_channel_logout false with secure false should pass",
			config: AuthenticationConfig{
				Session: AuthenticationSessionConfig{
					Domain: "localhost",
					Secure: false,
				},
				Methods: AuthenticationMethodsConfig{
					OIDC: AuthenticationMethod[AuthenticationMethodOIDCConfig]{
						Enabled: true,
						Method: AuthenticationMethodOIDCConfig{
							Providers: map[string]AuthenticationMethodOIDCProvider{
								"google": {
									ClientID:                "client-id",
									ClientSecret:            "client-secret",
									RedirectAddress:         "http://localhost:8080",
									AllowFrontChannelLogout: false,
								},
							},
						},
					},
				},
			},
			expectError: false,
		},
		{
			name: "oidc disabled should skip validation",
			config: AuthenticationConfig{
				Session: AuthenticationSessionConfig{
					Domain: "localhost",
					Secure: false,
				},
				Methods: AuthenticationMethodsConfig{
					OIDC: AuthenticationMethod[AuthenticationMethodOIDCConfig]{
						Enabled: false,
						Method: AuthenticationMethodOIDCConfig{
							Providers: map[string]AuthenticationMethodOIDCProvider{
								"google": {
									ClientID:                "client-id",
									ClientSecret:            "client-secret",
									RedirectAddress:         "http://localhost:8080",
									AllowFrontChannelLogout: true,
								},
							},
						},
					},
				},
			},
			expectError: false,
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
