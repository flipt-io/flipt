package config

import (
	"fmt"
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

	Session AuthenticationSession `json:"session,omitempty" mapstructure:"session"`
	Methods AuthenticationMethods `json:"methods,omitempty" mapstructure:"methods"`
}

// ShouldRunCleanup returns true if the cleanup background process should be started.
// It returns true given at-least 1 method is enabled and it's associated schedule
// has been configured (non-nil).
func (c AuthenticationConfig) ShouldRunCleanup() bool {
	return (c.Methods.Token.Enabled && c.Methods.Token.Cleanup != nil) ||
		(c.Methods.OIDC.Enabled && c.Methods.OIDC.Cleanup != nil)
}

func (c *AuthenticationConfig) setDefaults(v *viper.Viper) []string {
	methods := map[string]any{
		"token": nil,
		"oidc":  nil,
	}

	// set default for each methods
	for k := range methods {
		method := map[string]any{"enabled": false}
		// if the method has been enabled then set the defaults
		// for its cleanup strategy
		prefix := fmt.Sprintf("authentication.methods.%s", k)
		if v.GetBool(prefix + ".enabled") {
			method["cleanup"] = map[string]any{
				"interval":     time.Hour,
				"grace_period": 30 * time.Minute,
			}
		}

		methods[k] = method
	}

	v.SetDefault("authentication", map[string]any{
		"required": false,
		"session": map[string]any{
			"token_lifetime": "24h",
			"state_lifetime": "10m",
		},
		"methods": methods,
	})

	return nil
}

func (c *AuthenticationConfig) validate() error {
	for _, cleanup := range []struct {
		name     string
		schedule *AuthenticationCleanupSchedule
	}{
		// add additional schedules as token methods are created
		{"token", c.Methods.Token.Cleanup},
		{"oidc", c.Methods.OIDC.Cleanup},
	} {
		if cleanup.schedule == nil {
			continue
		}

		field := "authentication.method" + cleanup.name
		if cleanup.schedule.Interval <= 0 {
			return errFieldWrap(field+".cleanup.interval", errPositiveNonZeroDuration)
		}

		if cleanup.schedule.GracePeriod <= 0 {
			return errFieldWrap(field+".cleanup.grace_period", errPositiveNonZeroDuration)
		}
	}

	// ensure that when a session compatible authentication method has been
	// enabled that the session cookie domain has been configured with a non
	// empty value.
	if c.Methods.OIDC.Enabled {
		if c.Session.Domain == "" {
			err := errFieldWrap("authentication.session.domain", errValidationRequired)
			return fmt.Errorf("when session compatible auth method enabled: %w", err)
		}
	}

	return nil
}

// AuthenticationSession configures the session produced for browsers when
// establishing authentication via HTTP.
type AuthenticationSession struct {
	// Domain is the domain on which to register session cookies.
	Domain string `json:"domain,omitempty" mapstructure:"domain"`
	// Secure sets the secure property (i.e. HTTPS only) on both the state and token cookies.
	Secure bool `json:"secure" mapstructure:"secure"`
	// TokenLifetime is the duration of the flipt client token generated once
	// authentication has been established via a session compatible method.
	TokenLifetime time.Duration `json:"tokenLifetime,omitempty" mapstructure:"token_lifetime"`
	// StateLifetime is the lifetime duration of the state cookie.
	StateLifetime time.Duration `json:"stateLifetime,omitempty" mapstructure:"state_lifetime"`
}

// AuthenticationMethods is a set of configuration for each authentication
// method available for use within Flipt.
type AuthenticationMethods struct {
	Token AuthenticationMethodTokenConfig `json:"token,omitempty" mapstructure:"token"`
	OIDC  AuthenticationMethodOIDCConfig  `json:"oidc,omitempty" mapstructure:"oidc"`
}

// AuthenticationMethodTokenConfig contains fields used to configure the authentication
// method "token".
// This authentication method supports the ability to create static tokens via the
// /auth/v1/method/token prefix of endpoints.
type AuthenticationMethodTokenConfig struct {
	// Enabled designates whether or not static token authentication is enabled
	// and whether Flipt will mount the "token" method APIs.
	Enabled bool                           `json:"enabled,omitempty" mapstructure:"enabled"`
	Cleanup *AuthenticationCleanupSchedule `json:"cleanup,omitempty" mapstructure:"cleanup"`
}

// AuthenticationMethodOIDCConfig configures the OIDC authentication method.
// This method can be used to establish browser based sessions.
type AuthenticationMethodOIDCConfig struct {
	Enabled   bool                                        `json:"enabled,omitempty" mapstructure:"enabled"`
	Providers map[string]AuthenticationMethodOIDCProvider `json:"providers,omitempty" mapstructure:"providers"`
	Cleanup   *AuthenticationCleanupSchedule              `json:"cleanup,omitempty" mapstructure:"cleanup"`
}

// AuthenticationOIDCProviderGoogle configures the Google OIDC provider credentials
type AuthenticationMethodOIDCProvider struct {
	IssuerURL       string   `json:"issuerURL,omitempty" mapstructure:"issuer_url"`
	ClientID        string   `json:"clientID,omitempty" mapstructure:"client_id"`
	ClientSecret    string   `json:"clientSecret,omitempty" mapstructure:"client_secret"`
	RedirectAddress string   `json:"redirectAddress,omitempty" mapstructure:"redirect_address"`
	Scopes          []string `json:"scopes,omitempty" mapstructure:"scopes"`
}

// AuthenticationCleanupSchedule is used to configure a cleanup goroutine.
type AuthenticationCleanupSchedule struct {
	Interval    time.Duration `json:"interval,omitempty" mapstructure:"interval"`
	GracePeriod time.Duration `json:"gracePeriod,omitempty" mapstructure:"grace_period"`
}
