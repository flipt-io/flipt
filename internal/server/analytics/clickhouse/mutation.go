package clickhouse

import (
	"context"
	"fmt"
	"time"

	"go.flipt.io/flipt/rpc/flipt/evaluation"
)

func (c *Client) IncrementVariantFlagEvaluation(ctx context.Context, namespaceKey string, resp *evaluation.VariantEvaluationResponse) error {
	_, err := c.Conn.ExecContext(ctx, fmt.Sprintf("INSERT INTO %s VALUES (toDateTime(?),?,?,?,?,?,?)", counterAnalyticsTable), time.Now().Format(time.DateTime), counterAnalyticsName, namespaceKey, resp.FlagKey, resp.Reason.String(), resp.Match, 1)

	return err
}

func (c *Client) IncrementBooleanFlagEvaluation(ctx context.Context, namespaceKey string, resp *evaluation.BooleanEvaluationResponse) error {
	return nil
}
