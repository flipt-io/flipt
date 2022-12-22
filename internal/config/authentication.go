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
	for _, v := range auth.Method_value {
		method := auth.Method(v)
		if method == auth.Method_METHOD_NONE {
			continue
		}

		stringToAuthMethod[methodName(method)] = method
	}
}

func methodName(method auth.Method) string {
	return strings.ToLower(strings.TrimPrefix(auth.Method_name[int32(method)], "METHOD_"))
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
func (c AuthenticationConfig) ShouldRunCleanup() (shouldCleanup bool) {
	for _, info := range c.Methods.AllMethods() {
		shouldCleanup = shouldCleanup || (info.Enabled && info.Cleanup != nil)
	}

	return
}

func (c *AuthenticationConfig) setDefaults(v *viper.Viper) {
	methods := map[string]any{}

	// set default for each methods
	for _, info := range c.Methods.AllMethods() {
		method := map[string]any{"enabled": false}
		// if the method has been enabled then set the defaults
		// for its cleanup strategy
		prefix := fmt.Sprintf("authentication.methods.%s", info.Name())
		if v.GetBool(prefix + ".enabled") {
			method["cleanup"] = map[string]any{
				"interval":     time.Hour,
				"grace_period": 30 * time.Minute,
			}
		}

		methods[info.Name()] = method
	}

	v.SetDefault("authentication", map[string]any{
		"required": false,
		"session": map[string]any{
			"token_lifetime": "24h",
			"state_lifetime": "10m",
		},
		"methods": methods,
	})
}

func (c *AuthenticationConfig) validate() error {
	var sessionEnabled bool
	for _, info := range c.Methods.AllMethods() {
		sessionEnabled = sessionEnabled || (info.Enabled && info.SessionCompatible)
		if info.Cleanup == nil {
			continue
		}

		field := "authentication.method" + info.Name()
		if info.Cleanup.Interval <= 0 {
			return errFieldWrap(field+".cleanup.interval", errPositiveNonZeroDuration)
		}

		if info.Cleanup.GracePeriod <= 0 {
			return errFieldWrap(field+".cleanup.grace_period", errPositiveNonZeroDuration)
		}
	}

	// ensure that when a session compatible authentication method has been
	// enabled that the session cookie domain has been configured with a non
	// empty value.
	if sessionEnabled {
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
	Token AuthenticationMethod[AuthenticationMethodTokenConfig] `json:"token,omitempty" mapstructure:"token"`
	OIDC  AuthenticationMethod[AuthenticationMethodOIDCConfig]  `json:"oidc,omitempty" mapstructure:"oidc"`
}

// AllMethods returns all the AuthenticationMethod instances available.
func (a AuthenticationMethods) AllMethods() []StaticAuthenticationMethodInfo {
	return []StaticAuthenticationMethodInfo{
		a.Token.Info(),
		a.OIDC.Info(),
	}
}

// StaticAuthenticationMethodInfo embeds an AuthenticationMethodInfo alongside
// the other properties of an AuthenticationMethod.
type StaticAuthenticationMethodInfo struct {
	AuthenticationMethodInfo
	Enabled bool
	Cleanup *AuthenticationCleanupSchedule
}

// AuthenticationMethodInfo is a structure which describes properties
// of a particular authentication method.
// i.e. the name and whether or not the method is session compatible.
type AuthenticationMethodInfo struct {
	Method            auth.Method
	SessionCompatible bool
}

// Name returns the friendly lower-case name for the authentication method.
func (a AuthenticationMethodInfo) Name() string {
	return methodName(a.Method)
}

// AuthenticationMethodInfoProvider is a type with a single method Info
// which returns an AuthenticationMethodInfo describing the underlying
// methods properties.
type AuthenticationMethodInfoProvider interface {
	Info() AuthenticationMethodInfo
}

// AuthenticationMethod is a container for authentication methods.
// It describes the common properties of all authentication methods.
// Along with leaving a generic slot for the particular method to declare
// its own structural fields. This generic field (Method) must implement
// the AuthenticationMethodInfoProvider to be valid at compile time.
type AuthenticationMethod[C AuthenticationMethodInfoProvider] struct {
	Method  C                              `mapstructure:",squash"`
	Enabled bool                           `json:"enabled,omitempty" mapstructure:"enabled"`
	Cleanup *AuthenticationCleanupSchedule `json:"cleanup,omitempty" mapstructure:"cleanup"`
}

func (a AuthenticationMethod[C]) Info() StaticAuthenticationMethodInfo {
	return StaticAuthenticationMethodInfo{
		AuthenticationMethodInfo: a.Method.Info(),
		Enabled:                  a.Enabled,
		Cleanup:                  a.Cleanup,
	}
}

// AuthenticationMethodTokenConfig contains fields used to configure the authentication
// method "token".
// This authentication method supports the ability to create static tokens via the
// /auth/v1/method/token prefix of endpoints.
type AuthenticationMethodTokenConfig struct{}

// Info describes properties of the authentication method "token".
func (a AuthenticationMethodTokenConfig) Info() AuthenticationMethodInfo {
	return AuthenticationMethodInfo{
		Method:            auth.Method_METHOD_TOKEN,
		SessionCompatible: false,
	}
}

// AuthenticationMethodOIDCConfig configures the OIDC authentication method.
// This method can be used to establish browser based sessions.
type AuthenticationMethodOIDCConfig struct {
	Providers map[string]AuthenticationMethodOIDCProvider `json:"providers,omitempty" mapstructure:"providers"`
}

// Info describes properties of the authentication method "oidc".
func (a AuthenticationMethodOIDCConfig) Info() AuthenticationMethodInfo {
	return AuthenticationMethodInfo{
		Method:            auth.Method_METHOD_OIDC,
		SessionCompatible: true,
	}
}

// AuthenticationOIDCProvider configures provider credentials
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
