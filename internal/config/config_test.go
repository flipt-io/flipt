package config

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/iancoleman/strcase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gocloud.dev/blob"
	"gocloud.dev/blob/memblob"
	"gopkg.in/yaml.v2"
)

func TestScheme(t *testing.T) {
	tests := []struct {
		name   string
		scheme Scheme
		want   string
	}{
		{
			name:   "https",
			scheme: HTTPS,
			want:   "https",
		},
		{
			name:   "http",
			scheme: HTTP,
			want:   "http",
		},
	}

	for _, tt := range tests {
		var (
			scheme = tt.scheme
			want   = tt.want
		)

		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, want, scheme.String())
		})
	}
}

func TestLogEncoding(t *testing.T) {
	tests := []struct {
		name     string
		encoding LogEncoding
		want     string
	}{
		{
			name:     "console",
			encoding: LogEncodingConsole,
			want:     "console",
		},
		{
			name:     "json",
			encoding: LogEncodingJSON,
			want:     "json",
		},
	}

	for _, tt := range tests {
		var (
			encoding = tt.encoding
			want     = tt.want
		)

		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, want, encoding.String())
		})
	}
}

func TestLoad(t *testing.T) {
	tests := []struct {
		name         string
		path         string
		wantErr      error
		envOverrides map[string]string
		expected     func() *Config
		warnings     []string
	}{
		{
			name:     "defaults",
			path:     "",
			expected: Default,
		},
		{
			name: "defaults with env overrides",
			path: "",
			envOverrides: map[string]string{
				"FLIPT_LOG_LEVEL":        "DEBUG",
				"FLIPT_SERVER_HTTP_PORT": "8081",
			},
			expected: func() *Config {
				cfg := Default()
				cfg.Log.Level = "DEBUG"
				cfg.Server.HTTPPort = 8081
				return cfg
			},
		},
		{
			name: "metrics disabled",
			path: "./testdata/metrics/disabled.yml",
			expected: func() *Config {
				cfg := Default()
				cfg.Metrics.Enabled = false
				return cfg
			},
		},
		{
			name: "metrics OTLP",
			path: "./testdata/metrics/otlp.yml",
			expected: func() *Config {
				cfg := Default()
				cfg.Metrics.Enabled = true
				cfg.Metrics.Exporter = MetricsOTLP
				return cfg
			},
		},
		{
			name:    "server https missing cert file",
			path:    "./testdata/server/https_missing_cert_file.yml",
			wantErr: errors.New("server: cert_file non-empty value is required"),
		},
		{
			name:    "server https missing cert key",
			path:    "./testdata/server/https_missing_cert_key.yml",
			wantErr: errors.New("server: cert_key non-empty value is required"),
		},
		{
			name:    "server https defined but not found cert file",
			path:    "./testdata/server/https_not_found_cert_file.yml",
			wantErr: errors.New("server: cert_file stat ./testdata/ssl_cert_not_exist.pem: no such file or directory"),
		},
		{
			name:    "server https defined but not found cert key",
			path:    "./testdata/server/https_not_found_cert_key.yml",
			wantErr: errors.New("server: cert_key stat ./testdata/ssl_cert_not_exist.key: no such file or directory"),
		},
		{
			name: "authentication token with provided bootstrap token",
			path: "./testdata/authentication/token_bootstrap_token.yml",
			expected: func() *Config {
				cfg := Default()
				cfg.Authentication.Required = true
				cfg.Authentication.Methods = AuthenticationMethodsConfig{
					Token: AuthenticationMethod[AuthenticationMethodTokenConfig]{
						Enabled: true,
						Method: AuthenticationMethodTokenConfig{
							Storage: AuthenticationMethodTokenStorage{
								Type: AuthenticationMethodTokenStorageTypeStatic,
								Tokens: map[string]AuthenticationMethodStaticToken{
									"bootstrap": {
										Credential: "s3cr3t!",
									},
								},
							},
						},
					},
				}
				return cfg
			},
		},
		{
			name: "authentication session strip domain scheme/port",
			path: "./testdata/authentication/session_domain_scheme_port.yml",
			expected: func() *Config {
				cfg := Default()
				cfg.Authentication.Required = true
				cfg.Authentication.Session = AuthenticationSessionConfig{
					Domain: "localhost",
					Storage: AuthenticationSessionStorageConfig{
						Type: AuthenticationSessionStorageTypeMemory,
						Cleanup: AuthenticationSessionStorageCleanupConfig{
							GracePeriod: 30 * time.Minute,
						},
					},
					TokenLifetime: 24 * time.Hour,
					StateLifetime: 10 * time.Minute,
					CSRF: AuthenticationSessionCSRFConfig{
						Secure: true,
					},
				}
				cfg.Authentication.Methods = AuthenticationMethodsConfig{
					Token: AuthenticationMethod[AuthenticationMethodTokenConfig]{
						Enabled: true,
						Method: AuthenticationMethodTokenConfig{
							Storage: AuthenticationMethodTokenStorage{
								Type: AuthenticationMethodTokenStorageTypeStatic,
							},
						},
					},
					OIDC: AuthenticationMethod[AuthenticationMethodOIDCConfig]{
						Enabled: true,
					},
				}
				return cfg
			},
		},
		{
			name: "authentication oidc with nonce",
			path: "./testdata/authentication/oidc.yml",
			expected: func() *Config {
				cfg := Default()
				cfg.Authentication.Required = true
				cfg.Authentication.Session = AuthenticationSessionConfig{
					Domain: "localhost",
					Storage: AuthenticationSessionStorageConfig{
						Type: AuthenticationSessionStorageTypeMemory,
						Cleanup: AuthenticationSessionStorageCleanupConfig{
							GracePeriod: 30 * time.Minute,
						},
					},
					TokenLifetime: 24 * time.Hour,
					StateLifetime: 10 * time.Minute,
					CSRF: AuthenticationSessionCSRFConfig{
						Secure: true,
					},
				}
				cfg.Authentication.Methods = AuthenticationMethodsConfig{
					OIDC: AuthenticationMethod[AuthenticationMethodOIDCConfig]{
						Enabled: true,
						Method: AuthenticationMethodOIDCConfig{
							Providers: map[string]AuthenticationMethodOIDCProvider{
								"foo": {
									ClientID:        "client_id",
									ClientSecret:    "client_secret",
									RedirectAddress: "http://localhost:8080",
									Nonce:           "nonce-asdf8080",
									Scopes:          []string{"user:email"},
								},
								"bar": {
									ClientID:        "client_id",
									ClientSecret:    "client_secret",
									RedirectAddress: "http://localhost:8080",
									Nonce:           "",
									Scopes:          []string{"user:email"},
								},
							},
						},
					},
				}
				return cfg
			},
		},
		{
			name: "authentication kubernetes defaults when enabled",
			path: "./testdata/authentication/kubernetes.yml",
			expected: func() *Config {
				cfg := Default()
				cfg.Authentication.Required = true
				cfg.Authentication.Methods = AuthenticationMethodsConfig{
					Kubernetes: AuthenticationMethod[AuthenticationMethodKubernetesConfig]{
						Enabled: true,
						Method: AuthenticationMethodKubernetesConfig{
							DiscoveryURL:            "https://kubernetes.default.svc.cluster.local",
							CAPath:                  "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt",
							ServiceAccountTokenPath: "/var/run/secrets/kubernetes.io/serviceaccount/token",
						},
					},
				}
				return cfg
			},
		},
		{
			name:    "authentication github requires read:org scope when allowing orgs",
			path:    "./testdata/authentication/github_missing_org_scope.yml",
			wantErr: errors.New("authentication: scopes must contain read:org when allowed_organizations is not empty"),
		},
		{
			name:    "authentication github missing client id",
			path:    "./testdata/authentication/github_missing_client_id.yml",
			wantErr: errors.New("authentication: client_id non-empty value is required"),
		},
		{
			name:    "authentication github missing client secret",
			path:    "./testdata/authentication/github_missing_client_secret.yml",
			wantErr: errors.New("authentication: client_secret non-empty value is required"),
		},
		{
			name:    "authentication github missing client id",
			path:    "./testdata/authentication/github_missing_redirect_address.yml",
			wantErr: errors.New("authentication: redirect_address non-empty value is required"),
		},
		{
			name:    "authentication github has non declared org in allowed_teams",
			path:    "./testdata/authentication/github_missing_org_when_declaring_allowed_teams.yml",
			wantErr: errors.New("authentication: allowed_teams the organization 'my-other-org' was not declared in 'allowed_organizations' field"),
		},
		{
			name:    "authentication oidc missing client id",
			path:    "./testdata/authentication/oidc_missing_client_id.yml",
			wantErr: errors.New("authentication: client_id non-empty value is required"),
		},
		{
			name:    "authentication oidc missing client secret",
			path:    "./testdata/authentication/oidc_missing_client_secret.yml",
			wantErr: errors.New("authentication: client_secret non-empty value is required"),
		},
		{
			name:    "authentication oidc missing client id",
			path:    "./testdata/authentication/oidc_missing_redirect_address.yml",
			wantErr: errors.New("authentication: redirect_address non-empty value is required"),
		},
		{
			name:    "authentication jwt public key file or jwks url required",
			path:    "./testdata/authentication/jwt_missing_key_file_and_jwks_url.yml",
			wantErr: errors.New("authentication: one of jwks_url or public_key_file is required"),
		},
		{
			name:    "authentication jwt public key file and jwks url mutually exclusive",
			path:    "./testdata/authentication/jwt_key_file_and_jwks_url.yml",
			wantErr: errors.New("authentication: only one of jwks_url or public_key_file can be set"),
		},
		{
			name:    "authentication jwks invalid url",
			path:    "./testdata/authentication/jwt_invalid_jwks_url.yml",
			wantErr: errors.New("authentication: jwks_url parse \" http://localhost:8080/.well-known/jwks.json\": first path segment in URL cannot contain colon"),
		},
		{
			name:    "authentication jwt public key file not found",
			path:    "./testdata/authentication/jwt_key_file_not_found.yml",
			wantErr: errors.New("authentication: public_key_file stat testdata/authentication/jwt_key_file.pem: no such file or directory"),
		},
		{
			name:    "authorization required without authentication",
			path:    "./testdata/authorization/authentication_not_required.yml",
			wantErr: errors.New("authorization: requires authentication to also be required"),
		},
		{
			name: "authorization with all authentication methods enabled",
			path: "./testdata/authorization/all_authentication_methods_enabled.yml",
			expected: func() *Config {
				cfg := Default()

				cfg.Authorization = AuthorizationConfig{
					Required: true,
					Backend:  AuthorizationBackendLocal,
					Local: &AuthorizationLocalConfig{
						Policy: &AuthorizationSourceLocalConfig{
							Path:         "/path/to/policy.rego",
							PollInterval: 5 * time.Minute,
						},
					},
				}

				cfg.Authentication = AuthenticationConfig{
					Required: true,
					Session: AuthenticationSessionConfig{
						Domain:        "auth.flipt.io",
						Secure:        true,
						TokenLifetime: 24 * time.Hour,
						StateLifetime: 10 * time.Minute,
						CSRF: AuthenticationSessionCSRFConfig{
							Key:    "abcdefghijklmnopqrstuvwxyz1234567890", //gitleaks:allow
							Secure: true,
						},
						Storage: AuthenticationSessionStorageConfig{
							Type: AuthenticationSessionStorageTypeMemory,
							Cleanup: AuthenticationSessionStorageCleanupConfig{
								GracePeriod: 30 * time.Minute,
							},
						},
					},
					Methods: AuthenticationMethodsConfig{
						Token: AuthenticationMethod[AuthenticationMethodTokenConfig]{
							Enabled: true,
							Method: AuthenticationMethodTokenConfig{
								Storage: AuthenticationMethodTokenStorage{
									Type: AuthenticationMethodTokenStorageTypeStatic,
								},
							},
						},
						OIDC: AuthenticationMethod[AuthenticationMethodOIDCConfig]{
							Method: AuthenticationMethodOIDCConfig{
								Providers: map[string]AuthenticationMethodOIDCProvider{
									"google": {
										IssuerURL:       "http://accounts.google.com",
										ClientID:        "abcdefg",
										ClientSecret:    "bcdefgh",
										RedirectAddress: "http://auth.flipt.io",
									},
								},
							},
							Enabled: true,
						},
						Kubernetes: AuthenticationMethod[AuthenticationMethodKubernetesConfig]{
							Enabled: true,
							Method: AuthenticationMethodKubernetesConfig{
								DiscoveryURL:            "https://some-other-k8s.namespace.svc",
								CAPath:                  "/path/to/ca/certificate/ca.pem",
								ServiceAccountTokenPath: "/path/to/sa/token",
							},
						},
						Github: AuthenticationMethod[AuthenticationMethodGithubConfig]{
							Method: AuthenticationMethodGithubConfig{
								ClientId:        "abcdefg",
								ClientSecret:    "bcdefgh",
								RedirectAddress: "http://auth.flipt.io",
							},
							Enabled: true,
						},
					},
				}

				return cfg
			},
		},
		{
			name: "advanced",
			path: "./testdata/advanced.yml",
			expected: func() *Config {
				cfg := Default()

				cfg.Log = LogConfig{
					Level:     "WARN",
					File:      "testLogFile.txt",
					Encoding:  LogEncodingJSON,
					GRPCLevel: "ERROR",
					Keys: LogKeys{
						Time:    "time",
						Level:   "level",
						Message: "msg",
					},
				}

				cfg.Cors = CorsConfig{
					Enabled:        true,
					AllowedOrigins: []string{"foo.com", "bar.com", "baz.com"},
					AllowedHeaders: []string{"X-Some-Header", "X-Some-Other-Header"},
				}

				cfg.Server = ServerConfig{
					Host:      "127.0.0.1",
					Protocol:  HTTPS,
					HTTPPort:  8081,
					HTTPSPort: 8080,
					GRPCPort:  9001,
					CertFile:  "./testdata/ssl_cert.pem",
					CertKey:   "./testdata/ssl_key.pem",
				}

				cfg.Storage = StoragesConfig{
					"default": {
						Backend: StorageBackendConfig{
							Type: MemoryStorageBackendType,
						},
						Remote:       "https://github.com/flipt-io/flipt.git",
						FetchPolicy:  "strict",
						Branch:       "main",
						PollInterval: 5 * time.Second,
						Credentials:  "git",
					},
				}

				cfg.Credentials = CredentialsConfig{
					"git": {
						Type: "basic",
						Basic: &BasicAuthConfig{
							Username: "user",
							Password: "pass",
						},
					},
				}

				cfg.Meta = MetaConfig{
					CheckForUpdates:  false,
					TelemetryEnabled: false,
				}

				cfg.Authentication = AuthenticationConfig{
					Required: true,
					Session: AuthenticationSessionConfig{
						Domain:        "auth.flipt.io",
						Secure:        true,
						TokenLifetime: 24 * time.Hour,
						StateLifetime: 10 * time.Minute,
						CSRF: AuthenticationSessionCSRFConfig{
							Key:    "abcdefghijklmnopqrstuvwxyz1234567890", //gitleaks:allow
							Secure: true,
						},
						Storage: AuthenticationSessionStorageConfig{
							Type: AuthenticationSessionStorageTypeMemory,
							Cleanup: AuthenticationSessionStorageCleanupConfig{
								GracePeriod: 30 * time.Minute,
							},
						},
					},
					Methods: AuthenticationMethodsConfig{
						Token: AuthenticationMethod[AuthenticationMethodTokenConfig]{
							Enabled: true,
							Method: AuthenticationMethodTokenConfig{
								Storage: AuthenticationMethodTokenStorage{
									Type: AuthenticationMethodTokenStorageTypeStatic,
									Tokens: map[string]AuthenticationMethodStaticToken{
										"static": {
											Credential: "abcdefg",
										},
									},
								},
							},
						},
						OIDC: AuthenticationMethod[AuthenticationMethodOIDCConfig]{
							Method: AuthenticationMethodOIDCConfig{
								Providers: map[string]AuthenticationMethodOIDCProvider{
									"google": {
										IssuerURL:       "http://accounts.google.com",
										ClientID:        "abcdefg",
										ClientSecret:    "bcdefgh",
										RedirectAddress: "http://auth.flipt.io",
									},
								},
							},
							Enabled: true,
						},
						Kubernetes: AuthenticationMethod[AuthenticationMethodKubernetesConfig]{
							Enabled: true,
							Method: AuthenticationMethodKubernetesConfig{
								DiscoveryURL:            "https://some-other-k8s.namespace.svc",
								CAPath:                  "/path/to/ca/certificate/ca.pem",
								ServiceAccountTokenPath: "/path/to/sa/token",
							},
						},
						Github: AuthenticationMethod[AuthenticationMethodGithubConfig]{
							Method: AuthenticationMethodGithubConfig{
								ClientId:        "abcdefg",
								ClientSecret:    "bcdefgh",
								RedirectAddress: "http://auth.flipt.io",
							},
							Enabled: true,
						},
					},
				}

				cfg.Authorization = AuthorizationConfig{
					Required: true,
					Backend:  AuthorizationBackendLocal,
					Local: &AuthorizationLocalConfig{
						Policy: &AuthorizationSourceLocalConfig{
							Path:         "/path/to/policy.rego",
							PollInterval: time.Minute,
						},
						Data: &AuthorizationSourceLocalConfig{
							Path:         "/path/to/policy/data.json",
							PollInterval: time.Minute,
						},
					},
				}

				return cfg
			},
		},
		{
			name:    "version v1",
			path:    "./testdata/version/v1.yml",
			wantErr: errors.New("invalid version: 1.0"),
		},
		{
			name: "git config provided",
			path: "./testdata/storage/git_provided.yml",
			expected: func() *Config {
				cfg := Default()
				cfg.Storage = StoragesConfig{
					"default": {
						Backend: StorageBackendConfig{
							Type: MemoryStorageBackendType,
						},
						Remote:       "git@github.com:foo/bar.git",
						FetchPolicy:  "strict",
						Branch:       "main",
						PollInterval: 30 * time.Second,
					},
				}
				return cfg
			},
		},
		{
			name: "git config provided with directory",
			path: "./testdata/storage/git_provided_with_directory.yml",
			expected: func() *Config {
				cfg := Default()
				cfg.Storage = StoragesConfig{
					"default": {
						Backend: StorageBackendConfig{
							Type: LocalStorageBackendType,
							Path: "/path/to/gitdir",
						},
						Remote:       "git@github.com:foo/bar.git",
						FetchPolicy:  "lenient",
						Branch:       "main",
						PollInterval: 30 * time.Second,
					},
				}
				return cfg
			},
		},
		{
			name:    "git basic auth partially provided",
			path:    "./testdata/storage/git_basic_auth_invalid.yml",
			wantErr: errors.New("credentials: default credentials: basic credentials: both username and password for basic auth non-empty value is required"),
		},
		{
			name:    "git ssh auth missing password",
			path:    "./testdata/storage/git_ssh_auth_invalid_missing_password.yml",
			wantErr: errors.New("credentials: default credentials: ssh credentials: ssh authentication credentials: password non-empty value is required"),
		},
		{
			name:    "git ssh auth missing private key parts",
			path:    "./testdata/storage/git_ssh_auth_invalid_private_key_missing.yml",
			wantErr: errors.New("credentials: default credentials: ssh credentials: ssh authentication credentials: private_key please provide exclusively one of private_key_bytes or private_key_path"),
		},
		{
			name:    "git ssh auth provided both private key forms",
			path:    "./testdata/storage/git_ssh_auth_invalid_private_key_both.yml",
			wantErr: errors.New("credentials: default credentials: ssh credentials: ssh authentication credentials: private_key please provide exclusively one of private_key_bytes or private_key_path"),
		},
		{
			name: "git valid with ssh auth",
			path: "./testdata/storage/git_ssh_auth_valid_with_path.yml",
			expected: func() *Config {
				cfg := Default()
				cfg.Storage = StoragesConfig{
					"default": {
						Backend: StorageBackendConfig{
							Type: MemoryStorageBackendType,
						},
						Remote:       "git@github.com:foo/bar.git",
						FetchPolicy:  "strict",
						Branch:       "main",
						PollInterval: 30 * time.Second,
						Credentials:  "git",
					},
				}
				cfg.Credentials = CredentialsConfig{
					"git": {
						Type: "ssh",
						SSH: &SSHAuthConfig{
							User:           "git",
							Password:       "bar",
							PrivateKeyPath: "/path/to/pem.key",
						},
					},
				}
				return cfg
			},
		},
		{
			name:    "clickhouse enabled but no URL set",
			path:    "./testdata/analytics/invalid_clickhouse_configuration_empty_url.yml",
			wantErr: errors.New("analytics: clickhouse url non-empty value is required"),
		},
		{
			name:    "analytics flush period too low",
			path:    "./testdata/analytics/invalid_buffer_configuration_flush_period.yml",
			wantErr: errors.New("analytics: buffer flush period below 10 seconds"),
		},
		{
			name: "ui topbar with correct hex color",
			path: "./testdata/ui/topbar_color.yml",
			expected: func() *Config {
				cfg := Default()
				cfg.UI.Topbar.Color = "#42bda0"
				return cfg
			},
		},
		{
			name:    "ui topbar with invalid hex color",
			path:    "./testdata/ui/topbar_invalid_color.yml",
			wantErr: errors.New("ui: topbar expected valid hex color, got invalid"),
		},
		{
			name:    "environments no matching storage",
			path:    "./testdata/environments/no_matching_storage.yml",
			wantErr: errors.New("environments: \"default\" references undefined storage \"non_existent\""),
		},
		{
			name: "environments missing name",
			path: "./testdata/environments/missing_name.yml",
			expected: func() *Config {
				cfg := Default()
				cfg.Environments = EnvironmentsConfig{
					"default": {
						Name:    "default",
						Storage: "default",
						Default: true,
					},
				}
				return cfg
			},
		},
		{
			name: "environments single environment is default",
			path: "./testdata/environments/single_is_default.yml",
			expected: func() *Config {
				cfg := Default()
				cfg.Environments = EnvironmentsConfig{
					"default": {
						Name:    "default",
						Storage: "default",
						Default: true,
					},
				}
				return cfg
			},
		},
		{
			name:    "environments no default",
			path:    "./testdata/environments/no_default.yml",
			wantErr: errors.New("environments: default environment non-empty value is required"),
		},
		{
			name:    "environments sharing storage without distinct directories",
			path:    "./testdata/environments/shared_storage_no_directories.yml",
			wantErr: errors.New("environments: directory environments [prod, staging] share the same storage \"git\" and directory \"\""),
		},
		{
			name:    "environments sharing storage with same directory",
			path:    "./testdata/environments/shared_storage_same_directory.yml",
			wantErr: errors.New("environments: directory environments [prod, staging] share the same storage \"git\" and directory \"repo\""),
		},
		{
			name: "environments sharing storage with different directories",
			path: "./testdata/environments/shared_storage_different_directories.yml",
			expected: func() *Config {
				cfg := Default()
				cfg.Storage = StoragesConfig{
					"git": {
						Backend: StorageBackendConfig{
							Type: MemoryStorageBackendType,
						},
						Branch:       "main",
						PollInterval: 30 * time.Second,
					},
				}
				cfg.Environments = EnvironmentsConfig{
					"prod": {
						Name:      "prod",
						Storage:   "git",
						Directory: "prod",
						Default:   true,
					},
					"staging": {
						Name:      "staging",
						Storage:   "git",
						Directory: "staging",
					},
				}
				return cfg
			},
		},
		{
			name:    "environments invalid scm",
			path:    "./testdata/environments/invalid_scm.yml",
			wantErr: errors.New("environments: default environments: scm unexpected SCM type: \"invalid\""),
		},
		{
			name: "environments valid scm",
			path: "./testdata/environments/valid_scm.yml",
			expected: func() *Config {
				cfg := Default()
				cfg.Environments = EnvironmentsConfig{
					"default": {
						Name:    "default",
						Storage: "default",
						Default: true,
						SCM: &SCMConfig{
							Type:        GitHubSCMType,
							Credentials: ptr("git"),
						},
					},
				}
				cfg.Credentials = CredentialsConfig{
					"git": {
						Type: "basic",
						Basic: &BasicAuthConfig{
							Username: "user",
							Password: "pass",
						},
					},
				}
				return cfg
			},
		},
		{
			name: "secrets config with file provider",
			path: "./testdata/secrets/file_provider.yml",
			expected: func() *Config {
				cfg := Default()
				cfg.Secrets = SecretsConfig{
					Providers: ProvidersConfig{
						File: &FileProviderConfig{
							Enabled:  true,
							BasePath: "/etc/flipt/secrets",
						},
					},
				}
				return cfg
			},
		},
		{
			name: "secrets config with vault provider",
			path: "./testdata/secrets/vault_provider.yml",
			expected: func() *Config {
				cfg := Default()
				cfg.Secrets = SecretsConfig{
					Providers: ProvidersConfig{
						Vault: &VaultProviderConfig{
							Enabled:    true,
							Address:    "https://vault.example.com",
							AuthMethod: "token",
							Token:      "hvs.test_token",
							Mount:      "secret",
						},
					},
				}
				return cfg
			},
		},
		{
			name: "secrets config with multiple providers",
			path: "./testdata/secrets/multiple_providers.yml",
			expected: func() *Config {
				cfg := Default()
				cfg.Secrets = SecretsConfig{
					Providers: ProvidersConfig{
						File: &FileProviderConfig{
							Enabled:  true,
							BasePath: "/etc/flipt/secrets",
						},
						Vault: &VaultProviderConfig{
							Enabled:    true,
							Address:    "https://vault.example.com",
							AuthMethod: "kubernetes",
							Role:       "flipt-role",
							Mount:      "secret",
						},
					},
				}
				return cfg
			},
		},
		{
			name:    "secrets config file provider missing base path",
			path:    "./testdata/secrets/file_provider_missing_base_path.yml",
			wantErr: errors.New("secrets.providers.file: base_path non-empty value is required"),
		},
		{
			name:    "secrets config vault provider missing address",
			path:    "./testdata/secrets/vault_provider_missing_address.yml",
			wantErr: errors.New("secrets.providers.vault: address non-empty value is required"),
		},
		{
			name:    "secrets config vault provider missing token for token auth",
			path:    "./testdata/secrets/vault_provider_missing_token.yml",
			wantErr: errors.New("secrets.providers.vault: token non-empty value is required"),
		},
		{
			name:    "secrets config vault provider missing role for kubernetes auth",
			path:    "./testdata/secrets/vault_provider_missing_role_kubernetes.yml",
			wantErr: errors.New("secrets.providers.vault: role non-empty value is required"),
		},
		{
			name:    "secrets config vault provider missing role for approle auth",
			path:    "./testdata/secrets/vault_provider_missing_role_approle.yml",
			wantErr: errors.New("secrets.providers.vault: role non-empty value is required"),
		},
		{
			name:    "secrets config vault provider unsupported auth method",
			path:    "./testdata/secrets/vault_provider_unsupported_auth.yml",
			wantErr: errors.New("secrets.providers.vault: auth_method unsupported auth method: ldap"),
		},
		{
			name: "git signing config enabled",
			path: "./testdata/storage/git_signing_enabled.yml",
			expected: func() *Config {
				cfg := Default()
				cfg.Storage = StoragesConfig{
					"default": {
						Backend: StorageBackendConfig{
							Type: MemoryStorageBackendType,
						},
						Remote:       "git@github.com:foo/bar.git",
						FetchPolicy:  "strict",
						Branch:       "main",
						PollInterval: 30 * time.Second,
						Signature: SignatureConfig{
							Name:    "Flipt Bot",
							Email:   "bot@flipt.io",
							Enabled: true,
							Type:    "gpg",
							KeyRef: &SecretReference{
								Provider: "vault",
								Path:     "secret/data/gpg",
								Key:      "private_key",
							},
							KeyID: "1234567890ABCDEF",
						},
					},
				}
				return cfg
			},
		},
		{
			name:    "git signing missing signing type",
			path:    "./testdata/storage/git_signing_missing_type.yml",
			wantErr: errors.New("storage: default storage: signature type non-empty value is required"),
		},
		{
			name:    "git signing unsupported signing type",
			path:    "./testdata/storage/git_signing_unsupported_type.yml",
			wantErr: errors.New("storage: default storage: signature type: unsupported signing type: x509 (only 'gpg' is supported)"),
		},
		{
			name:    "git signing missing key reference",
			path:    "./testdata/storage/git_signing_missing_key_ref.yml",
			wantErr: errors.New("storage: default storage: signature key_ref non-empty value is required"),
		},
		{
			name:    "git signing invalid key reference missing provider",
			path:    "./testdata/storage/git_signing_invalid_key_ref_provider.yml",
			wantErr: errors.New("storage: default storage: signature key_ref: secret_reference: provider non-empty value is required"),
		},
		{
			name:    "git signing invalid key reference missing path",
			path:    "./testdata/storage/git_signing_invalid_key_ref_path.yml",
			wantErr: errors.New("storage: default storage: signature key_ref: secret_reference: path non-empty value is required"),
		},
		{
			name:    "git signing invalid key reference missing key",
			path:    "./testdata/storage/git_signing_invalid_key_ref_key.yml",
			wantErr: errors.New("storage: default storage: signature key_ref: secret_reference: key non-empty value is required"),
		},
		{
			name: "license valid with key only",
			path: "./testdata/license/valid_with_key_only.yml",
			expected: func() *Config {
				cfg := Default()
				cfg.License = LicenseConfig{
					Key:  "test-license-key-12345",
					File: "",
				}
				return cfg
			},
		},
		{
			name: "license valid with key and file",
			path: "./testdata/license/valid_with_key_and_file.yml",
			expected: func() *Config {
				cfg := Default()
				cfg.License = LicenseConfig{
					Key:  "test-license-key-12345",
					File: "./testdata/ssl_cert.pem",
				}
				return cfg
			},
		},
		{
			name:    "license invalid - file specified but no key",
			path:    "./testdata/license/invalid_file_no_key.yml",
			wantErr: errors.New("license: key non-empty value is required"),
		},
		{
			name:    "license invalid - file not found",
			path:    "./testdata/license/invalid_file_not_found.yml",
			wantErr: errors.New("license: file stat ./testdata/nonexistent_license.cert: no such file or directory"),
		},
	}

	for _, tt := range tests {
		var (
			path     = tt.path
			wantErr  = tt.wantErr
			expected *Config
			warnings = tt.warnings
		)

		if tt.expected != nil {
			expected = tt.expected()
		}

		t.Run(tt.name+" (YAML)", func(t *testing.T) {
			// backup and restore environment
			backup := os.Environ()
			defer func() {
				os.Clearenv()
				for _, env := range backup {
					key, value, _ := strings.Cut(env, "=")
					os.Setenv(key, value)
				}
			}()

			for key, value := range tt.envOverrides {
				t.Logf("Setting env '%s=%s'\n", key, value)
				os.Setenv(key, value)
			}

			res, err := Load(context.Background(), path)

			if wantErr != nil {
				t.Log(err)
				if err == nil {
					require.Failf(t, "expected error", "expected %q, found <nil>", wantErr)
				}
				if errors.Is(err, wantErr) {
					return
				} else if err.Error() == wantErr.Error() {
					return
				}
				require.Fail(t, "expected error", "expected %q, found %q", wantErr, err)
			}

			require.NoError(t, err)

			assert.NotNil(t, res)
			assert.Equal(t, expected, res.Config)
			assert.Equal(t, warnings, res.Warnings)
		})

		t.Run(tt.name+" (ENV)", func(t *testing.T) {
			// backup and restore environment
			backup := os.Environ()
			defer func() {
				os.Clearenv()
				for _, env := range backup {
					key, value, _ := strings.Cut(env, "=")
					os.Setenv(key, value)
				}
			}()

			if path != "" {
				// read the input config file into equivalent envs
				envs := readYAMLIntoEnv(t, path)
				for _, env := range envs {
					t.Logf("Setting env '%s=%s'\n", env[0], env[1])
					os.Setenv(env[0], env[1])
				}
			}

			for key, value := range tt.envOverrides {
				t.Logf("Setting env '%s=%s'\n", key, value)
				os.Setenv(key, value)
			}

			// load default (empty) config
			res, err := Load(context.Background(), "./testdata/default.yml")

			if wantErr != nil {
				t.Log(err)
				if err == nil {
					require.Failf(t, "expected error", "expected %q, found <nil>", wantErr)
				}
				if errors.Is(err, wantErr) {
					return
				} else if err.Error() == wantErr.Error() {
					return
				}
				require.Fail(t, "expected error", "expected %q, found %q", wantErr, err)
			}

			require.NoError(t, err)

			assert.NotNil(t, res)
			assert.Equal(t, expected, res.Config)
		})
	}
}

