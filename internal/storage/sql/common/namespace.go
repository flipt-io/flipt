package common

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	errs "go.flipt.io/flipt/errors"
	"go.flipt.io/flipt/internal/storage"
	fliptsql "go.flipt.io/flipt/internal/storage/sql"
	flipt "go.flipt.io/flipt/rpc/flipt"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Store) GetNamespace(ctx context.Context, key string) (*flipt.Namespace, error) {
	var (
		createdAt fliptsql.Timestamp
		updatedAt fliptsql.Timestamp

		namespace = &flipt.Namespace{}

		err = s.builder.Select("\"key\", name, description, protected, created_at, updated_at").
			From("namespaces").
			Where(sq.Eq{"\"key\"": key}).
			QueryRowContext(ctx).
			Scan(
				&namespace.Key,
				&namespace.Name,
				&namespace.Description,
				&namespace.Protected,
				&createdAt,
				&updatedAt)
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errs.ErrNotFoundf(`namespace "%s"`, key)
		}

		return nil, err
	}

	namespace.CreatedAt = createdAt.Timestamp
	namespace.UpdatedAt = updatedAt.Timestamp

	return namespace, nil
}

func (s *Store) ListNamespaces(ctx context.Context, opts ...storage.QueryOption) (storage.ResultSet[*flipt.Namespace], error) {
	params := &storage.QueryParams{}

	for _, opt := range opts {
		opt(params)
	}

	var (
		namespaces []*flipt.Namespace
		results    = storage.ResultSet[*flipt.Namespace]{}

		query = s.builder.Select("\"key\", name, description, protected, created_at, updated_at").
			From("namespaces").
			OrderBy(fmt.Sprintf("created_at %s", params.Order))
	)

	if params.Limit > 0 {
		query = query.Limit(params.Limit + 1)
	}

	var offset uint64

	if params.PageToken != "" {
		var token PageToken

		if err := json.Unmarshal([]byte(params.PageToken), &token); err != nil {
			return results, fmt.Errorf("decoding page token %w", err)
		}

		offset = token.Offset
		query = query.Offset(offset)
	} else if params.Offset > 0 {
		offset = params.Offset
		query = query.Offset(offset)
	}

	rows, err := query.QueryContext(ctx)
	if err != nil {
		return results, err
	}

	defer func() {
		if cerr := rows.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	for rows.Next() {
		var (
			namespace = &flipt.Namespace{}

			createdAt fliptsql.Timestamp
			updatedAt fliptsql.Timestamp
		)

		if err := rows.Scan(
			&namespace.Key,
			&namespace.Name,
			&namespace.Description,
			&namespace.Protected,
			&createdAt,
			&updatedAt,
		); err != nil {
			return results, err
		}

		namespace.CreatedAt = createdAt.Timestamp
		namespace.UpdatedAt = updatedAt.Timestamp

		namespaces = append(namespaces, namespace)
	}

	if err := rows.Err(); err != nil {
		return results, err
	}

	if err := rows.Close(); err != nil {
		return results, err
	}

	var next *flipt.Namespace

	if len(namespaces) > int(params.Limit) && params.Limit > 0 {
		next = namespaces[len(namespaces)-1]
		namespaces = namespaces[:params.Limit]
	}

	results.Results = namespaces

	if next != nil {
		out, err := json.Marshal(PageToken{Key: next.Key, Offset: offset + uint64(len(namespaces))})
		if err != nil {
			return results, fmt.Errorf("encoding page token %w", err)
		}
		results.NextPageToken = string(out)
	}

	return results, nil
}

func (s *Store) CountNamespaces(ctx context.Context) (uint64, error) {
	var count uint64

	if err := s.builder.Select("COUNT(*)").
		From("namespaces").
		QueryRowContext(ctx).
		Scan(&count); err != nil {
		return 0, err
	}

	return count, nil
}

func (s *Store) CreateNamespace(ctx context.Context, r *flipt.CreateNamespaceRequest) (*flipt.Namespace, error) {
	var (
		now       = timestamppb.Now()
		namespace = &flipt.Namespace{
			Key:         r.Key,
			Name:        r.Name,
			Description: r.Description,
			Protected:   r.Protected,
			CreatedAt:   now,
			UpdatedAt:   now,
		}
	)

	if _, err := s.builder.Insert("namespaces").
		Columns("\"key\"", "name", "description", "protected", "created_at", "updated_at").
		Values(
			namespace.Key,
			namespace.Name,
			namespace.Description,
			namespace.Protected,
			&fliptsql.Timestamp{Timestamp: namespace.CreatedAt},
			&fliptsql.Timestamp{Timestamp: namespace.UpdatedAt},
		).
		ExecContext(ctx); err != nil {
		return nil, err
	}

	return namespace, nil
}

func (s *Store) UpdateNamespace(ctx context.Context, r *flipt.UpdateNamespaceRequest) (*flipt.Namespace, error) {
	query := s.builder.Update("namespaces").
		Set("name", r.Name).
		Set("description", r.Description).
		Set("updated_at", &fliptsql.Timestamp{Timestamp: timestamppb.Now()}).
		Where(sq.Eq{"\"key\"": r.Key})

	res, err := query.ExecContext(ctx)
	if err != nil {
		return nil, err
	}

	count, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}

	if count != 1 {
		return nil, errs.ErrNotFoundf(`namespace "%s"`, r.Key)
	}

	return s.GetNamespace(ctx, r.Key)
}

func (s *Store) DeleteNamespace(ctx context.Context, r *flipt.DeleteNamespaceRequest) error {
	_, err := s.builder.Delete("namespaces").
		Where(sq.Eq{"\"key\"": r.Key}).
		ExecContext(ctx)

	return err
}
