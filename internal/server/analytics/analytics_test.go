package analytics

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/rpc/v2/analytics"
	"go.uber.org/zap/zaptest"
)

func TestGetStepFromDuration(t *testing.T) {
	cases := []struct {
		name     string
		duration time.Duration
		want     int
	}{
		{
			name:     "1 hour duration",
			duration: time.Hour,
			want:     1,
		},
		{
			name:     "4 hour duration",
			duration: 4 * time.Hour,
			want:     5,
		},
		{
			name:     "12 hour duration",
			duration: 12 * time.Hour,
			want:     15,
		},
		{
			name:     "24 hour duration",
			duration: 24 * time.Hour,
			want:     30,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			step := getStepFromDuration(tt.duration)
			assert.Equal(t, tt.want, step)
		})
	}
}

func TestGetFlagEvaluationsCountWithInvalidInput(t *testing.T) {
	logger := zaptest.NewLogger(t)
	client := NewMockClient(t)
	service := New(logger, client)

	_, err := service.GetFlagEvaluationsCount(context.Background(), &analytics.GetFlagEvaluationsCountRequest{
		EnvironmentKey: "default",
		NamespaceKey:   "bar",
		FlagKey:        "foo",
	})
	require.Error(t, err)
	_, err = service.GetFlagEvaluationsCount(context.Background(), &analytics.GetFlagEvaluationsCountRequest{
		EnvironmentKey: "default",
		NamespaceKey:   "bar",
		FlagKey:        "foo",
		From:           time.Now().Format(time.DateTime),
	})
	require.Error(t, err)
}

func TestGetFlagEvaluationsCountClientError(t *testing.T) {
	logger := zaptest.NewLogger(t)
	client := NewMockClient(t)

	ctx := context.Background()

	client.EXPECT().GetFlagEvaluationsCount(ctx, &FlagEvaluationsCountRequest{
		EnvironmentKey: "default",
		NamespaceKey:   "bar",
		FlagKey:        "foo",
		From:           time.Date(2022, 6, 9, 11, 0, 0, 0, time.UTC),
		To:             time.Date(2022, 6, 9, 11, 30, 0, 0, time.UTC),
		StepMinutes:    1,
	}).Return(nil, nil, errors.New("client error"))

	t.Run("old date format", func(t *testing.T) {
		service := New(logger, client)
		from := "2022-06-09 11:00:00"
		to := "2022-06-09 11:30:00"
		_, err := service.GetFlagEvaluationsCount(ctx, &analytics.GetFlagEvaluationsCountRequest{
			EnvironmentKey: "default",
			NamespaceKey:   "bar",
			FlagKey:        "foo",
			From:           from,
			To:             to,
		})

		require.Error(t, err)
		require.ErrorContains(t, err, "client error")
	})

	t.Run("rfc 3339 date format", func(t *testing.T) {
		service := New(logger, client)
		from := "2022-06-09T11:00:00Z"
		to := "2022-06-09T11:30:00.000Z"
		_, err := service.GetFlagEvaluationsCount(ctx, &analytics.GetFlagEvaluationsCountRequest{
			EnvironmentKey: "default",
			NamespaceKey:   "bar",
			FlagKey:        "foo",
			From:           from,
			To:             to,
		})

		require.Error(t, err)
		require.ErrorContains(t, err, "client error")
	})
}