func TestServeHTTP(t *testing.T) {
	var (
		cfg = Default()
		req = httptest.NewRequest("GET", "http://example.com/foo", nil)
		w   = httptest.NewRecorder()
	)

	cfg.ServeHTTP(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.NotEmpty(t, body)
}

// readyYAMLIntoEnv parses the file provided at path as YAML.
// It walks the keys and values and builds up a set of environment variables
// compatible with viper's expectations for automatic env capability.
func readYAMLIntoEnv(t *testing.T, path string) [][2]string {
	t.Helper()

	configFile, err := os.ReadFile(path)
	require.NoError(t, err)

	var config map[any]any
	err = yaml.Unmarshal(configFile, &config)
	require.NoError(t, err)

	return getEnvVars("flipt", config)
}

func getEnvVars(prefix string, v map[any]any) (vals [][2]string) {
OUTER:
	for key, value := range v {
		switch v := value.(type) {
		case map[any]any:
			vals = append(vals, getEnvVars(fmt.Sprintf("%s_%v", prefix, key), v)...)
		case []any:
			builder := strings.Builder{}
			for i, s := range v {
				switch s := s.(type) {
				case map[any]any:
					// descend into slices of objects
					vals = append(vals, getEnvVars(fmt.Sprintf("%s_%v", prefix, key), s)...)
					continue OUTER
				default:
					builder.WriteString(fmt.Sprintf("%v", s))
					if i < len(v)-1 {
						builder.WriteByte(' ')
					}
				}
			}

			vals = append(vals, [2]string{
				fmt.Sprintf("%s_%s", strings.ToUpper(prefix), strings.ToUpper(fmt.Sprintf("%v", key))),
				builder.String(),
			})
		default:
			vals = append(vals, [2]string{
				fmt.Sprintf("%s_%s", strings.ToUpper(prefix), strings.ToUpper(fmt.Sprintf("%v", key))),
				fmt.Sprintf("%v", value),
			})
		}
	}

	return vals
}

func TestMarshalYAML(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr error
		cfg     func() *Config
	}{
		{
			name: "defaults",
			path: "./testdata/marshal/yaml/default.yml",
			cfg:  Default,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wantErr := tt.wantErr

			expected, err := os.ReadFile(tt.path)
			require.NoError(t, err)

			out, err := yaml.Marshal(tt.cfg())
			if wantErr != nil {
				t.Log(err)
				match := false
				if errors.Is(err, wantErr) {
					match = true
				} else if err.Error() == wantErr.Error() {
					match = true
				}
				require.True(t, match, "expected error %v to match: %v", err, wantErr)
				return
			}

			require.NoError(t, err)

			assert.YAMLEq(t, string(expected), string(out))
		})
	}
}

