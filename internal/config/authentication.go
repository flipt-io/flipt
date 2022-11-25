package config

import (
	"strings"
	"time"

	"github.com/spf13/viper"
	"go.flipt.io/flipt/rpc/flipt/auth"
)

var (
	_                  defaulter = (*AuthenticationConfig)(nil)
	stringToAuthMethod           = map[string]auth.Method{}
)

func init() {
	for method, v := range auth.Method_value {
		if auth.Method(v) == auth.Method_METHOD_NONE {
			continue
		}

		name := strings.ToLower(strings.TrimPrefix(method, "METHOD_"))
		stringToAuthMethod[name] = auth.Method(v)
	}
}

// AuthenticationConfig configures Flipts authentication mechanisms
type AuthenticationConfig struct {
	// Required designates whether authentication credentials are validated.
	// If required == true, then authentication is required for all API endpoints.
	// Else, authentication is not required and Flipt's APIs are not secured.
	Required bool `json:"required,omitempty" mapstructure:"required"`

	Cleanup AuthenticationCleanupSchedules `json:"cleanup,omitempty" mapstructure:"cleanup"`

	Methods AuthenticationMethods `json:"methods,omitempty" mapstructure:"methods"`
}

func (a *AuthenticationConfig) setDefaults(v *viper.Viper) []string {
	cleanup := map[string]any{}
	if v.GetBool("authentication.methods.token.enabled") {
		cleanup["token"] = map[string]any{
			"interval":     time.Hour,
			"grace_period": 30 * time.Minute,
		}
	}

	defaults := map[string]any{
		"required": false,
		"methods": map[string]any{
			"token": map[string]any{
				"enabled": false,
			},
		},
	}

	if len(cleanup) > 0 {
		defaults["cleanup"] = cleanup
	}

	v.SetDefault("authentication", defaults)

	return nil
}

// AuthenticationMethods is a set of configuration for each authentication
// method available for use within Flipt.
type AuthenticationMethods struct {
	Token AuthenticationMethodTokenConfig `json:"token,omitempty" mapstructure:"token"`
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

// AuthenticationCleanupSchedules is a map of method to schedule
type AuthenticationCleanupSchedules struct {
	Token *AuthenticationCleanupSchedule `json:"token,omitempty" mapstructure:"token"`
}

// ShouldRun returns true if the cleanup background process should be started.
// It returns true given at-least 1 schedule has been configured (not-nil).
func (a AuthenticationCleanupSchedules) ShouldRun() bool {
	return a.Token != nil
}

// AuthenticationCleanupSchedule is used to configure a cleanup goroutine.
type AuthenticationCleanupSchedule struct {
	Interval    time.Duration `json:"interval,omitempty" mapstructure:"interval"`
	GracePeriod time.Duration `json:"gracePeriod,omitempty" mapstructure:"grace_period"`
}
