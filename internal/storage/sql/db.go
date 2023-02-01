package sql

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"net/url"

	sq "github.com/Masterminds/squirrel"
	"github.com/XSAM/otelsql"
	"github.com/go-sql-driver/mysql"
	"github.com/lib/pq"
	"github.com/mattn/go-sqlite3"
	"github.com/xo/dburl"
	"go.flipt.io/flipt/internal/config"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

// Open opens a connection to the db
func Open(cfg config.Config, opts ...Option) (*sql.DB, Driver, error) {
	var options Options
	for _, opt := range opts {
		opt(&options)
	}

	sql, driver, err := open(cfg, options)
	if err != nil {
		return nil, 0, err
	}

	sql.SetMaxIdleConns(cfg.Database.MaxIdleConn)

	var maxOpenConn int 
	if cfg.Database.MaxOpenConn > 0 {
		maxOpenConn = cfg.Database.MaxOpenConn
	}

	// if we're using sqlite, we need to set always set the max open connections to 1
	// see: https://github.com/mattn/go-sqlite3/issues/274
	if driver == SQLite {
		if maxOpenConn > 1 {
			log.Warning("Ignoring config.db.max_open_conn due to driver limitation (sqlite)", zap.String("attempted_max_conn", maxOpenConn))
		}
		
		maxOpenConn = 1
	}
	
	sql.SetMaxOpenConns(maxOpenConn)

	if cfg.Database.ConnMaxLifetime > 0 {
		sql.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)
	}

	err = otelsql.RegisterDBStatsMetrics(sql,
		otelsql.WithAttributes(
			attribute.Key("driver").String(driver.String()),
		))

	return sql, driver, err
}

// BuilderFor returns a squirrel statement builder which decorates
// the provided sql.DB configured for the provided driver.
func BuilderFor(db *sql.DB, driver Driver) sq.StatementBuilderType {
	builder := sq.StatementBuilder.RunWith(sq.NewStmtCacher(db))
	if driver == Postgres || driver == CockroachDB {
		builder = builder.PlaceholderFormat(sq.Dollar)
	}

	return builder
}

type Options struct {
	sslDisabled bool
	migrate     bool
}

type Option func(*Options)

func WithSSLDisabled(o *Options) {
	o.sslDisabled = true
}

func WithMigrate(o *Options) {
	o.migrate = true
}

func open(cfg config.Config, opts Options) (*sql.DB, Driver, error) {
	d, url, err := parse(cfg, opts)
	if err != nil {
		return nil, 0, err
	}

	driverName := fmt.Sprintf("instrumented-%s", d)

	var (
		dr    driver.Driver
		attrs []attribute.KeyValue
	)

	switch d {
	case SQLite:
		dr = &sqlite3.SQLiteDriver{}
		attrs = []attribute.KeyValue{semconv.DBSystemSqlite}
	case Postgres:
		dr = &pq.Driver{}
		attrs = []attribute.KeyValue{semconv.DBSystemPostgreSQL}
	case CockroachDB:
		dr = &pq.Driver{}
		attrs = []attribute.KeyValue{semconv.DBSystemCockroachdb}
	case MySQL:
		dr = &mysql.MySQLDriver{}
		attrs = []attribute.KeyValue{semconv.DBSystemMySQL}
	}

	registered := false

	for _, dd := range sql.Drivers() {
		if dd == driverName {
			registered = true
			break
		}
	}

	if !registered {
		sql.Register(driverName, otelsql.WrapDriver(dr, otelsql.WithAttributes(attrs...)))
	}

	db, err := sql.Open(driverName, url.DSN)
	if err != nil {
		return nil, 0, fmt.Errorf("opening db for driver: %s %w", d, err)
	}

	return db, d, nil
}

var (
	driverToString = map[Driver]string{
		SQLite:      "sqlite3",
		Postgres:    "postgres",
		MySQL:       "mysql",
		CockroachDB: "cockroachdb",
	}

	stringToDriver = map[string]Driver{
		"sqlite3":     SQLite,
		"postgres":    Postgres,
		"mysql":       MySQL,
		"cockroachdb": CockroachDB,
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
	// CockroachDB ...
	CockroachDB
)

func parse(cfg config.Config, opts Options) (Driver, *dburl.URL, error) {
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

	url, err := dburl.Parse(u)
	if err != nil {
		return 0, nil, fmt.Errorf("error parsing url: %q, %w", url, err)
	}

	driver := stringToDriver[url.Unaliased]
	if driver == 0 {
		return 0, nil, fmt.Errorf("unknown database driver for: %q", url.Driver)
	}

	switch driver {
	case Postgres, CockroachDB:
		if opts.sslDisabled {
			v := url.Query()
			v.Set("sslmode", "disable")
			url.RawQuery = v.Encode()
			// we need to re-parse since we modified the query params
			url, err = dburl.Parse(url.URL.String())
		}
	case MySQL:
		v := url.Query()
		v.Set("multiStatements", "true")
		v.Set("parseTime", "true")
		if !opts.migrate {
			v.Set("sql_mode", "ANSI")
		}
		url.RawQuery = v.Encode()
		// we need to re-parse since we modified the query params
		url, err = dburl.Parse(url.URL.String())

	case SQLite:
		v := url.Query()
		v.Set("cache", "shared")
		v.Set("mode", "rwc")
		v.Set("_fk", "true")
		url.RawQuery = v.Encode()

		// we need to re-parse since we modified the query params
		url, err = dburl.Parse(url.URL.String())
	}

	return driver, url, err
}
