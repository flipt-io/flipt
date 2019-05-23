package storage

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"net/url"
	"path/filepath"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/database"
	"github.com/golang-migrate/migrate/database/postgres"
	"github.com/golang-migrate/migrate/database/sqlite3"
	"github.com/golang/protobuf/ptypes"
	proto "github.com/golang/protobuf/ptypes/timestamp"
	"github.com/pkg/errors"
)

type timestamp struct {
	*proto.Timestamp
}

func (t *timestamp) Scan(value interface{}) error {
	v, ok := value.(time.Time)
	if ok {
		val, err := ptypes.TimestampProto(v)
		if err != nil {
			return err
		}
		t.Timestamp = val
	}
	return nil
}

func (t *timestamp) Value() (driver.Value, error) {
	return ptypes.Timestamp(t.Timestamp)
}

const pgIntegrityConstraint = "integrity_constraint_violation"

// DB is an abstraction for a database
type DB struct {
	dbType dbType
	url    *url.URL

	builder sq.StatementBuilderType
	db      *sql.DB
}

// Open opens a connection to the db given a URL
func Open(url string) (*DB, error) {
	dbType, u, err := parse(url)
	if err != nil {
		return nil, err
	}

	db, err := sql.Open(dbType.String(), u.String())
	if err != nil {
		return nil, err
	}

	var (
		cacher  = sq.NewStmtCacher(db)
		builder sq.StatementBuilderType
	)

	switch dbType {
	case dbSQLite:
		builder = sq.StatementBuilder.RunWith(cacher)
	case dbPostgres:
		builder = sq.StatementBuilder.PlaceholderFormat(sq.Dollar).RunWith(cacher)
	}

	return &DB{
		dbType:  dbType,
		url:     u,
		builder: builder,
		db:      db,
	}, nil
}

// Close closes a connection to the db
func (d *DB) Close() error {
	return d.db.Close()
}

// Migrate runs database migrations that exist in the given path on our db
func (d *DB) Migrate(path string) error {
	var (
		dr  database.Driver
		err error
	)

	switch d.dbType {
	case dbSQLite:
		dr, err = sqlite3.WithInstance(d.db, &sqlite3.Config{})
	case dbPostgres:
		dr, err = postgres.WithInstance(d.db, &postgres.Config{})
	}

	if err != nil {
		return errors.Wrapf(err, "getting db driver: %s for migrations", d.dbType)
	}

	f := filepath.Clean(fmt.Sprintf("%s/%s", path, d.dbType))
	mm, err := migrate.NewWithDatabaseInstance(fmt.Sprintf("file://%s", f), d.dbType.String(), dr)
	if err != nil {
		return errors.Wrap(err, "opening migrations")
	}

	if err := mm.Up(); err != nil && err != migrate.ErrNoChange {
		return errors.Wrap(err, "running migrations")
	}

	return nil
}

var (
	dbTypeToString = map[dbType]string{
		dbSQLite:   "sqlite3",
		dbPostgres: "postgres",
	}

	stringToDBType = map[string]dbType{
		"file":     dbSQLite,
		"postgres": dbPostgres,
	}
)

type dbType uint8

func (d dbType) String() string {
	return dbTypeToString[d]
}

const (
	_ dbType = iota
	dbSQLite
	dbPostgres
)

func parse(in string) (dbType, *url.URL, error) {
	u, err := url.Parse(in)
	if err != nil {
		return 0, u, errors.Wrapf(err, "parsing url: %q", in)
	}

	dbType := stringToDBType[u.Scheme]
	if dbType == 0 {
		return 0, u, fmt.Errorf("unknown database type: %s", u.Scheme)
	}

	switch dbType {
	case dbSQLite:
		v := u.Query()
		v.Set("cache", "shared")
		v.Set("_fk", "true")
		u.RawQuery = v.Encode()
	case dbPostgres:
		// do nothing
	}

	return dbType, u, nil
}
