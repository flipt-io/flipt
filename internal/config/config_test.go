package config

import (
	"errors"
	"fmt"
	"io/fs"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/santhosh-tekuri/jsonschema/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uber/jaeger-client-go"
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
			assert.NoError(t, err)
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
			assert.NoError(t, err)
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
			assert.NoError(t, err)
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
			want:     "file",
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
			assert.NoError(t, err)
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
			assert.NoError(t, err)
			assert.JSONEq(t, fmt.Sprintf("%q", want), string(json))
		})
	}
}

func defaultConfig() *Config {
	return &Config{
		Log: LogConfig{
			Level:     "INFO",
			Encoding:  LogEncodingConsole,
			GRPCLevel: "ERROR",
			Keys: LogKeys{
				Time:    "T",
				Level:   "L",
				Message: "M",
			},
		},

		UI: UIConfig{
			Enabled: true,
		},

		Cors: CorsConfig{
			Enabled:        false,
			AllowedOrigins: []string{"*"},
		},

		Cache: CacheConfig{
			Enabled: false,
			Backend: CacheMemory,
			TTL:     1 * time.Minute,
			Memory: MemoryCacheConfig{
				EvictionInterval: 5 * time.Minute,
			},
			Redis: RedisCacheConfig{
				Host:     "localhost",
				Port:     6379,
				Password: "",
				DB:       0,
			},
		},

		Server: ServerConfig{
			Host:      "0.0.0.0",
			Protocol:  HTTP,
			HTTPPort:  8080,
			HTTPSPort: 443,
			GRPCPort:  9000,
		},

		Tracing: TracingConfig{
			Enabled:  false,
			Exporter: TracingJaeger,
			Jaeger: JaegerTracingConfig{
				Host: jaeger.DefaultUDPSpanServerHost,
				Port: jaeger.DefaultUDPSpanServerPort,
			},
			Zipkin: ZipkinTracingConfig{
				Endpoint: "http://localhost:9411/api/v2/spans",
			},
			OTLP: OTLPTracingConfig{
				Endpoint: "localhost:4317",
			},
		},

		Database: DatabaseConfig{
			URL:         "file:/var/opt/flipt/flipt.db",
			MaxIdleConn: 2,
		},

		Meta: MetaConfig{
			CheckForUpdates:  true,
			TelemetryEnabled: true,
			StateDirectory:   "",
		},

		Authentication: AuthenticationConfig{
			Session: AuthenticationSession{
				TokenLifetime: 24 * time.Hour,
				StateLifetime: 10 * time.Minute,
			},
		},

		Audit: AuditConfig{
			Sinks: SinksConfig{
				LogFile: LogFileSinkConfig{
					Enabled: false,
					File:    "",
				},
			},
			Buffer: BufferConfig{
				Capacity:    2,
				FlushPeriod: 2 * time.Minute,
			},
		},
	}
}

