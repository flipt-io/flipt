package db

import (
	"context"
	"database/sql"

	sq "github.com/Masterminds/squirrel"
	"github.com/gofrs/uuid"

	proto "github.com/golang/protobuf/ptypes"
	errs "github.com/markphelps/flipt/errors"
	flipt "github.com/markphelps/flipt/rpc"
	"github.com/markphelps/flipt/storage"
)

var _ storage.FlagStore = &FlagStore{}

// FlagStore is a SQL FlagStore
type FlagStore struct {
	builder sq.StatementBuilderType
}

// NewFlagStore creates a FlagStore
func NewFlagStore(builder sq.StatementBuilderType) *FlagStore {
	return &FlagStore{
		builder: builder,
	}
}

// GetFlag gets a flag
func (s *FlagStore) GetFlag(ctx context.Context, key string) (*flipt.Flag, error) {
	var (
		createdAt timestamp
		updatedAt timestamp

		flag = &flipt.Flag{}

		query = s.builder.Select("key, name, description, enabled, created_at, updated_at").
			From("flags").
			Where(sq.Eq{"key": key})

		err = query.QueryRowContext(ctx).Scan(
			&flag.Key,
			&flag.Name,
			&flag.Description,
			&flag.Enabled,
			&createdAt,
			&updatedAt)
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errs.ErrNotFoundf("flag %q", key)
		}

		return nil, err
	}

	flag.CreatedAt = createdAt.Timestamp
	flag.UpdatedAt = updatedAt.Timestamp

	if err := s.variants(ctx, flag); err != nil {
		return nil, err
	}

	return flag, nil
}

// ListFlags lists all flags
func (s *FlagStore) ListFlags(ctx context.Context, opts ...storage.QueryOption) ([]*flipt.Flag, error) {
	var (
		flags []*flipt.Flag

		query = s.builder.Select("key, name, description, enabled, created_at, updated_at").
			From("flags").
			OrderBy("created_at ASC")
	)

	params := &storage.QueryParams{}

	for _, opt := range opts {
		opt(params)
	}

	if params.Limit > 0 {
		query = query.Limit(params.Limit)
	}

	if params.Offset > 0 {
		query = query.Offset(params.Offset)
	}

	rows, err := query.QueryContext(ctx)
	if err != nil {
		return nil, err
	}

	defer func() {
		if cerr := rows.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	for rows.Next() {
		var (
			flag      = &flipt.Flag{}
			createdAt timestamp
			updatedAt timestamp
		)

		if err := rows.Scan(
			&flag.Key,
			&flag.Name,
			&flag.Description,
			&flag.Enabled,
			&createdAt,
			&updatedAt); err != nil {
			return nil, err
		}

		flag.CreatedAt = createdAt.Timestamp
		flag.UpdatedAt = updatedAt.Timestamp

		if err := s.variants(ctx, flag); err != nil {
			return nil, err
		}

		flags = append(flags, flag)
	}

	return flags, rows.Err()
}

// CreateFlag creates a flag
func (s *FlagStore) CreateFlag(ctx context.Context, r *flipt.CreateFlagRequest) (*flipt.Flag, error) {
	var (
		now  = proto.TimestampNow()
		flag = &flipt.Flag{
			Key:         r.Key,
			Name:        r.Name,
			Description: r.Description,
			Enabled:     r.Enabled,
			CreatedAt:   now,
			UpdatedAt:   now,
		}
	)

	if _, err := s.builder.Insert("flags").
		Columns("key", "name", "description", "enabled", "created_at", "updated_at").
		Values(flag.Key, flag.Name, flag.Description, flag.Enabled, &timestamp{flag.CreatedAt}, &timestamp{flag.UpdatedAt}).
		ExecContext(ctx); err != nil {
		return nil, err
	}

	return flag, nil
}

// UpdateFlag updates an existing flag
func (s *FlagStore) UpdateFlag(ctx context.Context, r *flipt.UpdateFlagRequest) (*flipt.Flag, error) {
	query := s.builder.Update("flags").
		Set("name", r.Name).
		Set("description", r.Description).
		Set("enabled", r.Enabled).
		Set("updated_at", &timestamp{proto.TimestampNow()}).
		Where(sq.Eq{"key": r.Key})

	res, err := query.ExecContext(ctx)
	if err != nil {
		return nil, err
	}

	count, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}

	if count != 1 {
		return nil, errs.ErrNotFoundf("flag %q", r.Key)
	}

	return s.GetFlag(ctx, r.Key)
}

