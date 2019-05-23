package storage

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"path/filepath"
	"strings"
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

// DB is an abstraction for a database
type DB struct {
	dbType dbType
	uri    string

	builder sq.StatementBuilderType
	db      *sql.DB
}

// Open opens a connection to the db given a URL
func Open(url string) (*DB, error) {
	dbType, uri, err := parse(url)
	if err != nil {
		return nil, err
	}

	db, err := sql.Open(dbType.String(), uri)
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
		uri:     uri,
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
		"sqlite3":  dbSQLite,
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

func parse(url string) (dbType, string, error) {
	parts := strings.SplitN(url, "://", 2)
	// TODO: check parts

	dbType := stringToDBType[parts[0]]
	if dbType == 0 {
		return 0, "", fmt.Errorf("unknown database type: %s", parts[0])
	}

	uri := parts[1]

	switch dbType {
	case dbSQLite:
		uri = fmt.Sprintf("%s?cache=shared&_fk=true", parts[1])
	case dbPostgres:
		uri = fmt.Sprintf("postgres://%s", parts[1])
	}

	return dbType, uri, nil
}
