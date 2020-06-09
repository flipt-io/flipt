package db

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/golang/protobuf/ptypes"
	proto "github.com/golang/protobuf/ptypes/timestamp"
	"github.com/markphelps/flipt/errors"
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
func Open(rawurl string) (*sql.DB, Driver, error) {
	driver, url, err := parse(rawurl)
	if err != nil {
		return nil, 0, err
	}

	db, err := sql.Open(driver.String(), url)
	if err != nil {
		return nil, 0, err
	}

	return db, driver, nil
}

var (
	driverToString = map[Driver]string{
		SQLite:   "sqlite3",
		Postgres: "postgres",
		MySQL:    "mysql",
	}

	schemeToDriver = map[string]Driver{
		"file":     SQLite,
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

// copied from net/url
func getScheme(rawurl string) (scheme, path string, err error) {
	for i := 0; i < len(rawurl); i++ {
		c := rawurl[i]
		switch {
		case 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z':
		// do nothing
		case '0' <= c && c <= '9' || c == '+' || c == '-' || c == '.':
			if i == 0 {
				return "", rawurl, nil
			}
		case c == ':':
			if i == 0 {
				return "", "", errors.New("missing protocol scheme")
			}
			return rawurl[:i], rawurl[i+1:], nil
		default:
			// we have encountered an invalid character,
			// so there is no valid scheme
			return "", rawurl, nil
		}
	}
	return "", rawurl, nil
}

func parse(rawurl string) (Driver, string, error) {
	scheme, _, err := getScheme(rawurl)
	if err != nil {
		return 0, "", fmt.Errorf("getting scheme from url: %q", rawurl)
	}

	driver := schemeToDriver[scheme]
	if driver == 0 {
		return 0, "", fmt.Errorf("unknown database driver for: %q", scheme)
	}

	errURL := func(rawurl string, err error) error {
		return fmt.Errorf("error parsing url: %q, %v", rawurl, err)
	}

	switch driver {
	case MySQL:
		cfg, err := mysql.ParseDSN(strings.TrimPrefix(rawurl, "mysql://"))
		if err != nil {
			return 0, "", errURL(rawurl, err)
		}
		cfg.MultiStatements = true

		return driver, cfg.FormatDSN(), nil
	default:
		u, err := url.Parse(rawurl)
		if err != nil {
			return 0, "", errURL(rawurl, err)
		}

		if driver == SQLite {
			v := u.Query()
			v.Set("cache", "shared")
			v.Set("_fk", "true")
			u.RawQuery = v.Encode()
		}

		return driver, u.String(), nil
	}
}