// DeleteFlag deletes a flag
func (s *FlagStore) DeleteFlag(ctx context.Context, r *flipt.DeleteFlagRequest) error {
	_, err := s.builder.Delete("flags").
		Where(sq.Eq{"key": r.Key}).
		ExecContext(ctx)

	return err
}

// CreateVariant creates a variant
func (s *FlagStore) CreateVariant(ctx context.Context, r *flipt.CreateVariantRequest) (*flipt.Variant, error) {
	var (
		now = proto.TimestampNow()
		v   = &flipt.Variant{
			Id:          uuid.Must(uuid.NewV4()).String(),
			FlagKey:     r.FlagKey,
			Key:         r.Key,
			Name:        r.Name,
			Description: r.Description,
			CreatedAt:   now,
			UpdatedAt:   now,
		}
	)

	if _, err := s.builder.Insert("variants").
		Columns("id", "flag_key", "key", "name", "description", "created_at", "updated_at").
		Values(v.Id, v.FlagKey, v.Key, v.Name, v.Description, &timestamp{v.CreatedAt}, &timestamp{v.UpdatedAt}).
		ExecContext(ctx); err != nil {
		return nil, err
	}

	return v, nil
}

// UpdateVariant updates an existing variant
func (s *FlagStore) UpdateVariant(ctx context.Context, r *flipt.UpdateVariantRequest) (*flipt.Variant, error) {
	query := s.builder.Update("variants").
		Set("key", r.Key).
		Set("name", r.Name).
		Set("description", r.Description).
		Set("updated_at", &timestamp{proto.TimestampNow()}).
		Where(sq.And{sq.Eq{"id": r.Id}, sq.Eq{"flag_key": r.FlagKey}})

	res, err := query.ExecContext(ctx)
	if err != nil {
		return nil, err
	}

	count, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}

	if count != 1 {
		return nil, errs.ErrNotFoundf("variant %q", r.Key)
	}

	var (
		createdAt timestamp
		updatedAt timestamp

		v = &flipt.Variant{}
	)

	if err := s.builder.Select("id, key, flag_key, name, description, created_at, updated_at").
		From("variants").
		Where(sq.And{sq.Eq{"id": r.Id}, sq.Eq{"flag_key": r.FlagKey}}).
		QueryRowContext(ctx).
		Scan(&v.Id, &v.Key, &v.FlagKey, &v.Name, &v.Description, &createdAt, &updatedAt); err != nil {
		return nil, err
	}

	v.CreatedAt = createdAt.Timestamp
	v.UpdatedAt = updatedAt.Timestamp

	return v, nil
}

// DeleteVariant deletes a variant
func (s *FlagStore) DeleteVariant(ctx context.Context, r *flipt.DeleteVariantRequest) error {
	_, err := s.builder.Delete("variants").
		Where(sq.And{sq.Eq{"id": r.Id}, sq.Eq{"flag_key": r.FlagKey}}).
		ExecContext(ctx)

	return err
}

func (s *FlagStore) variants(ctx context.Context, flag *flipt.Flag) (err error) {
	query := s.builder.Select("id, flag_key, key, name, description, created_at, updated_at").
		From("variants").
		Where(sq.Eq{"flag_key": flag.Key}).
		OrderBy("created_at ASC")

	rows, err := query.QueryContext(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if cerr := rows.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	for rows.Next() {
		var (
			variant              flipt.Variant
			createdAt, updatedAt timestamp
		)

		if err := rows.Scan(
			&variant.Id,
			&variant.FlagKey,
			&variant.Key,
			&variant.Name,
			&variant.Description,
			&createdAt,
			&updatedAt); err != nil {
			return err
		}

		variant.CreatedAt = createdAt.Timestamp
		variant.UpdatedAt = updatedAt.Timestamp
		flag.Variants = append(flag.Variants, &variant)
	}

	return rows.Err()
}
