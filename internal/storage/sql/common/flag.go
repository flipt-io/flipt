package common

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/gofrs/uuid"

	errs "go.flipt.io/flipt/errors"
	"go.flipt.io/flipt/internal/storage"
	fliptsql "go.flipt.io/flipt/internal/storage/sql"
	flipt "go.flipt.io/flipt/rpc/flipt"
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

// GetFlag gets a flag with variants by key
func (s *Store) GetFlag(ctx context.Context, p storage.ResourceRequest) (*flipt.Flag, error) {
	var (
		createdAt fliptsql.Timestamp
		updatedAt fliptsql.Timestamp

		flag = &flipt.Flag{}

		err = s.builder.Select("namespace_key, \"key\", \"type\", name, description, enabled, created_at, updated_at").
			From("flags").
			Where(sq.Eq{"namespace_key": p.Namespace(), "\"key\"": p.Key}).
			QueryRowContext(ctx).
			Scan(
				&flag.NamespaceKey,
				&flag.Key,
				&flag.Type,
				&flag.Name,
				&flag.Description,
				&flag.Enabled,
				&createdAt,
				&updatedAt)
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errs.ErrNotFoundf("flag %q", p)
		}

		return nil, err
	}

	flag.CreatedAt = createdAt.Timestamp
	flag.UpdatedAt = updatedAt.Timestamp

	query := s.builder.Select("id, namespace_key, flag_key, \"key\", name, description, attachment, created_at, updated_at").
		From("variants").
		Where(sq.Eq{"namespace_key": flag.NamespaceKey, "flag_key": flag.Key}).
		OrderBy("created_at ASC")

	rows, err := query.QueryContext(ctx)
	if err != nil {
		return flag, err
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
			&variant.NamespaceKey,
			&variant.FlagKey,
			&variant.Key,
			&variant.Name,
			&variant.Description,
			&attachment,
			&createdAt,
			&updatedAt); err != nil {
			return flag, err
		}

		variant.CreatedAt = createdAt.Timestamp
		variant.UpdatedAt = updatedAt.Timestamp
		if attachment.Valid {
			compactedAttachment, err := compactJSONString(attachment.String)
			if err != nil {
				return flag, err
			}
			variant.Attachment = compactedAttachment
		}

		flag.Variants = append(flag.Variants, &variant)
	}

	return flag, rows.Err()
}

type optionalVariant struct {
	Id           sql.NullString
	NamespaceKey sql.NullString
	Key          sql.NullString
	FlagKey      sql.NullString
	Name         sql.NullString
	Description  sql.NullString
	Attachment   sql.NullString
	CreatedAt    fliptsql.NullableTimestamp
	UpdatedAt    fliptsql.NullableTimestamp
}