type sliceEnvBinder []string

func (s *sliceEnvBinder) MustBindEnv(v ...string) {
	*s = append(*s, v...)
}

func Test_mustBindEnv(t *testing.T) {
	for _, test := range []struct {
		name string
		// inputs
		env []string
		typ any
		// expected outputs
		bound []string
	}{
		{
			name: "simple struct",
			env:  []string{},
			typ: struct {
				A string `mapstructure:"a"`
				B string `mapstructure:"b"`
				C string `mapstructure:"c"`
			}{},
			bound: []string{"a", "b", "c"},
		},
		{
			name: "nested structs with pointers",
			env:  []string{},
			typ: struct {
				A string `mapstructure:"a"`
				B struct {
					C *struct {
						D int
					} `mapstructure:"c"`
					E []string `mapstructure:"e"`
				} `mapstructure:"b"`
			}{},
			bound: []string{"a", "b.c.d", "b.e"},
		},
		{
			name: "structs with maps and no environment variables",
			env:  []string{},
			typ: struct {
				A struct {
					B map[string]string `mapstructure:"b"`
				} `mapstructure:"a"`
			}{},
			// no environment variable to direct mappings
			bound: []string{},
		},
		{
			name: "structs with maps with env",
			env:  []string{"A_B_FOO", "A_B_BAR", "A_B_BAZ"},
			typ: struct {
				A struct {
					B map[string]string `mapstructure:"b"`
				} `mapstructure:"a"`
			}{},
			// no environment variable to direct mappings
			bound: []string{"a.b.foo", "a.b.bar", "a.b.baz"},
		},
		{
			name: "structs with maps of structs (env not specific enough)",
			env:  []string{"A_B_FOO", "A_B_BAR"},
			typ: struct {
				A struct {
					B map[string]struct {
						C string `mapstructure:"c"`
						D struct {
							E int `mapstructure:"e"`
						} `mapstructure:"d"`
					} `mapstructure:"b"`
				} `mapstructure:"a"`
			}{},
			// no environment variable to direct mappings
			bound: []string{},
		},
		{
			name: "structs with maps of structs",
			env: []string{
				"A_B_FOO_C",
				"A_B_FOO_D",
				"A_B_BAR_BAZ_C",
				"A_B_BAR_BAZ_D",
			},
			typ: struct {
				A struct {
					B map[string]struct {
						C string `mapstructure:"c"`
						D struct {
							E int `mapstructure:"e"`
						} `mapstructure:"d"`
					} `mapstructure:"b"`
				} `mapstructure:"a"`
			}{},
			bound: []string{
				"a.b.foo.c",
				"a.b.bar_baz.c",
				"a.b.foo.d.e",
				"a.b.bar_baz.d.e",
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			binder := sliceEnvBinder{}

			typ := reflect.TypeOf(test.typ)
			bindEnvVars(&binder, test.env, []string{}, typ)

			assert.Equal(t, test.bound, []string(binder))
		})
	}
}

