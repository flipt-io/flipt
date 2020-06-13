package mysql

import (
	"database/sql"

	sq "github.com/Masterminds/squirrel"
	"github.com/markphelps/flipt/storage"
	"github.com/markphelps/flipt/storage/db/common"
)

var _ storage.Store = &Store{}

func NewStore(db *sql.DB) *Store {
	builder := sq.StatementBuilder.RunWith(sq.NewStmtCacher(db))

	return &Store{
		Store: common.NewStore(db, builder),
	}
}

type Store struct {
	*common.Store
}
