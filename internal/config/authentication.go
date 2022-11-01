package config

import "github.com/spf13/viper"

var _ defaulter = (*AuthenticationConfig)(nil)

// AuthenticationConfig configures Flipts authentication mechanisms
type AuthenticationConfig struct {
	// Required designates whether authentication credentials are validated.
	// If required == true, then authentication is required for all API endpoints.
	// Else, authentication is not required and Flipt's APIs are not secured.
	Required bool `json:"required,omitempty" mapstructure:"required"`

	Methods struct {
		Token AuthenticationMethodTokenConfig `json:"token,omitempty" mapstructure:"token"`
	} `json:"methods,omitempty" mapstructure:"methods"`
}

func (a *AuthenticationConfig) setDefaults(v *viper.Viper) []string {
	v.SetDefault("authentication", map[string]any{
		"required": false,
		"methods": map[string]any{
			"token": map[string]any{
				"enabled": false,
			},
		},
	})

	return nil
}

// AuthenticationMethodTokenConfig contains fields used to configure the authentication
// method "token".
// This authentication method supports the ability to create static tokens via the
// /auth/v1/method/token prefix of endpoints.
type AuthenticationMethodTokenConfig struct {
	// Enabled designates whether or not static token authentication is enabled
	// and whether Flipt will mount the "token" method APIs.
	Enabled bool `json:"enabled,omitempty" mapstructure:"enabled"`
}
