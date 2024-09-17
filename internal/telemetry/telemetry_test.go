package telemetry

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/internal/info"
	"go.uber.org/zap/zaptest"

	segment "gopkg.in/segmentio/analytics-go.v3"
)

var _ segment.Client = &mockAnalytics{}

type mockAnalytics struct {
	msg        segment.Message
	enqueueErr error
	closed     bool
}

func (m *mockAnalytics) Enqueue(msg segment.Message) error {
	m.msg = msg
	return m.enqueueErr
}

func (m *mockAnalytics) Close() error {
	m.closed = true
	return nil
}

type mockFile struct {
	io.Reader
	io.Writer
}

func (m *mockFile) Seek(offset int64, whence int) (int64, error) {
	return 0, nil
}

func (m *mockFile) Truncate(_ int64) error {
	return nil
}

func TestNewReporter(t *testing.T) {
	var (
		cfg = config.Config{
			Meta: config.MetaConfig{
				TelemetryEnabled: true,
			},
		}

		logger        = zaptest.NewLogger(t)
		reporter, err = NewReporter(cfg, logger, "foo", "http://localhost:8080", info.Flipt{})
	)
	require.NoError(t, err)
	assert.NotNil(t, reporter)
}

func TestShutdown(t *testing.T) {
	var (
		logger        = zaptest.NewLogger(t)
		mockAnalytics = &mockAnalytics{}

		reporter = &Reporter{
			cfg: config.Config{
				Meta: config.MetaConfig{
					TelemetryEnabled: true,
				},
			},
			logger:   logger,
			client:   mockAnalytics,
			shutdown: make(chan struct{}),
		}
	)

	err := reporter.Shutdown()
	require.NoError(t, err)

	assert.True(t, mockAnalytics.closed)
}

var experimental = map[string]any{
	"cloud": map[string]any{},
}

