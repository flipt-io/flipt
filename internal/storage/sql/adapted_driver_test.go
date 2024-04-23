package sql

import (
	"context"

	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestAdaptedDriver(t *testing.T) {
	mockDriver := NewMockDriverContext(t)
	t.Run("failure", func(t *testing.T) {
		name := "pgx://failure"
		mockDriver.On("OpenConnector", name).Return(nil, &pgconn.PgError{})
		d := &adaptedDriver{origin: mockDriver, adapter: Postgres}
		_, err := d.Open(name)
		require.Error(t, err)
	})
	t.Run("success", func(t *testing.T) {
		o := newMockConnector(t)
		var mockConn = &mockDriverConn{}
		o.On("Connect", mock.Anything).Once().Return(mockConn, nil)
		name := "pgx://success"
		mockDriver.On("OpenConnector", name).Return(o, nil)
		d := &adaptedDriver{origin: mockDriver, adapter: Postgres}
		conn, err := d.Open(name)
		require.NoError(t, err)
		require.Equal(t, mockConn, conn)
	})
}

func TestAdaptedConnectorConnect(t *testing.T) {
	o := newMockConnector(t)
	d := &adaptedDriver{}
	c := &adaptedConnector{
		origin:  o,
		adapter: Postgres,
		driver:  d,
	}
	require.Equal(t, d, c.Driver())
	t.Run("failure", func(t *testing.T) {
		var mockConn *mockDriverConn
		ctx := context.Background()
		o.On("Connect", ctx).Once().Return(mockConn, &pgconn.ConnectError{})
		_, err := c.Connect(ctx)
		require.Error(t, err)
		require.Equal(t, err, errConnectionFailed)
	})

	t.Run("success", func(t *testing.T) {
		var mockConn = &mockDriverConn{}
		ctx := context.Background()
		o.On("Connect", ctx).Once().Return(mockConn, nil)
		conn, err := c.Connect(ctx)
		require.NoError(t, err)
		require.Equal(t, mockConn, conn)
	})
}
