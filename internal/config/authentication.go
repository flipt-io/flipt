package config

import (
	"fmt"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/spf13/viper"
	"go.flipt.io/flipt/rpc/flipt/auth"
	"google.golang.org/protobuf/types/known/structpb"
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

// Enabled returns true if authentication is marked as required
// or any of the authentication methods are enabled.
func (c AuthenticationConfig) Enabled() bool {
	if c.Required {
		return true
	}

	for _, info := range c.Methods.AllMethods() {
		if info.Enabled {
			return true
		}
	}

	return false
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
			// apply any method specific defaults
			info.setDefaults(method)
			// set default cleanup
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

		host, err := getHostname(c.Session.Domain)
		if err != nil {
			return fmt.Errorf("invalid domain: %w", err)
		}

		// strip scheme and port from domain
		// domain cookies are not allowed to have a scheme or port
		// https://github.com/golang/go/issues/28297
		c.Session.Domain = host
	}

	return nil
}

func getHostname(rawurl string) (string, error) {
	if !strings.Contains(rawurl, "://") {
		rawurl = "http://" + rawurl
	}
	u, err := url.Parse(rawurl)
	if err != nil {
		return "", err
	}
	return strings.Split(u.Host, ":")[0], nil
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
	// CSRF configures CSRF provention mechanisms.
	CSRF AuthenticationSessionCSRF `json:"csrf,omitempty" mapstructure:"csrf"`
}

// AuthenticationSessionCSRF configures cross-site request forgery prevention.
type AuthenticationSessionCSRF struct {
	// Key is the private key string used to authenticate csrf tokens.
	Key string `json:"-" mapstructure:"key"`
}

// AuthenticationMethods is a set of configuration for each authentication
// method available for use within Flipt.
type AuthenticationMethods struct {
	Token      AuthenticationMethod[AuthenticationMethodTokenConfig]      `json:"token,omitempty" mapstructure:"token"`
	OIDC       AuthenticationMethod[AuthenticationMethodOIDCConfig]       `json:"oidc,omitempty" mapstructure:"oidc"`
	Kubernetes AuthenticationMethod[AuthenticationMethodKubernetesConfig] `json:"kubernetes,omitempty" mapstructure:"kubernetes"`
}

// AllMethods returns all the AuthenticationMethod instances available.
func (a *AuthenticationMethods) AllMethods() []StaticAuthenticationMethodInfo {
	return []StaticAuthenticationMethodInfo{
		a.Token.info(),
		a.OIDC.info(),
		a.Kubernetes.info(),
	}
}

// EnabledMethods returns all the AuthenticationMethod instances that have been enabled.
func (a *AuthenticationMethods) EnabledMethods() []StaticAuthenticationMethodInfo {
	var enabled []StaticAuthenticationMethodInfo
	for _, info := range a.AllMethods() {
		if info.Enabled {
			enabled = append(enabled, info)
		}
	}

	return enabled
}

// StaticAuthenticationMethodInfo embeds an AuthenticationMethodInfo alongside
// the other properties of an AuthenticationMethod.
type StaticAuthenticationMethodInfo struct {
	AuthenticationMethodInfo
	Enabled bool
	Cleanup *AuthenticationCleanupSchedule

	// used for bootstrapping defaults
	setDefaults func(map[string]any)
	// used for testing purposes to ensure all methods
	// are appropriately cleaned up via the background process.
	setEnabled func()
	setCleanup func(AuthenticationCleanupSchedule)
}

// Enable can only be called in a testing scenario.
// It is used to enable a target method without having a concrete reference.
func (s StaticAuthenticationMethodInfo) Enable(t *testing.T) {
	s.setEnabled()
}

// SetCleanup can only be called in a testing scenario.
// It is used to configure cleanup for a target method without having a concrete reference.
func (s StaticAuthenticationMethodInfo) SetCleanup(t *testing.T, c AuthenticationCleanupSchedule) {
	s.setCleanup(c)
}

// AuthenticationMethodInfo is a structure which describes properties
// of a particular authentication method.
// i.e. the name and whether or not the method is session compatible.
type AuthenticationMethodInfo struct {
	Method            auth.Method
	SessionCompatible bool
	Metadata          *structpb.Struct
}

// Name returns the friendly lower-case name for the authentication method.
func (a AuthenticationMethodInfo) Name() string {
	return methodName(a.Method)
}

// AuthenticationMethodInfoProvider is a type with a single method Info
// which returns an AuthenticationMethodInfo describing the underlying
// methods properties.
type AuthenticationMethodInfoProvider interface {
	setDefaults(map[string]any)
	info() AuthenticationMethodInfo
}

// AuthenticationMethod is a container for authentication methods.
// It describes the common properties of all authentication methods.
// Along with leaving a generic slot for the particular method to declare
// its own structural fields. This generic field (Method) must implement
// the AuthenticationMethodInfoProvider to be valid at compile time.
// nolint:musttag
type AuthenticationMethod[C AuthenticationMethodInfoProvider] struct {
	Method  C                              `mapstructure:",squash"`
	Enabled bool                           `json:"enabled,omitempty" mapstructure:"enabled"`
	Cleanup *AuthenticationCleanupSchedule `json:"cleanup,omitempty" mapstructure:"cleanup"`
}

