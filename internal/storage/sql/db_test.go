//nolint:gosec
package sql_test

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/internal/storage"
	fliptsql "go.flipt.io/flipt/internal/storage/sql"
	"go.flipt.io/flipt/internal/storage/sql/mysql"
	"go.flipt.io/flipt/internal/storage/sql/postgres"
	"go.flipt.io/flipt/internal/storage/sql/sqlite"
	fliptsqltesting "go.flipt.io/flipt/internal/storage/sql/testing"
	"go.flipt.io/flipt/rpc/flipt"
	"go.uber.org/zap/zaptest"

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
			name: "libsql url",
			cfg: config.DatabaseConfig{
				URL:             "libsql://file:/flipt.db",
				MaxOpenConn:     5,
				ConnMaxLifetime: 30 * time.Minute,
			},
			driver: fliptsql.LibSQL,
		},
		{
			name: "libsql with http",
			cfg: config.DatabaseConfig{
				URL:             "http://127.0.0.1:8000",
				MaxOpenConn:     5,
				ConnMaxLifetime: 30 * time.Minute,
			},
			driver: fliptsql.LibSQL,
		},
		{
			name: "libsql with https",
			cfg: config.DatabaseConfig{
				URL:             "https://turso.remote",
				MaxOpenConn:     5,
				ConnMaxLifetime: 30 * time.Minute,
			},
			driver: fliptsql.LibSQL,
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
			name: "cockroachdb url",
			cfg: config.DatabaseConfig{
				URL: "cockroachdb://cockroachdb@localhost:26257/flipt?sslmode=disable",
			},
			driver: fliptsql.CockroachDB,
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

func TestDBTestSuite(t *testing.T) {
	suite.Run(t, new(DBTestSuite))
}

type DBTestSuite struct {
	suite.Suite
	db        *fliptsqltesting.Database
	store     storage.Store
	namespace string
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

func (s *DBTestSuite) SetupSuite() {
	setup := func() error {
		logger := zaptest.NewLogger(s.T())

		db, err := fliptsqltesting.Open()
		if err != nil {
			return err
		}

		s.db = db

		builder := sq.StatementBuilder.RunWith(sq.NewStmtCacher(db.DB))

		var store storage.Store

		switch db.Driver {
		case fliptsql.SQLite, fliptsql.LibSQL:
			store = sqlite.NewStore(db.DB, builder, logger)
		case fliptsql.Postgres, fliptsql.CockroachDB:
			store = postgres.NewStore(db.DB, builder, logger)
		case fliptsql.MySQL:
			store = mysql.NewStore(db.DB, builder, logger)
		}

		namespace := randomString(6)

		if _, err := store.CreateNamespace(context.Background(), &flipt.CreateNamespaceRequest{
			Key: namespace,
		}); err != nil {
			return fmt.Errorf("failed to create namespace: %w", err)
		}

		s.namespace = namespace
		s.store = store
		return nil
	}

	s.Require().NoError(setup())
}

func randomString(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}

	return string(b)
}

func (s *DBTestSuite) TearDownSuite() {
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if s.db != nil {
		s.db.Shutdown(shutdownCtx)
	}
}
