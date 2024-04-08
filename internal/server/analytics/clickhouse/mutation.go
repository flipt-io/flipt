package clickhouse

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.flipt.io/flipt/internal/server/analytics"
)

// IncrementFlagEvaluationCounts does a batch insert of flag values from evaluations.
func (c *Client) IncrementFlagEvaluationCounts(ctx context.Context, responses []*analytics.EvaluationResponse) error {
	if len(responses) < 1 {
		return nil
	}

	valuePlaceHolders := make([]string, 0, len(responses))
	valueArgs := make([]interface{}, 0, len(responses)*10)

	for _, response := range responses {
		valuePlaceHolders = append(valuePlaceHolders, "(toDateTime(?, 'UTC'),?,?,?,?,?,?,?,?,?)")
		valueArgs = append(valueArgs, response.Timestamp.Format(time.DateTime))
		valueArgs = append(valueArgs, counterAnalyticsName)
		valueArgs = append(valueArgs, response.NamespaceKey)
		valueArgs = append(valueArgs, response.FlagKey)
		valueArgs = append(valueArgs, response.FlagType)
		valueArgs = append(valueArgs, response.Reason)
		valueArgs = append(valueArgs, response.Match)
		valueArgs = append(valueArgs, response.EvaluationValue)
		valueArgs = append(valueArgs, response.EntityId)
		valueArgs = append(valueArgs, 1)
	}

	//nolint:gosec
	stmt := fmt.Sprintf("INSERT INTO %s VALUES %s", counterAnalyticsTable, strings.Join(valuePlaceHolders, ","))

	_, err := c.Conn.ExecContext(ctx, stmt, valueArgs...)
	return err
}
