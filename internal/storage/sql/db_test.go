//nolint:gosec
package sql_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/config"
	fliptsql "go.flipt.io/flipt/internal/storage/sql"

	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func TestOpen(t *testing.T) {
	tests := []struct {
		name    string
		cfg     config.DatabaseConfig
		driver  fliptsql.Driver
		wantErr bool
	}{
		{
			name: "sqlite url",
			cfg: config.DatabaseConfig{
				URL:             "file:/flipt.db",
				MaxOpenConn:     5,
				ConnMaxLifetime: 30 * time.Minute,
			},
			driver: fliptsql.SQLite,
		},
		{
			name: "sqlite url (without slash)",
			cfg: config.DatabaseConfig{
				URL:             "file:flipt.db",
				MaxOpenConn:     5,
				ConnMaxLifetime: 30 * time.Minute,
			},
			driver: fliptsql.SQLite,
		},
		{
			name: "postgres url",
			cfg: config.DatabaseConfig{
				URL: "postgres://postgres@localhost:5432/flipt?sslmode=disable",
			},
			driver: fliptsql.Postgres,
		},
		{
			name: "mysql url",
			cfg: config.DatabaseConfig{
				URL: "mysql://mysql@localhost:3306/flipt",
			},
			driver: fliptsql.MySQL,
		},
		{
			name: "invalid url",
			cfg: config.DatabaseConfig{
				URL: "tcp://a b",
			},
			wantErr: true,
		},
		{
			name: "unknown driver",
			cfg: config.DatabaseConfig{
				URL: "mongo://127.0.0.1",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		var (
			cfg     = tt.cfg
			driver  = tt.driver
			wantErr = tt.wantErr
		)

		t.Run(tt.name, func(t *testing.T) {
			db, d, err := fliptsql.Open(config.Config{
				Database: cfg,
			})

			if wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, db)

			defer db.Close()

			assert.Equal(t, driver, d)
		})
	}
}