func (a *AuthenticationMethod[C]) setDefaults(defaults map[string]any) {
	a.Method.setDefaults(defaults)
}

func (a *AuthenticationMethod[C]) info() StaticAuthenticationMethodInfo {
	return StaticAuthenticationMethodInfo{
		AuthenticationMethodInfo: a.Method.info(),
		Enabled:                  a.Enabled,
		Cleanup:                  a.Cleanup,

		setDefaults: a.setDefaults,
		setEnabled: func() {
			a.Enabled = true
		},
		setCleanup: func(c AuthenticationCleanupSchedule) {
			a.Cleanup = &c
		},
	}
}

// AuthenticationMethodTokenConfig contains fields used to configure the authentication
// method "token".
// This authentication method supports the ability to create static tokens via the
// /auth/v1/method/token prefix of endpoints.
type AuthenticationMethodTokenConfig struct {
	Bootstrap AuthenticationMethodTokenBootstrapConfig `json:"bootstrap" mapstructure:"bootstrap"`
}

func (a AuthenticationMethodTokenConfig) setDefaults(map[string]any) {}

// info describes properties of the authentication method "token".
func (a AuthenticationMethodTokenConfig) info() AuthenticationMethodInfo {
	return AuthenticationMethodInfo{
		Method:            auth.Method_METHOD_TOKEN,
		SessionCompatible: false,
	}
}

// AuthenticationMethodTokenBootstrapConfig contains fields used to configure the
// bootstrap process for the authentication method "token".
type AuthenticationMethodTokenBootstrapConfig struct {
	Token      string        `json:"-" mapstructure:"token"`
	Expiration time.Duration `json:"expiration,omitempty" mapstructure:"expiration"`
}

// AuthenticationMethodOIDCConfig configures the OIDC authentication method.
// This method can be used to establish browser based sessions.
type AuthenticationMethodOIDCConfig struct {
	Providers map[string]AuthenticationMethodOIDCProvider `json:"providers,omitempty" mapstructure:"providers"`
}

func (a AuthenticationMethodOIDCConfig) setDefaults(map[string]any) {}

// info describes properties of the authentication method "oidc".
func (a AuthenticationMethodOIDCConfig) info() AuthenticationMethodInfo {
	info := AuthenticationMethodInfo{
		Method:            auth.Method_METHOD_OIDC,
		SessionCompatible: true,
	}

	var (
		metadata  = make(map[string]any)
		providers = make(map[string]any, len(a.Providers))
	)

	// this ensures we expose the authorize and callback URL endpoint
	// to the UI via the /auth/v1/method endpoint
	for provider := range a.Providers {
		providers[provider] = map[string]any{
			"authorize_url": fmt.Sprintf("/auth/v1/method/oidc/%s/authorize", provider),
			"callback_url":  fmt.Sprintf("/auth/v1/method/oidc/%s/callback", provider),
		}
	}

	metadata["providers"] = providers
	info.Metadata, _ = structpb.NewStruct(metadata)
	return info
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

// AuthenticationMethodKubernetesConfig contains the fields necessary for the Kubernetes authentication
// method to be performed. This method supports Flipt being deployed in a Kubernetes environment
// and allowing it to exchange client tokens for valid service account tokens presented via this method.
type AuthenticationMethodKubernetesConfig struct {
	// DiscoveryURL is the URL to the local Kubernetes cluster serving the "well-known" OIDC discovery endpoint.
	// https://openid.net/specs/openid-connect-discovery-1_0.html
	// The URL is used to fetch the OIDC configuration and subsequently the JWKS certificates.
	DiscoveryURL string `json:"discoveryURL,omitempty" mapstructure:"discovery_url"`
	// CAPath is the path on disk to the trusted certificate authority certificate for validating
	// HTTPS requests to the issuer.
	CAPath string `json:"caPath,omitempty" mapstructure:"ca_path"`
	// ServiceAccountTokenPath is the location on disk to the Flipt instances service account token.
	// This should be the token issued for the service account associated with Flipt in the environment.
	ServiceAccountTokenPath string `json:"serviceAccountTokenPath,omitempty" mapstructure:"service_account_token_path"`
}

func (a AuthenticationMethodKubernetesConfig) setDefaults(defaults map[string]any) {
	defaults["discovery_url"] = "https://kubernetes.default.svc.cluster.local"
	defaults["ca_path"] = "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"
	defaults["service_account_token_path"] = "/var/run/secrets/kubernetes.io/serviceaccount/token"
}

// info describes properties of the authentication method "kubernetes".
func (a AuthenticationMethodKubernetesConfig) info() AuthenticationMethodInfo {
	return AuthenticationMethodInfo{
		Method:            auth.Method_METHOD_KUBERNETES,
		SessionCompatible: false,
	}
}
