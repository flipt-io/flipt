package sql

import (
	"context"
	"database/sql/driver"
	"time"

	pgx "github.com/jackc/pgx/v5/stdlib"
)

const adaptedDriverOpenTimeout = 60 * time.Second

// This is the wrapper around sql driver. By default, pgx driver returns connection
// error with the host, username and password. `adaptedDriver` and `postgresConnector`
// allow to customize errors and preventing leakage of the credentials to outside.
func newAdaptedPostgresDriver(d Driver) driver.Driver {
	return &adaptedDriver{origin: &pgx.Driver{}, adapter: d}
}

var _ driver.Driver = (*adaptedDriver)(nil)
var _ driver.DriverContext = (*adaptedDriver)(nil)

type adaptedDriver struct {
	adapter Driver
	origin  driver.DriverContext
}

func (d *adaptedDriver) Open(name string) (driver.Conn, error) {
	connector, err := d.OpenConnector(name)
	if err != nil {
		return nil, d.adapter.AdaptError(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), adaptedDriverOpenTimeout)
	defer cancel()
	return connector.Connect(ctx)
}

func (d *adaptedDriver) OpenConnector(name string) (driver.Connector, error) {
	connector, err := d.origin.OpenConnector(name)
	if err != nil {
		return nil, d.adapter.AdaptError(err)
	}
	return &adaptedConnector{origin: connector, driver: d, adapter: d.adapter}, nil
}

var _ driver.Connector = (*adaptedConnector)(nil)

type adaptedConnector struct {
	origin  driver.Connector
	driver  driver.Driver
	adapter Driver
}

func (c *adaptedConnector) Driver() driver.Driver {
	return c.driver
}

func (c *adaptedConnector) Connect(ctx context.Context) (driver.Conn, error) {
	conn, err := c.origin.Connect(ctx)
	if err != nil {
		return nil, c.adapter.AdaptError(err)
	}
	return conn, nil
}
