package common

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	errs "go.flipt.io/flipt/errors"
	"go.flipt.io/flipt/internal/storage"
	fliptsql "go.flipt.io/flipt/internal/storage/sql"
	flipt "go.flipt.io/flipt/rpc/flipt"
)

func (s *Store) GetNamespace(ctx context.Context, p storage.NamespaceRequest) (*flipt.Namespace, error) {
	var (
		createdAt fliptsql.Timestamp
		updatedAt fliptsql.Timestamp

		namespace = &flipt.Namespace{}

		err = s.builder.Select("\"key\", name, description, protected, created_at, updated_at").
			From("namespaces").
			Where(sq.Eq{"\"key\"": p.Namespace()}).
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
			return nil, errs.ErrNotFoundf("namespace %q", p)
		}

		return nil, err
	}

	namespace.CreatedAt = createdAt.Timestamp
	namespace.UpdatedAt = updatedAt.Timestamp

	return namespace, nil
}

func (s *Store) ListNamespaces(ctx context.Context, req *storage.ListRequest[storage.ReferenceRequest]) (storage.ResultSet[*flipt.Namespace], error) {
	var (
		namespaces []*flipt.Namespace
		results    = storage.ResultSet[*flipt.Namespace]{}

		query = s.builder.Select("\"key\", name, description, protected, created_at, updated_at").
			From("namespaces").
			OrderBy(fmt.Sprintf("created_at %s", req.QueryParams.Order))
	)

	if req.QueryParams.Limit > 0 {
		query = query.Limit(req.QueryParams.Limit + 1)
	}

	var offset uint64

	if req.QueryParams.PageToken != "" {
		token, err := decodePageToken(s.logger, req.QueryParams.PageToken)
		if err != nil {
			return results, err
		}

		offset = token.Offset
		query = query.Offset(offset)
	} else if req.QueryParams.Offset > 0 {
		offset = req.QueryParams.Offset
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

	if len(namespaces) > int(req.QueryParams.Limit) && req.QueryParams.Limit > 0 {
		next = namespaces[len(namespaces)-1]
		namespaces = namespaces[:req.QueryParams.Limit]
	}

	results.Results = namespaces

	if next != nil {
		out, err := json.Marshal(PageToken{Key: next.Key, Offset: offset + uint64(len(namespaces))})
		if err != nil {
			return results, fmt.Errorf("encoding page token %w", err)
		}
		results.NextPageToken = base64.StdEncoding.EncodeToString(out)
	}

	return results, nil
}

func (s *Store) CountNamespaces(ctx context.Context, _ storage.ReferenceRequest) (uint64, error) {
	var count uint64

	if err := s.builder.Select("COUNT(*)").
		From("namespaces").
		QueryRowContext(ctx).
		Scan(&count); err != nil {
		return 0, err
	}

	return count, nil
}

func (s *Store) CreateNamespace(ctx context.Context, r *flipt.CreateNamespaceRequest) (_ *flipt.Namespace, err error) {
	defer func() {
		if err == nil {
			err = s.setVersion(ctx, r.Key)
		}
	}()

	var (
		now       = flipt.Now()
		namespace = &flipt.Namespace{
			Key:         r.Key,
			Name:        r.Name,
			Description: r.Description,
			CreatedAt:   now,
			UpdatedAt:   now,
		}
	)

	if _, err := s.builder.Insert("namespaces").
		Columns("\"key\"", "name", "description", "created_at", "updated_at").
		Values(
			namespace.Key,
			namespace.Name,
			namespace.Description,
			&fliptsql.Timestamp{Timestamp: namespace.CreatedAt},
			&fliptsql.Timestamp{Timestamp: namespace.UpdatedAt},
		).
		ExecContext(ctx); err != nil {
		return nil, err
	}

	return namespace, nil
}

func (s *Store) UpdateNamespace(ctx context.Context, r *flipt.UpdateNamespaceRequest) (_ *flipt.Namespace, err error) {
	defer func() {
		if err == nil {
			err = s.setVersion(ctx, r.Key)
		}
	}()

	query := s.builder.Update("namespaces").
		Set("name", r.Name).
		Set("description", r.Description).
		Set("updated_at", &fliptsql.Timestamp{Timestamp: flipt.Now()}).
		Where(sq.Eq{"\"key\"": r.Key})

	res, err := query.ExecContext(ctx)
	if err != nil {
		return nil, err
	}

	count, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}

	p := storage.NewNamespace(r.Key)

	if count != 1 {
		return nil, errs.ErrNotFoundf("namespace %q", p)
	}

	return s.GetNamespace(ctx, p)
}

func (s *Store) DeleteNamespace(ctx context.Context, r *flipt.DeleteNamespaceRequest) (err error) {

	_, err = s.builder.Delete("namespaces").
		Where(sq.Eq{"\"key\"": r.Key}).
		ExecContext(ctx)

	return err
}
