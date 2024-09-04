package sql

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"io/fs"
	"net/url"

	"github.com/ClickHouse/clickhouse-go/v2"
	sq "github.com/Masterminds/squirrel"
	"github.com/XSAM/otelsql"
	"github.com/go-sql-driver/mysql"

	"github.com/libsql/libsql-client-go/libsql"
	"github.com/mattn/go-sqlite3"
	"github.com/xo/dburl"
	"go.flipt.io/flipt/internal/config"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

func init() {
	// we do a bit of surgery in dburl to stop it from walking
	// up the provided file:/path to see if any parent directories
	// exist, else dburl assumes the postgres protocol.
	// see: https://github.com/xo/dburl/issues/35
	stat := dburl.Stat
	dburl.Stat = func(name string) (fs.FileInfo, error) {
		fi, err := stat(name)
		if err == nil {
			return fi, nil
		}

		if errors.Is(err, fs.ErrNotExist) {
			return fileInfo(name), nil
		}

		return nil, err
	}

	// register libsql driver with dburl
	dburl.Register(dburl.Scheme{
		Driver:    "libsql",
		Generator: dburl.GenOpaque,
		Transport: 0,
		Opaque:    true,
		Aliases:   []string{"libsql", "http", "https"},
		Override:  "file",
	})
	// drop references to lib/pq and relay on pgx
	dburl.Unregister("postgres")
	dburl.RegisterAlias("pgx", "postgres")
	dburl.RegisterAlias("pgx", "postgresql")
}

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

	err = otelsql.RegisterDBStatsMetrics(sql,
		otelsql.WithAttributes(
			attribute.Key("driver").String(driver.String()),
		))

	return sql, driver, err
}

// BuilderFor returns a squirrel statement builder which decorates
// the provided sql.DB configured for the provided driver.
func BuilderFor(db *sql.DB, driver Driver, preparedStatementsEnabled bool) sq.StatementBuilderType {
	var brdb sq.BaseRunner = db
	if preparedStatementsEnabled {
		brdb = sq.NewStmtCacher(db)
	}

	builder := sq.StatementBuilder.RunWith(brdb)
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
	case LibSQL:
		dr = &libsql.Driver{}
		attrs = []attribute.KeyValue{semconv.DBSystemSqlite}
	case Postgres:
		dr = newAdaptedPostgresDriver(d)
		attrs = []attribute.KeyValue{semconv.DBSystemPostgreSQL}
	case CockroachDB:
		dr = newAdaptedPostgresDriver(d)
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

	db.SetMaxIdleConns(cfg.Database.MaxIdleConn)

	var maxOpenConn int
	if cfg.Database.MaxOpenConn > 0 {
		maxOpenConn = cfg.Database.MaxOpenConn
	}

	// if we're using sqlite, we need to set always set the max open connections to 1
	// see: https://github.com/mattn/go-sqlite3/issues/274
	if d == SQLite || d == LibSQL {
		maxOpenConn = 1
	}

	db.SetMaxOpenConns(maxOpenConn)

	if cfg.Database.ConnMaxLifetime > 0 {
		db.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)
	}

	return db, d, nil
}

// openAnalytics is a convenience function of providing a database.sql instance for
// an analytics database.
func openAnalytics(cfg config.Config) (*sql.DB, Driver, error) {
	if cfg.Analytics.Storage.Clickhouse.Enabled {
		clickhouseOptions, err := cfg.Analytics.Storage.Clickhouse.Options()
		if err != nil {
			return nil, 0, err
		}

		db := clickhouse.OpenDB(clickhouseOptions)

		return db, Clickhouse, nil
	}

	return nil, 0, errors.New("no analytics db provided")
}

var (
	driverToString = map[Driver]string{
		SQLite:      "sqlite3",
		LibSQL:      "libsql",
		Postgres:    "postgres",
		MySQL:       "mysql",
		CockroachDB: "cockroachdb",
		Clickhouse:  "clickhouse",
	}

	stringToDriver = map[string]Driver{
		"sqlite3":     SQLite,
		"libsql":      LibSQL,
		"pgx":         Postgres,
		"mysql":       MySQL,
		"cockroachdb": CockroachDB,
		"clickhouse":  Clickhouse,
	}
)

// Driver represents a database driver
type Driver uint8

func (d Driver) String() string {
	return driverToString[d]
}

func (d Driver) Migrations() string {
	if d == LibSQL {
		return "sqlite3"
	}
	return d.String()
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
	// LibSQL ...
	LibSQL
	// Clickhouse ...
	Clickhouse
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
		return 0, nil, fmt.Errorf("error parsing url: %w", err)
	}

	driver := stringToDriver[url.UnaliasedDriver]
	if driver == 0 {
		return 0, nil, fmt.Errorf("unknown database driver for: %q", url.Driver)
	}

	v := url.Query()
	switch driver {
	case Postgres, CockroachDB:
		if opts.sslDisabled {
			v.Set("sslmode", "disable")
		}

		if !cfg.Database.PreparedStatementsEnabled {
			v.Set("default_query_exec_mode", "simple_protocol")
		}
	case MySQL:
		v.Set("multiStatements", "true")
		v.Set("parseTime", "true")
		if !opts.migrate {
			v.Set("sql_mode", "ANSI")
		}
	case SQLite, LibSQL:
		if url.Scheme != "http" && url.Scheme != "https" {
			v.Set("cache", "shared")
			v.Set("mode", "rwc")
			v.Set("_fk", "true")
		}
	}

	url.RawQuery = v.Encode()
	// we need to re-parse since we modified the query params
	url, err = dburl.Parse(url.URL.String())

	if url.Scheme == "http" {
		url.DSN = "http://" + url.DSN
	} else if url.Scheme == "https" {
		url.DSN = "https://" + url.DSN
	}

	return driver, url, err
}
