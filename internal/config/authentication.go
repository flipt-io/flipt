package config

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/spf13/viper"
	"go.flipt.io/flipt/rpc/flipt/auth"
	"google.golang.org/protobuf/types/known/structpb"
)

var (
	_                  defaulter = (*AuthenticationConfig)(nil)
	_                  validator = (*AuthenticationConfig)(nil)
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
	Required bool `json:"required" mapstructure:"required" yaml:"required"`

	// Exclude allows you to skip enforcing authentication on the different
	// top-level sections of the API.
	// By default, given required == true, the API is fully protected.
	Exclude struct {
		// Management refers to the section of the API with the prefix /api/v1
		Management bool `json:"management,omitempty" mapstructure:"management" yaml:"management,omitempty"`
		// Metadata refers to the section of the API with the prefix /meta
		Metadata bool `json:"metadata,omitempty" mapstructure:"metadata" yaml:"metadata,omitempty"`
		// Evaluation refers to the section of the API with the prefix /evaluation/v1
		Evaluation bool `json:"evaluation,omitempty" mapstructure:"evaluation" yaml:"evaluation,omitempty"`
		// OFREP refers to the section of the API with the prefix /ofrep
		OFREP bool `json:"ofrep,omitempty" mapstructure:"ofrep" yaml:"ofrep,omitempty"`
	} `json:"exclude,omitempty" mapstructure:"exclude" yaml:"exclude,omitempty"`

	Session AuthenticationSession `json:"session,omitempty" mapstructure:"session" yaml:"session,omitempty"`
	Methods AuthenticationMethods `json:"methods,omitempty" mapstructure:"methods" yaml:"methods,omitempty"`
}

// Enabled returns true if authentication is marked as required
// or any of the authentication methods are enabled.
func (c AuthenticationConfig) Enabled() bool {
	if c.Required {
		return true
	}

	for _, info := range c.Methods.AllMethods(context.Background()) {
		if info.Enabled {
			return true
		}
	}

	return false
}

// RequiresDatabase returns true if any of the enabled authentication
// methods requires a database connection
func (c AuthenticationConfig) RequiresDatabase() bool {
	for _, info := range c.Methods.AllMethods(context.Background()) {
		if info.Enabled && info.RequiresDatabase {
			return true
		}
	}

	return false
}

// IsZero returns true if the authentication config is not enabled.
// This is used for marshalling to YAML for `config init`.
func (c AuthenticationConfig) IsZero() bool {
	return !c.Enabled()
}

// ShouldRunCleanup returns true if the cleanup background process should be started.
// It returns true given at-least 1 method is enabled and it's associated schedule
// has been configured (non-nil).
func (c AuthenticationConfig) ShouldRunCleanup() (shouldCleanup bool) {
	for _, info := range c.Methods.AllMethods(context.Background()) {
		shouldCleanup = shouldCleanup || info.RequiresCleanup()
	}

	return
}

