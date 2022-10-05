package config

import (
	"encoding/json"
	"time"

	"github.com/spf13/viper"
)

const (
	// configuration keys
	dbURL             = "db.url"
	dbMigrationsPath  = "db.migrations.path"
	dbMaxIdleConn     = "db.max_idle_conn"
	dbMaxOpenConn     = "db.max_open_conn"
	dbConnMaxLifetime = "db.conn_max_lifetime"
	dbName            = "db.name"
	dbUser            = "db.user"
	dbPassword        = "db.password"
	dbHost            = "db.host"
	dbPort            = "db.port"
	dbProtocol        = "db.protocol"

	// database protocol enum
	_ DatabaseProtocol = iota
	// DatabaseSQLite ...
	DatabaseSQLite
	// DatabasePostgres ...
	DatabasePostgres
	// DatabaseMySQL ...
	DatabaseMySQL
)

// DatabaseConfig contains fields, which configure the various relational database backends.
//
// Flipt currently supports SQLite, Postgres and MySQL backends.
type DatabaseConfig struct {
	MigrationsPath  string           `json:"migrationsPath,omitempty"`
	URL             string           `json:"url,omitempty"`
	MaxIdleConn     int              `json:"maxIdleConn,omitempty"`
	MaxOpenConn     int              `json:"maxOpenConn,omitempty"`
	ConnMaxLifetime time.Duration    `json:"connMaxLifetime,omitempty"`
	Name            string           `json:"name,omitempty"`
	User            string           `json:"user,omitempty"`
	Password        string           `json:"password,omitempty"`
	Host            string           `json:"host,omitempty"`
	Port            int              `json:"port,omitempty"`
	Protocol        DatabaseProtocol `json:"protocol,omitempty"`
}

func (c *DatabaseConfig) init() (warnings []string, _ error) {
	// read in configuration via viper
	if viper.IsSet(dbURL) {
		c.URL = viper.GetString(dbURL)

	} else if viper.IsSet(dbProtocol) || viper.IsSet(dbName) || viper.IsSet(dbUser) || viper.IsSet(dbPassword) || viper.IsSet(dbHost) || viper.IsSet(dbPort) {
		c.URL = ""

		if viper.IsSet(dbProtocol) {
			c.Protocol = stringToDatabaseProtocol[viper.GetString(dbProtocol)]
		}

		if viper.IsSet(dbName) {
			c.Name = viper.GetString(dbName)
		}

		if viper.IsSet(dbUser) {
			c.User = viper.GetString(dbUser)
		}

		if viper.IsSet(dbPassword) {
			c.Password = viper.GetString(dbPassword)
		}

		if viper.IsSet(dbHost) {
			c.Host = viper.GetString(dbHost)
		}

		if viper.IsSet(dbPort) {
			c.Port = viper.GetInt(dbPort)
		}

	}

	if viper.IsSet(dbMigrationsPath) {
		c.MigrationsPath = viper.GetString(dbMigrationsPath)
	}

	if viper.IsSet(dbMaxIdleConn) {
		c.MaxIdleConn = viper.GetInt(dbMaxIdleConn)
	}

	if viper.IsSet(dbMaxOpenConn) {
		c.MaxOpenConn = viper.GetInt(dbMaxOpenConn)
	}

	if viper.IsSet(dbConnMaxLifetime) {
		c.ConnMaxLifetime = viper.GetDuration(dbConnMaxLifetime)
	}

	// validation
	if c.URL == "" {
		if c.Protocol == 0 {
			return nil, errFieldRequired("database.protocol")
		}

		if c.Host == "" {
			return nil, errFieldRequired("database.host")
		}

		if c.Name == "" {
			return nil, errFieldRequired("database.name")
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

const ()

var (
	databaseProtocolToString = map[DatabaseProtocol]string{
		DatabaseSQLite:   "file",
		DatabasePostgres: "postgres",
		DatabaseMySQL:    "mysql",
	}

	stringToDatabaseProtocol = map[string]DatabaseProtocol{
		"file":     DatabaseSQLite,
		"sqlite":   DatabaseSQLite,
		"postgres": DatabasePostgres,
		"mysql":    DatabaseMySQL,
	}
)
