package clickhouse

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/internal/server/analytics"
	analyticsstorage "go.flipt.io/flipt/internal/storage/analytics"
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
	counterAnalyticsTable           = "flipt_counter_analytics_v2"
	counterAggregatedAnalyticsTable = "flipt_counter_aggregated_analytics_v2"
	counterAnalyticsName            = "flag_evaluation_count"
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

		connection, err := connect(cfg.Analytics.Storage.Clickhouse)
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
	m, err := analyticsstorage.NewMigrator(*cfg, logger)
	if err != nil {
		return err
	}

	if err := m.Up(forceMigrate); err != nil {
		return err
	}

	return nil
}

func connect(clickhouseConfig config.ClickhouseConfig) (*sql.DB, error) {
	options, err := clickhouseConfig.Options()
	if err != nil {
		return nil, err
	}

	conn := clickhouse.OpenDB(options)

	if err := conn.Ping(); err != nil {
		return nil, err
	}

	return conn, nil
}

func (c *Client) GetFlagEvaluationsCount(ctx context.Context, req *analytics.FlagEvaluationsCountRequest) ([]string, []float32, error) {
	step := &Step{
		intervalValue: req.StepMinutes,
		intervalStep:  "MINUTE",
	}

	rows, err := c.Conn.QueryContext(ctx, fmt.Sprintf(`
		SELECT
			sum(value) AS value,
			toStartOfInterval(timestamp, INTERVAL %[4]d %[5]s) AS timestamp
		FROM %[1]s
		WHERE
			environment_key = ? AND namespace_key = ? AND flag_key = ? AND
			timestamp >= toStartOfInterval(toDateTime('%[2]s', 'UTC'),  INTERVAL %[4]d %[5]s) AND
			timestamp < timestamp_add(toStartOfInterval(toDateTime('%[3]s', 'UTC'), INTERVAL %[4]d %[5]s), INTERVAL %[4]d %[5]s)
		GROUP BY timestamp
		ORDER BY timestamp ASC
		WITH FILL FROM toStartOfInterval(toDateTime('%[2]s', 'UTC'),  INTERVAL %[4]d %[5]s) TO timestamp_add(toStartOfInterval(toDateTime('%[3]s', 'UTC'), INTERVAL %[4]d %[5]s), INTERVAL %[4]d %[5]s) STEP INTERVAL %[4]d %[5]s
	`,
		counterAggregatedAnalyticsTable,
		req.From.UTC().Format(time.DateTime),
		req.To.UTC().Format(time.DateTime),
		step.intervalValue,
		step.intervalStep,
	),
		req.EnvironmentKey,
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

// GetBatchFlagEvaluationsCount retrieves evaluation counts for multiple flags in a single query
func (c *Client) GetBatchFlagEvaluationsCount(ctx context.Context, req *analytics.BatchFlagEvaluationsCountRequest) (map[string]analytics.FlagEvaluationData, error) {
	if len(req.FlagKeys) == 0 {
		return map[string]analytics.FlagEvaluationData{}, nil
	}

	step := &Step{
		intervalValue: req.StepMinutes,
		intervalStep:  "MINUTE",
	}

	// Create a parameterized query with placeholders for the flag keys
	placeholders := make([]string, len(req.FlagKeys))
	args := make([]interface{}, len(req.FlagKeys)+2) // +2 for environment and namespace keys

	args[0] = req.EnvironmentKey
	args[1] = req.NamespaceKey

	for i, key := range req.FlagKeys {
		placeholders[i] = "?"
		args[i+2] = key
	}

	// This query gets data for each flag individually
	query := fmt.Sprintf(`
		SELECT
			flag_key,
			sum(value) AS value,
			toStartOfInterval(timestamp, INTERVAL %[4]d %[5]s) AS timestamp
		FROM %[1]s
		WHERE
			environment_key = ? AND 
			namespace_key = ? AND 
			flag_key IN (%[6]s) AND
			timestamp >= toStartOfInterval(toDateTime('%[2]s', 'UTC'), INTERVAL %[4]d %[5]s) AND
			timestamp < timestamp_add(toStartOfInterval(toDateTime('%[3]s', 'UTC'), INTERVAL %[4]d %[5]s), INTERVAL %[4]d %[5]s)
		GROUP BY flag_key, timestamp
		ORDER BY flag_key, timestamp ASC
		WITH FILL FROM toStartOfInterval(toDateTime('%[2]s', 'UTC'), INTERVAL %[4]d %[5]s) TO timestamp_add(toStartOfInterval(toDateTime('%[3]s', 'UTC'), INTERVAL %[4]d %[5]s), INTERVAL %[4]d %[5]s) STEP INTERVAL %[4]d %[5]s
	`,
		counterAggregatedAnalyticsTable,
		req.From.UTC().Format(time.DateTime),
		req.To.UTC().Format(time.DateTime),
		step.intervalValue,
		step.intervalStep,
		strings.Join(placeholders, ","),
	)

	rows, err := c.Conn.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = rows.Close()
	}()

	// Map to store results by flag key
	result := make(map[string]map[string]float32) // flagKey -> timestamp -> value
	timestamps := make(map[string][]string)       // flagKey -> []timestamps
	orderedTimestamps := make(map[string]struct{})

	for rows.Next() {
		var (
			flagKey   string
			value     int
			timestamp string
		)

		if err := rows.Scan(&flagKey, &value, &timestamp); err != nil {
			return nil, err
		}

		// Initialize maps for this flag if not exist
		if _, ok := result[flagKey]; !ok {
			result[flagKey] = make(map[string]float32)
			timestamps[flagKey] = []string{}
		}

		result[flagKey][timestamp] = float32(value)

		// Keep track of all timestamps to ensure we have the same across all flags
		if _, exists := orderedTimestamps[timestamp]; !exists {
			orderedTimestamps[timestamp] = struct{}{}
			timestamps[flagKey] = append(timestamps[flagKey], timestamp)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Initialize result map with empty arrays for each requested flag
	flagData := make(map[string]analytics.FlagEvaluationData)
	for _, flagKey := range req.FlagKeys {
		// Skip flags with no data
		if _, ok := result[flagKey]; !ok {
			continue
		}

		ts := timestamps[flagKey]
		values := make([]float32, len(ts))

		for i, timestamp := range ts {
			values[i] = result[flagKey][timestamp]
		}

		// If limit is specified and we have more data points than the limit, downsample
		if req.Limit > 0 && len(ts) > req.Limit {
			downTs, downValues, err := downsampleData(ts, values, req.Limit)
			if err != nil {
				return nil, err
			}
			ts = downTs
			values = downValues
		}

		flagData[flagKey] = analytics.FlagEvaluationData{
			Timestamps: ts,
			Values:     values,
		}
	}

	return flagData, nil
}

// downsampleData reduces the number of data points to match the requested limit
func downsampleData(timestamps []string, values []float32, limit int) ([]string, []float32, error) {
	if len(timestamps) <= limit {
		return timestamps, values, nil
	}

	step := len(timestamps) / limit
	if step < 1 {
		step = 1
	}

	newTimestamps := make([]string, 0, limit)
	newValues := make([]float32, 0, limit)

	for i := 0; i < len(timestamps); i += step {
		newTimestamps = append(newTimestamps, timestamps[i])
		newValues = append(newValues, values[i])

		if len(newTimestamps) >= limit {
			break
		}
	}

	return newTimestamps, newValues, nil
}

// Close will close the DB connection.
func (c *Client) Close() error {
	if c != nil && c.Conn != nil {
		return c.Conn.Close()
	}

	return nil
}

func (c *Client) String() string {
	return "clickhouse"
}
