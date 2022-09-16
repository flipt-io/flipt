package common

import (
	"database/sql"
	"time"

	sq "github.com/Masterminds/squirrel"
)

type Store struct {
	builder sq.StatementBuilderType
	db      *sql.DB
}

func NewStore(db *sql.DB, builder sq.StatementBuilderType) *Store {
	return &Store{
		db:      db,
		builder: builder,
	}
}

type pageToken struct {
	Key       string    `json:"key,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"`
}
