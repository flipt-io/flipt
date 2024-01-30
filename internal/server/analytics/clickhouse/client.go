package clickhouse

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"go.uber.org/zap"
)

// Step defines the value and interval name of the time windows
// for the Clickhouse query.
type Step struct {
	intervalValue int
	intervalStep  string
}

const (
	counterAnalyticsTable = "flipt_counter_analytics"
	counterAnalyticsName  = "flag_evaluation_count"
	timeFormat            = "2006-01-02 15:04:05"
)

type Client struct {
	conn *sql.DB
}

// New constructs a new clickhouse client that conforms to the analytics.Client contract.
func New(logger *zap.Logger, connectionString string) (*Client, error) {
	conn, err := connect(connectionString)
	if err != nil {
		return nil, err
	}

	return &Client{conn: conn}, nil
}

func connect(connectionString string) (*sql.DB, error) {
	conn := clickhouse.OpenDB(&clickhouse.Options{
		Addr:     []string{connectionString},
		Protocol: clickhouse.HTTP,
	})

	if err := conn.Ping(); err != nil {
		return nil, err
	}

	return conn, nil
}

func (c *Client) GetFlagEvaluationsCount(ctx context.Context, flagKey string, from time.Duration) ([]string, []float32, error) {
	step := getStepFromDuration(from)

	rows, err := c.conn.QueryContext(ctx, fmt.Sprintf(`SELECT sum(value) AS value, toStartOfInterval(time, INTERVAL %d %s) AS timestamp
		FROM %s WHERE flag_key = ? AND time >= now() - toIntervalMinute(%f) GROUP BY timestamp ORDER BY timestamp`,
		step.intervalValue,
		step.intervalStep,
		counterAnalyticsTable,
		from.Seconds()),
		flagKey,
	)
	if err != nil {
		return nil, nil, err
	}

	defer func() {
		_ = rows.Close()
	}()

	var (
		timestamps = make([]string, 0)
		values     = make([]float32, 0)
	)
	for rows.Next() {
		var (
			timestamp string
			value     int
		)
		if err := rows.Scan(&value, &timestamp); err != nil {
			return nil, nil, err
		}

		timestamps = append(timestamps, timestamp)
		values = append(values, float32(value))
	}

	return timestamps, values, nil
}

// getStepFromDuration is a utility function that translates the duration passed in from the client
// to determine the interval steps we should use for the Clickhouse query.
func getStepFromDuration(from time.Duration) *Step {
	if from <= time.Hour {
		return &Step{
			intervalValue: 15,
			intervalStep:  "SECOND",
		}
	}

	if from > time.Hour && from <= 4*time.Hour {
		return &Step{
			intervalValue: 1,
			intervalStep:  "MINUTE",
		}
	}

	return &Step{
		intervalValue: 15,
		intervalStep:  "MINUTE",
	}
}

// IncrementFlagEvaluation inserts a row into Clickhouse that corresponds to a time when a flag was evaluated.
// This acts as a "prometheus-like" counter metric.
func (c *Client) IncrementFlagEvaluation(ctx context.Context, flagKey string) error {
	_, err := c.conn.ExecContext(ctx, fmt.Sprintf("INSERT INTO %s VALUES (toDateTime(?),?,?,?)", counterAnalyticsTable), time.Now().Format(timeFormat), counterAnalyticsName, flagKey, 1)

	return err
}
