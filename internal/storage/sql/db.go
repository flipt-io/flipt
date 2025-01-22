package sql

import (
	"database/sql"
	"errors"

	"github.com/ClickHouse/clickhouse-go/v2"

	"go.flipt.io/flipt/internal/config"
)

// openAnalytics is a convenience function of providing a database.sql instance for
// an analytics database.
func openAnalytics(cfg config.Config) (*sql.DB, Driver, error) {
	if cfg.Analytics.Storage.Clickhouse.Enabled {
		clickhouseOptions, err := cfg.Analytics.Storage.Clickhouse.Options()
		if err != nil {
			return nil, "", err
		}

		db := clickhouse.OpenDB(clickhouseOptions)

		return db, Clickhouse, nil
	}

	return nil, "", errors.New("no analytics db provided")
}

var (
	stringToDriver = map[string]Driver{
		"clickhouse": Clickhouse,
	}
)

// Driver represents a database driver
type Driver string

func (d Driver) String() string {
	return string(d)
}

func (d Driver) Migrations() string {
	return d.String()
}

const (
	Clickhouse Driver = "clickhouse"
)