type mockURLOpener struct {
	bucket *blob.Bucket
}

// OpenBucketURL opens a blob.Bucket based on u.
func (c *mockURLOpener) OpenBucketURL(ctx context.Context, u *url.URL) (*blob.Bucket, error) {
	for param := range u.Query() {
		return nil, fmt.Errorf("open bucket %v: invalid query parameter %q", u, param)
	}
	return c.bucket, nil
}

func TestGetConfigFile(t *testing.T) {
	blob.DefaultURLMux().RegisterBucket("mock", &mockURLOpener{
		bucket: memblob.OpenBucket(nil),
	})

	var (
		ctx        = context.Background()
		configData = []byte("some config data")
		b, err     = blob.OpenBucket(ctx, "mock://mybucket")
	)

	require.NoError(t, err)
	t.Cleanup(func() { b.Close() })

	w, err := b.NewWriter(ctx, "config/local.yml", nil)
	require.NoError(t, err)

	_, err = w.Write(configData)
	require.NoError(t, err)

	err = w.Close()
	require.NoError(t, err)

	t.Run("successful", func(t *testing.T) {
		f, err := getConfigFile(ctx, "mock://mybucket/config/local.yml")
		require.NoError(t, err)
		t.Cleanup(func() { f.Close() })

		s, err := f.Stat()
		require.NoError(t, err)
		require.Equal(t, "local.yml", s.Name())

		data, err := io.ReadAll(f)
		require.NoError(t, err)
		require.Equal(t, configData, data)
	})

	for _, tt := range []struct {
		name string
		path string
	}{
		{"unknown bucket", "mock://otherbucket/config.yml"},
		{"unknown scheme", "unknown://otherbucket/config.yml"},
		{"no bucket", "mock://"},
		{"no key", "mock://mybucket"},
		{"no data", ""},
	} {
		t.Run(tt.name, func(t *testing.T) {
			_, err = getConfigFile(ctx, tt.path)
			require.Error(t, err)
		})
	}
}

