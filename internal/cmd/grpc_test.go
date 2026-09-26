package cmd

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/internal/server/analytics"
	"go.flipt.io/flipt/internal/server/environments"
	"go.flipt.io/flipt/internal/server/evaluation"
	"go.flipt.io/flipt/internal/storage"
	"go.flipt.io/flipt/rpc/flipt/core"
	rpcevaluation "go.flipt.io/flipt/rpc/flipt/evaluation"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/zap/zaptest"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	grpchealth "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/test/bufconn"
)

func checkEvaluationStatus(t *testing.T, healthSrv *health.Server) grpchealth.HealthCheckResponse_ServingStatus {
	t.Helper()

	resp, err := healthSrv.Check(t.Context(), &grpchealth.HealthCheckRequest{Service: healthServiceEvaluation})
	require.NoError(t, err)
	require.NotNil(t, resp)

	return resp.Status
}

func TestEvaluationHealthAggregator(t *testing.T) {
	tests := []struct {
		name     string
		expected []string
		reports  []string
		want     grpchealth.HealthCheckResponse_ServingStatus
	}{
		{
			name:     "single env ready serves",
			expected: []string{"production"},
			reports:  []string{"production"},
			want:     grpchealth.HealthCheckResponse_SERVING,
		},
		{
			name:     "partial readiness stays unknown",
			expected: []string{"production", "staging"},
			reports:  []string{"production"},
			want:     grpchealth.HealthCheckResponse_UNKNOWN,
		},
		{
			name:     "all expected ready serves",
			expected: []string{"production", "staging"},
			reports:  []string{"production", "staging"},
			want:     grpchealth.HealthCheckResponse_SERVING,
		},
		{
			name:     "unknown key ignored",
			expected: []string{"production"},
			reports:  []string{"unknown-env"},
			want:     grpchealth.HealthCheckResponse_UNKNOWN,
		},
		{
			name:     "unknown then expected serves",
			expected: []string{"production"},
			reports:  []string{"unknown-env", "production"},
			want:     grpchealth.HealthCheckResponse_SERVING,
		},
		{
			name:     "duplicate reports count once",
			expected: []string{"production", "staging"},
			reports:  []string{"production", "production"},
			want:     grpchealth.HealthCheckResponse_UNKNOWN,
		},
		{
			name:     "duplicates then all ready serves",
			expected: []string{"production", "staging"},
			reports:  []string{"production", "production", "staging"},
			want:     grpchealth.HealthCheckResponse_SERVING,
		},
		{
			name:     "no reports stays unknown",
			expected: []string{"production", "staging"},
			reports:  nil,
			want:     grpchealth.HealthCheckResponse_UNKNOWN,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			healthSrv := health.NewServer()
			healthSrv.SetServingStatus(healthServiceEvaluation, grpchealth.HealthCheckResponse_UNKNOWN)

			agg := newEvaluationHealthAggregator(healthSrv, tt.expected)
			for _, env := range tt.reports {
				agg.ReportSnapshotReady(env)
			}

			assert.Equal(t, tt.want, checkEvaluationStatus(t, healthSrv))
		})
	}
}

func TestServerSpansRequired(t *testing.T) {
	tests := []struct {
		name       string
		tracing    bool
		clickhouse bool
		expected   bool
	}{
		{name: "neither", tracing: false, clickhouse: false, expected: false},
		{name: "tracing only", tracing: true, clickhouse: false, expected: true},
		{name: "clickhouse only", tracing: false, clickhouse: true, expected: true},
		{name: "both", tracing: true, clickhouse: true, expected: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.Default()
			cfg.Tracing.Enabled = tt.tracing
			cfg.Analytics.Storage.Clickhouse.Enabled = tt.clickhouse
			assert.Equal(t, tt.expected, serverSpansRequired(cfg))
		})
	}
}

// analyticsCaptureMutator is a test AnalyticsStoreMutator that captures the
// evaluation responses forwarded by AnalyticsSinkSpanExporter in place of the
// ClickHouse sink so tests can assert on them via ch.
type analyticsCaptureMutator struct {
	ch chan []*analytics.EvaluationResponse
}

var _ analytics.AnalyticsStoreMutator = (*analyticsCaptureMutator)(nil)

func (m *analyticsCaptureMutator) IncrementFlagEvaluationCounts(_ context.Context, responses []*analytics.EvaluationResponse) error {
	m.ch <- responses
	return nil
}

func (m *analyticsCaptureMutator) Close() error { return nil }

