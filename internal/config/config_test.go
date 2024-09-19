package config

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/iancoleman/strcase"
	"github.com/santhosh-tekuri/jsonschema/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/oci"
	"gocloud.dev/blob"
	"gocloud.dev/blob/memblob"
	"gopkg.in/yaml.v2"
)

func TestJSONSchema(t *testing.T) {
	_, err := jsonschema.Compile("../../config/flipt.schema.json")
	require.NoError(t, err)
}

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
			json, err := scheme.MarshalJSON()
			require.NoError(t, err)
			assert.JSONEq(t, fmt.Sprintf("%q", want), string(json))
		})
	}
}

func TestCacheBackend(t *testing.T) {
	tests := []struct {
		name    string
		backend CacheBackend
		want    string
	}{
		{
			name:    "memory",
			backend: CacheMemory,
			want:    "memory",
		},
		{
			name:    "redis",
			backend: CacheRedis,
			want:    "redis",
		},
	}

	for _, tt := range tests {
		var (
			backend = tt.backend
			want    = tt.want
		)

		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, want, backend.String())
			json, err := backend.MarshalJSON()
			require.NoError(t, err)
			assert.JSONEq(t, fmt.Sprintf("%q", want), string(json))
		})
	}
}

func TestTracingExporter(t *testing.T) {
	tests := []struct {
		name     string
		exporter TracingExporter
		want     string
	}{
		{
			name:     "jaeger",
			exporter: TracingJaeger,
			want:     "jaeger",
		},
		{
			name:     "zipkin",
			exporter: TracingZipkin,
			want:     "zipkin",
		},
		{
			name:     "otlp",
			exporter: TracingOTLP,
			want:     "otlp",
		},
	}

	for _, tt := range tests {
		var (
			exporter = tt.exporter
			want     = tt.want
		)

		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, want, exporter.String())
			json, err := exporter.MarshalJSON()
			require.NoError(t, err)
			assert.JSONEq(t, fmt.Sprintf("%q", want), string(json))
		})
	}
}

