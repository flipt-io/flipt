package db

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"net/url"
	"time"

	"github.com/golang/protobuf/ptypes"
	proto "github.com/golang/protobuf/ptypes/timestamp"
)

type timestamp struct {
	*proto.Timestamp
}

func (t *timestamp) Scan(value interface{}) error {
	if v, ok := value.(time.Time); ok {
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

// Open opens a connection to the db given a URL
func Open(url string) (*sql.DB, Dialect, error) {
	driver, u, err := parse(url)
	if err != nil {
		return nil, 0, err
	}

	db, err := sql.Open(driver.String(), u.String())
	if err != nil {
		return nil, 0, err
	}

	return db, driver, nil
}

var (
	dialectToString = map[Dialect]string{
		SQLite:   "sqlite3",
		Postgres: "postgres",
	}

	schemeToDialect = map[string]Dialect{
		"file":     SQLite,
		"postgres": Postgres,
	}
)

// Dialect represents a database dialect
type Dialect uint8

func (d Dialect) String() string {
	return dialectToString[d]
}

const (
	_ Dialect = iota
	// SQLite ...
	SQLite
	// Postgres ...
	Postgres
)

func parse(in string) (Dialect, *url.URL, error) {
	u, err := url.Parse(in)
	if err != nil {
		return 0, u, fmt.Errorf("parsing url: %q: %w", in, err)
	}

	dialect := schemeToDialect[u.Scheme]
	if dialect == 0 {
		return 0, u, fmt.Errorf("unknown database dialect for: %s", u.Scheme)
	}

	if dialect == SQLite {
		v := u.Query()
		v.Set("cache", "shared")
		v.Set("_fk", "true")
		u.RawQuery = v.Encode()
	}

	return dialect, u, nil
}
