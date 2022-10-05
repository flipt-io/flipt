package common

import (
	"database/sql"

	sq "github.com/Masterminds/squirrel"
	"go.uber.org/zap"
)

type Store struct {
	builder sq.StatementBuilderType
	db      *sql.DB
	logger  *zap.Logger
}

func NewStore(db *sql.DB, builder sq.StatementBuilderType, logger *zap.Logger) *Store {
	return &Store{
		db:      db,
		builder: builder,
		logger:  logger,
	}
}

type PageToken struct {
	Key    string `json:"key,omitempty"`
	Offset uint64 `json:"offset,omitempty"`
}
