package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/health"
	grpchealth "google.golang.org/grpc/health/grpc_health_v1"
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