func TestAuthenticationGithubCaseSensitivity(t *testing.T) {
	expected := Default()
	expected.Authentication.Required = true
	expected.Authentication.Session.Domain = "localhost"
	expected.Authentication.Methods = AuthenticationMethodsConfig{
		Github: AuthenticationMethod[AuthenticationMethodGithubConfig]{
			Enabled: true,
			Method: AuthenticationMethodGithubConfig{
				ClientId:             "client_id",
				ClientSecret:         "client_secret",
				RedirectAddress:      "http://localhost:8080",
				Scopes:               []string{"read:org"},
				AllowedOrganizations: []string{"MyOrg"},
				AllowedTeams:         map[string][]string{"MyOrg": {"team"}},
			},
		},
	}
	res, err := Load(t.Context(), "./testdata/authentication/github_case_sensitive.yml")
	require.NoError(t, err)

	assert.NotNil(t, res)
	assert.Equal(t, expected, res.Config)
}

// add any struct tags to match their camelCase equivalents here.
var camelCaseMatchers = map[string]string{
	"requireTLS":   "requireTLS",
	"discoveryURL": "discoveryURL",
}

func TestStructTags(t *testing.T) {
	configType := reflect.TypeOf(Config{})
	configTags := getStructTags(configType)

	for k, v := range camelCaseMatchers {
		strcase.ConfigureAcronym(k, v)
	}

	// Validate the struct tags for the Config struct.
	// recursively validate the struct tags for all sub-structs.
	validateStructTags(t, configTags, configType)
}

