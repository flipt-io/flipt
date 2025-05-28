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
		// Evaluation refers to the section of the API with the prefix /evaluation/v1
		Evaluation bool `json:"evaluation,omitempty" mapstructure:"evaluation" yaml:"evaluation,omitempty"`
		// OFREP refers to the section of the API with the prefix /ofrep
		OFREP bool `json:"ofrep,omitempty" mapstructure:"ofrep" yaml:"ofrep,omitempty"`
	} `json:"exclude,omitempty" mapstructure:"exclude" yaml:"exclude,omitempty"`

	Session AuthenticationSessionConfig `json:"session,omitempty" mapstructure:"session" yaml:"session,omitempty"`
	Methods AuthenticationMethodsConfig `json:"methods,omitempty" mapstructure:"methods" yaml:"methods,omitempty"`
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

// IsZero returns true if the authentication config is not enabled.
// This is used for marshalling to YAML for `config init`.
func (c AuthenticationConfig) IsZero() bool {
	return !c.Enabled()
}

func (c *AuthenticationConfig) setDefaults(v *viper.Viper) error {
	// Set top-level authentication defaults
	v.SetDefault("authentication.required", false)
	v.SetDefault("authentication.session.storage.type", "memory")
	v.SetDefault("authentication.session.storage.cleanup.grace_period", "30m")
	v.SetDefault("authentication.session.token_lifetime", "24h")
	v.SetDefault("authentication.session.state_lifetime", "10m")

	// Set defaults for each authentication method
	for _, info := range c.Methods.AllMethods(context.Background()) {
		prefix := "authentication.methods." + info.Name()
		v.SetDefault(prefix+".enabled", false)
		if v.GetBool(prefix + ".enabled") {
			// If enabled, apply method-specific defaults
			methodDefaults := map[string]any{}
			info.setDefaults(methodDefaults)
			for k, val := range methodDefaults {
				v.SetDefault(prefix+"."+k, val)
			}
		}
	}

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
	// ensure that when a session compatible authentication method has been
	// enabled that the session cookie domain has been configured with a non
	// empty value.
	if c.SessionEnabled() {
		if c.Session.Domain == "" {
			return errFieldRequired("authentication", "session_domain")
		}

		host, err := getHostname(c.Session.Domain)
		if err != nil {
			return errFieldWrap("authentication", "session_domain", err)
		}

		// strip scheme and port from domain
		// domain cookies are not allowed to have a scheme or port
		// https://github.com/golang/go/issues/28297
		c.Session.Domain = host

		if err := c.Session.Storage.validate(); err != nil {
			return errFieldWrap("authentication", "session_storage", err)
		}
	}

	for _, info := range c.Methods.AllMethods(context.Background()) {
		if err := info.validate(); err != nil {
			return err
		}
	}

	if c.Methods.Kubernetes.Enabled && c.Methods.JWT.Enabled {
		return errFieldWrap("authentication", "methods", fmt.Errorf("kubernetes and jwt methods cannot currently both be enabled at the same time"))
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

// AuthenticationSessionConfig configures the session produced for browsers when
// establishing authentication via HTTP.
type AuthenticationSessionConfig struct {
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
	CSRF AuthenticationSessionCSRFConfig `json:"csrf,omitempty" mapstructure:"csrf" yaml:"csrf,omitempty"`
	// Storage configures the storage mechanism for the session.
	Storage AuthenticationSessionStorageConfig `json:"storage,omitempty" mapstructure:"storage" yaml:"storage,omitempty"`
}

type AuthenticationSessionStorageType string

const (
	AuthenticationSessionStorageTypeMemory = AuthenticationSessionStorageType("memory")
	AuthenticationSessionStorageTypeRedis  = AuthenticationSessionStorageType("redis")
)

var (
	_ validator = (*AuthenticationSessionStorageConfig)(nil)
	_ defaulter = (*AuthenticationSessionStorageConfig)(nil)
)

type AuthenticationSessionStorageConfig struct {
	Type    AuthenticationSessionStorageType          `json:"type" mapstructure:"type" yaml:"type"`
	Redis   AuthenticationSessionStorageRedisConfig   `json:"redis,omitempty" mapstructure:"redis" yaml:"redis,omitempty"`
	Cleanup AuthenticationSessionStorageCleanupConfig `json:"cleanup,omitempty" mapstructure:"cleanup" yaml:"cleanup,omitempty"`
}

func (c AuthenticationSessionStorageConfig) validate() error {
	if err := c.Cleanup.validate(); err != nil {
		return err
	}

	if c.Type == AuthenticationSessionStorageTypeRedis {
		return c.Redis.validate()
	}

	return nil
}

func (c *AuthenticationSessionStorageConfig) setDefaults(v *viper.Viper) error {
	v.SetDefault("type", AuthenticationSessionStorageTypeMemory)

	return c.Redis.setDefaults(v)
}

// AuthenticationSessionStorageCleanupConfig configures the schedule for cleaning up expired authentication records.
type AuthenticationSessionStorageCleanupConfig struct {
	GracePeriod time.Duration `json:"gracePeriod,omitempty" mapstructure:"grace_period" yaml:"grace_period,omitempty"`
}

func (c AuthenticationSessionStorageCleanupConfig) validate() error {
	if c.GracePeriod < 0 {
		return errFieldPositiveDuration("", "cleanup_grace_period")
	}

	return nil
}

type RedisCacheMode string

const (
	RedisCacheModeSingle  RedisCacheMode = "single"
	RedisCacheModeCluster RedisCacheMode = "cluster"
)

var (
	_ validator = (*AuthenticationSessionStorageRedisConfig)(nil)
	_ defaulter = (*AuthenticationSessionStorageRedisConfig)(nil)
)

// AuthenticationSessionStorageRedisConfig contains fields, which configure the connection
// credentials for redis backed session storage.
type AuthenticationSessionStorageRedisConfig struct {
	Host            string         `json:"host,omitempty" mapstructure:"host" yaml:"host,omitempty"`
	Port            int            `json:"port,omitempty" mapstructure:"port" yaml:"port,omitempty"`
	RequireTLS      bool           `json:"requireTLS,omitempty" mapstructure:"require_tls" yaml:"require_tls,omitempty"`
	Username        string         `json:"-" mapstructure:"username" yaml:"-"`
	Password        string         `json:"-" mapstructure:"password" yaml:"-"`
	DB              int            `json:"db,omitempty" mapstructure:"db" yaml:"db,omitempty"`
	PoolSize        int            `json:"poolSize" mapstructure:"pool_size" yaml:"pool_size"`
	MinIdleConn     int            `json:"minIdleConn" mapstructure:"min_idle_conn" yaml:"min_idle_conn"`
	ConnMaxIdleTime time.Duration  `json:"connMaxIdleTime" mapstructure:"conn_max_idle_time" yaml:"conn_max_idle_time"`
	NetTimeout      time.Duration  `json:"netTimeout" mapstructure:"net_timeout" yaml:"net_timeout"`
	CaCertBytes     string         `json:"-" mapstructure:"ca_cert_bytes" yaml:"-"`
	CaCertPath      string         `json:"-" mapstructure:"ca_cert_path" yaml:"-"`
	InsecureSkipTLS bool           `json:"-" mapstructure:"insecure_skip_tls" yaml:"-"`
	Prefix          string         `json:"prefix" mapstructure:"prefix" yaml:"prefix"`
	Mode            RedisCacheMode `json:"mode" mapstructure:"mode" yaml:"mode"`
}

func (cfg *AuthenticationSessionStorageRedisConfig) validate() error {
	if cfg.CaCertBytes != "" && cfg.CaCertPath != "" {
		return errString("", "please provide exclusively one of ca_cert_bytes or ca_cert_path")
	}

	return nil
}

func (cfg *AuthenticationSessionStorageRedisConfig) setDefaults(v *viper.Viper) error {
	v.SetDefault("mode", RedisCacheModeSingle)
	v.SetDefault("prefix", "flipt")
	return nil
}

// AuthenticationSessionCSRFConfig configures cross-site request forgery prevention.
type AuthenticationSessionCSRFConfig struct {
	// Key is the private key string used to authenticate csrf tokens.
	Key string `json:"-" mapstructure:"key"`
}

// AuthenticationMethodsConfig is a set of configuration for each authentication
// method available for use within Flipt.
type AuthenticationMethodsConfig struct {
	Token      AuthenticationMethod[AuthenticationMethodTokenConfig]      `json:"token,omitempty" mapstructure:"token" yaml:"token,omitempty"`
	Github     AuthenticationMethod[AuthenticationMethodGithubConfig]     `json:"github,omitempty" mapstructure:"github" yaml:"github,omitempty"`
	OIDC       AuthenticationMethod[AuthenticationMethodOIDCConfig]       `json:"oidc,omitempty" mapstructure:"oidc" yaml:"oidc,omitempty"`
	Kubernetes AuthenticationMethod[AuthenticationMethodKubernetesConfig] `json:"kubernetes,omitempty" mapstructure:"kubernetes" yaml:"kubernetes,omitempty"`
	JWT        AuthenticationMethod[AuthenticationMethodJWTConfig]        `json:"jwt,omitempty" mapstructure:"jwt" yaml:"jwt,omitempty"`
}

// AllMethods returns all the AuthenticationMethod instances available.
func (a *AuthenticationMethodsConfig) AllMethods(ctx context.Context) []StaticAuthenticationMethodInfo {
	return []StaticAuthenticationMethodInfo{
		a.Token.info(ctx),
		a.Github.info(ctx),
		a.OIDC.info(ctx),
		a.Kubernetes.info(ctx),
		a.JWT.info(ctx),
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
func (a *AuthenticationMethodsConfig) EnabledMethods() []StaticAuthenticationMethodInfo {
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

	// used for bootstrapping defaults
	setDefaults func(map[string]any)
	// used for auth method specific validation
	validate func() error

	// used for testing purposes to ensure all methods
	// are appropriately cleaned up via the background process.
	setEnabled func()
}

// Enable can only be called in a testing scenario.
// It is used to enable a target method without having a concrete reference.
func (s StaticAuthenticationMethodInfo) Enable(t *testing.T) {
	s.setEnabled()
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
	Method  C    `mapstructure:",squash"`
	Enabled bool `json:"enabled,omitempty" mapstructure:"enabled" yaml:"enabled,omitempty"`
}

func (a *AuthenticationMethod[C]) setDefaults(defaults map[string]any) {
	a.Method.setDefaults(defaults)
}

func (a *AuthenticationMethod[C]) info(ctx context.Context) StaticAuthenticationMethodInfo {
	return StaticAuthenticationMethodInfo{
		AuthenticationMethodInfo: a.Method.info(ctx),
		Enabled:                  a.Enabled,

		setDefaults: a.setDefaults,
		validate:    a.validate,
		setEnabled: func() {
			a.Enabled = true
		},
	}
}

func (a *AuthenticationMethod[C]) validate() error {
	if !a.Enabled {
		return nil
	}

	return a.Method.validate()
}

var (
	_ validator = (*AuthenticationMethodTokenConfig)(nil)
	_ validator = (*AuthenticationMethodOIDCConfig)(nil)
	_ validator = (*AuthenticationMethodKubernetesConfig)(nil)
	_ validator = (*AuthenticationMethodGithubConfig)(nil)
	_ validator = (*AuthenticationMethodJWTConfig)(nil)
)

// AuthenticationMethodTokenConfig contains fields used to configure the authentication
// method "token".
type AuthenticationMethodTokenConfig struct {
	Storage AuthenticationMethodTokenStorage `json:"storage" mapstructure:"storage" yaml:"storage"`
}

func (a AuthenticationMethodTokenConfig) setDefaults(defaults map[string]any) {
	defaults["storage"] = map[string]any{
		"type": "static",
	}
}

// info describes properties of the authentication method "token".
func (a AuthenticationMethodTokenConfig) info(_ context.Context) AuthenticationMethodInfo {
	return AuthenticationMethodInfo{
		Method:            auth.Method_METHOD_TOKEN,
		SessionCompatible: false,
	}
}

func (a AuthenticationMethodTokenConfig) validate() error {
	if a.Storage.Type != AuthenticationMethodTokenStorageTypeStatic {
		return errFieldWrap("storage type", "authentication.method.token.storage.type", fmt.Errorf("unexpected type: %q", a.Storage.Type))
	}

	return nil
}

type AuthenticationMethodTokenStorageType string

const (
	AuthenticationMethodTokenStorageTypeStatic = AuthenticationMethodTokenStorageType("static")
)

type AuthenticationMethodTokenStorage struct {
	Type   AuthenticationMethodTokenStorageType       `json:"type" mapstructure:"type" yaml:"type"`
	Tokens map[string]AuthenticationMethodStaticToken `json:"tokens" mapstructure:"tokens" yaml:"tokens"`
}

// AuthenticationMethodStaticToken contains fields used to configure the
// static token authentication method.
type AuthenticationMethodStaticToken struct {
	Credential string            `json:"-" mapstructure:"credential" yaml:"-"`
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
	for _, config := range a.Providers {
		if err := config.validate(); err != nil {
			return err
		}
	}

	return nil
}

// AuthenticationOIDCProvider configures provider credentials
type AuthenticationMethodOIDCProvider struct {
	IssuerURL       string   `json:"issuerURL,omitempty" mapstructure:"issuer_url" yaml:"issuer_url,omitempty"`
	ClientID        string   `json:"-" mapstructure:"client_id" yaml:"-"`
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
		return errFieldRequired("authentication", "client_id")
	}

	if a.ClientSecret == "" {
		return errFieldRequired("authentication", "client_secret")
	}

	if a.RedirectAddress == "" {
		return errFieldRequired("authentication", "redirect_address")
	}

	return nil
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
	}

	metadata := make(map[string]any)

	metadata["authorize_url"] = path.Join(getForwardPrefix(ctx), "/auth/v1/method/github/authorize")
	metadata["callback_url"] = path.Join(getForwardPrefix(ctx), "/auth/v1/method/github/callback")

	info.Metadata, _ = structpb.NewStruct(metadata)

	return info
}

func (a AuthenticationMethodGithubConfig) validate() error {
	if a.ClientId == "" {
		return errFieldRequired("authentication", "client_id")
	}

	if a.ClientSecret == "" {
		return errFieldRequired("authentication", "client_secret")
	}

	if a.RedirectAddress == "" {
		return errFieldRequired("authentication", "redirect_address")
	}

	// ensure scopes contain read:org if allowed organizations is not empty
	if len(a.AllowedOrganizations) > 0 && !slices.Contains(a.Scopes, "read:org") {
		return errFieldWrap("authentication", "scopes", fmt.Errorf("must contain read:org when allowed_organizations is not empty"))
	}

	// ensure all the declared organizations were declared in allowed organizations
	if len(a.AllowedTeams) > 0 {
		for org := range a.AllowedTeams {
			if !slices.Contains(a.AllowedOrganizations, org) {
				return errFieldWrap(
					"authentication",
					"allowed_teams",
					fmt.Errorf("the organization '%s' was not declared in 'allowed_organizations' field", org),
				)
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
	}
}

func (a AuthenticationMethodJWTConfig) validate() error {
	setFields := nonEmpty([]string{a.JWKSURL, a.PublicKeyFile})
	if setFields < 1 {
		return errString("authentication", "one of jwks_url or public_key_file is required")
	}

	if setFields > 1 {
		return errString("authentication", "only one of jwks_url or public_key_file can be set")
	}

	if a.JWKSURL != "" {
		// ensure jwks url is valid
		if _, err := url.Parse(a.JWKSURL); err != nil {
			return errFieldWrap("authentication", "jwks_url", err)
		}
	}

	if a.PublicKeyFile != "" {
		// ensure public key file exists
		if _, err := os.Stat(a.PublicKeyFile); err != nil {
			return errFieldWrap("authentication", "public_key_file", err)
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
