package common

import (
	"context"
	"database/sql"
	"time"

	sq "github.com/Masterminds/squirrel"
	"go.flipt.io/flipt/internal/storage"
	"go.uber.org/zap"
)

var _ storage.Store = &Store{}

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

func (s *Store) String() string {
	return ""
}

func (s *Store) GetVersion(ctx context.Context) (string, error) {
	var version string
	err := s.builder.
		Select("version").
		From("metadata").
		RunWith(s.db).
		QueryRowContext(ctx).
		Scan(&version)
	return version, err
}

func (s *Store) setVersion(ctx context.Context) error {
	version := time.Now().UTC().Format(time.RFC3339)
	_, err := s.builder.
		Update("metadata").
		Set("version", version).
		ExecContext(ctx)
	return err
}
