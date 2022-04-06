package config

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	jaeger "github.com/uber/jaeger-client-go"
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

func TestLoad(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		wantErr  bool
		expected *Config
	}{
		{
			name:     "defaults",
			path:     "./testdata/default.yml",
			expected: Default(),
		},
		{
			name:     "deprecated defaults",
			path:     "./testdata/deprecated.yml",
			expected: Default(),
		},
		{
			name: "database key/value",
			path: "./testdata/database.yml",
			expected: &Config{
				Log: LogConfig{
					Level: "INFO",
				},

				UI: UIConfig{
					Enabled: true,
				},

				Cors: CorsConfig{
					Enabled:        false,
					AllowedOrigins: []string{"*"},
				},

				Cache: CacheConfig{
					Memory: MemoryCacheConfig{
						Enabled:          false,
						Expiration:       -1,
						EvictionInterval: 10 * time.Minute,
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
					Protocol:       DatabaseMySQL,
					Host:           "localhost",
					Port:           3306,
					User:           "flipt",
					Password:       "s3cr3t!",
					Name:           "flipt",
					MigrationsPath: "/etc/flipt/config/migrations",
					MaxIdleConn:    2,
				},

				Meta: MetaConfig{
					CheckForUpdates:  true,
					TelemetryEnabled: true,
				},
			},
		},
		{
			name: "advanced",
			path: "./testdata/advanced.yml",
			expected: &Config{
				Log: LogConfig{
					Level: "WARN",
					File:  "testLogFile.txt",
				},
				UI: UIConfig{
					Enabled: false,
				},
				Cors: CorsConfig{
					Enabled:        true,
					AllowedOrigins: []string{"foo.com"},
				},
				Cache: CacheConfig{
					Memory: MemoryCacheConfig{
						Enabled:          true,
						Expiration:       5 * time.Minute,
						EvictionInterval: 1 * time.Minute,
					},
				},
				Server: ServerConfig{
					Host:      "127.0.0.1",
					Protocol:  HTTPS,
					HTTPPort:  8081,
					HTTPSPort: 8080,
					GRPCPort:  9001,
					CertFile:  "./testdata/ssl_cert.pem",
					CertKey:   "./testdata/ssl_key.pem",
				},
				Tracing: TracingConfig{
					Jaeger: JaegerTracingConfig{
						Enabled: true,
						Host:    "localhost",
						Port:    6831,
					},
				},
				Database: DatabaseConfig{
					MigrationsPath:  "./config/migrations",
					URL:             "postgres://postgres@localhost:5432/flipt?sslmode=disable",
					MaxIdleConn:     10,
					MaxOpenConn:     50,
					ConnMaxLifetime: 30 * time.Minute,
				},
				Meta: MetaConfig{
					CheckForUpdates:  false,
					TelemetryEnabled: false,
				},
			},
		},
	}

	for _, tt := range tests {
		var (
			path     = tt.path
			wantErr  = tt.wantErr
			expected = tt.expected
		)

		t.Run(tt.name, func(t *testing.T) {
			cfg, err := Load(path)

			if wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			assert.NotNil(t, cfg)
			assert.Equal(t, expected, cfg)
		})
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name       string
		cfg        *Config
		wantErrMsg string
	}{
		{
			name: "https: valid",
			cfg: &Config{
				Server: ServerConfig{
					Protocol: HTTPS,
					CertFile: "./testdata/ssl_cert.pem",
					CertKey:  "./testdata/ssl_key.pem",
				},
				Database: DatabaseConfig{
					URL: "localhost",
				},
			},
		},
		{
			name: "http: valid",
			cfg: &Config{
				Server: ServerConfig{
					Protocol: HTTP,
				},
				Database: DatabaseConfig{
					URL: "localhost",
				},
			},
		},
		{
			name: "https: empty cert_file path",
			cfg: &Config{
				Server: ServerConfig{
					Protocol: HTTPS,
					CertFile: "",
					CertKey:  "./testdata/ssl_key.pem",
				},
			},
			wantErrMsg: "server.cert_file cannot be empty when using HTTPS",
		},
		{
			name: "https: empty key_file path",
			cfg: &Config{
				Server: ServerConfig{
					Protocol: HTTPS,
					CertFile: "./testdata/ssl_cert.pem",
					CertKey:  "",
				},
			},
			wantErrMsg: "server.cert_key cannot be empty when using HTTPS",
		},
		{
			name: "https: missing cert_file",
			cfg: &Config{
				Server: ServerConfig{
					Protocol: HTTPS,
					CertFile: "foo.pem",
					CertKey:  "./testdata/ssl_key.pem",
				},
			},
			wantErrMsg: "cannot find TLS server.cert_file at \"foo.pem\"",
		},
		{
			name: "https: missing key_file",
			cfg: &Config{
				Server: ServerConfig{
					Protocol: HTTPS,
					CertFile: "./testdata/ssl_cert.pem",
					CertKey:  "bar.pem",
				},
			},
			wantErrMsg: "cannot find TLS server.cert_key at \"bar.pem\"",
		},
		{
			name: "db: missing protocol",
			cfg: &Config{
				Server: ServerConfig{
					Protocol: HTTP,
				},
				Database: DatabaseConfig{},
			},
			wantErrMsg: "database.protocol cannot be empty",
		},
		{
			name: "db: missing host",
			cfg: &Config{
				Server: ServerConfig{
					Protocol: HTTP,
				},
				Database: DatabaseConfig{
					Protocol: DatabaseSQLite,
				},
			},
			wantErrMsg: "database.host cannot be empty",
		},
		{
			name: "db: missing name",
			cfg: &Config{
				Server: ServerConfig{
					Protocol: HTTP,
				},
				Database: DatabaseConfig{
					Protocol: DatabaseSQLite,
					Host:     "localhost",
				},
			},
			wantErrMsg: "database.name cannot be empty",
		},
	}

	for _, tt := range tests {
		var (
			cfg        = tt.cfg
			wantErrMsg = tt.wantErrMsg
		)

		t.Run(tt.name, func(t *testing.T) {
			err := cfg.validate()

			if wantErrMsg != "" {
				require.Error(t, err)
				assert.EqualError(t, err, wantErrMsg)
				return
			}

			require.NoError(t, err)
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

	body, _ := ioutil.ReadAll(resp.Body)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.NotEmpty(t, body)
}
