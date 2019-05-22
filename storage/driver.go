package storage

import (
	"fmt"
	"strings"
)

var (
	driverToString = map[Driver]string{
		SQLite:   "sqlite3",
		Postgres: "postgres",
	}

	stringToDriver = map[string]Driver{
		"sqlite3":  SQLite,
		"postgres": Postgres,
	}
)

type Driver uint8

func (d Driver) String() string {
	return driverToString[d]
}

const (
	_ Driver = iota
	SQLite
	Postgres
)

func Parse(url string) (Driver, string, error) {
	parts := strings.SplitN(url, "://", 2)
	// TODO: check parts

	driver := stringToDriver[parts[0]]
	if driver == 0 {
		return 0, "", fmt.Errorf("unknown database driver: %s", parts[0])
	}

	uri := parts[1]

	switch driver {
	case SQLite:
		uri = fmt.Sprintf("%s?cache=shared&_fk=true", parts[1])
	case Postgres:
		uri = fmt.Sprintf("postgres://%s?sslmode=disable", parts[1])
	}

	return driver, uri, nil
}
