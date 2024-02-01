package analytics_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.flipt.io/flipt/internal/server/analytics/clickhouse"
	analyticstesting "go.flipt.io/flipt/internal/server/analytics/testing"
	"go.flipt.io/flipt/rpc/flipt/analytics"
)

type AnalyticsDBTestSuite struct {
	suite.Suite
	client *clickhouse.Client
}

func TestAnalyticsDBTestSuite(t *testing.T) {
	suite.Run(t, new(AnalyticsDBTestSuite))
}

func (a *AnalyticsDBTestSuite) SetupSuite() {
	setup := func() error {
		db, err := analyticstesting.Open()
		if err != nil {
			return err
		}

		c := &clickhouse.Client{
			Conn: db.DB,
		}

		a.client = c

		return nil
	}

	a.Require().NoError(setup())
}

func (a *AnalyticsDBTestSuite) TestSomething() {
	t := a.T()

	for i := 0; i < 5; i++ {
		err := a.client.IncrementFlagEvaluation(context.TODO(), "default", "flag1")
		require.Nil(t, err)
	}

	now := time.Now()
	_, values, err := a.client.GetFlagEvaluationsCount(context.TODO(), &analytics.GetFlagEvaluationsCountRequest{
		NamespaceKey: "default",
		FlagKey:      "flag1",
		From:         now.Add(-time.Hour).Format(time.DateTime),
		To:           now.Format(time.DateTime),
	})
	require.Nil(t, err)

	assert.Len(t, values, 1)
	assert.Equal(t, values[0], float32(5))
}