// ListFlags lists all flags with variants
func (s *Store) ListFlags(ctx context.Context, req *storage.ListRequest[storage.NamespaceRequest]) (storage.ResultSet[*flipt.Flag], error) {
	var (
		flags   []*flipt.Flag
		results = storage.ResultSet[*flipt.Flag]{}

		query = s.builder.Select("namespace_key, \"key\", \"type\", name, description, enabled, created_at, updated_at").
			From("flags").
			Where(sq.Eq{"namespace_key": req.Predicate.Namespace()}).
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

	// keep track of flags so we can associated variants in second query.
	flagsByKey := make(map[string]*flipt.Flag)
	for rows.Next() {
		var (
			flag = &flipt.Flag{}

			fCreatedAt fliptsql.Timestamp
			fUpdatedAt fliptsql.Timestamp
		)

		if err := rows.Scan(
			&flag.NamespaceKey,
			&flag.Key,
			&flag.Type,
			&flag.Name,
			&flag.Description,
			&flag.Enabled,
			&fCreatedAt,
			&fUpdatedAt); err != nil {
			return results, err
		}

		flag.CreatedAt = fCreatedAt.Timestamp
		flag.UpdatedAt = fUpdatedAt.Timestamp

		flags = append(flags, flag)
		flagsByKey[flag.Key] = flag
	}

	if err := rows.Err(); err != nil {
		return results, err
	}

	if err := rows.Close(); err != nil {
		return results, err
	}

	if err := s.setVariants(ctx, req.Predicate.Namespace(), flagsByKey); err != nil {
		return results, err
	}

	var next *flipt.Flag

	if len(flags) > int(req.QueryParams.Limit) && req.QueryParams.Limit > 0 {
		next = flags[len(flags)-1]
		flags = flags[:req.QueryParams.Limit]
	}

	results.Results = flags

	if next != nil {
		out, err := json.Marshal(PageToken{Key: next.Key, Offset: offset + uint64(len(flags))})
		if err != nil {
			return results, fmt.Errorf("encoding page token %w", err)
		}
		results.NextPageToken = base64.StdEncoding.EncodeToString(out)
	}

	return results, nil
}

func (s *Store) setVariants(ctx context.Context, namespaceKey string, flagsByKey map[string]*flipt.Flag) error {
	allFlagKeys := make([]string, 0, len(flagsByKey))
	for k := range flagsByKey {
		allFlagKeys = append(allFlagKeys, k)
	}

	query := s.builder.Select("id, namespace_key, \"key\", flag_key, name, description, attachment, created_at, updated_at").
		From("variants").
		Where(sq.Eq{"namespace_key": namespaceKey, "flag_key": allFlagKeys}).
		OrderBy("created_at")

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
			variant    optionalVariant
			vCreatedAt fliptsql.NullableTimestamp
			vUpdatedAt fliptsql.NullableTimestamp
		)

		if err := rows.Scan(
			&variant.Id,
			&variant.NamespaceKey,
			&variant.Key,
			&variant.FlagKey,
			&variant.Name,
			&variant.Description,
			&variant.Attachment,
			&vCreatedAt,
			&vUpdatedAt); err != nil {
			return err
		}

		if flag, ok := flagsByKey[variant.FlagKey.String]; ok {
			flag.Variants = append(flag.Variants, &flipt.Variant{
				Id:           variant.Id.String,
				NamespaceKey: variant.NamespaceKey.String,
				Key:          variant.Key.String,
				FlagKey:      variant.FlagKey.String,
				Name:         variant.Name.String,
				Description:  variant.Description.String,
				Attachment:   variant.Attachment.String,
				CreatedAt:    vCreatedAt.Timestamp,
				UpdatedAt:    vUpdatedAt.Timestamp,
			})
		}
	}

	if err := rows.Err(); err != nil {
		return err
	}

	return rows.Close()
}

// CountFlags counts all flags
func (s *Store) CountFlags(ctx context.Context, p storage.NamespaceRequest) (uint64, error) {
	var count uint64

	if err := s.builder.Select("COUNT(*)").
		From("flags").
		Where(sq.Eq{"namespace_key": p.Namespace()}).
		QueryRowContext(ctx).
		Scan(&count); err != nil {
		return 0, err
	}

	return count, nil
}

