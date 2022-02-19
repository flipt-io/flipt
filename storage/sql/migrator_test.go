package sql

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/golang-migrate/migrate"
	stubDB "github.com/golang-migrate/migrate/database/stub"
	"github.com/golang-migrate/migrate/source"
	stubSource "github.com/golang-migrate/migrate/source/stub"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		l, _     = test.NewNullLogger()
		logger   = logrus.NewEntry(l)
		migrator = Migrator{
			migrator: m,
			logger:   logger,
		}
	)

	defer migrator.Close()

	err = migrator.Run(false)
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
		l, _     = test.NewNullLogger()
		logger   = logrus.NewEntry(l)
		migrator = Migrator{
			migrator: m,
			logger:   logger,
		}
	)

	defer migrator.Close()

	err = migrator.Run(false)
	assert.NoError(t, err)
}

func TestMigratorExpectedVersions(t *testing.T) {
	for db, driver := range stringToDriver {
		migrations, err := ioutil.ReadDir(filepath.Join("../../config/migrations", db))
		require.NoError(t, err)

		count := len(migrations)
		require.True(t, count > 0, "no migrations found for %s", db)

		// 1 is the up migration and 1 is the down migration
		// so we should have count/2 migrations
		// and they start at 0
		actual := uint((count / 2) - 1)
		assert.Equal(t, actual, expectedVersions[driver], "expectedVersions for %s should be set to %d. you need to increment expectedVersions after adding a new migration", db, actual)
	}
}
