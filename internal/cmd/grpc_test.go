package cmd

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/internal/info"
	"go.uber.org/zap/zaptest"
)

func TestGetTraceExporter(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *config.Config
		wantErr error
	}{
		{
			name: "Jaeger",
			cfg: &config.Config{
				Tracing: config.TracingConfig{
					Exporter: config.TracingJaeger,
					Jaeger: config.JaegerTracingConfig{
						Host: "localhost",
						Port: 6831,
					},
				},
			},
		},
		{
			name: "Zipkin",
			cfg: &config.Config{
				Tracing: config.TracingConfig{
					Exporter: config.TracingZipkin,
					Zipkin: config.ZipkinTracingConfig{
						Endpoint: "http://localhost:9411/api/v2/spans",
					},
				},
			},
		},
		{
			name: "OTLP HTTP",
			cfg: &config.Config{
				Tracing: config.TracingConfig{
					Exporter: config.TracingOTLP,
					OTLP: config.OTLPTracingConfig{
						Endpoint: "http://localhost:4317",
						Headers:  map[string]string{"key": "value"},
					},
				},
			},
		},
		{
			name: "OTLP HTTPS",
			cfg: &config.Config{
				Tracing: config.TracingConfig{
					Exporter: config.TracingOTLP,
					OTLP: config.OTLPTracingConfig{
						Endpoint: "https://localhost:4317",
						Headers:  map[string]string{"key": "value"},
					},
				},
			},
		},
		{
			name: "OTLP GRPC",
			cfg: &config.Config{
				Tracing: config.TracingConfig{
					Exporter: config.TracingOTLP,
					OTLP: config.OTLPTracingConfig{
						Endpoint: "grpc://localhost:4317",
						Headers:  map[string]string{"key": "value"},
					},
				},
			},
		},
		{
			name: "OTLP default",
			cfg: &config.Config{
				Tracing: config.TracingConfig{
					Exporter: config.TracingOTLP,
					OTLP: config.OTLPTracingConfig{
						Endpoint: "localhost:4317",
						Headers:  map[string]string{"key": "value"},
					},
				},
			},
		},
		{
			name: "Unsupported Exporter",
			cfg: &config.Config{
				Tracing: config.TracingConfig{},
			},
			wantErr: errors.New("unsupported tracing exporter: "),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			traceExpOnce = sync.Once{}
			exp, expFunc, err := getTraceExporter(context.Background(), tt.cfg)
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
				return
			}
			t.Cleanup(func() {
				err := expFunc(context.Background())
				assert.NoError(t, err)
			})
			assert.NoError(t, err)
			assert.NotNil(t, exp)
			assert.NotNil(t, expFunc)

		})
	}
}

func TestNewGRPCServer(t *testing.T) {
	tmp := t.TempDir()
	cfg := &config.Config{}
	cfg.Database.URL = fmt.Sprintf("file:%s", filepath.Join(tmp, "flipt.db"))
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	s, err := NewGRPCServer(ctx, zaptest.NewLogger(t), cfg, info.Flipt{}, false)
	assert.NoError(t, err)
	t.Cleanup(func() {
		err := s.Shutdown(ctx)
		assert.NoError(t, err)
	})
	assert.NotEmpty(t, s.Server.GetServiceInfo())
}
