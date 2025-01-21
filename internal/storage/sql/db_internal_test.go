package sql

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/config"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		cfg     config.DatabaseConfig
		dsn     string
		driver  Driver
		options []Option
		wantErr bool
	}{
		{
			name: "sqlite url",
			cfg: config.DatabaseConfig{
				URL: "file:flipt.db",
			},
			driver: SQLite,
			dsn:    "flipt.db?_fk=true&cache=shared&mode=rwc",
		},
		{
			name: "sqlite",
			cfg: config.DatabaseConfig{
				Protocol: config.DatabaseSQLite,
				Host:     "flipt.db",
			},
			driver: SQLite,
			dsn:    "flipt.db?_fk=true&cache=shared&mode=rwc",
		},
		{
			name: "postgresql url",
			cfg: config.DatabaseConfig{
				URL: "postgresql://postgres@localhost:5432/flipt?sslmode=disable",
			},
			driver: Postgres,
			dsn:    "postgres://postgres@localhost:5432/flipt?default_query_exec_mode=simple_protocol&sslmode=disable",
		},
		{
			name: "postgresql url prepared statements enabled",
			cfg: config.DatabaseConfig{
				URL:                       "postgresql://postgres@localhost:5432/flipt?sslmode=disable",
				PreparedStatementsEnabled: true,
			},
			driver: Postgres,
			dsn:    "postgres://postgres@localhost:5432/flipt?sslmode=disable",
		},
		{
			name: "postgresql no disable sslmode",
			cfg: config.DatabaseConfig{
				URL: "postgresql://postgres@localhost:5432/flipt",
			},
			driver: Postgres,
			dsn:    "postgres://postgres@localhost:5432/flipt?default_query_exec_mode=simple_protocol",
		},
		{
			name: "postgres url prepared statements enabled",
			cfg: config.DatabaseConfig{
				URL:                       "postgres://postgres@localhost:5432/flipt?sslmode=disable",
				PreparedStatementsEnabled: true,
			},
			driver: Postgres,
			dsn:    "postgres://postgres@localhost:5432/flipt?sslmode=disable",
		},
		{
			name: "postgres url",
			cfg: config.DatabaseConfig{
				URL: "postgres://postgres@localhost:5432/flipt?sslmode=disable",
			},
			driver: Postgres,
			dsn:    "postgres://postgres@localhost:5432/flipt?default_query_exec_mode=simple_protocol&sslmode=disable",
		},
		{
			name: "postgres no disable sslmode",
			cfg: config.DatabaseConfig{
				URL: "postgres://postgres@localhost:5432/flipt",
			},
			driver: Postgres,
			dsn:    "postgres://postgres@localhost:5432/flipt?default_query_exec_mode=simple_protocol",
		},
		{
			name: "postgres disable sslmode via opts",
			cfg: config.DatabaseConfig{
				Protocol: config.DatabasePostgres,
				Name:     "flipt",
				Host:     "localhost",
				Port:     5432,
				User:     "postgres",
			},
			options: []Option{WithSSLDisabled},
			driver:  Postgres,
			dsn:     "postgres://postgres@localhost:5432/flipt?default_query_exec_mode=simple_protocol&sslmode=disable",
		},
		{
			name: "postgres no port",
			cfg: config.DatabaseConfig{
				Protocol: config.DatabasePostgres,
				Name:     "flipt",
				Host:     "localhost",
				User:     "postgres",
			},
			driver: Postgres,
			dsn:    "postgres://postgres@localhost:5432/flipt?default_query_exec_mode=simple_protocol",
		},
		{
			name: "postgres no password",
			cfg: config.DatabaseConfig{
				Protocol: config.DatabasePostgres,
				Name:     "flipt",
				Host:     "localhost",
				Port:     5432,
				User:     "postgres",
			},
			driver: Postgres,
			dsn:    "postgres://postgres@localhost:5432/flipt?default_query_exec_mode=simple_protocol",
		},
		{
			name: "postgres with password",
			cfg: config.DatabaseConfig{
				Protocol: config.DatabasePostgres,
				Name:     "flipt",
				Host:     "localhost",
				Port:     5432,
				User:     "postgres",
				Password: "foo",
			},
			driver: Postgres,
			dsn:    "postgres://postgres:foo@localhost:5432/flipt?default_query_exec_mode=simple_protocol",
		},
		{
			name: "mysql url",
			cfg: config.DatabaseConfig{
				URL: "mysql://mysql@localhost:3306/flipt",
			},
			driver: MySQL,
			dsn:    "mysql@tcp(localhost:3306)/flipt?multiStatements=true&parseTime=true&sql_mode=ANSI",
		},
		{
			name: "mysql no ANSI sql mode via opts",
			cfg: config.DatabaseConfig{
				Protocol: config.DatabaseMySQL,
				Name:     "flipt",
				Host:     "localhost",
				Port:     3306,
				User:     "mysql",
			},
			options: []Option{
				WithMigrate,
			},
			driver: MySQL,
			dsn:    "mysql@tcp(localhost:3306)/flipt?multiStatements=true&parseTime=true",
		},
		{
			name: "mysql no port",
			cfg: config.DatabaseConfig{
				Protocol: config.DatabaseMySQL,
				Name:     "flipt",
				Host:     "localhost",
				User:     "mysql",
				Password: "foo",
			},
			driver: MySQL,
			dsn:    "mysql:foo@tcp(localhost:3306)/flipt?multiStatements=true&parseTime=true&sql_mode=ANSI",
		},
		{
			name: "mysql no password",
			cfg: config.DatabaseConfig{
				Protocol: config.DatabaseMySQL,
				Name:     "flipt",
				Host:     "localhost",
				Port:     3306,
				User:     "mysql",
			},
			driver: MySQL,
			dsn:    "mysql@tcp(localhost:3306)/flipt?multiStatements=true&parseTime=true&sql_mode=ANSI",
		},
		{
			name: "mysql with password",
			cfg: config.DatabaseConfig{
				Protocol: config.DatabaseMySQL,
				Name:     "flipt",
				Host:     "localhost",
				Port:     3306,
				User:     "mysql",
				Password: "foo",
			},
			driver: MySQL,
			dsn:    "mysql:foo@tcp(localhost:3306)/flipt?multiStatements=true&parseTime=true&sql_mode=ANSI",
		},
		{
			name: "invalid url",
			cfg: config.DatabaseConfig{
				URL: "http://a b",
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
		{
			name: "postgres multi hosts url",
			cfg: config.DatabaseConfig{
				URL:                       "postgres://user:pass@host1:5432,host2:2345/flipt?application_name=flipt&target_session_attrs=primary",
				PreparedStatementsEnabled: false,
			},
			driver: Postgres,
			dsn:    "postgres://user:pass@host1:5432,host2:2345/flipt?application_name=flipt&default_query_exec_mode=simple_protocol&target_session_attrs=primary",
		},
	}

	for _, tt := range tests {
		tt := tt

		var (
			cfg     = tt.cfg
			driver  = tt.driver
			url     = tt.dsn
			wantErr = tt.wantErr
			opts    Options
		)

		for _, opt := range tt.options {
			opt(&opts)
		}

		t.Run(tt.name, func(t *testing.T) {
			d, u, err := parse(config.Config{
				Database: cfg,
			}, opts)

			if wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, driver, d)
			assert.Equal(t, url, u.DSN)
		})
	}
}
