package sql

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"net/url"

	"github.com/go-sql-driver/mysql"
	"github.com/lib/pq"
	"github.com/luna-duclos/instrumentedsql"
	"github.com/luna-duclos/instrumentedsql/opentracing"
	"github.com/markphelps/flipt/config"
	"github.com/mattn/go-sqlite3"
	"github.com/xo/dburl"
)

// Open opens a connection to the db
func Open(cfg config.Config) (*sql.DB, Driver, error) {
	sql, driver, err := open(cfg, false)
	if err != nil {
		return nil, 0, err
	}

	sql.SetMaxIdleConns(cfg.Database.MaxIdleConn)

	if cfg.Database.MaxOpenConn > 0 {
		sql.SetMaxOpenConns(cfg.Database.MaxOpenConn)
	}
	if cfg.Database.ConnMaxLifetime > 0 {
		sql.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)
	}

	registerMetrics(driver, sql)

	return sql, driver, nil
}

func open(cfg config.Config, migrate bool) (*sql.DB, Driver, error) {
	d, url, err := parse(cfg, migrate)
	if err != nil {
		return nil, 0, err
	}

	driverName := fmt.Sprintf("instrumented-%s", d)

	var dr driver.Driver

	switch d {
	case SQLite:
		dr = &sqlite3.SQLiteDriver{}
	case Postgres:
		dr = &pq.Driver{}
	case MySQL:
		dr = &mysql.MySQLDriver{}
	}

	registered := false

	for _, dd := range sql.Drivers() {
		if dd == driverName {
			registered = true
			break
		}
	}

	if !registered {
		sql.Register(driverName, instrumentedsql.WrapDriver(dr, instrumentedsql.WithTracer(opentracing.NewTracer(false))))
	}

	db, err := sql.Open(driverName, url.DSN)
	if err != nil {
		return nil, 0, fmt.Errorf("opening db for driver: %s %w", d, err)
	}

	return db, d, nil
}

var (
	driverToString = map[Driver]string{
		SQLite:   "sqlite3",
		Postgres: "postgres",
		MySQL:    "mysql",
	}

	stringToDriver = map[string]Driver{
		"sqlite3":  SQLite,
		"postgres": Postgres,
		"mysql":    MySQL,
	}
)

// Driver represents a database driver
type Driver uint8

func (d Driver) String() string {
	return driverToString[d]
}

const (
	_ Driver = iota
	// SQLite ...
	SQLite
	// Postgres ...
	Postgres
	// MySQL ...
	MySQL
)

func parse(cfg config.Config, migrate bool) (Driver, *dburl.URL, error) {
	u := cfg.Database.URL

	if u == "" {
		host := cfg.Database.Host

		if cfg.Database.Port > 0 {
			host = fmt.Sprintf("%s:%d", host, cfg.Database.Port)
		}

		uu := url.URL{
			Scheme: cfg.Database.Protocol.String(),
			Host:   host,
			Path:   cfg.Database.Name,
		}

		if cfg.Database.User != "" {
			if cfg.Database.Password != "" {
				uu.User = url.UserPassword(cfg.Database.User, cfg.Database.Password)
			} else {
				uu.User = url.User(cfg.Database.User)
			}
		}

		u = uu.String()
	}

	errURL := func(rawurl string, err error) error {
		return fmt.Errorf("error parsing url: %q, %w", rawurl, err)
	}

	url, err := dburl.Parse(u)
	if err != nil {
		return 0, nil, errURL(u, err)
	}

	driver := stringToDriver[url.Driver]
	if driver == 0 {
		return 0, nil, fmt.Errorf("unknown database driver for: %q", url.Driver)
	}

	switch driver {
	case MySQL:
		v := url.Query()
		v.Set("multiStatements", "true")
		v.Set("parseTime", "true")
		if !migrate {
			v.Set("sql_mode", "ANSI")
		}
		url.RawQuery = v.Encode()
		// we need to re-parse since we modified the query params
		url, err = dburl.Parse(url.URL.String())

	case SQLite:
		v := url.Query()
		v.Set("cache", "shared")
		v.Set("_fk", "true")
		url.RawQuery = v.Encode()

		// we need to re-parse since we modified the query params
		url, err = dburl.Parse(url.URL.String())
	}

	return driver, url, err
}