func validateStructTags(t *testing.T, tags map[string]map[string]string, tType reflect.Type) {
	tName := tType.Name()
	for fieldName, fieldTags := range tags {
		fieldType, ok := tType.FieldByName(fieldName)
		require.True(t, ok, "field %s not found in type %s", fieldName, tName)

		// Validate the `json` struct tag.
		jsonTag, ok := fieldTags["json"]
		if ok {
			require.True(t, isCamelCase(jsonTag), "json tag for field '%s.%s' should be camelCase but is '%s'", tName, fieldName, jsonTag)
		}

		// Validate the `mapstructure` struct tag.
		mapstructureTag, ok := fieldTags["mapstructure"]
		if ok {
			require.True(t, isSnakeCase(mapstructureTag), "mapstructure tag for field '%s.%s' should be snake_case but is '%s'", tName, fieldName, mapstructureTag)
		}

		// Validate the `yaml` struct tag.
		yamlTag, ok := fieldTags["yaml"]
		if ok {
			require.True(t, isSnakeCase(yamlTag), "yaml tag for field '%s.%s' should be snake_case but is '%s'", tName, fieldName, yamlTag)
		}

		// recursively validate the struct tags for all sub-structs.
		if fieldType.Type.Kind() == reflect.Struct {
			validateStructTags(t, getStructTags(fieldType.Type), fieldType.Type)
		}
	}
}

