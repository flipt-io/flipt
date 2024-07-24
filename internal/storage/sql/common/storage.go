package common

import (
	"context"
	"database/sql"
	"time"

	sq "github.com/Masterminds/squirrel"
	"go.flipt.io/flipt/internal/storage"
	fliptsql "go.flipt.io/flipt/internal/storage/sql"
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

func (s *Store) GetVersion(ctx context.Context, ns storage.NamespaceRequest) (string, error) {
	var stateModifiedAt fliptsql.NullableTimestamp

	err := s.builder.
		Select("state_modified_at").
		From("namespaces").
		Where(sq.Eq{"\"key\"": ns.Namespace()}).
		Limit(1).
		RunWith(s.db).
		QueryRowContext(ctx).
		Scan(&stateModifiedAt)

	if err != nil {
		return "", err
	}

	if !stateModifiedAt.IsValid() {
		return "", nil
	}

	return stateModifiedAt.Timestamp.String(), nil
}

func (s *Store) setVersion(ctx context.Context, namespace string) error {
	_, err := s.builder.
		Update("namespaces").
		Set("state_modified_at", time.Now().UTC()).
		Where(sq.Eq{"\"key\"": namespace}).
		ExecContext(ctx)
	return err
}