func TestPing(t *testing.T) {
	test := []struct {
		name string
		cfg  config.Config
		want map[string]any
	}{
		{
			name: "basic",
			cfg: config.Config{
				Database: config.DatabaseConfig{
					Protocol: config.DatabaseSQLite,
				},
			},
			want: map[string]any{
				"version": "1.0.0",
				"os":      "linux",
				"arch":    "amd64",
				"storage": map[string]any{
					"database": "sqlite",
				},
				"experimental": experimental,
			},
		},
		{
			name: "with db url",
			cfg: config.Config{
				Database: config.DatabaseConfig{
					URL: "sqlite:///foo.db",
				},
			},
			want: map[string]any{
				"version": "1.0.0",
				"os":      "linux",
				"arch":    "amd64",
				"storage": map[string]any{
					"database": "sqlite",
				},
				"experimental": experimental,
			},
		},
		{
			name: "with unknown db url",
			cfg: config.Config{
				Database: config.DatabaseConfig{
					URL: "foo:///foo.db",
				},
			},
			want: map[string]any{
				"version": "1.0.0",
				"os":      "linux",
				"arch":    "amd64",
				"storage": map[string]any{
					"database": "unknown",
				},
				"experimental": experimental,
			},
		},
		{
			name: "with cache not enabled",
			cfg: config.Config{
				Database: config.DatabaseConfig{
					Protocol: config.DatabaseSQLite,
				},
				Cache: config.CacheConfig{
					Enabled: false,
					Backend: config.CacheRedis,
				},
			},
			want: map[string]any{
				"version": "1.0.0",
				"os":      "linux",
				"arch":    "amd64",
				"storage": map[string]any{
					"database": "sqlite",
				},
				"experimental": experimental,
			},
		},
		{
			name: "with cache",
			cfg: config.Config{
				Database: config.DatabaseConfig{
					Protocol: config.DatabaseSQLite,
				},
				Cache: config.CacheConfig{
					Enabled: true,
					Backend: config.CacheRedis,
				},
			},
			want: map[string]any{
				"version": "1.0.0",
				"os":      "linux",
				"arch":    "amd64",
				"storage": map[string]any{
					"database": "sqlite",
					"cache":    "redis",
				},
				"experimental": experimental,
			},
		},
		{
			name: "with auth not enabled",
			cfg: config.Config{
				Database: config.DatabaseConfig{
					Protocol: config.DatabaseSQLite,
				},
				Authentication: config.AuthenticationConfig{
					Required: false,
					Methods: config.AuthenticationMethods{
						Token: config.AuthenticationMethod[config.AuthenticationMethodTokenConfig]{
							Enabled: false,
						},
					},
				},
			},
			want: map[string]any{
				"version": "1.0.0",
				"os":      "linux",
				"arch":    "amd64",
				"storage": map[string]any{
					"database": "sqlite",
				},
				"experimental": experimental,
			},
		},
		{
			name: "with auth",
			cfg: config.Config{
				Database: config.DatabaseConfig{
					Protocol: config.DatabaseSQLite,
				},
				Authentication: config.AuthenticationConfig{
					Required: false,
					Methods: config.AuthenticationMethods{
						Token: config.AuthenticationMethod[config.AuthenticationMethodTokenConfig]{
							Enabled: true,
						},
					},
				},
			},
			want: map[string]any{
				"version": "1.0.0",
				"os":      "linux",
				"arch":    "amd64",
				"storage": map[string]any{
					"database": "sqlite",
				},
				"authentication": map[string]any{
					"methods": []any{
						"token",
					},
				},
				"experimental": experimental,
			},
		},
		{
			name: "with audit logfile disabled",
			cfg: config.Config{
				Database: config.DatabaseConfig{
					Protocol: config.DatabaseSQLite,
				},
				Audit: config.AuditConfig{
					Sinks: config.SinksConfig{
						Log: config.LogSinkConfig{
							Enabled: false,
						},
					},
				},
			},
			want: map[string]any{
				"version": "1.0.0",
				"os":      "linux",
				"arch":    "amd64",
				"storage": map[string]any{
					"database": "sqlite",
				},
				"experimental": experimental,
			},
		},
		{
			name: "with audit logfile enabled",
			cfg: config.Config{
				Database: config.DatabaseConfig{
					Protocol: config.DatabaseSQLite,
				},
				Audit: config.AuditConfig{
					Sinks: config.SinksConfig{
						Log: config.LogSinkConfig{
							Enabled: true,
						},
					},
				},
			},
			want: map[string]any{
				"version": "1.0.0",
				"os":      "linux",
				"arch":    "amd64",
				"storage": map[string]any{
					"database": "sqlite",
				},
				"audit": map[string]any{
					"sinks": []any{
						"log",
					},
				},
				"experimental": experimental,
			},
		},
		{
			name: "with audit webhook disabled",
			cfg: config.Config{
				Database: config.DatabaseConfig{
					Protocol: config.DatabaseSQLite,
				},
				Audit: config.AuditConfig{
					Sinks: config.SinksConfig{
						Webhook: config.WebhookSinkConfig{
							Enabled: false,
						},
					},
				},
			},
			want: map[string]any{
				"version": "1.0.0",
				"os":      "linux",
				"arch":    "amd64",
				"storage": map[string]any{
					"database": "sqlite",
				},
				"experimental": experimental,
			},
		},
		{
			name: "with audit webhook enabled",
			cfg: config.Config{
				Database: config.DatabaseConfig{
					Protocol: config.DatabaseSQLite,
				},
				Audit: config.AuditConfig{
					Sinks: config.SinksConfig{
						Webhook: config.WebhookSinkConfig{
							Enabled: true,
						},
					},
				},
			},
			want: map[string]any{
				"version": "1.0.0",
				"os":      "linux",
				"arch":    "amd64",
				"storage": map[string]any{
					"database": "sqlite",
				},
				"audit": map[string]any{
					"sinks": []any{
						"webhook",
					},
				},
				"experimental": experimental,
			},
		},
		{
			name: "with audit logfile and webhook enabled",
			cfg: config.Config{
				Database: config.DatabaseConfig{
					Protocol: config.DatabaseSQLite,
				},
				Audit: config.AuditConfig{
					Sinks: config.SinksConfig{
						Log: config.LogSinkConfig{
							Enabled: true,
						},
						Webhook: config.WebhookSinkConfig{
							Enabled: true,
						},
					},
				},
			},
			want: map[string]any{
				"version": "1.0.0",
				"os":      "linux",
				"arch":    "amd64",
				"storage": map[string]any{
					"database": "sqlite",
				},
				"audit": map[string]any{
					"sinks": []any{
						"log", "webhook",
					},
				},
				"experimental": experimental,
			},
		},
		{
			name: "with tracing not enabled",
			cfg: config.Config{
				Database: config.DatabaseConfig{
					Protocol: config.DatabaseSQLite,
				},
				Tracing: config.TracingConfig{
					Enabled:  false,
					Exporter: config.TracingOTLP,
				},
			},
			want: map[string]any{
				"version": "1.0.0",
				"os":      "linux",
				"arch":    "amd64",
				"storage": map[string]any{
					"database": "sqlite",
				},
				"experimental": experimental,
			},
		},
		{
			name: "with tracing enabled",
			cfg: config.Config{
				Database: config.DatabaseConfig{
					Protocol: config.DatabaseSQLite,
				},
				Tracing: config.TracingConfig{
					Enabled:  true,
					Exporter: config.TracingOTLP,
				},
			},
			want: map[string]any{
				"version": "1.0.0",
				"os":      "linux",
				"arch":    "amd64",
				"storage": map[string]any{
					"database": "sqlite",
				},
				"tracing": map[string]any{
					"exporter": "otlp",
				},
				"experimental": experimental,
			},
		},
		{
			name: "with analytics not enabled",
			cfg: config.Config{
				Database: config.DatabaseConfig{
					Protocol: config.DatabaseSQLite,
				},
				Analytics: config.AnalyticsConfig{
					Storage: config.AnalyticsStorageConfig{
						Clickhouse: config.ClickhouseConfig{
							Enabled: false,
							URL:     "http://localhost:8123",
						},
					},
				},
			},
			want: map[string]any{
				"version": "1.0.0",
				"os":      "linux",
				"arch":    "amd64",
				"storage": map[string]any{
					"database": "sqlite",
				},
				"experimental": experimental,
			},
		},
		{
			name: "with analytics enabled",
			cfg: config.Config{
				Database: config.DatabaseConfig{
					Protocol: config.DatabaseSQLite,
				},
				Analytics: config.AnalyticsConfig{
					Storage: config.AnalyticsStorageConfig{
						Clickhouse: config.ClickhouseConfig{
							Enabled: true,
							URL:     "http://localhost:8123",
						},
					},
				},
			},
			want: map[string]any{
				"version": "1.0.0",
				"os":      "linux",
				"arch":    "amd64",
				"storage": map[string]any{
					"database": "sqlite",
				},
				"analytics": map[string]any{
					"storage": "clickhouse",
				},
				"experimental": experimental,
			},
		},
	}

	for _, tt := range test {
		t.Run(tt.name, func(t *testing.T) {
			var (
				logger        = zaptest.NewLogger(t)
				mockAnalytics = &mockAnalytics{}
			)

			cfg := tt.cfg
			cfg.Meta.TelemetryEnabled = true

			var (
				reporter = &Reporter{
					cfg:    cfg,
					logger: logger,
					client: mockAnalytics,
					info: info.Flipt{
						Version: "1.0.0",
						OS:      "linux",
						Arch:    "amd64",
					},
				}

				in       = bytes.NewBuffer(nil)
				out      = bytes.NewBuffer(nil)
				mockFile = &mockFile{
					Reader: in,
					Writer: out,
				}
			)

			err := reporter.ping(context.Background(), mockFile)
			require.NoError(t, err)

			msg, ok := mockAnalytics.msg.(segment.Track)
			require.True(t, ok)
			assert.Equal(t, "flipt.ping", msg.Event)
			assert.NotEmpty(t, msg.AnonymousId)
			assert.Equal(t, msg.AnonymousId, msg.Properties["uuid"])
			assert.Equal(t, "1.5", msg.Properties["version"])
			assert.Equal(t, tt.want, msg.Properties["flipt"])

			assert.NotEmpty(t, out.String())
		})
	}
}

