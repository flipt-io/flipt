package sql

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	stubDB "github.com/golang-migrate/migrate/v4/database/stub"
	"github.com/golang-migrate/migrate/v4/source"
	stubSource "github.com/golang-migrate/migrate/v4/source/stub"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestMigratorRun(t *testing.T) {
	s := &stubDB.Stub{}
	d, err := s.Open("")
	require.NoError(t, err)

	stubMigrations := source.NewMigrations()
	stubMigrations.Append(&source.Migration{Version: 1, Direction: source.Up, Identifier: "CREATE 1"})
	stubMigrations.Append(&source.Migration{Version: 1, Direction: source.Down, Identifier: "DROP 1"})

	src := &stubSource.Stub{}
	srcDrv, err := src.Open("")
	require.NoError(t, err)

	srcDrv.(*stubSource.Stub).Migrations = stubMigrations

	m, err := migrate.NewWithInstance("stub", srcDrv, "", d)
	require.NoError(t, err)

	var (
		logger   = zaptest.NewLogger(t)
		migrator = Migrator{
			migrator: m,
			logger:   logger,
		}
	)

	defer migrator.Close()

	err = migrator.Up(false)
	assert.NoError(t, err)
}

func TestMigratorRun_NoChange(t *testing.T) {
	s := &stubDB.Stub{}
	d, err := s.Open("")
	require.NoError(t, err)

	err = d.SetVersion(1, false)
	require.NoError(t, err)

	stubMigrations := source.NewMigrations()
	stubMigrations.Append(&source.Migration{Version: 1, Direction: source.Up, Identifier: "CREATE 1"})
	stubMigrations.Append(&source.Migration{Version: 1, Direction: source.Down, Identifier: "DROP 1"})

	src := &stubSource.Stub{}
	srcDrv, err := src.Open("")
	require.NoError(t, err)

	srcDrv.(*stubSource.Stub).Migrations = stubMigrations

	m, err := migrate.NewWithInstance("stub", srcDrv, "", d)
	require.NoError(t, err)

	var (
		logger   = zaptest.NewLogger(t)
		migrator = Migrator{
			migrator: m,
			logger:   logger,
		}
	)

	defer migrator.Close()

	err = migrator.Up(false)
	assert.NoError(t, err)
}

func TestMigratorExpectedVersions(t *testing.T) {
	for _, driver := range stringToDriver {
		migrations, err := os.ReadDir(filepath.Join("../../../config/migrations", driver.Migrations()))
		require.NoError(t, err)

		count := len(migrations)
		require.Positive(t, count, "no migrations found for %s", driver)

		// migrations start at 0
		actual := uint(count - 1)
		assert.Equal(t, actual, expectedVersions[driver], "expectedVersions for %s should be set to %d. you need to increment expectedVersions after adding a new migration", driver, actual)
	}
}