func isCamelCase(s string) bool {
	return s == strcase.ToLowerCamel(s)
}

func isSnakeCase(s string) bool {
	return s == strcase.ToSnake(s)
}

func getStructTags(t reflect.Type) map[string]map[string]string {
	tags := make(map[string]map[string]string)

	for i := range t.NumField() {
		field := t.Field(i)

		// Get the field name.
		fieldName := field.Name

		// Get the field tags.
		fieldTags := make(map[string]string)
		for _, tag := range []string{"json", "mapstructure", "yaml"} {
			tagValue := field.Tag.Get(tag)
			if tagValue == "-" {
				fieldTags[tag] = "skip"
				continue
			}
			values := strings.Split(tagValue, ",")
			if len(values) > 1 {
				tagValue = values[0]
			}
			if tagValue != "" {
				fieldTags[tag] = tagValue
			}
		}

		tags[fieldName] = fieldTags
	}

	return tags
}

func ptr[T any](v T) *T {
	return &v
}

func TestLicenseConfig_validate(t *testing.T) {
	tests := []struct {
		name      string
		config    *LicenseConfig
		setupFile func(t *testing.T) string // Returns file path if file setup is needed
		wantErr   bool
		errMsg    string
	}{
		{
			name: "valid config with empty values",
			config: &LicenseConfig{
				Key:  "",
				File: "",
			},
			wantErr: false,
		},
		{
			name: "valid config with key only",
			config: &LicenseConfig{
				Key:  "test-license-key",
				File: "",
			},
			wantErr: false,
		},
		{
			name: "valid config with both key and file",
			config: &LicenseConfig{
				Key: "test-license-key",
				// File will be set by setupFile
			},
			setupFile: func(t *testing.T) string {
				tmpDir := t.TempDir()
				licenseFile := filepath.Join(tmpDir, "license.cert")
				err := os.WriteFile(licenseFile, []byte("test license content"), 0o600)
				require.NoError(t, err)
				return licenseFile
			},
			wantErr: false,
		},
		{
			name: "error when file specified but key is missing",
			config: &LicenseConfig{
				Key: "", // Missing key
				// File will be set by setupFile
			},
			setupFile: func(t *testing.T) string {
				tmpDir := t.TempDir()
				licenseFile := filepath.Join(tmpDir, "license.cert")
				err := os.WriteFile(licenseFile, []byte("test license content"), 0o600)
				require.NoError(t, err)
				return licenseFile
			},
			wantErr: true,
			errMsg:  "license: key non-empty value is required",
		},
		{
			name: "error when file doesn't exist",
			config: &LicenseConfig{
				Key:  "test-license-key",
				File: "/nonexistent/path/license.cert",
			},
			wantErr: true,
			errMsg:  "license: file stat /nonexistent/path/license.cert: no such file or directory",
		},
		{
			name: "error when file path is a directory",
			config: &LicenseConfig{
				Key: "test-license-key",
				// File will be set to a directory by setupFile
			},
			setupFile: func(t *testing.T) string {
				tmpDir := t.TempDir()
				// Return the directory path itself instead of a file
				return tmpDir
			},
			wantErr: false, // os.Stat on directory succeeds, so this is actually valid
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := tt.config

			// Setup file if needed
			if tt.setupFile != nil {
				config.File = tt.setupFile(t)
			}

			err := config.validate()

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestStringToReferenceHookFunc(t *testing.T) {
	t.Run("with environment variables only", func(t *testing.T) {
		hook := stringToReferenceHookFunc()

		tests := []struct {
			name     string
			input    string
			envVars  map[string]string
			expected any
		}{
			{
				name:     "explicit environment reference",
				input:    "${env:TEST_VAR}",
				envVars:  map[string]string{"TEST_VAR": "test_value"},
				expected: "test_value",
			},
			{
				name:     "explicit environment reference - missing var",
				input:    "${env:MISSING_VAR}",
				envVars:  nil,
				expected: "",
			},
			{
				name:     "secret reference left as-is",
				input:    "${secret:github-token}",
				expected: "${secret:github-token}",
			},
			{
				name:     "non-reference string",
				input:    "normal_string",
				expected: "normal_string",
			},
			{
				name:     "invalid reference format",
				input:    "${invalid}",
				expected: "${invalid}", // Not a valid reference format, so returned as-is
			},
			{
				name:     "empty string",
				input:    "",
				expected: "",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				// Set environment variables
				for key, value := range tt.envVars {
					t.Setenv(key, value)
				}

				hookFunc := hook.(func(reflect.Type, reflect.Type, any) (any, error))
				result, err := hookFunc(
					reflect.TypeOf(""),
					reflect.TypeOf(""),
					tt.input,
				)

				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			})
		}
	})

	t.Run("with new approach - secrets left as-is", func(t *testing.T) {
		hook := stringToReferenceHookFunc()

		tests := []struct {
			name     string
			input    string
			envVars  map[string]string
			expected any
		}{
			{
				name:     "explicit environment reference",
				input:    "${env:TEST_VAR}",
				envVars:  map[string]string{"TEST_VAR": "test_value"},
				expected: "test_value",
			},
			{
				name:     "secret reference left as-is",
				input:    "${secret:github-token}",
				expected: "${secret:github-token}",
			},
			{
				name:     "complex secret reference left as-is",
				input:    "${secret:vault:auth/oauth:client_secret}",
				expected: "${secret:vault:auth/oauth:client_secret}",
			},
			{
				name:     "non-reference string",
				input:    "normal_string",
				expected: "normal_string",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				// Set environment variables
				for key, value := range tt.envVars {
					t.Setenv(key, value)
				}

				hookFunc := hook.(func(reflect.Type, reflect.Type, any) (any, error))
				result, err := hookFunc(
					reflect.TypeOf(""),
					reflect.TypeOf(""),
					tt.input,
				)

				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			})
		}
	})
}

func TestConfigSecretReferences(t *testing.T) {
	t.Run("load config with secret references preserved", func(t *testing.T) {
		// Set up environment variables
		t.Setenv("HTTP_PORT", "8080")
		t.Setenv("GITHUB_CLIENT_ID", "test_client_id")
		t.Setenv("GITHUB_REDIRECT_URL", "http://localhost:8080/auth/callback")

		// Load config normally - secret references should remain as-is
		result, err := Load(context.Background(), "./testdata/secret_references.yml")
		require.NoError(t, err)
		require.NotNil(t, result)
		require.NotNil(t, result.Config)

		// Verify environment variables were resolved during load
		assert.Equal(t, 8080, result.Config.Server.HTTPPort)
		assert.Equal(t, "test_client_id", result.Config.Authentication.Methods.Github.Method.ClientId)
		assert.Equal(t, "http://localhost:8080/auth/callback", result.Config.Authentication.Methods.Github.Method.RedirectAddress)

		// Verify secret references remain unresolved (as raw strings)
		assert.Equal(t, "${secret:github-client-secret}", result.Config.Authentication.Methods.Github.Method.ClientSecret)
	})

	t.Run("load config with mixed reference types", func(t *testing.T) {
		// Set up environment variables
		t.Setenv("HTTP_PORT", "8080")
		t.Setenv("GRPC_PORT", "9000")
		t.Setenv("GITHUB_CLIENT_ID", "test_client_id")
		t.Setenv("GITHUB_REDIRECT_URL", "http://localhost:8080/auth/callback")

		// Load config normally
		result, err := Load(context.Background(), "./testdata/mixed_references.yml")
		require.NoError(t, err)
		require.NotNil(t, result)

		// Verify environment references were resolved during load
		assert.Equal(t, 8080, result.Config.Server.HTTPPort)                                                                       // env reference
		assert.Equal(t, 9000, result.Config.Server.GRPCPort)                                                                       // env reference
		assert.Equal(t, "test_client_id", result.Config.Authentication.Methods.Github.Method.ClientId)                             // env reference
		assert.Equal(t, "http://localhost:8080/auth/callback", result.Config.Authentication.Methods.Github.Method.RedirectAddress) // env reference

		// Verify secret references remain unresolved
		assert.Equal(t, "${secret:github-token}", result.Config.Authentication.Methods.Github.Method.ClientSecret) // secret reference
	})
}
