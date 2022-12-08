package memory

import (
	"testing"

	oplocktesting "go.flipt.io/flipt/internal/storage/oplock/testing"
	storagesql "go.flipt.io/flipt/internal/storage/sql"
	sqltesting "go.flipt.io/flipt/internal/storage/sql/testing"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

func Test_Harness(t *testing.T) {
	logger := zaptest.NewLogger(t, zaptest.Level(zap.WarnLevel))
	db, err := sqltesting.Open()
	if err != nil {
		t.Fatal(err)
	}

	oplocktesting.Harness(
		t,
		New(
			logger,
			db.Driver,
			storagesql.BuilderFor(db.DB, db.Driver),
		))
}