func (c *AuthenticationConfig) setDefaults(v *viper.Viper) error {
	methods := map[string]any{}

	// set default for each methods
	for _, info := range c.Methods.AllMethods(context.Background()) {
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

	return nil
}

func (c *AuthenticationConfig) SessionEnabled() bool {
	var sessionEnabled bool
	for _, info := range c.Methods.AllMethods(context.Background()) {
		sessionEnabled = sessionEnabled || (info.Enabled && info.SessionCompatible)
	}

	return sessionEnabled
}

func (c *AuthenticationConfig) validate() error {
	var sessionEnabled bool

	for _, info := range c.Methods.AllMethods(context.Background()) {
		if !info.RequiresCleanup() {
			continue
		}

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

	for _, info := range c.Methods.AllMethods(context.Background()) {
		if err := info.validate(); err != nil {
			return err
		}
	}

	return nil
}

func (c *AuthenticationConfig) deprecations(v *viper.Viper) []deprecated {
	if v.Get("authentication.exclude.metadata") != nil {
		return []deprecated{deprecateAuthenticationExcludeMetdata}
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
	Domain string `json:"domain,omitempty" mapstructure:"domain" yaml:"domain,omitempty"`
	// Secure sets the secure property (i.e. HTTPS only) on both the state and token cookies.
	Secure bool `json:"secure,omitempty" mapstructure:"secure" yaml:"secure,omitempty"`
	// TokenLifetime is the duration of the flipt client token generated once
	// authentication has been established via a session compatible method.
	TokenLifetime time.Duration `json:"tokenLifetime,omitempty" mapstructure:"token_lifetime" yaml:"token_lifetime,omitempty"`
	// StateLifetime is the lifetime duration of the state cookie.
	StateLifetime time.Duration `json:"stateLifetime,omitempty" mapstructure:"state_lifetime" yaml:"state_lifetime,omitempty"`
	// CSRF configures CSRF provention mechanisms.
	CSRF AuthenticationSessionCSRF `json:"csrf,omitempty" mapstructure:"csrf" yaml:"csrf,omitempty"`
}

// AuthenticationSessionCSRF configures cross-site request forgery prevention.
type AuthenticationSessionCSRF struct {
	// Key is the private key string used to authenticate csrf tokens.
	Key string `json:"-" mapstructure:"key"`
}

// AuthenticationMethods is a set of configuration for each authentication
// method available for use within Flipt.
type AuthenticationMethods struct {
	Token      AuthenticationMethod[AuthenticationMethodTokenConfig]      `json:"token,omitempty" mapstructure:"token" yaml:"token,omitempty"`
	Github     AuthenticationMethod[AuthenticationMethodGithubConfig]     `json:"github,omitempty" mapstructure:"github" yaml:"github,omitempty"`
	OIDC       AuthenticationMethod[AuthenticationMethodOIDCConfig]       `json:"oidc,omitempty" mapstructure:"oidc" yaml:"oidc,omitempty"`
	Kubernetes AuthenticationMethod[AuthenticationMethodKubernetesConfig] `json:"kubernetes,omitempty" mapstructure:"kubernetes" yaml:"kubernetes,omitempty"`
	JWT        AuthenticationMethod[AuthenticationMethodJWTConfig]        `json:"jwt,omitempty" mapstructure:"jwt" yaml:"jwt,omitempty"`
	Cloud      AuthenticationMethod[AuthenticationMethodCloudConfig]      `json:"cloud,omitempty" mapstructure:"cloud" yaml:"cloud,omitempty"`
}

// AllMethods returns all the AuthenticationMethod instances available.
func (a *AuthenticationMethods) AllMethods(ctx context.Context) []StaticAuthenticationMethodInfo {
	return []StaticAuthenticationMethodInfo{
		a.Token.info(ctx),
		a.Github.info(ctx),
		a.OIDC.info(ctx),
		a.Kubernetes.info(ctx),
		a.JWT.info(ctx),
		a.Cloud.info(ctx),
	}
}

type forwardPrefixContext struct{}

func WithForwardPrefix(ctx context.Context, prefix string) context.Context {
	return context.WithValue(ctx, forwardPrefixContext{}, prefix)
}

func getForwardPrefix(ctx context.Context) string {
	prefix, _ := ctx.Value(forwardPrefixContext{}).(string)
	return prefix
}

// EnabledMethods returns all the AuthenticationMethod instances that have been enabled.
func (a *AuthenticationMethods) EnabledMethods() []StaticAuthenticationMethodInfo {
	var enabled []StaticAuthenticationMethodInfo
	for _, info := range a.AllMethods(context.Background()) {
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
	// used for auth method specific validation
	validate func() error

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

// RequiresCleanup returns true if the method is enabled and requires cleanup.
func (s StaticAuthenticationMethodInfo) RequiresCleanup() bool {
	return s.Enabled && s.RequiresDatabase && s.Cleanup != nil
}

// AuthenticationMethodInfo is a structure which describes properties
// of a particular authentication method.
// i.e. the name and whether or not the method is session compatible.
type AuthenticationMethodInfo struct {
	Method            auth.Method
	SessionCompatible bool
	RequiresDatabase  bool
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
	info(context.Context) AuthenticationMethodInfo
	validate() error
}

// AuthenticationMethod is a container for authentication methods.
// It describes the common properties of all authentication methods.
// Along with leaving a generic slot for the particular method to declare
// its own structural fields. This generic field (Method) must implement
// the AuthenticationMethodInfoProvider to be valid at compile time.
// nolint:musttag
type AuthenticationMethod[C AuthenticationMethodInfoProvider] struct {
	Method  C                              `mapstructure:",squash"`
	Enabled bool                           `json:"enabled,omitempty" mapstructure:"enabled" yaml:"enabled,omitempty"`
	Cleanup *AuthenticationCleanupSchedule `json:"cleanup,omitempty" mapstructure:"cleanup,omitempty" yaml:"cleanup,omitempty"`
}

func (a *AuthenticationMethod[C]) setDefaults(defaults map[string]any) {
	a.Method.setDefaults(defaults)
}

func (a *AuthenticationMethod[C]) info(ctx context.Context) StaticAuthenticationMethodInfo {
	return StaticAuthenticationMethodInfo{
		AuthenticationMethodInfo: a.Method.info(ctx),
		Enabled:                  a.Enabled,
		Cleanup:                  a.Cleanup,

		setDefaults: a.setDefaults,
		validate:    a.validate,
		setEnabled: func() {
			a.Enabled = true
		},
		setCleanup: func(c AuthenticationCleanupSchedule) {
			a.Cleanup = &c
		},
	}
}

func (a *AuthenticationMethod[C]) validate() error {
	if !a.Enabled {
		return nil
	}

	return a.Method.validate()
}

type AuthenticationMethodCloudConfig struct{}

func (a AuthenticationMethodCloudConfig) setDefaults(map[string]any) {}

// info describes properties of the authentication method "cloud".
func (a AuthenticationMethodCloudConfig) info(_ context.Context) AuthenticationMethodInfo {
	return AuthenticationMethodInfo{
		Method:            auth.Method_METHOD_CLOUD,
		SessionCompatible: true,
		RequiresDatabase:  false,
	}
}

func (a AuthenticationMethodCloudConfig) validate() error { return nil }

// AuthenticationMethodTokenConfig contains fields used to configure the authentication
// method "token".
// This authentication method supports the ability to create static tokens via the
// /auth/v1/method/token prefix of endpoints.
type AuthenticationMethodTokenConfig struct {
	Bootstrap AuthenticationMethodTokenBootstrapConfig `json:"bootstrap" mapstructure:"bootstrap" yaml:"bootstrap"`
}

func (a AuthenticationMethodTokenConfig) setDefaults(map[string]any) {}

// info describes properties of the authentication method "token".
func (a AuthenticationMethodTokenConfig) info(_ context.Context) AuthenticationMethodInfo {
	return AuthenticationMethodInfo{
		Method:            auth.Method_METHOD_TOKEN,
		SessionCompatible: false,
		RequiresDatabase:  true,
	}
}

func (a AuthenticationMethodTokenConfig) validate() error { return nil }

// AuthenticationMethodTokenBootstrapConfig contains fields used to configure the
// bootstrap process for the authentication method "token".
type AuthenticationMethodTokenBootstrapConfig struct {
	Token      string            `json:"-" mapstructure:"token" yaml:"token"`
	Expiration time.Duration     `json:"expiration,omitempty" mapstructure:"expiration" yaml:"expiration,omitempty"`
	Metadata   map[string]string `json:"metadata,omitempty" mapstructure:"metadata" yaml:"metadata,omitempty"`
}

// AuthenticationMethodOIDCConfig configures the OIDC authentication method.
// This method can be used to establish browser based sessions.
type AuthenticationMethodOIDCConfig struct {
	EmailMatches []string                                    `json:"emailMatches,omitempty" mapstructure:"email_matches" yaml:"email_matches,omitempty"`
	Providers    map[string]AuthenticationMethodOIDCProvider `json:"providers,omitempty" mapstructure:"providers" yaml:"providers,omitempty"`
}

func (a AuthenticationMethodOIDCConfig) setDefaults(defaults map[string]any) {
	for provider := range a.Providers {
		providerDefaults := map[string]any{}
		a.Providers[provider].setDefaults(providerDefaults)
		defaults[provider] = providerDefaults
	}
}

// info describes properties of the authentication method "oidc".
func (a AuthenticationMethodOIDCConfig) info(ctx context.Context) AuthenticationMethodInfo {
	info := AuthenticationMethodInfo{
		Method:            auth.Method_METHOD_OIDC,
		SessionCompatible: true,
		RequiresDatabase:  true,
	}

	var (
		metadata  = make(map[string]any)
		providers = make(map[string]any, len(a.Providers))
	)

	// this ensures we expose the authorize and callback URL endpoint
	// to the UI via the /auth/v1/method endpoint
	for provider := range a.Providers {
		providers[provider] = map[string]any{
			"authorize_url": path.Join(
				getForwardPrefix(ctx),
				fmt.Sprintf("/auth/v1/method/oidc/%s/authorize", provider),
			),
			"callback_url": path.Join(
				getForwardPrefix(ctx),
				fmt.Sprintf("/auth/v1/method/oidc/%s/callback", provider),
			),
		}
	}

	metadata["providers"] = providers

	info.Metadata, _ = structpb.NewStruct(metadata)

	return info
}

func (a AuthenticationMethodOIDCConfig) validate() error {
	for provider, config := range a.Providers {
		if err := config.validate(); err != nil {
			return fmt.Errorf("provider %q: %w", provider, err)
		}
	}

	return nil
}

// AuthenticationOIDCProvider configures provider credentials
type AuthenticationMethodOIDCProvider struct {
	IssuerURL       string   `json:"issuerURL,omitempty" mapstructure:"issuer_url" yaml:"issuer_url,omitempty"`
	ClientID        string   `json:"-,omitempty" mapstructure:"client_id" yaml:"-"`
	ClientSecret    string   `json:"-" mapstructure:"client_secret" yaml:"-"`
	RedirectAddress string   `json:"redirectAddress,omitempty" mapstructure:"redirect_address" yaml:"redirect_address,omitempty"`
	Nonce           string   `json:"nonce,omitempty" mapstructure:"nonce" yaml:"nonce,omitempty"`
	Scopes          []string `json:"scopes,omitempty" mapstructure:"scopes" yaml:"scopes,omitempty"`
	UsePKCE         bool     `json:"usePKCE,omitempty" mapstructure:"use_pkce" yaml:"use_pkce,omitempty"`
}

func (a AuthenticationMethodOIDCProvider) setDefaults(defaults map[string]any) {
	defaults["nonce"] = "static"
}

func (a AuthenticationMethodOIDCProvider) validate() error {
	if a.ClientID == "" {
		return errFieldWrap("client_id", errValidationRequired)
	}

	if a.ClientSecret == "" {
		return errFieldWrap("client_secret", errValidationRequired)
	}

	if a.RedirectAddress == "" {
		return errFieldWrap("redirect_address", errValidationRequired)
	}

	return nil
}

// AuthenticationCleanupSchedule is used to configure a cleanup goroutine.
type AuthenticationCleanupSchedule struct {
	Interval    time.Duration `json:"interval,omitempty" mapstructure:"interval" yaml:"interval,omitempty"`
	GracePeriod time.Duration `json:"gracePeriod,omitempty" mapstructure:"grace_period" yaml:"grace_period,omitempty"`
}

// AuthenticationMethodKubernetesConfig contains the fields necessary for the Kubernetes authentication
// method to be performed. This method supports Flipt being deployed in a Kubernetes environment
// and allowing it to exchange client tokens for valid service account tokens presented via this method.
type AuthenticationMethodKubernetesConfig struct {
	// DiscoveryURL is the URL to the local Kubernetes cluster serving the "well-known" OIDC discovery endpoint.
	// https://openid.net/specs/openid-connect-discovery-1_0.html
	// The URL is used to fetch the OIDC configuration and subsequently the JWKS certificates.
	DiscoveryURL string `json:"discoveryURL,omitempty" mapstructure:"discovery_url" yaml:"discovery_url,omitempty"`
	// CAPath is the path on disk to the trusted certificate authority certificate for validating
	// HTTPS requests to the issuer.
	CAPath string `json:"caPath,omitempty" mapstructure:"ca_path" yaml:"ca_path,omitempty"`
	// ServiceAccountTokenPath is the location on disk to the Flipt instances service account token.
	// This should be the token issued for the service account associated with Flipt in the environment.
	ServiceAccountTokenPath string `json:"serviceAccountTokenPath,omitempty" mapstructure:"service_account_token_path" yaml:"service_account_token_path,omitempty"`
}

func (a AuthenticationMethodKubernetesConfig) setDefaults(defaults map[string]any) {
	defaults["discovery_url"] = "https://kubernetes.default.svc.cluster.local"
	defaults["ca_path"] = "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"
	defaults["service_account_token_path"] = "/var/run/secrets/kubernetes.io/serviceaccount/token"
}

// info describes properties of the authentication method "kubernetes".
func (a AuthenticationMethodKubernetesConfig) info(_ context.Context) AuthenticationMethodInfo {
	return AuthenticationMethodInfo{
		Method:            auth.Method_METHOD_KUBERNETES,
		SessionCompatible: false,
		RequiresDatabase:  true,
	}
}

func (a AuthenticationMethodKubernetesConfig) validate() error { return nil }

// AuthenticationMethodGithubConfig contains configuration and information for completing an OAuth
// 2.0 flow with GitHub as a provider.
type AuthenticationMethodGithubConfig struct {
	ServerURL            string              `json:"serverUrl,omitempty" mapstructure:"server_url" yaml:"server_url,omitempty"`
	ApiURL               string              `json:"apiUrl,omitempty" mapstructure:"api_url" yaml:"api_url,omitempty"`
	ClientId             string              `json:"-" mapstructure:"client_id" yaml:"-"`
	ClientSecret         string              `json:"-" mapstructure:"client_secret" yaml:"-"`
	RedirectAddress      string              `json:"redirectAddress,omitempty" mapstructure:"redirect_address" yaml:"redirect_address,omitempty"`
	Scopes               []string            `json:"scopes,omitempty" mapstructure:"scopes" yaml:"scopes,omitempty"`
	AllowedOrganizations []string            `json:"allowedOrganizations,omitempty" mapstructure:"allowed_organizations" yaml:"allowed_organizations,omitempty"`
	AllowedTeams         map[string][]string `json:"allowedTeams,omitempty" mapstructure:"allowed_teams" yaml:"allowed_teams,omitempty"`
}

func (a AuthenticationMethodGithubConfig) setDefaults(defaults map[string]any) {}

// info describes properties of the authentication method "github".
func (a AuthenticationMethodGithubConfig) info(ctx context.Context) AuthenticationMethodInfo {
	info := AuthenticationMethodInfo{
		Method:            auth.Method_METHOD_GITHUB,
		SessionCompatible: true,
		RequiresDatabase:  true,
	}

	metadata := make(map[string]any)

	metadata["authorize_url"] = path.Join(getForwardPrefix(ctx), "/auth/v1/method/github/authorize")
	metadata["callback_url"] = path.Join(getForwardPrefix(ctx), "/auth/v1/method/github/callback")

	info.Metadata, _ = structpb.NewStruct(metadata)

	return info
}

func (a AuthenticationMethodGithubConfig) validate() error {
	errWrap := func(err error) error {
		return fmt.Errorf("provider %q: %w", "github", err)
	}

	if a.ClientId == "" {
		return errWrap(errFieldWrap("client_id", errValidationRequired))
	}

	if a.ClientSecret == "" {
		return errWrap(errFieldWrap("client_secret", errValidationRequired))
	}

	if a.RedirectAddress == "" {
		return errWrap(errFieldWrap("redirect_address", errValidationRequired))
	}

	// ensure scopes contain read:org if allowed organizations is not empty
	if len(a.AllowedOrganizations) > 0 && !slices.Contains(a.Scopes, "read:org") {
		return errWrap(errFieldWrap("scopes", fmt.Errorf("must contain read:org when allowed_organizations is not empty")))
	}

	// ensure all the declared organizations were declared in allowed organizations
	if len(a.AllowedTeams) > 0 {
		for org := range a.AllowedTeams {
			if !slices.Contains(a.AllowedOrganizations, org) {
				return errWrap(errFieldWrap(
					"allowed_teams",
					fmt.Errorf("the organization '%s' was not declared in 'allowed_organizations' field", org),
				))
			}
		}
	}

	return nil
}

type AuthenticationMethodJWTConfig struct {
	// ValidateClaims is used to validate the claims of the JWT token.
	ValidateClaims struct {
		// Issuer is the issuer of the JWT token.
		Issuer string `json:"-" mapstructure:"issuer" yaml:"issuer,omitempty"`
		// Subject is the subject of the JWT token.
		Subject string `json:"-" mapstructure:"subject" yaml:"subject,omitempty"`
		// Audiences is the audience of the JWT token.
		Audiences []string `json:"-" mapstructure:"audiences" yaml:"audiences,omitempty"`
	} `json:"-" mapstructure:"validate_claims" yaml:"validate_claims,omitempty"`
	// JWKsURL is the URL to the JWKS endpoint.
	// This is used to fetch the public keys used to validate the JWT token.
	JWKSURL string `json:"-" mapstructure:"jwks_url" yaml:"jwks_url,omitempty"`
	// PublicKeyFile is the path to the public PEM encoded key file on disk.
	PublicKeyFile string `json:"-" mapstructure:"public_key_file" yaml:"public_key_file,omitempty"`
}

func (a AuthenticationMethodJWTConfig) setDefaults(map[string]any) {}

// info describes properties of the authentication method "jwt".
func (a AuthenticationMethodJWTConfig) info(_ context.Context) AuthenticationMethodInfo {
	return AuthenticationMethodInfo{
		Method:            auth.Method_METHOD_JWT,
		SessionCompatible: false,
		RequiresDatabase:  false,
	}
}

func (a AuthenticationMethodJWTConfig) validate() error {
	setFields := nonEmpty([]string{a.JWKSURL, a.PublicKeyFile})
	if setFields < 1 {
		return fmt.Errorf("one of jwks_url or public_key_file is required")
	}

	if setFields > 1 {
		return fmt.Errorf("only one of jwks_url or public_key_file can be set")
	}

	if a.JWKSURL != "" {
		// ensure jwks url is valid
		if _, err := url.Parse(a.JWKSURL); err != nil {
			return errFieldWrap("jwks_url", err)
		}
	}

	if a.PublicKeyFile != "" {
		// ensure public key file exists
		if _, err := os.Stat(a.PublicKeyFile); err != nil {
			return errFieldWrap("public_key_file", err)
		}
	}

	return nil
}

func nonEmpty(values []string) int {
	set := filter(values, func(s string) bool {
		return s != ""
	})

	return len(set)
}

func filter[T any](ss []T, test func(T) bool) (ret []T) {
	for _, s := range ss {
		if test(s) {
			ret = append(ret, s)
		}
	}
	return
}