// CreateFlag creates a flag
func (s *Store) CreateFlag(ctx context.Context, r *flipt.CreateFlagRequest) (_ *flipt.Flag, err error) {
	defer func() {
		if err == nil {
			err = s.setVersion(ctx, r.NamespaceKey)
		}
	}()

	if r.NamespaceKey == "" {
		r.NamespaceKey = storage.DefaultNamespace
	}

	var (
		now  = flipt.Now()
		flag = &flipt.Flag{
			NamespaceKey: r.NamespaceKey,
			Key:          r.Key,
			Type:         r.Type,
			Name:         r.Name,
			Description:  r.Description,
			Enabled:      r.Enabled,
			CreatedAt:    now,
			UpdatedAt:    now,
		}
	)

	if _, err := s.builder.Insert("flags").
		Columns("namespace_key", "\"key\"", "\"type\"", "name", "description", "enabled", "created_at", "updated_at").
		Values(
			flag.NamespaceKey,
			flag.Key,
			flag.Type,
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
func (s *Store) UpdateFlag(ctx context.Context, r *flipt.UpdateFlagRequest) (_ *flipt.Flag, err error) {
	defer func() {
		if err == nil {
			err = s.setVersion(ctx, r.NamespaceKey)
		}
	}()

	if r.NamespaceKey == "" {
		r.NamespaceKey = storage.DefaultNamespace
	}

	query := s.builder.Update("flags").
		Set("name", r.Name).
		Set("description", r.Description).
		Set("enabled", r.Enabled).
		Set("updated_at", &fliptsql.Timestamp{Timestamp: flipt.Now()}).
		Where(sq.And{sq.Eq{"namespace_key": r.NamespaceKey}, sq.Eq{"\"key\"": r.Key}})

	res, err := query.ExecContext(ctx)
	if err != nil {
		return nil, err
	}

	count, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}

	p := storage.NewResource(r.NamespaceKey, r.Key)

	if count != 1 {
		return nil, errs.ErrNotFoundf("flag %q", p)
	}

	return s.GetFlag(ctx, p)
}

// DeleteFlag deletes a flag
func (s *Store) DeleteFlag(ctx context.Context, r *flipt.DeleteFlagRequest) error {
	defer func() {
		_ = s.setVersion(ctx, r.NamespaceKey)
	}()

	if r.NamespaceKey == "" {
		r.NamespaceKey = storage.DefaultNamespace
	}

	_, err := s.builder.Delete("flags").
		Where(sq.And{sq.Eq{"namespace_key": r.NamespaceKey}, sq.Eq{"\"key\"": r.Key}}).
		ExecContext(ctx)

	return err
}

// CreateVariant creates a variant
func (s *Store) CreateVariant(ctx context.Context, r *flipt.CreateVariantRequest) (*flipt.Variant, error) {
	defer func() {
		_ = s.setVersion(ctx, r.NamespaceKey)
	}()

	if r.NamespaceKey == "" {
		r.NamespaceKey = storage.DefaultNamespace
	}

	var (
		now = flipt.Now()
		v   = &flipt.Variant{
			Id:           uuid.Must(uuid.NewV4()).String(),
			NamespaceKey: r.NamespaceKey,
			FlagKey:      r.FlagKey,
			Key:          r.Key,
			Name:         r.Name,
			Description:  r.Description,
			Attachment:   r.Attachment,
			CreatedAt:    now,
			UpdatedAt:    now,
		}
	)

	attachment := emptyAsNil(r.Attachment)
	if _, err := s.builder.Insert("variants").
		Columns("id", "namespace_key", "flag_key", "\"key\"", "name", "description", "attachment", "created_at", "updated_at").
		Values(
			v.Id,
			v.NamespaceKey,
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
func (s *Store) UpdateVariant(ctx context.Context, r *flipt.UpdateVariantRequest) (_ *flipt.Variant, err error) {
	defer func() {
		if err == nil {
			err = s.setVersion(ctx, r.NamespaceKey)
		}
	}()

	if r.NamespaceKey == "" {
		r.NamespaceKey = storage.DefaultNamespace
	}

	whereClause := sq.And{sq.Eq{"id": r.Id}, sq.Eq{"flag_key": r.FlagKey}, sq.Eq{"namespace_key": r.NamespaceKey}}

	query := s.builder.Update("variants").
		Set("\"key\"", r.Key).
		Set("name", r.Name).
		Set("description", r.Description).
		Set("attachment", emptyAsNil(r.Attachment)).
		Set("updated_at", &fliptsql.Timestamp{Timestamp: flipt.Now()}).
		Where(whereClause)

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
		createdAt  fliptsql.Timestamp
		updatedAt  fliptsql.Timestamp

		v = &flipt.Variant{}
	)

	if err := s.builder.Select("id, namespace_key, \"key\", flag_key, name, description, attachment, created_at, updated_at").
		From("variants").
		Where(whereClause).
		QueryRowContext(ctx).
		Scan(&v.Id, &v.NamespaceKey, &v.Key, &v.FlagKey, &v.Name, &v.Description, &attachment, &createdAt, &updatedAt); err != nil {
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
func (s *Store) DeleteVariant(ctx context.Context, r *flipt.DeleteVariantRequest) (err error) {
	defer func() {
		if err == nil {
			err = s.setVersion(ctx, r.NamespaceKey)
		}
	}()

	if r.NamespaceKey == "" {
		r.NamespaceKey = storage.DefaultNamespace
	}

	_, err = s.builder.Delete("variants").
		Where(sq.And{sq.Eq{"id": r.Id}, sq.Eq{"flag_key": r.FlagKey}, sq.Eq{"namespace_key": r.NamespaceKey}}).
		ExecContext(ctx)

	return err
}
