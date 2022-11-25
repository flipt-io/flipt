package config

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uber/jaeger-client-go"
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
			Jaeger: JaegerTracingConfig{
				Enabled: false,
				Host:    jaeger.DefaultUDPSpanServerHost,
				Port:    jaeger.DefaultUDPSpanServerPort,
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
	}
}

func TestLoad(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		wantErr  error
		expected func() *Config
	}{
		{
			name:     "defaults",
			path:     "./testdata/default.yml",
			expected: defaultConfig,
		},
		{
			name:     "deprecated - cache memory items defaults",
			path:     "./testdata/deprecated/cache_memory_items.yml",
			expected: defaultConfig,
		},
		{
			name: "deprecated - cache memory enabled",
			path: "./testdata/deprecated/cache_memory_enabled.yml",
			expected: func() *Config {
				cfg := defaultConfig()
				cfg.Cache.Enabled = true
				cfg.Cache.Backend = CacheMemory
				cfg.Cache.TTL = -time.Second
				cfg.Warnings = append(cfg.Warnings, deprecatedMsgMemoryEnabled, deprecatedMsgMemoryExpiration)
				return cfg
			},
		},
		{
			name: "deprecated - database migrations path",
			path: "./testdata/deprecated/database_migrations_path.yml",
			expected: func() *Config {
				cfg := defaultConfig()
				cfg.Warnings = append(cfg.Warnings, deprecatedMsgDatabaseMigrations)
				return cfg
			},
		},
		{
			name: "deprecated - database migrations path legacy",
			path: "./testdata/deprecated/database_migrations_path_legacy.yml",
			expected: func() *Config {
				cfg := defaultConfig()
				cfg.Warnings = append(cfg.Warnings, deprecatedMsgDatabaseMigrations)
				return cfg
			},
		},
		{
			name: "cache - no backend set",
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
			name: "cache - memory",
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
			name: "cache - redis",
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
			name:    "server - https missing cert file",
			path:    "./testdata/server/https_missing_cert_file.yml",
			wantErr: errValidationRequired,
		},
		{
			name:    "server - https missing cert key",
			path:    "./testdata/server/https_missing_cert_key.yml",
			wantErr: errValidationRequired,
		},
		{
			name:    "server - https defined but not found cert file",
			path:    "./testdata/server/https_not_found_cert_file.yml",
			wantErr: fs.ErrNotExist,
		},
		{
			name:    "server - https defined but not found cert key",
			path:    "./testdata/server/https_not_found_cert_key.yml",
			wantErr: fs.ErrNotExist,
		},
		{
			name:    "database - protocol required",
			path:    "./testdata/database/missing_protocol.yml",
			wantErr: errValidationRequired,
		},
		{
			name:    "database - host required",
			path:    "./testdata/database/missing_host.yml",
			wantErr: errValidationRequired,
		},
		{
			name:    "database - name required",
			path:    "./testdata/database/missing_name.yml",
			wantErr: errValidationRequired,
		},
		{
			name: "advanced",
			path: "./testdata/advanced.yml",
			expected: func() *Config {
				cfg := defaultConfig()
				cfg.Log = LogConfig{
					Level:     "WARN",
					File:      "testLogFile.txt",
					Encoding:  LogEncodingJSON,
					GRPCLevel: "ERROR",
				}
				cfg.UI = UIConfig{
					Enabled: false,
				}
				cfg.Cors = CorsConfig{
					Enabled:        true,
					AllowedOrigins: []string{"foo.com"},
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
					Jaeger: JaegerTracingConfig{
						Enabled: true,
						Host:    "localhost",
						Port:    6831,
					},
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
					Methods: AuthenticationMethods{
						Token: AuthenticationMethodTokenConfig{
							Enabled: true,
						},
					},
					Cleanup: AuthenticationCleanupSchedules{
						Token: &AuthenticationCleanupSchedule{
							Interval:    2 * time.Hour,
							GracePeriod: 48 * time.Hour,
						},
					},
				}
				return cfg
			},
		},
	}

	for _, tt := range tests {
		var (
			path     = tt.path
			wantErr  = tt.wantErr
			expected *Config
		)

		if tt.expected != nil {
			expected = tt.expected()
		}

		t.Run(tt.name+" (YAML)", func(t *testing.T) {
			cfg, err := Load(path)

			if wantErr != nil {
				t.Log(err)
				require.ErrorIs(t, err, wantErr)
				return
			}

			require.NoError(t, err)

			assert.NotNil(t, cfg)
			assert.Equal(t, expected, cfg)
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
			cfg, err := Load("./testdata/default.yml")

			if wantErr != nil {
				t.Log(err)
				require.ErrorIs(t, err, wantErr)
				return
			}

			require.NoError(t, err)

			assert.NotNil(t, cfg)
			assert.Equal(t, expected, cfg)
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