// TestAnalyticsEventsRequireServerStatsHandler pins the mechanism behind
// serverSpansRequired: evaluation events are span events, so without the OTel
// server stats handler no span exists in the request context and the
// analytics exporter never receives them (no ClickHouse rows).
// It exercises the real evaluation.Server.Boolean over bufconn so attribute
// renames and event-shape changes in production break here instead of
// silently drifting from a hand-rolled fake.
func TestAnalyticsEventsRequireServerStatsHandler(t *testing.T) {
	const bufSize = 1024 * 1024

	run := func(t *testing.T, withStatsHandler bool) []*analytics.EvaluationResponse {
		t.Helper()

		mut := &analyticsCaptureMutator{ch: make(chan []*analytics.EvaluationResponse, 10)}

		tp := tracesdk.NewTracerProvider()
		t.Cleanup(func() { _ = tp.Shutdown(context.Background()) }) //nolint:usetesting
		tp.RegisterSpanProcessor(tracesdk.NewSimpleSpanProcessor(
			analytics.NewAnalyticsSinkSpanExporter(zaptest.NewLogger(t), mut),
		))

		prev := otel.GetTracerProvider()
		otel.SetTracerProvider(tp)
		t.Cleanup(func() { otel.SetTracerProvider(prev) })

		var (
			envStore    = evaluation.NewMockEnvironmentStore(t)
			environment = environments.NewMockEnvironment(t)
			store       = storage.NewMockReadOnlyStore(t)
		)

		environment.On("Key").Return("test-environment")
		envStore.On("Get", mock.Anything, mock.Anything).Return(environment, nil)
		environment.On("EvaluationStore").Return(store, nil)

		store.On("GetFlag", mock.Anything, storage.NewResource("test-namespace", "test-flag")).Return(&core.Flag{
			Key:     "test-flag",
			Enabled: true,
			Type:    core.FlagType_BOOLEAN_FLAG_TYPE,
		}, nil)

		// 100% threshold rollout always matches, deterministically producing
		// reason=match, enabled=true without segment fixtures or sampling flake.
		store.On("GetEvaluationRollouts", mock.Anything, storage.NewResource("test-namespace", "test-flag")).Return([]*storage.EvaluationRollout{
			{
				NamespaceKey: "test-namespace",
				Rank:         1,
				RolloutType:  core.RolloutType_THRESHOLD_ROLLOUT_TYPE,
				Threshold: &storage.RolloutThreshold{
					Percentage: 100,
					Value:      true,
				},
			},
		}, nil)

		var opts []grpc.ServerOption
		if withStatsHandler {
			opts = append(opts, grpc.StatsHandler(otelgrpc.NewServerHandler()))
		}
		srv := grpc.NewServer(opts...)
		evaluation.New(zaptest.NewLogger(t), envStore, evaluation.WithTracing(true)).RegisterGRPC(srv)

		lis := bufconn.Listen(bufSize)
		go func() { _ = srv.Serve(lis) }()
		t.Cleanup(srv.Stop)

		conn, err := grpc.NewClient(
			"passthrough:///bufnet",
			grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
				return lis.DialContext(ctx)
			}),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		require.NoError(t, err)
		t.Cleanup(func() { _ = conn.Close() })

		_, err = rpcevaluation.NewEvaluationServiceClient(conn).Boolean(t.Context(), &rpcevaluation.EvaluationRequest{
			FlagKey:        "test-flag",
			NamespaceKey:   "test-namespace",
			EnvironmentKey: "test-environment",
			EntityId:       "test-entity",
		})
		require.NoError(t, err)

		select {
		case got := <-mut.ch:
			return got
		case <-time.After(2 * time.Second):
			return nil
		}
	}

	t.Run("with stats handler events reach analytics", func(t *testing.T) {
		got := run(t, true)
		require.Len(t, got, 1)
		require.NotNil(t, got[0].EvaluationValue)
		assert.Equal(t, "true", *got[0].EvaluationValue)
		require.NotNil(t, got[0].Match)
		assert.True(t, *got[0].Match)
		assert.Equal(t, "test-flag", got[0].FlagKey)
		assert.Equal(t, "test-environment", got[0].EnvironmentKey)
		assert.Equal(t, "test-namespace", got[0].NamespaceKey)
	})

	t.Run("without stats handler events are dropped", func(t *testing.T) {
		assert.Empty(t, run(t, false))
	})
}

func TestEvaluationHealthAggregator_StickyServing(t *testing.T) {
	tests := []struct {
		name         string
		expected     []string
		reachReady   []string
		extraReports []string
	}{
		{
			name:         "duplicate report stays serving",
			expected:     []string{"production"},
			reachReady:   []string{"production"},
			extraReports: []string{"production"},
		},
		{
			name:         "unknown key stays serving",
			expected:     []string{"production"},
			reachReady:   []string{"production"},
			extraReports: []string{"unknown-env"},
		},
		{
			name:         "duplicate and unknown stay serving",
			expected:     []string{"production", "staging"},
			reachReady:   []string{"production", "staging"},
			extraReports: []string{"production", "unknown-env"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			healthSrv := health.NewServer()
			healthSrv.SetServingStatus(healthServiceEvaluation, grpchealth.HealthCheckResponse_UNKNOWN)

			agg := newEvaluationHealthAggregator(healthSrv, tt.expected)
			for _, env := range tt.reachReady {
				agg.ReportSnapshotReady(env)
			}
			require.Equal(t, grpchealth.HealthCheckResponse_SERVING, checkEvaluationStatus(t, healthSrv))

			for _, env := range tt.extraReports {
				agg.ReportSnapshotReady(env)
			}
			assert.Equal(t, grpchealth.HealthCheckResponse_SERVING, checkEvaluationStatus(t, healthSrv))
		})
	}
}
