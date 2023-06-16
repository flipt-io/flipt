package config

import (
	"encoding/json"
	"time"

	"github.com/spf13/viper"
)

// cheers up the unparam linter
var _ defaulter = (*DatabaseConfig)(nil)

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
)

// DatabaseConfig contains fields, which configure the various relational database backends.
//
// Flipt currently supports SQLite, Postgres and MySQL backends.
type DatabaseConfig struct {
	URL                       string           `json:"url,omitempty" mapstructure:"url,omitempty"`
	MaxIdleConn               int              `json:"maxIdleConn,omitempty" mapstructure:"max_idle_conn"`
	MaxOpenConn               int              `json:"maxOpenConn,omitempty" mapstructure:"max_open_conn"`
	ConnMaxLifetime           time.Duration    `json:"connMaxLifetime,omitempty" mapstructure:"conn_max_lifetime"`
	Name                      string           `json:"name,omitempty" mapstructure:"name,omitempty"`
	User                      string           `json:"user,omitempty" mapstructure:"user,omitempty"`
	Password                  string           `json:"password,omitempty" mapstructure:"password,omitempty"`
	Host                      string           `json:"host,omitempty" mapstructure:"host,omitempty"`
	Port                      int              `json:"port,omitempty" mapstructure:"port,omitempty"`
	Protocol                  DatabaseProtocol `json:"protocol,omitempty" mapstructure:"protocol,omitempty"`
	PreparedStatementsEnabled bool             `json:"preparedStatementsEnabled,omitempty" mapstructure:"prepared_statements_enabled"`
}

func (c *DatabaseConfig) setDefaults(v *viper.Viper) {
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
		v.SetDefault("db.url", "file:/var/opt/flipt/flipt.db")
	}

	v.SetDefault("db.prepared_statements_enabled", true)
}

func (c *DatabaseConfig) deprecations(v *viper.Viper) []deprecation {
	var deprecations []deprecation

	if v.IsSet("db.migrations.path") || v.IsSet("db.migrations_path") {
		deprecations = append(deprecations, deprecation{
			option:            "db.migrations.path",
			additionalMessage: deprecatedMsgDatabaseMigrations,
		})
	}

	return deprecations
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
		DatabaseSQLite:      "file",
		DatabasePostgres:    "postgres",
		DatabaseMySQL:       "mysql",
		DatabaseCockroachDB: "cockroachdb",
	}

	stringToDatabaseProtocol = map[string]DatabaseProtocol{
		"file":        DatabaseSQLite,
		"sqlite":      DatabaseSQLite,
		"postgres":    DatabasePostgres,
		"mysql":       DatabaseMySQL,
		"cockroachdb": DatabaseCockroachDB,
		"cockroach":   DatabaseCockroachDB,
	}
)
