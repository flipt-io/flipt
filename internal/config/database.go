package config

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"

	"github.com/spf13/viper"
)

var (
	_ defaulter = (*DatabaseConfig)(nil)
	_ validator = (*DatabaseConfig)(nil)
)

const (
	// database protocol enum
	_ DatabaseProtocol = iota
	// DatabaseSQLite ...
	DatabaseSQLite
	// DatabasePostgres ...
	DatabasePostgres
	// DatabaseMySQL ...
	DatabaseMySQL
	// DatabaseCockroachDB ...
	DatabaseCockroachDB
	// DatabaseLibSQL ...
	DatabaseLibSQL
)

// DatabaseConfig contains fields, which configure the various relational database backends.
//
// Flipt currently supports SQLite, Postgres and MySQL backends.
type DatabaseConfig struct {
	URL                       string           `json:"url,omitempty" mapstructure:"url,omitempty" yaml:"url,omitempty"`
	MaxIdleConn               int              `json:"maxIdleConn,omitempty" mapstructure:"max_idle_conn" yaml:"max_idle_conn,omitempty"`
	MaxOpenConn               int              `json:"maxOpenConn,omitempty" mapstructure:"max_open_conn" yaml:"max_open_conn,omitempty"`
	ConnMaxLifetime           time.Duration    `json:"connMaxLifetime,omitempty" mapstructure:"conn_max_lifetime" yaml:"conn_max_lifetime,omitempty"`
	Name                      string           `json:"name,omitempty" mapstructure:"name,omitempty" yaml:"name,omitempty"`
	User                      string           `json:"user,omitempty" mapstructure:"user,omitempty" yaml:"user,omitempty"`
	Password                  string           `json:"-" mapstructure:"password,omitempty" yaml:"-"`
	Host                      string           `json:"host,omitempty" mapstructure:"host,omitempty" yaml:"host,omitempty"`
	Port                      int              `json:"port,omitempty" mapstructure:"port,omitempty" yaml:"port,omitempty"`
	Protocol                  DatabaseProtocol `json:"protocol,omitempty" mapstructure:"protocol,omitempty" yaml:"protocol,omitempty"`
	PreparedStatementsEnabled bool             `json:"preparedStatementsEnabled,omitempty" mapstructure:"prepared_statements_enabled" yaml:"prepared_statements_enabled,omitempty"`
}

func (c *DatabaseConfig) setDefaults(v *viper.Viper) error {
	v.SetDefault("db", map[string]any{
		"max_idle_conn": 2,
	})

	// URL default is only set given that none of the alternative
	// database connections parameters are provided
	setDefaultURL := true
	for _, field := range []string{"name", "user", "password", "host", "port", "protocol"} {
		setDefaultURL = setDefaultURL && !v.IsSet("db."+field)
	}

	if setDefaultURL {
		dbRoot, err := defaultDatabaseRoot()
		if err != nil {
			return fmt.Errorf("getting default database directory: %w", err)
		}

		path := filepath.ToSlash(filepath.Join(dbRoot, "flipt.db"))
		v.SetDefault("db.url", "file:"+path)
	}

	v.SetDefault("db.prepared_statements_enabled", true)
	return nil
}

func (c *DatabaseConfig) validate() (err error) {
	if c.URL == "" {
		if c.Protocol == 0 {
			return errFieldRequired("db.protocol")
		}

		if c.Host == "" {
			return errFieldRequired("db.host")
		}

		if c.Name == "" {
			return errFieldRequired("db.name")
		}
	}

	return
}

// DatabaseProtocol represents a database protocol
type DatabaseProtocol uint8

func (d DatabaseProtocol) String() string {
	return databaseProtocolToString[d]
}

func (d DatabaseProtocol) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

var (
	databaseProtocolToString = map[DatabaseProtocol]string{
		DatabaseSQLite:      "sqlite",
		DatabaseLibSQL:      "libsql",
		DatabasePostgres:    "postgres",
		DatabaseMySQL:       "mysql",
		DatabaseCockroachDB: "cockroachdb",
	}

	stringToDatabaseProtocol = map[string]DatabaseProtocol{
		"file":        DatabaseSQLite,
		"sqlite":      DatabaseSQLite,
		"libsql":      DatabaseLibSQL,
		"postgres":    DatabasePostgres,
		"mysql":       DatabaseMySQL,
		"cockroachdb": DatabaseCockroachDB,
		"cockroach":   DatabaseCockroachDB,
	}
)
