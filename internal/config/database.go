package config

import (
	"encoding/json"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
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
)

var dbProtocolDecodeHooks = mapstructure.ComposeDecodeHookFunc(
	mapstructure.StringToTimeDurationHookFunc(),
	mapstructure.StringToSliceHookFunc(","),
	StringToEnumHookFunc(stringToDatabaseProtocol),
)

// DatabaseConfig contains fields, which configure the various relational database backends.
//
// Flipt currently supports SQLite, Postgres and MySQL backends.
type DatabaseConfig struct {
	MigrationsPath  string           `json:"migrationsPath,omitempty" mapstructure:"migrations_path"`
	URL             string           `json:"url,omitempty" mapstructure:"url"`
	MaxIdleConn     int              `json:"maxIdleConn,omitempty" mapstructure:"max_idle_conn"`
	MaxOpenConn     int              `json:"maxOpenConn,omitempty" mapstructure:"max_open_conn"`
	ConnMaxLifetime time.Duration    `json:"connMaxLifetime,omitempty" mapstructure:"conn_max_lifetime"`
	Name            string           `json:"name,omitempty" mapstructure:"name"`
	User            string           `json:"user,omitempty" mapstructure:"user"`
	Password        string           `json:"password,omitempty" mapstructure:"password"`
	Host            string           `json:"host,omitempty" mapstructure:"host"`
	Port            int              `json:"port,omitempty" mapstructure:"port"`
	Protocol        DatabaseProtocol `json:"protocol,omitempty" mapstructure:"protocol"`
}

func (c *DatabaseConfig) viperKey() string {
	return "db"
}

func (c *DatabaseConfig) unmarshalViper(v *viper.Viper) (warnings []string, err error) {
	// supports nesting `path` beneath `migrations` key
	v.RegisterAlias("migrations_path", "migrations.path")
	v.SetDefault("migrations_path", "/etc/flipt/config/migrations")
	v.SetDefault("max_idle_conn", 2)

	// URL default is only set given that none of the alternative
	// database connections parameters are provided
	setDefaultURL := true
	for _, field := range []string{"name", "user", "password", "host", "port", "protocol"} {
		setDefaultURL = setDefaultURL && !v.IsSet(field)
	}

	if setDefaultURL {
		v.SetDefault("url", "file:/var/opt/flipt/flipt.db")
	}

	if err = v.Unmarshal(c, viper.DecodeHook(dbProtocolDecodeHooks)); err != nil {
		return
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
