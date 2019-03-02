package storage

import (
	"context"
	"database/sql"

	sq "github.com/Masterminds/squirrel"
	uuid "github.com/satori/go.uuid"

	proto "github.com/golang/protobuf/ptypes"
	flipt "github.com/markphelps/flipt/proto"
	sqlite3 "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
)

var _ FlagStore = &FlagStorage{}

// FlagStorage is a SQL FlagStore
type FlagStorage struct {
	logger  logrus.FieldLogger
	builder sq.StatementBuilderType
}

// NewFlagStorage creates a FlagStorage
func NewFlagStorage(logger logrus.FieldLogger, builder sq.StatementBuilderType) *FlagStorage {
	return &FlagStorage{
		logger:  logger.WithField("storage", "flag"),
		builder: builder,
	}
}

// GetFlag gets a flag
func (s *FlagStorage) GetFlag(ctx context.Context, r *flipt.GetFlagRequest) (*flipt.Flag, error) {
	s.logger.WithField("request", r).Debug("get flag")
	flag, err := s.flag(ctx, r.Key)
	s.logger.WithField("response", flag).Debug("get flag")
	return flag, err
}

func (s *FlagStorage) flag(ctx context.Context, key string) (*flipt.Flag, error) {
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
			return nil, ErrNotFoundf("flag %q", key)
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
func (s *FlagStorage) ListFlags(ctx context.Context, r *flipt.ListFlagRequest) ([]*flipt.Flag, error) {
	s.logger.WithField("request", r).Debug("list flags")

	var (
		flags []*flipt.Flag

		query = s.builder.Select("key, name, description, enabled, created_at, updated_at").
			From("flags").
			OrderBy("created_at ASC")
	)

	if r.Limit > 0 {
		query = query.Limit(uint64(r.Limit))
	}
	if r.Offset > 0 {
		query = query.Offset(uint64(r.Offset))
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

	s.logger.WithField("response", flags).Debug("list flags")
	return flags, rows.Err()
}

// CreateFlag creates a flag
func (s *FlagStorage) CreateFlag(ctx context.Context, r *flipt.CreateFlagRequest) (*flipt.Flag, error) {
	s.logger.WithField("request", r).Debug("create flag")

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

		query = s.builder.Insert("flags").
			Columns("key", "name", "description", "enabled", "created_at", "updated_at").
			Values(flag.Key, flag.Name, flag.Description, flag.Enabled, &timestamp{flag.CreatedAt}, &timestamp{flag.UpdatedAt})
	)

	if _, err := query.ExecContext(ctx); err != nil {
		if sqliteErr, ok := err.(sqlite3.Error); ok {
			if sqliteErr.Code == sqlite3.ErrConstraint {
				return nil, ErrInvalidf("flag %q is not unique", r.Key)
			}
		}
		return nil, err
	}

	s.logger.WithField("response", flag).Debug("create flag")
	return flag, nil
}

// UpdateFlag updates an existing flag
func (s *FlagStorage) UpdateFlag(ctx context.Context, r *flipt.UpdateFlagRequest) (*flipt.Flag, error) {
	s.logger.WithField("request", r).Debug("update flag")

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
		return nil, ErrNotFoundf("flag %q", r.Key)
	}

	flag, err := s.flag(ctx, r.Key)
	s.logger.WithField("response", flag).Debug("update flag")
	return flag, err
}

// DeleteFlag deletes a flag
func (s *FlagStorage) DeleteFlag(ctx context.Context, r *flipt.DeleteFlagRequest) error {
	s.logger.WithField("request", r).Debug("delete flag")

	_, err := s.builder.Delete("flags").
		Where(sq.Eq{"key": r.Key}).
		ExecContext(ctx)

	return err
}

// CreateVariant creates a variant
func (s *FlagStorage) CreateVariant(ctx context.Context, r *flipt.CreateVariantRequest) (*flipt.Variant, error) {
	s.logger.WithField("request", r).Debug("create variant")

	var (
		now = proto.TimestampNow()
		v   = &flipt.Variant{
			Id:          uuid.NewV4().String(),
			FlagKey:     r.FlagKey,
			Key:         r.Key,
			Name:        r.Name,
			Description: r.Description,
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		query = s.builder.Insert("variants").
			Columns("id", "flag_key", "key", "name", "description", "created_at", "updated_at").
			Values(v.Id, v.FlagKey, v.Key, v.Name, v.Description, &timestamp{v.CreatedAt}, &timestamp{v.UpdatedAt})
	)

	if _, err := query.ExecContext(ctx); err != nil {
		if sqliteErr, ok := err.(sqlite3.Error); ok {
			if sqliteErr.ExtendedCode == sqlite3.ErrConstraintForeignKey {
				return nil, ErrNotFoundf("flag %q", r.FlagKey)
			}
		}
		return nil, err
	}

	s.logger.WithField("response", v).Debug("create variant")
	return v, nil
}

// UpdateVariant updates an existing variant
func (s *FlagStorage) UpdateVariant(ctx context.Context, r *flipt.UpdateVariantRequest) (*flipt.Variant, error) {
	s.logger.WithField("request", r).Debug("update variant")

	res, err := s.builder.Update("variants").
		Set("key", r.Key).
		Set("name", r.Name).
		Set("description", r.Description).
		Set("updated_at", &timestamp{proto.TimestampNow()}).
		Where(sq.And{sq.Eq{"id": r.Id}, sq.Eq{"flag_key": r.FlagKey}}).
		ExecContext(ctx)

	if err != nil {
		return nil, err
	}

	count, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}

	if count != 1 {
		return nil, ErrNotFoundf("variant %q", r.Key)
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

	s.logger.WithField("response", v).Debug("update variant")
	return v, nil
}

// DeleteVariant deletes a variant
func (s *FlagStorage) DeleteVariant(ctx context.Context, r *flipt.DeleteVariantRequest) error {
	s.logger.WithField("request", r).Debug("delete variant")

	_, err := s.builder.Delete("variants").
		Where(sq.And{sq.Eq{"id": r.Id}, sq.Eq{"flag_key": r.FlagKey}}).
		ExecContext(ctx)

	return err
}

func (s *FlagStorage) variants(ctx context.Context, flag *flipt.Flag) (err error) {
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
