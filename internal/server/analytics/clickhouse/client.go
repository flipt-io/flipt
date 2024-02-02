package clickhouse

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"go.flipt.io/flipt/internal/config"
	fliptsql "go.flipt.io/flipt/internal/storage/sql"
	"go.flipt.io/flipt/rpc/flipt/analytics"
	"go.uber.org/zap"
)

// Step defines the value and interval name of the time windows
// for the Clickhouse query.
type Step struct {
	intervalValue int
	intervalStep  string
}

var dbOnce sync.Once

const (
	counterAnalyticsTable = "flipt_counter_analytics"
	counterAnalyticsName  = "flag_evaluation_count"
)

type Client struct {
	Conn *sql.DB
}

// New constructs a new clickhouse client that conforms to the analytics.Client contract.
func New(logger *zap.Logger, cfg *config.Config, forceMigrate bool) (*Client, error) {
	var (
		conn          *sql.DB
		clickhouseErr error
	)

	dbOnce.Do(func() {
		err := runMigrations(logger, cfg, forceMigrate)
		if err != nil {
			clickhouseErr = err
			return
		}

		connection, err := connect(cfg.Analytics.Clickhouse.URL)
		if err != nil {
			clickhouseErr = err
			return
		}

		conn = connection
	})

	if clickhouseErr != nil {
		return nil, clickhouseErr
	}

	return &Client{Conn: conn}, nil
}

// runMigrations will run migrations for clickhouse if enabled from the client.
func runMigrations(logger *zap.Logger, cfg *config.Config, forceMigrate bool) error {
	m, err := fliptsql.NewMigrator(*cfg, logger)
	if err != nil {
		return err
	}

	if err := m.Up(forceMigrate); err != nil {
		return err
	}

	return nil
}

func connect(connectionString string) (*sql.DB, error) {
	conn := clickhouse.OpenDB(&clickhouse.Options{
		Addr: []string{connectionString},
		Auth: clickhouse.Auth{
			Database: "flipt_analytics",
		},
	})

	if err := conn.Ping(); err != nil {
		return nil, err
	}

	return conn, nil
}

func (c *Client) GetFlagEvaluationsCount(ctx context.Context, req *analytics.GetFlagEvaluationsCountRequest) ([]string, []float32, error) {
	fromTime, err := time.Parse(time.DateTime, req.From)
	if err != nil {
		return nil, nil, err
	}

	toTime, err := time.Parse(time.DateTime, req.To)
	if err != nil {
		return nil, nil, err
	}

	duration := toTime.Sub(fromTime)

	step := getStepFromDuration(duration)

	rows, err := c.Conn.QueryContext(ctx, fmt.Sprintf(`
		SELECT
			sum(value) AS value,
			toStartOfInterval(timestamp, INTERVAL %d %s) AS timestamp
		FROM %s WHERE namespace_key = ? AND flag_key = ? AND timestamp >= '%s' AND timestamp < '%s'
		GROUP BY timestamp
		ORDER BY timestamp ASC
		`,
		step.intervalValue,
		step.intervalStep,
		counterAnalyticsTable,
		fromTime.Format(time.DateTime),
		toTime.Format(time.DateTime),
	),
		req.NamespaceKey,
		req.FlagKey,
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

	if err := rows.Err(); err != nil {
		return nil, nil, err
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
