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

const (
	pgConstraintForeignKey = "foreign_key_violation"
	pgConstraintUnique     = "unique_violation"
)

// Open opens a connection to the db given a URL
func Open(url string) (*sql.DB, Driver, error) {
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
	driverToString = map[Driver]string{
		SQLite:   "sqlite3",
		Postgres: "postgres",
	}

	schemeToDriver = map[string]Driver{
		"file":     SQLite,
		"postgres": Postgres,
	}
)

// Driver represents a database driver type
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
)

func parse(in string) (Driver, *url.URL, error) {
	u, err := url.Parse(in)
	if err != nil {
		return 0, u, fmt.Errorf("parsing url: %q: %w", in, err)
	}

	driver := schemeToDriver[u.Scheme]
	if driver == 0 {
		return 0, u, fmt.Errorf("unknown database driver for: %s", u.Scheme)
	}

	if driver == SQLite {
		v := u.Query()
		v.Set("cache", "shared")
		v.Set("_fk", "true")
		u.RawQuery = v.Encode()
	}

	return driver, u, nil
}