func TestPing_Existing(t *testing.T) {
	var (
		logger        = zaptest.NewLogger(t)
		mockAnalytics = &mockAnalytics{}

		reporter = &Reporter{
			cfg: config.Config{
				Meta: config.MetaConfig{
					TelemetryEnabled: true,
				},
			},
			logger: logger,
			client: mockAnalytics,
			info: info.Flipt{
				Version: "1.0.0",
				OS:      "linux",
				Arch:    "amd64",
			},
		}

		b, _     = os.ReadFile("./testdata/telemetry_v1.json")
		in       = bytes.NewReader(b)
		out      = bytes.NewBuffer(nil)
		mockFile = &mockFile{
			Reader: in,
			Writer: out,
		}
	)

	err := reporter.ping(context.Background(), mockFile)
	require.NoError(t, err)

	msg, ok := mockAnalytics.msg.(segment.Track)
	require.True(t, ok)
	assert.Equal(t, "flipt.ping", msg.Event)
	assert.Equal(t, "1545d8a8-7a66-4d8d-a158-0a1c576c68a6", msg.AnonymousId)
	assert.Equal(t, "1545d8a8-7a66-4d8d-a158-0a1c576c68a6", msg.Properties["uuid"])
	assert.Equal(t, "1.5", msg.Properties["version"])
	assert.Equal(t, "1.0.0", msg.Properties["flipt"].(map[string]any)["version"])

	assert.NotEmpty(t, out.String())
}

