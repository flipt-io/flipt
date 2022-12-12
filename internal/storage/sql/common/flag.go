package common

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/gofrs/uuid"

	errs "go.flipt.io/flipt/errors"
	"go.flipt.io/flipt/internal/storage"
	fliptsql "go.flipt.io/flipt/internal/storage/sql"
	flipt "go.flipt.io/flipt/rpc/flipt"
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
		createdAt fliptsql.Timestamp
		updatedAt fliptsql.Timestamp

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
			return nil, errs.NewErrorf[errs.ErrNotFound]("flag %q", key)
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
func (s *Store) ListFlags(ctx context.Context, opts ...storage.QueryOption) (storage.ResultSet[*flipt.Flag], error) {
	params := &storage.QueryParams{}

	for _, opt := range opts {
		opt(params)
	}

	var (
		flags   []*flipt.Flag
		results = storage.ResultSet[*flipt.Flag]{}

		query = s.builder.Select("\"key\", name, description, enabled, created_at, updated_at").
			From("flags").
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
			flag      = &flipt.Flag{}
			createdAt fliptsql.Timestamp
			updatedAt fliptsql.Timestamp
		)

		if err := rows.Scan(
			&flag.Key,
			&flag.Name,
			&flag.Description,
			&flag.Enabled,
			&createdAt,
			&updatedAt); err != nil {
			return results, err
		}

		flag.CreatedAt = createdAt.Timestamp
		flag.UpdatedAt = updatedAt.Timestamp

		if err := s.variants(ctx, flag); err != nil {
			return results, err
		}

		flags = append(flags, flag)
	}

	if err := rows.Err(); err != nil {
		return results, err
	}

	var next *flipt.Flag

	if len(flags) > int(params.Limit) && params.Limit > 0 {
		next = flags[len(flags)-1]
		flags = flags[:params.Limit]
	}

	results.Results = flags

	if next != nil {
		out, err := json.Marshal(PageToken{Key: next.Key, Offset: offset + uint64(len(flags))})
		if err != nil {
			return results, fmt.Errorf("encoding page token %w", err)
		}
		results.NextPageToken = string(out)
	}

	return results, nil
}

// CountFlags counts all flags
func (s *Store) CountFlags(ctx context.Context) (uint64, error) {
	var count uint64

	if err := s.builder.Select("COUNT(*)").From("flags").QueryRowContext(ctx).Scan(&count); err != nil {
		return 0, err
	}

	return count, nil
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
		Values(
			flag.Key,
			flag.Name,
			flag.Description,
			flag.Enabled,
			&fliptsql.Timestamp{Timestamp: flag.CreatedAt},
			&fliptsql.Timestamp{Timestamp: flag.UpdatedAt},
		).
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
		return nil, errs.NewErrorf[errs.ErrNotFound]("flag %q", r.Key)
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
		Values(
			v.Id,
			v.FlagKey,
			v.Key,
			v.Name,
			v.Description,
			attachment,
			&fliptsql.Timestamp{Timestamp: v.CreatedAt},
			&fliptsql.Timestamp{Timestamp: v.UpdatedAt},
		).
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
		Set("updated_at", &fliptsql.Timestamp{Timestamp: timestamppb.Now()}).
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
		return nil, errs.NewErrorf[errs.ErrNotFound]("variant %q", r.Key)
	}

	var (
		attachment sql.NullString
		createdAt  fliptsql.Timestamp
		updatedAt  fliptsql.Timestamp

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
			createdAt, updatedAt fliptsql.Timestamp
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