func TestDatabaseProtocol(t *testing.T) {
	tests := []struct {
		name     string
		protocol DatabaseProtocol
		want     string
	}{
		{
			name:     "postgres",
			protocol: DatabasePostgres,
			want:     "postgres",
		},
		{
			name:     "mysql",
			protocol: DatabaseMySQL,
			want:     "mysql",
		},
		{
			name:     "sqlite",
			protocol: DatabaseSQLite,
			want:     "sqlite",
		},
		{
			name:     "cockroachdb",
			protocol: DatabaseCockroachDB,
			want:     "cockroachdb",
		},
		{
			name:     "libsql",
			protocol: DatabaseLibSQL,
			want:     "libsql",
		},
	}

	for _, tt := range tests {
		var (
			protocol = tt.protocol
			want     = tt.want
		)

		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, want, protocol.String())
			json, err := protocol.MarshalJSON()
			require.NoError(t, err)
			assert.JSONEq(t, fmt.Sprintf("%q", want), string(json))
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
			json, err := encoding.MarshalJSON()
			require.NoError(t, err)
			assert.JSONEq(t, fmt.Sprintf("%q", want), string(json))
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
			name: "environment variable substitution",
			path: "./testdata/envsubst.yml",
			envOverrides: map[string]string{
				"HTTP_PORT":  "18080",
				"LOG_FORMAT": "json",
			},
			expected: func() *Config {
				cfg := Default()
				cfg.Log.Encoding = "json"
				cfg.Server.HTTPPort = 18080
				return cfg
			},
		},
		{
			name: "deprecated tracing jaeger",
			path: "./testdata/deprecated/tracing_jaeger.yml",
			expected: func() *Config {
				cfg := Default()
				cfg.Tracing.Enabled = true
				cfg.Tracing.Exporter = TracingJaeger
				return cfg
			},
			warnings: []string{
				"\"tracing.exporter.jaeger\" is deprecated and will be removed in a future release.",
			},
		},
		{
			name: "deprecated authentication excluding metadata",
			path: "./testdata/deprecated/authentication_excluding_metadata.yml",
			expected: func() *Config {
				cfg := Default()
				cfg.Authentication.Required = true
				cfg.Authentication.Exclude.Metadata = true
				return cfg
			},
			warnings: []string{
				"\"authentication.exclude.metadata\" is deprecated and will be removed in a future release. This feature never worked as intended. Metadata can no longer be excluded from authentication (when required).",
			},
		},
		{
			name: "cache no backend set",
			path: "./testdata/cache/default.yml",
			expected: func() *Config {
				cfg := Default()
				cfg.Cache.Enabled = true
				cfg.Cache.Backend = CacheMemory
				cfg.Cache.TTL = 30 * time.Minute
				return cfg
			},
		},
		{
			name: "cache memory",
			path: "./testdata/cache/memory.yml",
			expected: func() *Config {
				cfg := Default()
				cfg.Cache.Enabled = true
				cfg.Cache.Backend = CacheMemory
				cfg.Cache.TTL = 5 * time.Minute
				cfg.Cache.Memory.EvictionInterval = 10 * time.Minute
				return cfg
			},
		},
		{
			name: "cache redis",
			path: "./testdata/cache/redis.yml",
			expected: func() *Config {
				cfg := Default()
				cfg.Cache.Enabled = true
				cfg.Cache.Backend = CacheRedis
				cfg.Cache.TTL = time.Minute
				cfg.Cache.Redis.Host = "localhost"
				cfg.Cache.Redis.Port = 6378
				cfg.Cache.Redis.RequireTLS = true
				cfg.Cache.Redis.DB = 1
				cfg.Cache.Redis.Password = "s3cr3t!"
				cfg.Cache.Redis.PoolSize = 50
				cfg.Cache.Redis.MinIdleConn = 2
				cfg.Cache.Redis.ConnMaxIdleTime = 10 * time.Minute
				cfg.Cache.Redis.NetTimeout = 500 * time.Millisecond
				return cfg
			},
		},
		{
			name: "cache redis with username",
			path: "./testdata/cache/redis-username.yml",
			expected: func() *Config {
				cfg := Default()
				cfg.Cache.Enabled = true
				cfg.Cache.Backend = CacheRedis
				cfg.Cache.Redis.Username = "app"
				cfg.Cache.Redis.Password = "s3cr3t!"
				return cfg
			},
		},
		{
			name: "cache redis with insecure skip verify tls",
			path: "./testdata/cache/redis-tls-insecure.yml",
			expected: func() *Config {
				cfg := Default()
				cfg.Cache.Enabled = true
				cfg.Cache.Backend = CacheRedis
				cfg.Cache.Redis.InsecureSkipTLS = true
				return cfg
			},
		},
		{
			name: "cache redis with ca path bundle",
			path: "./testdata/cache/redis-ca-path.yml",
			expected: func() *Config {
				cfg := Default()
				cfg.Cache.Enabled = true
				cfg.Cache.Backend = CacheRedis
				cfg.Cache.Redis.CaCertPath = "internal/config/testdata/ca.pem"
				return cfg
			},
		},
		{
			name: "cache redis with ca bytes bundle",
			path: "./testdata/cache/redis-ca-bytes.yml",
			expected: func() *Config {
				cfg := Default()
				cfg.Cache.Enabled = true
				cfg.Cache.Backend = CacheRedis
				cfg.Cache.Redis.CaCertBytes = "pemblock\n"
				return cfg
			},
		},
		{
			name:    "cache redis with ca path and bytes bundle",
			path:    "./testdata/cache/redis-ca-invalid.yml",
			wantErr: errors.New("please provide exclusively one of ca_cert_bytes or ca_cert_path"),
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
				cfg.Metrics.OTLP.Endpoint = "http://localhost:9999"
				cfg.Metrics.OTLP.Headers = map[string]string{"api-key": "test-key"}
				return cfg
			},
		},
		{
			name: "tracing zipkin",
			path: "./testdata/tracing/zipkin.yml",
			expected: func() *Config {
				cfg := Default()
				cfg.Tracing.Enabled = true
				cfg.Tracing.Exporter = TracingZipkin
				cfg.Tracing.Zipkin.Endpoint = "http://localhost:9999/api/v2/spans"
				return cfg
			},
		},
		{
			name:    "tracing with wrong sampling ration",
			path:    "./testdata/tracing/wrong_sampling_ratio.yml",
			wantErr: errors.New("sampling ratio should be a number between 0 and 1"),
		},
		{
			name:    "tracing with wrong propagator",
			path:    "./testdata/tracing/wrong_propagator.yml",
			wantErr: errors.New("invalid propagator option: wrong_propagator"),
		},
		{
			name: "tracing OTLP",
			path: "./testdata/tracing/otlp.yml",
			expected: func() *Config {
				cfg := Default()
				cfg.Tracing.SamplingRatio = 0.5
				cfg.Tracing.Enabled = true
				cfg.Tracing.Exporter = TracingOTLP
				cfg.Tracing.OTLP.Endpoint = "http://localhost:9999"
				cfg.Tracing.OTLP.Headers = map[string]string{"api-key": "test-key"}
				return cfg
			},
		},
		{
			name: "database key/value",
			path: "./testdata/database.yml",
			expected: func() *Config {
				cfg := Default()
				cfg.Database = DatabaseConfig{
					Protocol:                  DatabaseMySQL,
					Host:                      "localhost",
					Port:                      3306,
					User:                      "flipt",
					Password:                  "s3cr3t!",
					Name:                      "flipt",
					MaxIdleConn:               2,
					PreparedStatementsEnabled: true,
				}
				return cfg
			},
		},
		{
			name:    "server https missing cert file",
			path:    "./testdata/server/https_missing_cert_file.yml",
			wantErr: errValidationRequired,
		},
		{
			name:    "server https missing cert key",
			path:    "./testdata/server/https_missing_cert_key.yml",
			wantErr: errValidationRequired,
		},
		{
			name:    "server https defined but not found cert file",
			path:    "./testdata/server/https_not_found_cert_file.yml",
			wantErr: fs.ErrNotExist,
		},
		{
			name:    "server https defined but not found cert key",
			path:    "./testdata/server/https_not_found_cert_key.yml",
			wantErr: fs.ErrNotExist,
		},
		{
			name:    "database protocol required",
			path:    "./testdata/database/missing_protocol.yml",
			wantErr: errValidationRequired,
		},
		{
			name:    "database host required",
			path:    "./testdata/database/missing_host.yml",
			wantErr: errValidationRequired,
		},
		{
			name:    "database name required",
			path:    "./testdata/database/missing_name.yml",
			wantErr: errValidationRequired,
		},
		{
			name:    "authentication token negative interval",
			path:    "./testdata/authentication/token_negative_interval.yml",
			wantErr: errPositiveNonZeroDuration,
		},
		{
			name:    "authentication token zero grace_period",
			path:    "./testdata/authentication/token_zero_grace_period.yml",
			wantErr: errPositiveNonZeroDuration,
		},
		{
			name: "authentication token with provided bootstrap token",
			path: "./testdata/authentication/token_bootstrap_token.yml",
			expected: func() *Config {
				cfg := Default()
				cfg.Authentication.Required = true
				cfg.Authentication.Methods = AuthenticationMethods{
					Token: AuthenticationMethod[AuthenticationMethodTokenConfig]{
						Enabled: true,
						Method: AuthenticationMethodTokenConfig{
							Bootstrap: AuthenticationMethodTokenBootstrapConfig{
								Token:      "s3cr3t!",
								Expiration: 24 * time.Hour,
							},
						},
						Cleanup: &AuthenticationCleanupSchedule{
							Interval:    time.Hour,
							GracePeriod: 30 * time.Minute,
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
				cfg.Authentication.Session.Domain = "localhost"
				cfg.Authentication.Methods = AuthenticationMethods{
					Token: AuthenticationMethod[AuthenticationMethodTokenConfig]{
						Enabled: true,
						Cleanup: &AuthenticationCleanupSchedule{
							Interval:    time.Hour,
							GracePeriod: 30 * time.Minute,
						},
					},
					OIDC: AuthenticationMethod[AuthenticationMethodOIDCConfig]{
						Enabled: true,
						Cleanup: &AuthenticationCleanupSchedule{
							Interval:    time.Hour,
							GracePeriod: 30 * time.Minute,
						},
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
				cfg.Authentication.Session.Domain = "localhost"
				cfg.Authentication.Methods = AuthenticationMethods{
					OIDC: AuthenticationMethod[AuthenticationMethodOIDCConfig]{
						Enabled: true,
						Cleanup: &AuthenticationCleanupSchedule{
							Interval:    time.Hour,
							GracePeriod: 30 * time.Minute,
						},
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
					}}
				return cfg
			},
		},
		{
			name: "authentication kubernetes defaults when enabled",
			path: "./testdata/authentication/kubernetes.yml",
			expected: func() *Config {
				cfg := Default()
				cfg.Authentication.Required = true
				cfg.Authentication.Methods = AuthenticationMethods{
					Kubernetes: AuthenticationMethod[AuthenticationMethodKubernetesConfig]{
						Enabled: true,
						Method: AuthenticationMethodKubernetesConfig{
							DiscoveryURL:            "https://kubernetes.default.svc.cluster.local",
							CAPath:                  "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt",
							ServiceAccountTokenPath: "/var/run/secrets/kubernetes.io/serviceaccount/token",
						},
						Cleanup: &AuthenticationCleanupSchedule{
							Interval:    time.Hour,
							GracePeriod: 30 * time.Minute,
						},
					},
				}
				return cfg
			},
		},
		{
			name:    "authentication github requires read:org scope when allowing orgs",
			path:    "./testdata/authentication/github_missing_org_scope.yml",
			wantErr: errors.New("provider \"github\": field \"scopes\": must contain read:org when allowed_organizations is not empty"),
		},
		{
			name:    "authentication github missing client id",
			path:    "./testdata/authentication/github_missing_client_id.yml",
			wantErr: errors.New("provider \"github\": field \"client_id\": non-empty value is required"),
		},
		{
			name:    "authentication github missing client secret",
			path:    "./testdata/authentication/github_missing_client_secret.yml",
			wantErr: errors.New("provider \"github\": field \"client_secret\": non-empty value is required"),
		},
		{
			name:    "authentication github missing client id",
			path:    "./testdata/authentication/github_missing_redirect_address.yml",
			wantErr: errors.New("provider \"github\": field \"redirect_address\": non-empty value is required"),
		},
		{
			name:    "authentication github has non declared org in allowed_teams",
			path:    "./testdata/authentication/github_missing_org_when_declaring_allowed_teams.yml",
			wantErr: errors.New("provider \"github\": field \"allowed_teams\": the organization 'my-other-org' was not declared in 'allowed_organizations' field"),
		},
		{
			name:    "authentication oidc missing client id",
			path:    "./testdata/authentication/oidc_missing_client_id.yml",
			wantErr: errors.New("provider \"foo\": field \"client_id\": non-empty value is required"),
		},
		{
			name:    "authentication oidc missing client secret",
			path:    "./testdata/authentication/oidc_missing_client_secret.yml",
			wantErr: errors.New("provider \"foo\": field \"client_secret\": non-empty value is required"),
		},
		{
			name:    "authentication oidc missing client id",
			path:    "./testdata/authentication/oidc_missing_redirect_address.yml",
			wantErr: errors.New("provider \"foo\": field \"redirect_address\": non-empty value is required"),
		},
		{
			name:    "authentication jwt public key file or jwks url required",
			path:    "./testdata/authentication/jwt_missing_key_file_and_jwks_url.yml",
			wantErr: errors.New("one of jwks_url or public_key_file is required"),
		},
		{
			name:    "authentication jwt public key file and jwks url mutually exclusive",
			path:    "./testdata/authentication/jwt_key_file_and_jwks_url.yml",
			wantErr: errors.New("only one of jwks_url or public_key_file can be set"),
		},
		{
			name:    "authentication jwks invalid url",
			path:    "./testdata/authentication/jwt_invalid_jwks_url.yml",
			wantErr: errors.New(`field "jwks_url": parse " http://localhost:8080/.well-known/jwks.json": first path segment in URL cannot contain colon`),
		},
		{
			name:    "authentication jwt public key file not found",
			path:    "./testdata/authentication/jwt_key_file_not_found.yml",
			wantErr: errors.New(`field "public_key_file": stat testdata/authentication/jwt_key_file.pem: no such file or directory`),
		},
		{
			name: "server cloud missing port",
			path: "./testdata/server/cloud_missing_port.yml",
			expected: func() *Config {
				cfg := Default()
				cfg.Cloud.Host = "flipt.cloud"
				cfg.Server.Cloud.Enabled = true
				cfg.Server.Cloud.Port = 8443
				cfg.Experimental.Cloud.Enabled = true
				return cfg
			},
		},
		{
			name: "cloud trailing slash",
			path: "./testdata/cloud/trim_trailing_slash.yml",
			expected: func() *Config {
				cfg := Default()
				cfg.Cloud.Host = "flipt.cloud"
				cfg.Server.Cloud.Port = 8443
				cfg.Experimental.Cloud.Enabled = true
				return cfg
			},
		},
		{
			name:    "authorization required without authentication",
			path:    "./testdata/authorization/authentication_not_required.yml",
			wantErr: errors.New("authorization requires authentication also be required"),
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
					Session: AuthenticationSession{
						Domain:        "auth.flipt.io",
						Secure:        true,
						TokenLifetime: 24 * time.Hour,
						StateLifetime: 10 * time.Minute,
						CSRF: AuthenticationSessionCSRF{
							Key: "abcdefghijklmnopqrstuvwxyz1234567890", //gitleaks:allow
						},
					},
					Methods: AuthenticationMethods{
						Token: AuthenticationMethod[AuthenticationMethodTokenConfig]{
							Enabled: true,
							Cleanup: &AuthenticationCleanupSchedule{
								Interval:    2 * time.Hour,
								GracePeriod: 48 * time.Hour,
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
							Cleanup: &AuthenticationCleanupSchedule{
								Interval:    2 * time.Hour,
								GracePeriod: 48 * time.Hour,
							},
						},
						Kubernetes: AuthenticationMethod[AuthenticationMethodKubernetesConfig]{
							Enabled: true,
							Method: AuthenticationMethodKubernetesConfig{
								DiscoveryURL:            "https://some-other-k8s.namespace.svc",
								CAPath:                  "/path/to/ca/certificate/ca.pem",
								ServiceAccountTokenPath: "/path/to/sa/token",
							},
							Cleanup: &AuthenticationCleanupSchedule{
								Interval:    2 * time.Hour,
								GracePeriod: 48 * time.Hour,
							},
						},
						Github: AuthenticationMethod[AuthenticationMethodGithubConfig]{
							Method: AuthenticationMethodGithubConfig{
								ClientId:        "abcdefg",
								ClientSecret:    "bcdefgh",
								RedirectAddress: "http://auth.flipt.io",
							},
							Enabled: true,
							Cleanup: &AuthenticationCleanupSchedule{
								Interval:    2 * time.Hour,
								GracePeriod: 48 * time.Hour,
							},
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

				cfg.Audit = AuditConfig{
					Sinks: SinksConfig{
						Log: LogSinkConfig{
							Enabled: true,
							File:    "/path/to/logs.txt",
						},
						Kafka: KafkaSinkConfig{Encoding: "protobuf"},
					},
					Buffer: BufferConfig{
						Capacity:    10,
						FlushPeriod: 3 * time.Minute,
					},
					Events: []string{"*:*"},
				}

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

				cfg.Cache.Enabled = true
				cfg.Cache.Backend = CacheMemory
				cfg.Cache.TTL = 1 * time.Minute
				cfg.Cache.Memory = MemoryCacheConfig{
					EvictionInterval: 5 * time.Minute,
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

				cfg.Tracing = TracingConfig{
					Enabled:       true,
					Exporter:      TracingOTLP,
					SamplingRatio: 1,
					Propagators: []TracingPropagator{
						TracingPropagatorTraceContext,
						TracingPropagatorBaggage,
					},
					Jaeger: JaegerTracingConfig{
						Host: "localhost",
						Port: 6831,
					},
					Zipkin: ZipkinTracingConfig{
						Endpoint: "http://localhost:9411/api/v2/spans",
					},
					OTLP: OTLPTracingConfig{
						Endpoint: "localhost:4318",
					},
				}

				cfg.Storage = StorageConfig{
					Type: GitStorageType,
					Git: &StorageGitConfig{
						Backend: GitBackend{
							Type: GitBackendMemory,
						},
						Repository:   "https://github.com/flipt-io/flipt.git",
						Ref:          "production",
						RefType:      GitRefTypeStatic,
						PollInterval: 5 * time.Second,
						Authentication: Authentication{
							BasicAuth: &BasicAuth{
								Username: "user",
								Password: "pass",
							},
						},
					},
				}

				cfg.Database = DatabaseConfig{
					URL:                       "postgres://postgres@localhost:5432/flipt?sslmode=disable",
					MaxIdleConn:               10,
					MaxOpenConn:               50,
					ConnMaxLifetime:           30 * time.Minute,
					PreparedStatementsEnabled: true,
				}

				cfg.Meta = MetaConfig{
					CheckForUpdates:  false,
					TelemetryEnabled: false,
				}

				cfg.Authentication = AuthenticationConfig{
					Required: true,
					Session: AuthenticationSession{
						Domain:        "auth.flipt.io",
						Secure:        true,
						TokenLifetime: 24 * time.Hour,
						StateLifetime: 10 * time.Minute,
						CSRF: AuthenticationSessionCSRF{
							Key: "abcdefghijklmnopqrstuvwxyz1234567890", //gitleaks:allow
						},
					},
					Methods: AuthenticationMethods{
						Token: AuthenticationMethod[AuthenticationMethodTokenConfig]{
							Enabled: true,
							Cleanup: &AuthenticationCleanupSchedule{
								Interval:    2 * time.Hour,
								GracePeriod: 48 * time.Hour,
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
							Cleanup: &AuthenticationCleanupSchedule{
								Interval:    2 * time.Hour,
								GracePeriod: 48 * time.Hour,
							},
						},
						Kubernetes: AuthenticationMethod[AuthenticationMethodKubernetesConfig]{
							Enabled: true,
							Method: AuthenticationMethodKubernetesConfig{
								DiscoveryURL:            "https://some-other-k8s.namespace.svc",
								CAPath:                  "/path/to/ca/certificate/ca.pem",
								ServiceAccountTokenPath: "/path/to/sa/token",
							},
							Cleanup: &AuthenticationCleanupSchedule{
								Interval:    2 * time.Hour,
								GracePeriod: 48 * time.Hour,
							},
						},
						Github: AuthenticationMethod[AuthenticationMethodGithubConfig]{
							Method: AuthenticationMethodGithubConfig{
								ClientId:        "abcdefg",
								ClientSecret:    "bcdefgh",
								RedirectAddress: "http://auth.flipt.io",
							},
							Enabled: true,
							Cleanup: &AuthenticationCleanupSchedule{
								Interval:    2 * time.Hour,
								GracePeriod: 48 * time.Hour,
							},
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
			name: "version v1",
			path: "./testdata/version/v1.yml",
			expected: func() *Config {
				cfg := Default()
				cfg.Version = "1.0"
				return cfg
			},
		},
		{
			name:    "buffer size invalid capacity",
			path:    "./testdata/audit/invalid_buffer_capacity.yml",
			wantErr: errors.New("buffer capacity below 2 or above 10"),
		},
		{
			name:    "flush period invalid",
			path:    "./testdata/audit/invalid_flush_period.yml",
			wantErr: errors.New("flush period below 2 minutes or greater than 5 minutes"),
		},
		{
			name: "file not specified",
			path: "./testdata/audit/enable_without_file.yml",
			expected: func() *Config {
				cfg := Default()
				cfg.Audit.Sinks.Log.Enabled = true
				return cfg
			},
		},
		{
			name:    "url or template not specified",
			path:    "./testdata/audit/invalid_webhook_url_or_template_not_provided.yml",
			wantErr: errors.New("url or template(s) not provided"),
		},
		{
			name: "local config provided",
			path: "./testdata/storage/local_provided.yml",
			expected: func() *Config {
				cfg := Default()
				cfg.Storage = StorageConfig{
					Type: LocalStorageType,
					Local: &StorageLocalConfig{
						Path: ".",
					},
				}
				return cfg
			},
		},
		{
			name: "audit kafka valid",
			path: "./testdata/audit/valid_kafka.yml",
			expected: func() *Config {
				cfg := Default()
				cfg.Audit.Sinks.Kafka.Enabled = true
				cfg.Audit.Sinks.Kafka.Topic = "audit-topic"
				cfg.Audit.Sinks.Kafka.BootstrapServers = []string{"kafka-srv1", "kafka-srv2"}
				cfg.Audit.Sinks.Kafka.Encoding = "protobuf"
				cfg.Audit.Sinks.Kafka.Authentication = &KafkaAuthenticationConfig{
					Username: "user",
					Password: "passwd",
				}
				cfg.Audit.Sinks.Kafka.SchemaRegistry = &KafkaSchemaRegistryConfig{
					URL: "http://registry",
				}
				cfg.Audit.Sinks.Kafka.RequireTLS = true
				cfg.Audit.Sinks.Kafka.InsecureSkipTLS = true
				return cfg
			},
		},
		{
			name:    "audit kafka invalid topic",
			path:    "./testdata/audit/invalid_kafka_topic.yml",
			wantErr: errors.New("kafka topic not provided"),
		},
		{
			name:    "audit kafka invalid bootstrap servers",
			path:    "./testdata/audit/invalid_kafka_servers.yml",
			wantErr: errors.New("kafka bootstrap_servers not provided"),
		},
		{
			name: "git config provided",
			path: "./testdata/storage/git_provided.yml",
			expected: func() *Config {
				cfg := Default()
				cfg.Storage = StorageConfig{
					Type: GitStorageType,
					Git: &StorageGitConfig{
						Backend: GitBackend{
							Type: GitBackendMemory,
						},
						Ref:          "main",
						RefType:      GitRefTypeStatic,
						Repository:   "git@github.com:foo/bar.git",
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
				cfg.Storage = StorageConfig{
					Type: GitStorageType,
					Git: &StorageGitConfig{
						Backend: GitBackend{
							Type: GitBackendMemory,
						},
						Ref:          "main",
						RefType:      GitRefTypeStatic,
						Repository:   "git@github.com:foo/bar.git",
						Directory:    "baz",
						PollInterval: 30 * time.Second,
					},
				}
				return cfg
			},
		},
		{
			name: "git config provided with ref_type",
			path: "./testdata/storage/git_provided_with_ref_type.yml",
			expected: func() *Config {
				cfg := Default()
				cfg.Storage = StorageConfig{
					Type: GitStorageType,
					Git: &StorageGitConfig{
						Backend: GitBackend{
							Type: GitBackendMemory,
						},
						Ref:          "main",
						RefType:      GitRefTypeSemver,
						Repository:   "git@github.com:foo/bar.git",
						Directory:    "baz",
						PollInterval: 30 * time.Second,
					},
				}
				return cfg
			},
		},
		{
			name: "git config provided with ref_type",
			path: "./testdata/storage/git_provided_with_backend_type.yml",
			expected: func() *Config {
				cfg := Default()
				cfg.Storage = StorageConfig{
					Type: GitStorageType,
					Git: &StorageGitConfig{
						Backend: GitBackend{
							Type: GitBackendLocal,
							Path: "/path/to/gitdir",
						},
						Ref:          "main",
						RefType:      GitRefTypeStatic,
						Repository:   "git@github.com:foo/bar.git",
						PollInterval: 30 * time.Second,
					},
				}
				return cfg
			},
		},
		{
			name:    "git repository not provided",
			path:    "./testdata/storage/invalid_git_repo_not_specified.yml",
			wantErr: errors.New("git repository must be specified"),
		},
		{
			name:    "git invalid ref_type provided",
			path:    "./testdata/storage/git_invalid_ref_type.yml",
			wantErr: errors.New("invalid git storage reference type"),
		},
		{
			name:    "git basic auth partially provided",
			path:    "./testdata/storage/git_basic_auth_invalid.yml",
			wantErr: errors.New("both username and password need to be provided for basic auth"),
		},
		{
			name:    "git ssh auth missing password",
			path:    "./testdata/storage/git_ssh_auth_invalid_missing_password.yml",
			wantErr: errors.New("ssh authentication: password required"),
		},
		{
			name:    "git ssh auth missing private key parts",
			path:    "./testdata/storage/git_ssh_auth_invalid_private_key_missing.yml",
			wantErr: errors.New("ssh authentication: please provide exclusively one of private_key_bytes or private_key_path"),
		},
		{
			name:    "git ssh auth provided both private key forms",
			path:    "./testdata/storage/git_ssh_auth_invalid_private_key_both.yml",
			wantErr: errors.New("ssh authentication: please provide exclusively one of private_key_bytes or private_key_path"),
		},
		{
			name: "git valid with ssh auth",
			path: "./testdata/storage/git_ssh_auth_valid_with_path.yml",
			expected: func() *Config {
				cfg := Default()
				cfg.Storage = StorageConfig{
					Type: GitStorageType,
					Git: &StorageGitConfig{
						Backend: GitBackend{
							Type: GitBackendMemory,
						},
						Ref:          "main",
						RefType:      GitRefTypeStatic,
						Repository:   "git@github.com:foo/bar.git",
						PollInterval: 30 * time.Second,
						Authentication: Authentication{
							SSHAuth: &SSHAuth{
								User:           "git",
								Password:       "bar",
								PrivateKeyPath: "/path/to/pem.key",
							},
						},
					},
				}
				return cfg
			},
		},
		{
			name: "s3 config provided",
			path: "./testdata/storage/s3_provided.yml",
			expected: func() *Config {
				cfg := Default()
				cfg.Storage = StorageConfig{
					Type: ObjectStorageType,
					Object: &StorageObjectConfig{
						Type: S3ObjectSubStorageType,
						S3: &S3Storage{
							Bucket:       "testbucket",
							PollInterval: time.Minute,
						},
					},
				}
				return cfg
			},
		},
		{
			name: "s3 full config provided",
			path: "./testdata/storage/s3_full.yml",
			expected: func() *Config {
				cfg := Default()
				cfg.Storage = StorageConfig{
					Type: ObjectStorageType,
					Object: &StorageObjectConfig{
						Type: S3ObjectSubStorageType,
						S3: &S3Storage{
							Bucket:       "testbucket",
							Prefix:       "prefix",
							Region:       "region",
							PollInterval: 5 * time.Minute,
						},
					},
				}
				return cfg
			},
		},
		{
			name: "OCI config provided",
			path: "./testdata/storage/oci_provided.yml",
			expected: func() *Config {
				cfg := Default()
				cfg.Storage = StorageConfig{
					Type: OCIStorageType,
					OCI: &StorageOCIConfig{
						Repository:       "some.target/repository/abundle:latest",
						BundlesDirectory: "/tmp/bundles",
						Authentication: &OCIAuthentication{
							Type:     oci.AuthenticationTypeStatic,
							Username: "foo",
							Password: "bar",
						},
						PollInterval:    5 * time.Minute,
						ManifestVersion: "1.1",
					},
				}
				return cfg
			},
		},
		{
			name: "OCI config provided full",
			path: "./testdata/storage/oci_provided_full.yml",
			expected: func() *Config {
				cfg := Default()
				cfg.Storage = StorageConfig{
					Type: OCIStorageType,
					OCI: &StorageOCIConfig{
						Repository:       "some.target/repository/abundle:latest",
						BundlesDirectory: "/tmp/bundles",
						Authentication: &OCIAuthentication{
							Type:     oci.AuthenticationTypeStatic,
							Username: "foo",
							Password: "bar",
						},
						PollInterval:    5 * time.Minute,
						ManifestVersion: "1.0",
					},
				}
				return cfg
			},
		},
		{
			name: "OCI config provided AWS ECR",
			path: "./testdata/storage/oci_provided_aws_ecr.yml",
			expected: func() *Config {
				cfg := Default()
				cfg.Storage = StorageConfig{
					Type: OCIStorageType,
					OCI: &StorageOCIConfig{
						Repository:       "some.target/repository/abundle:latest",
						BundlesDirectory: "/tmp/bundles",
						Authentication: &OCIAuthentication{
							Type: oci.AuthenticationTypeAWSECR,
						},
						PollInterval:    5 * time.Minute,
						ManifestVersion: "1.1",
					},
				}
				return cfg
			},
		},
		{
			name: "OCI config provided with no authentication",
			path: "./testdata/storage/oci_provided_no_auth.yml",
			expected: func() *Config {
				cfg := Default()
				cfg.Storage = StorageConfig{
					Type: OCIStorageType,
					OCI: &StorageOCIConfig{
						Repository:       "some.target/repository/abundle:latest",
						BundlesDirectory: "/tmp/bundles",
						PollInterval:     5 * time.Minute,
						ManifestVersion:  "1.1",
					},
				}
				return cfg
			},
		},
		{
			name:    "OCI config provided with invalid authentication type",
			path:    "./testdata/storage/oci_provided_invalid_auth.yml",
			wantErr: errors.New("oci authentication type is not supported"),
		},
		{
			name:    "OCI invalid no repository",
			path:    "./testdata/storage/oci_invalid_no_repo.yml",
			wantErr: errors.New("oci storage repository must be specified"),
		},
		{
			name:    "OCI invalid unexpected scheme",
			path:    "./testdata/storage/oci_invalid_unexpected_scheme.yml",
			wantErr: errors.New("validating OCI configuration: unexpected repository scheme: \"unknown\" should be one of [http|https|flipt]"),
		},
		{
			name:    "OCI invalid wrong manifest version",
			path:    "./testdata/storage/oci_invalid_manifest_version.yml",
			wantErr: errors.New("wrong manifest version, it should be 1.0 or 1.1"),
		},
		{
			name:    "storage readonly config invalid",
			path:    "./testdata/storage/invalid_readonly.yml",
			wantErr: errors.New("setting read only mode is only supported with database storage"),
		},
		{
			name:    "s3 config invalid",
			path:    "./testdata/storage/s3_bucket_missing.yml",
			wantErr: errors.New("s3 bucket must be specified"),
		},
		{
			name:    "object storage type not provided",
			path:    "./testdata/storage/invalid_object_storage_type_not_specified.yml",
			wantErr: errors.New("object storage type must be specified"),
		},
		{
			name:    "azblob config invalid",
			path:    "./testdata/storage/azblob_invalid.yml",
			wantErr: errors.New("azblob container must be specified"),
		},
		{
			name: "azblob full config provided",
			path: "./testdata/storage/azblob_full.yml",
			expected: func() *Config {
				cfg := Default()
				cfg.Storage = StorageConfig{
					Type: ObjectStorageType,
					Object: &StorageObjectConfig{
						Type: AZBlobObjectSubStorageType,
						AZBlob: &AZBlobStorage{
							Container:    "testdata",
							Endpoint:     "https//devaccount.blob.core.windows.net",
							PollInterval: 5 * time.Minute,
						},
					},
				}
				return cfg
			},
		},
		{
			name:    "gs config invalid",
			path:    "./testdata/storage/gs_invalid.yml",
			wantErr: errors.New("googlecloud bucket must be specified"),
		},
		{
			name: "gs full config provided",
			path: "./testdata/storage/gs_full.yml",
			expected: func() *Config {
				cfg := Default()
				cfg.Storage = StorageConfig{
					Type: ObjectStorageType,
					Object: &StorageObjectConfig{
						Type: GSBlobObjectSubStorageType,
						GS: &GSStorage{
							Bucket:       "testdata",
							Prefix:       "prefix",
							PollInterval: 5 * time.Minute,
						},
					},
				}
				return cfg
			},
		},
		{
			name: "grpc keepalive config provided",
			path: "./testdata/server/grpc_keepalive.yml",
			expected: func() *Config {
				cfg := Default()
				cfg.Server.GRPCConnectionMaxIdleTime = 1 * time.Hour
				cfg.Server.GRPCConnectionMaxAge = 30 * time.Second
				cfg.Server.GRPCConnectionMaxAgeGrace = 10 * time.Second
				return cfg
			},
		},
		{
			name:    "clickhouse enabled but no URL set",
			path:    "./testdata/analytics/invalid_clickhouse_configuration_empty_url.yml",
			wantErr: errors.New("clickhouse url not provided"),
		},
		{
			name:    "analytics flush period too low",
			path:    "./testdata/analytics/invalid_buffer_configuration_flush_period.yml",
			wantErr: errors.New("flush period below 10 seconds"),
		},
		{
			name: "ui topbar with correct hex color",
			path: "./testdata/ui/topbar_color.yml",
			expected: func() *Config {
				cfg := Default()
				cfg.UI.Topbar.Color = "#42bda0"
				cfg.UI.Topbar.Label = "World"
				return cfg
			},
		},
		{
			name:    "ui topbar with invalid hex color",
			path:    "./testdata/ui/topbar_invalid_color.yml",
			wantErr: errors.New("expected valid hex color, got invalid"),
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
	for key, value := range v {
		switch v := value.(type) {
		case map[any]any:
			vals = append(vals, getEnvVars(fmt.Sprintf("%s_%v", prefix, key), v)...)
		case []any:
			builder := strings.Builder{}
			for i, s := range v {
				builder.WriteString(fmt.Sprintf("%v", s))
				if i < len(v)-1 {
					builder.WriteByte(' ')
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

	return
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
			cfg: func() *Config {
				cfg := Default()
				// override the database URL to a file path for testing
				cfg.Database.URL = "file:/tmp/flipt/flipt.db"
				return cfg
			},
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
		test := test
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

var (
	// add any struct tags to match their camelCase equivalents here.
	camelCaseMatchers = map[string]string{
		"requireTLS":   "requireTLS",
		"discoveryURL": "discoveryURL",
	}
)

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

	for i := 0; i < t.NumField(); i++ {
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
