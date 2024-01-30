package clickhouse

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/golang-migrate/migrate/v4"
	clickhouseMigrate "github.com/golang-migrate/migrate/v4/database/clickhouse"
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
	timeFormat            = "2006-01-02 15:04:05"
)

type Client struct {
	conn *sql.DB
}

// New constructs a new clickhouse client that conforms to the analytics.Client contract.
func New(logger *zap.Logger, connectionString string) (*Client, error) {
	var (
		conn          *sql.DB
		clickhouseErr error
	)

	dbOnce.Do(func() {
		err := runMigrations(logger, connectionString)
		if err != nil {
			clickhouseErr = err
			return
		}

		connection, err := connect(connectionString)
		if err != nil {
			clickhouseErr = err
			return
		}

		conn = connection
	})

	if clickhouseErr != nil {
		return nil, clickhouseErr
	}

	return &Client{conn: conn}, nil
}

func runMigrations(logger *zap.Logger, connectionString string) error {
	db := clickhouse.OpenDB(&clickhouse.Options{
		Addr: []string{connectionString},
	})

	driver, err := clickhouseMigrate.WithInstance(db, &clickhouseMigrate.Config{
		MigrationsTableEngine: "MergeTree",
	})
	if err != nil {
		logger.Error("error creating driver for clickhouse migrations", zap.Error(err))
		return err
	}

	m, err := migrate.NewWithDatabaseInstance("file://sql", "clickhouse", driver)
	if err != nil {
		logger.Error("error creating clickhouse DB instance for migrations", zap.Error(err))
		return err
	}

	if err := m.Up(); err != nil && errors.Is(err, migrate.ErrNoChange) {
		logger.Error("error running migrations on clickhouse", zap.Error(err))
		return err
	}

	return nil
}

func connect(connectionString string) (*sql.DB, error) {
	conn := clickhouse.OpenDB(&clickhouse.Options{
		Addr: []string{connectionString},
	})

	if err := conn.Ping(); err != nil {
		return nil, err
	}

	return conn, nil
}

func (c *Client) GetFlagEvaluationsCount(ctx context.Context, namespaceKey, flagKey string, from time.Duration) ([]string, []float32, error) {
	step := getStepFromDuration(from)

	rows, err := c.conn.QueryContext(ctx, fmt.Sprintf(`SELECT sum(value) AS value, toStartOfInterval(timestamp, INTERVAL %d %s) AS timestamp
		FROM %s WHERE namespaceKey = ? AND flag_key = ? AND timestamp >= now() - toIntervalMinute(%f) GROUP BY timestamp ORDER BY timestamp`,
		step.intervalValue,
		step.intervalStep,
		counterAnalyticsTable,
		from.Seconds()),
		namespaceKey,
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

// IncrementFlagEvaluation inserts a row into Clickhouse that corresponds to a time when a flag was evaluated.
// This acts as a "prometheus-like" counter metric.
func (c *Client) IncrementFlagEvaluation(ctx context.Context, namespaceKey, flagKey string) error {
	_, err := c.conn.ExecContext(ctx, fmt.Sprintf("INSERT INTO %s VALUES (toDateTime(?),?,?,?,?)", counterAnalyticsTable), time.Now().Format(timeFormat), counterAnalyticsName, namespaceKey, flagKey, 1)

	return err
}
