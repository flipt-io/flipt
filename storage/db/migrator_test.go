package db

import (
	"testing"

	"github.com/golang-migrate/migrate"
	stubDB "github.com/golang-migrate/migrate/database/stub"
	"github.com/golang-migrate/migrate/source"
	"github.com/golang-migrate/migrate/source/stub"
	stubSource "github.com/golang-migrate/migrate/source/stub"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMigratorCurrentVersion(t *testing.T) {
	s := &stubDB.Stub{}
	d, err := s.Open("")
	require.NoError(t, err)

	err = d.SetVersion(2, false)
	require.NoError(t, err)

	src := &stubSource.Stub{}
	srcDrv, err := src.Open("")
	require.NoError(t, err)

	m, err := migrate.NewWithInstance("stub", srcDrv, "", d)
	require.NoError(t, err)

	migrator := Migrator{
		migrator: m,
	}

	defer migrator.Close()

	v, err := migrator.CurrentVersion()
	assert.NoError(t, err)
	assert.Equal(t, uint(2), v)
}

func TestMigratorCurrentVersion_NilVersion(t *testing.T) {
	s := &stubDB.Stub{}
	d, err := s.Open("")
	require.NoError(t, err)

	src := &stubSource.Stub{}
	srcDrv, err := src.Open("")
	require.NoError(t, err)

	m, err := migrate.NewWithInstance("stub", srcDrv, "", d)
	require.NoError(t, err)

	migrator := Migrator{
		migrator: m,
	}

	defer migrator.Close()

	v, err := migrator.CurrentVersion()
	assert.EqualError(t, err, ErrMigrationsNilVersion.Error())
	assert.Equal(t, uint(0), v)
}

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

	srcDrv.(*stub.Stub).Migrations = stubMigrations

	m, err := migrate.NewWithInstance("stub", srcDrv, "", d)
	require.NoError(t, err)

	migrator := Migrator{
		migrator: m,
	}

	defer migrator.Close()

	err = migrator.Run()
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

	srcDrv.(*stub.Stub).Migrations = stubMigrations

	m, err := migrate.NewWithInstance("stub", srcDrv, "", d)
	require.NoError(t, err)

	migrator := Migrator{
		migrator: m,
	}

	defer migrator.Close()

	err = migrator.Run()
	assert.NoError(t, err)
}
