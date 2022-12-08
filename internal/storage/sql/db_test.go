package sql_test

import (
	"context"
	"os"
	"testing"
	"time"

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
	"go.uber.org/zap"
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
			name: "cockroachdb url",
			cfg: config.DatabaseConfig{
				URL: "cockroachdb://cockroachdb@localhost:26257/flipt?sslmode=disable",
			},
			driver: fliptsql.CockroachDB,
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
	db    *fliptsqltesting.Database
	store storage.Store
}

var dd string

func TestMain(m *testing.M) {
	dd = os.Getenv("FLIPT_TEST_DATABASE_PROTOCOL")
	os.Exit(m.Run())
}

func (s *DBTestSuite) SetupSuite() {
	setup := func() error {
		logger := zaptest.NewLogger(s.T(), zaptest.Level(zap.WarnLevel))

		db, err := fliptsqltesting.Open()
		if err != nil {
			return err
		}

		s.db = db

		var store storage.Store

		switch db.Driver {
		case fliptsql.SQLite:
			store = sqlite.NewStore(db.DB, logger)
		case fliptsql.Postgres, fliptsql.CockroachDB:
			store = postgres.NewStore(db.DB, logger)
		case fliptsql.MySQL:
			store = mysql.NewStore(db.DB, logger)
		}

		s.store = store
		return nil
	}

	s.Require().NoError(setup())
}

func (s *DBTestSuite) TearDownSuite() {
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if s.db != nil {
		s.db.Shutdown(shutdownCtx)
	}
}