func TestPing_Disabled(t *testing.T) {
	var (
		logger        = zaptest.NewLogger(t)
		mockAnalytics = &mockAnalytics{}

		reporter = &Reporter{
			cfg: config.Config{
				Meta: config.MetaConfig{
					TelemetryEnabled: false,
				},
			},
			logger: logger,
			client: mockAnalytics,
			info: info.Flipt{
				Version: "1.0.0",
				OS:      "linux",
				Arch:    "amd64",
			},
		}
	)

	err := reporter.ping(context.Background(), &mockFile{})
	require.NoError(t, err)

	assert.Nil(t, mockAnalytics.msg)
}

func TestPing_SpecifyStateDir(t *testing.T) {
	var (
		logger = zaptest.NewLogger(t)
		tmpDir = os.TempDir()

		mockAnalytics = &mockAnalytics{}

		reporter = &Reporter{
			cfg: config.Config{
				Meta: config.MetaConfig{
					TelemetryEnabled: true,
					StateDirectory:   tmpDir,
				},
			},
			logger: logger,
			client: mockAnalytics,
			info: info.Flipt{
				Version: "1.0.0",
				OS:      "linux",
				Arch:    "amd64",
			},
		}
	)

	path := filepath.Join(tmpDir, filename)
	defer os.Remove(path)

	err := reporter.report(context.Background())
	require.NoError(t, err)

	msg, ok := mockAnalytics.msg.(segment.Track)
	require.True(t, ok)
	assert.Equal(t, "flipt.ping", msg.Event)
	assert.NotEmpty(t, msg.AnonymousId)
	assert.Equal(t, msg.AnonymousId, msg.Properties["uuid"])
	assert.Equal(t, "1.5", msg.Properties["version"])
	assert.Equal(t, "1.0.0", msg.Properties["flipt"].(map[string]any)["version"])

	b, _ := os.ReadFile(path)
	assert.NotEmpty(t, b)
}
