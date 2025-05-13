package clickhouse

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	panalytics "go.flipt.io/flipt/internal/server/analytics"
	analyticstesting "go.flipt.io/flipt/internal/server/analytics/testing"
	"go.flipt.io/flipt/rpc/v2/analytics"
	"go.uber.org/zap"
)

type AnalyticsDBTestSuite struct {
	suite.Suite
	client  *Client
	service *panalytics.Server
}

func TestAnalyticsDBTestSuite(t *testing.T) {
	if os.Getenv("FLIPT_ANALYTICS_DATABASE_PROTOCOL") == "" {
		t.Skip("please provide an analytics database protocol to run tests")
	}
	suite.Run(t, new(AnalyticsDBTestSuite))
}

func (a *AnalyticsDBTestSuite) SetupSuite() {
	setup := func() error {
		db, err := analyticstesting.Open()
		if err != nil {
			return err
		}

		a.client = &Client{
			Conn: db.DB,
		}

		logger := zap.NewNop()
		a.service = panalytics.New(
			logger,
			a.client,
		)

		return nil
	}

	a.Require().NoError(setup())
}

func (a *AnalyticsDBTestSuite) TestAnalyticsMutationAndQuery() {
	t := a.T()

	now := time.Now().UTC()
	for range 5 {
		err := a.client.IncrementFlagEvaluationCounts(context.TODO(), []*panalytics.EvaluationResponse{
			{
				EnvironmentKey: "default",
				NamespaceKey:   "default",
				FlagKey:        "flag1",
				Reason:         "MATCH_EVALUATION_REASON",
				Timestamp:      now,
			},
			{
				EnvironmentKey: "default",
				NamespaceKey:   "default",
				FlagKey:        "flag1",
				Reason:         "MATCH_EVALUATION_REASON",
				Timestamp:      now,
			},
			{
				EnvironmentKey: "default",
				NamespaceKey:   "default",
				FlagKey:        "flag1",
				Reason:         "MATCH_EVALUATION_REASON",
				Timestamp:      now,
			},
			{
				EnvironmentKey: "default",
				NamespaceKey:   "default",
				FlagKey:        "flag1",
				Reason:         "MATCH_EVALUATION_REASON",
				Timestamp:      now,
			},
			{
				EnvironmentKey: "default",
				NamespaceKey:   "default",
				FlagKey:        "flag1",
				Reason:         "MATCH_EVALUATION_REASON",
				Timestamp:      now,
			},
		})
		require.NoError(t, err)
	}

	res, err := a.service.GetFlagEvaluationsCount(context.TODO(), &analytics.GetFlagEvaluationsCountRequest{
		EnvironmentKey: "default",
		NamespaceKey:   "default",
		FlagKey:        "flag1",
		From:           now.Add(-time.Hour).Format(time.DateTime),
		To:             now.Format(time.DateTime),
	})
	require.NoError(t, err)

	values := res.Values
	assert.Len(t, values, 1)
	assert.InDelta(t, 25, values[0], 0.0)
}