func TestLoad(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		wantErr  error
		expected func() *Config
		warnings []string
	}{
		{
			name:     "defaults",
			path:     "./testdata/default.yml",
			expected: defaultConfig,
		},
		{
			name: "deprecated tracing jaeger enabled",
			path: "./testdata/deprecated/tracing_jaeger_enabled.yml",
			expected: func() *Config {
				cfg := defaultConfig()
				cfg.Tracing.Enabled = true
				cfg.Tracing.Exporter = TracingJaeger
				return cfg
			},
			warnings: []string{
				"\"tracing.jaeger.enabled\" is deprecated and will be removed in a future version. Please use 'tracing.enabled' and 'tracing.exporter' instead.",
			},
		},
		{
			name: "deprecated cache memory enabled",
			path: "./testdata/deprecated/cache_memory_enabled.yml",
			expected: func() *Config {
				cfg := defaultConfig()
				cfg.Cache.Enabled = true
				cfg.Cache.Backend = CacheMemory
				cfg.Cache.TTL = -time.Second
				return cfg
			},
			warnings: []string{
				"\"cache.memory.enabled\" is deprecated and will be removed in a future version. Please use 'cache.enabled' and 'cache.backend' instead.",
				"\"cache.memory.expiration\" is deprecated and will be removed in a future version. Please use 'cache.ttl' instead.",
			},
		},
		{
			name:     "deprecated cache memory items defaults",
			path:     "./testdata/deprecated/cache_memory_items.yml",
			expected: defaultConfig,
			warnings: []string{
				"\"cache.memory.enabled\" is deprecated and will be removed in a future version. Please use 'cache.enabled' and 'cache.backend' instead.",
			},
		},
		{
			name:     "deprecated database migrations path",
			path:     "./testdata/deprecated/database_migrations_path.yml",
			expected: defaultConfig,
			warnings: []string{"\"db.migrations.path\" is deprecated and will be removed in a future version. Migrations are now embedded within Flipt and are no longer required on disk."},
		},
		{
			name:     "deprecated database migrations path legacy",
			path:     "./testdata/deprecated/database_migrations_path_legacy.yml",
			expected: defaultConfig,
			warnings: []string{"\"db.migrations.path\" is deprecated and will be removed in a future version. Migrations are now embedded within Flipt and are no longer required on disk."},
		},
		{
			name: "deprecated ui disabled",
			path: "./testdata/deprecated/ui_disabled.yml",
			expected: func() *Config {
				cfg := defaultConfig()
				cfg.UI.Enabled = false
				return cfg
			},
			warnings: []string{"\"ui.enabled\" is deprecated and will be removed in a future version."},
		},
		{
			name: "cache no backend set",
			path: "./testdata/cache/default.yml",
			expected: func() *Config {
				cfg := defaultConfig()
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
				cfg := defaultConfig()
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
				cfg := defaultConfig()
				cfg.Cache.Enabled = true
				cfg.Cache.Backend = CacheRedis
				cfg.Cache.TTL = time.Minute
				cfg.Cache.Redis.Host = "localhost"
				cfg.Cache.Redis.Port = 6378
				cfg.Cache.Redis.DB = 1
				cfg.Cache.Redis.Password = "s3cr3t!"
				return cfg
			},
		},
		{
			name: "tracing zipkin",
			path: "./testdata/tracing/zipkin.yml",
			expected: func() *Config {
				cfg := defaultConfig()
				cfg.Tracing.Enabled = true
				cfg.Tracing.Exporter = TracingZipkin
				cfg.Tracing.Zipkin.Endpoint = "http://localhost:9999/api/v2/spans"
				return cfg
			},
		},
		{
			name: "database key/value",
			path: "./testdata/database.yml",
			expected: func() *Config {
				cfg := defaultConfig()
				cfg.Database = DatabaseConfig{
					Protocol:    DatabaseMySQL,
					Host:        "localhost",
					Port:        3306,
					User:        "flipt",
					Password:    "s3cr3t!",
					Name:        "flipt",
					MaxIdleConn: 2,
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
				cfg := defaultConfig()
				cfg.Authentication.Methods.Token.Method.Bootstrap = AuthenticationMethodTokenBootstrapConfig{
					Token:      "s3cr3t!",
					Expiration: 24 * time.Hour,
				}
				return cfg
			},
		},
		{
			name: "authentication session strip domain scheme/port",
			path: "./testdata/authentication/session_domain_scheme_port.yml",
			expected: func() *Config {
				cfg := defaultConfig()
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
			name: "authentication kubernetes defaults when enabled",
			path: "./testdata/authentication/kubernetes.yml",
			expected: func() *Config {
				cfg := defaultConfig()
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
			name: "advanced",
			path: "./testdata/advanced.yml",
			expected: func() *Config {
				cfg := defaultConfig()

				cfg.Experimental.FilesystemStorage.Enabled = true

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
					Enabled:  true,
					Exporter: TracingJaeger,
					Jaeger: JaegerTracingConfig{
						Host: "localhost",
						Port: 6831,
					},
					Zipkin: ZipkinTracingConfig{
						Endpoint: "http://localhost:9411/api/v2/spans",
					},
					OTLP: OTLPTracingConfig{
						Endpoint: "localhost:4317",
					},
				}
				cfg.Storage = StorageConfig{
					Type: StorageType("database"),
				}
				cfg.Database = DatabaseConfig{
					URL:             "postgres://postgres@localhost:5432/flipt?sslmode=disable",
					MaxIdleConn:     10,
					MaxOpenConn:     50,
					ConnMaxLifetime: 30 * time.Minute,
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
					},
				}
				return cfg
			},
		},
		{
			name: "version v1",
			path: "./testdata/version/v1.yml",
			expected: func() *Config {
				cfg := defaultConfig()
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
			name:    "file not specified",
			path:    "./testdata/audit/invalid_enable_without_file.yml",
			wantErr: errors.New("file not specified"),
		},
		{
			name: "local config provided",
			path: "./testdata/storage/local_provided.yml",
			expected: func() *Config {
				cfg := defaultConfig()
				cfg.Experimental.FilesystemStorage.Enabled = true
				cfg.Storage = StorageConfig{
					Type: LocalStorageType,
					Local: Local{
						Path: ".",
					},
				}
				return cfg
			},
		},
		{
			name: "git config provided",
			path: "./testdata/storage/git_provided.yml",
			expected: func() *Config {
				cfg := defaultConfig()
				cfg.Experimental.FilesystemStorage.Enabled = true
				cfg.Storage = StorageConfig{
					Type: GitStorageType,
					Git: Git{
						Ref:          "main",
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
			name:    "git basic auth partially provided",
			path:    "./testdata/storage/git_basic_auth_invalid.yml",
			wantErr: errors.New("both username and password need to be provided for basic auth"),
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
			res, err := Load(path)

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

			// read the input config file into equivalent envs
			envs := readYAMLIntoEnv(t, path)
			for _, env := range envs {
				t.Logf("Setting env '%s=%s'\n", env[0], env[1])
				os.Setenv(env[0], env[1])
			}

			// load default (empty) config
			res, err := Load("./testdata/default.yml")

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

			assert.NotNil(t, res)
			assert.Equal(t, expected, res.Config)
		})
	}
}

func TestServeHTTP(t *testing.T) {
	var (
		cfg = defaultConfig()
		req = httptest.NewRequest("GET", "http://example.com/foo", nil)
		w   = httptest.NewRecorder()
	)

	cfg.ServeHTTP(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

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
		default:
			vals = append(vals, [2]string{
				fmt.Sprintf("%s_%s", strings.ToUpper(prefix), strings.ToUpper(fmt.Sprintf("%v", key))),
				fmt.Sprintf("%v", value),
			})
		}
	}

	return
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
