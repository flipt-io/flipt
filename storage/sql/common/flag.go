package common

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	sq "github.com/Masterminds/squirrel"
	"github.com/gofrs/uuid"

	errs "github.com/markphelps/flipt/errors"
	flipt "github.com/markphelps/flipt/rpc/flipt"
	"github.com/markphelps/flipt/storage"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func compactJSONString(jsonString string) (string, error) {
	var buf bytes.Buffer
	if err := json.Compact(&buf, []byte(jsonString)); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func emptyAsNil(str string) *string {
	if str == "" {
		return nil
	}
	return &str
}

// GetFlag gets a flag
func (s *Store) GetFlag(ctx context.Context, key string) (*flipt.Flag, error) {
	var (
		createdAt timestamp
		updatedAt timestamp

		flag = &flipt.Flag{}

		query = s.builder.Select("\"key\", name, description, enabled, created_at, updated_at").
			From("flags").
			Where(sq.Eq{"\"key\"": key})

		err = query.QueryRowContext(ctx).Scan(
			&flag.Key,
			&flag.Name,
			&flag.Description,
			&flag.Enabled,
			&createdAt,
			&updatedAt)
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
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
func (s *Store) ListFlags(ctx context.Context, opts ...storage.QueryOption) ([]*flipt.Flag, error) {
	var (
		flags []*flipt.Flag

		query = s.builder.Select("\"key\", name, description, enabled, created_at, updated_at").
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
func (s *Store) CreateFlag(ctx context.Context, r *flipt.CreateFlagRequest) (*flipt.Flag, error) {
	var (
		now  = timestamppb.Now()
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
		Columns("\"key\"", "name", "description", "enabled", "created_at", "updated_at").
		Values(flag.Key, flag.Name, flag.Description, flag.Enabled, &timestamp{flag.CreatedAt}, &timestamp{flag.UpdatedAt}).
		ExecContext(ctx); err != nil {
		return nil, err
	}

	return flag, nil
}

// UpdateFlag updates an existing flag
func (s *Store) UpdateFlag(ctx context.Context, r *flipt.UpdateFlagRequest) (*flipt.Flag, error) {
	query := s.builder.Update("flags").
		Set("name", r.Name).
		Set("description", r.Description).
		Set("enabled", r.Enabled).
		Set("updated_at", &timestamp{timestamppb.Now()}).
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
		return nil, errs.ErrNotFoundf("flag %q", r.Key)
	}

	return s.GetFlag(ctx, r.Key)
}

// DeleteFlag deletes a flag
func (s *Store) DeleteFlag(ctx context.Context, r *flipt.DeleteFlagRequest) error {
	_, err := s.builder.Delete("flags").
		Where(sq.Eq{"\"key\"": r.Key}).
		ExecContext(ctx)

	return err
}

// CreateVariant creates a variant
func (s *Store) CreateVariant(ctx context.Context, r *flipt.CreateVariantRequest) (*flipt.Variant, error) {
	var (
		now = timestamppb.Now()
		v   = &flipt.Variant{
			Id:          uuid.Must(uuid.NewV4()).String(),
			FlagKey:     r.FlagKey,
			Key:         r.Key,
			Name:        r.Name,
			Description: r.Description,
			Attachment:  r.Attachment,
			CreatedAt:   now,
			UpdatedAt:   now,
		}
	)

	attachment := emptyAsNil(r.Attachment)
	if _, err := s.builder.Insert("variants").
		Columns("id", "flag_key", "\"key\"", "name", "description", "attachment", "created_at", "updated_at").
		Values(v.Id, v.FlagKey, v.Key, v.Name, v.Description, attachment, &timestamp{v.CreatedAt}, &timestamp{v.UpdatedAt}).
		ExecContext(ctx); err != nil {
		return nil, err
	}

	if attachment != nil {
		compactedAttachment, err := compactJSONString(*attachment)
		if err != nil {
			return nil, err
		}
		v.Attachment = compactedAttachment
	}

	return v, nil
}

// UpdateVariant updates an existing variant
func (s *Store) UpdateVariant(ctx context.Context, r *flipt.UpdateVariantRequest) (*flipt.Variant, error) {
	query := s.builder.Update("variants").
		Set("\"key\"", r.Key).
		Set("name", r.Name).
		Set("description", r.Description).
		Set("attachment", emptyAsNil(r.Attachment)).
		Set("updated_at", &timestamp{timestamppb.Now()}).
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
		attachment sql.NullString
		createdAt  timestamp
		updatedAt  timestamp

		v = &flipt.Variant{}
	)

	if err := s.builder.Select("id, \"key\", flag_key, name, description, attachment, created_at, updated_at").
		From("variants").
		Where(sq.And{sq.Eq{"id": r.Id}, sq.Eq{"flag_key": r.FlagKey}}).
		QueryRowContext(ctx).
		Scan(&v.Id, &v.Key, &v.FlagKey, &v.Name, &v.Description, &attachment, &createdAt, &updatedAt); err != nil {
		return nil, err
	}

	v.CreatedAt = createdAt.Timestamp
	v.UpdatedAt = updatedAt.Timestamp
	if attachment.Valid {
		compactedAttachment, err := compactJSONString(attachment.String)
		if err != nil {
			return nil, err
		}
		v.Attachment = compactedAttachment
	}

	return v, nil
}

// DeleteVariant deletes a variant
func (s *Store) DeleteVariant(ctx context.Context, r *flipt.DeleteVariantRequest) error {
	_, err := s.builder.Delete("variants").
		Where(sq.And{sq.Eq{"id": r.Id}, sq.Eq{"flag_key": r.FlagKey}}).
		ExecContext(ctx)

	return err
}

func (s *Store) variants(ctx context.Context, flag *flipt.Flag) (err error) {
	query := s.builder.Select("id, flag_key, \"key\", name, description, attachment, created_at, updated_at").
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
			attachment           sql.NullString
		)

		if err := rows.Scan(
			&variant.Id,
			&variant.FlagKey,
			&variant.Key,
			&variant.Name,
			&variant.Description,
			&attachment,
			&createdAt,
			&updatedAt); err != nil {
			return err
		}

		variant.CreatedAt = createdAt.Timestamp
		variant.UpdatedAt = updatedAt.Timestamp
		if attachment.Valid {
			compactedAttachment, err := compactJSONString(attachment.String)
			if err != nil {
				return err
			}
			variant.Attachment = compactedAttachment
		}

		flag.Variants = append(flag.Variants, &variant)
	}

	return rows.Err()
}
