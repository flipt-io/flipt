package db

import (
	"testing"

	"github.com/golang-migrate/migrate"
	stubDB "github.com/golang-migrate/migrate/database/stub"
	stubSource "github.com/golang-migrate/migrate/source/stub"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

	v, err := migrator.CurrentVersion()
	assert.EqualError(t, err, ErrMigrationsNilVersion.Error())
	assert.Equal(t, uint(0), v)
}

func TestMigratorCurrentVersion(t *testing.T) {
	d := &stubDB.Stub{}
	d.SetVersion(2, false)

	src := &stubSource.Stub{}
	srcDrv, err := src.Open("")
	require.NoError(t, err)

	m, err := migrate.NewWithInstance("stub", srcDrv, "", d)
	require.NoError(t, err)

	migrator := Migrator{
		migrator: m,
	}

	v, err := migrator.CurrentVersion()
	assert.NoError(t, err)
	assert.Equal(t, uint(2), v)
}
