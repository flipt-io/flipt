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

// GetFlag gets a flag with variants by key
func (s *Store) GetFlag(ctx context.Context, namespaceKey, key string) (*flipt.Flag, error) {
	if namespaceKey == "" {
		namespaceKey = storage.DefaultNamespace
	}

	var (
		createdAt fliptsql.Timestamp
		updatedAt fliptsql.Timestamp

		flag = &flipt.Flag{}

		err = s.builder.Select("namespace_key, \"key\", name, description, enabled, created_at, updated_at").
			From("flags").
			Where(sq.And{sq.Eq{"namespace_key": namespaceKey}, sq.Eq{"\"key\"": key}}).
			QueryRowContext(ctx).
			Scan(
				&flag.NamespaceKey,
				&flag.Key,
				&flag.Name,
				&flag.Description,
				&flag.Enabled,
				&createdAt,
				&updatedAt)
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errs.ErrNotFoundf(`flag "%s/%s"`, namespaceKey, key)
		}

		return nil, err
	}

	flag.CreatedAt = createdAt.Timestamp
	flag.UpdatedAt = updatedAt.Timestamp

	query := s.builder.Select("id, namespace_key, flag_key, \"key\", name, description, attachment, created_at, updated_at").
		From("variants").
		Where(sq.And{sq.Eq{"namespace_key": flag.NamespaceKey}, sq.Eq{"flag_key": flag.Key}}).
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
func (s *Store) ListFlags(ctx context.Context, namespaceKey string, opts ...storage.QueryOption) (storage.ResultSet[*flipt.Flag], error) {
	if namespaceKey == "" {
		namespaceKey = storage.DefaultNamespace
	}

	params := &storage.QueryParams{}

	for _, opt := range opts {
		opt(params)
	}

	var (
		flags   []*flipt.Flag
		results = storage.ResultSet[*flipt.Flag]{}

		query = s.builder.Select("f.namespace_key, f.key, f.name, f.description, f.enabled, f.created_at, f.updated_at, v.id, v.namespace_key, v.key, v.flag_key, v.name, v.description, v.attachment, v.created_at, v.updated_at").
			From("flags f").
			Where(sq.Eq{"f.namespace_key": namespaceKey}).
			LeftJoin("variants v ON v.flag_key = f.key AND v.namespace_key = f.namespace_key").
			OrderBy(fmt.Sprintf("f.created_at %s", params.Order))
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

	// keep track of flags we've seen so we don't append duplicates because of the join
	uniqueFlags := make(map[string][]*flipt.Variant)

	for rows.Next() {
		var (
			flag = &flipt.Flag{}
			v    = &optionalVariant{}

			fCreatedAt fliptsql.Timestamp
			fUpdatedAt fliptsql.Timestamp
		)

		if err := rows.Scan(
			&flag.NamespaceKey,
			&flag.Key,
			&flag.Name,
			&flag.Description,
			&flag.Enabled,
			&fCreatedAt,
			&fUpdatedAt,
			&v.Id,
			&v.NamespaceKey,
			&v.Key,
			&v.FlagKey,
			&v.Name,
			&v.Description,
			&v.Attachment,
			&v.CreatedAt,
			&v.UpdatedAt); err != nil {
			return results, err
		}

		flag.CreatedAt = fCreatedAt.Timestamp
		flag.UpdatedAt = fUpdatedAt.Timestamp

		// append flag to output results if we haven't seen it yet, to maintain order
		if _, ok := uniqueFlags[flag.Key]; !ok {
			flags = append(flags, flag)
		}

		// append variant to flag if it exists (not null)
		if v.Id.Valid {
			variant := &flipt.Variant{
				Id: v.Id.String,
			}
			if v.NamespaceKey.Valid {
				variant.NamespaceKey = v.NamespaceKey.String
			}
			if v.Key.Valid {
				variant.Key = v.Key.String
			}
			if v.FlagKey.Valid {
				variant.FlagKey = v.FlagKey.String
			}
			if v.Name.Valid {
				variant.Name = v.Name.String
			}
			if v.Description.Valid {
				variant.Description = v.Description.String
			}
			if v.Attachment.Valid {
				compactedAttachment, err := compactJSONString(v.Attachment.String)
				if err != nil {
					return results, err
				}
				variant.Attachment = compactedAttachment
			}
			if v.CreatedAt.IsValid() {
				variant.CreatedAt = v.CreatedAt.Timestamp
			}
			if v.UpdatedAt.IsValid() {
				variant.UpdatedAt = v.UpdatedAt.Timestamp
			}

			uniqueFlags[flag.Key] = append(uniqueFlags[flag.Key], variant)
		}
	}

	if err := rows.Err(); err != nil {
		return results, err
	}

	if err := rows.Close(); err != nil {
		return results, err
	}

	// set variants on flags before returning results
	for _, f := range flags {
		f.Variants = uniqueFlags[f.Key]
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
func (s *Store) CountFlags(ctx context.Context, namespaceKey string) (uint64, error) {
	var count uint64

	if namespaceKey == "" {
		namespaceKey = storage.DefaultNamespace
	}

	if err := s.builder.Select("COUNT(*)").
		From("flags").
		Where(sq.Eq{"namespace_key": namespaceKey}).
		QueryRowContext(ctx).
		Scan(&count); err != nil {
		return 0, err
	}

	return count, nil
}

// CreateFlag creates a flag
func (s *Store) CreateFlag(ctx context.Context, r *flipt.CreateFlagRequest) (*flipt.Flag, error) {
	if r.NamespaceKey == "" {
		r.NamespaceKey = storage.DefaultNamespace
	}

	var (
		now  = timestamppb.Now()
		flag = &flipt.Flag{
			NamespaceKey: r.NamespaceKey,
			Key:          r.Key,
			Name:         r.Name,
			Description:  r.Description,
			Enabled:      r.Enabled,
			CreatedAt:    now,
			UpdatedAt:    now,
		}
	)

	if _, err := s.builder.Insert("flags").
		Columns("namespace_key", "\"key\"", "name", "description", "enabled", "created_at", "updated_at").
		Values(
			flag.NamespaceKey,
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
	if r.NamespaceKey == "" {
		r.NamespaceKey = storage.DefaultNamespace
	}

	query := s.builder.Update("flags").
		Set("name", r.Name).
		Set("description", r.Description).
		Set("enabled", r.Enabled).
		Set("updated_at", &fliptsql.Timestamp{Timestamp: timestamppb.Now()}).
		Where(sq.And{sq.Eq{"namespace_key": r.NamespaceKey}, sq.Eq{"\"key\"": r.Key}})

	res, err := query.ExecContext(ctx)
	if err != nil {
		return nil, err
	}

	count, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}

	if count != 1 {
		return nil, errs.ErrNotFoundf(`flag "%s/%s"`, r.NamespaceKey, r.Key)
	}

	return s.GetFlag(ctx, r.NamespaceKey, r.Key)
}

// DeleteFlag deletes a flag
func (s *Store) DeleteFlag(ctx context.Context, r *flipt.DeleteFlagRequest) error {
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
	if r.NamespaceKey == "" {
		r.NamespaceKey = storage.DefaultNamespace
	}

	var (
		now = timestamppb.Now()
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
func (s *Store) UpdateVariant(ctx context.Context, r *flipt.UpdateVariantRequest) (*flipt.Variant, error) {
	if r.NamespaceKey == "" {
		r.NamespaceKey = storage.DefaultNamespace
	}

	whereClause := sq.And{sq.Eq{"id": r.Id}, sq.Eq{"flag_key": r.FlagKey}, sq.Eq{"namespace_key": r.NamespaceKey}}

	query := s.builder.Update("variants").
		Set("\"key\"", r.Key).
		Set("name", r.Name).
		Set("description", r.Description).
		Set("attachment", emptyAsNil(r.Attachment)).
		Set("updated_at", &fliptsql.Timestamp{Timestamp: timestamppb.Now()}).
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
func (s *Store) DeleteVariant(ctx context.Context, r *flipt.DeleteVariantRequest) error {
	if r.NamespaceKey == "" {
		r.NamespaceKey = storage.DefaultNamespace
	}

	_, err := s.builder.Delete("variants").
		Where(sq.And{sq.Eq{"id": r.Id}, sq.Eq{"flag_key": r.FlagKey}, sq.Eq{"namespace_key": r.NamespaceKey}}).
		ExecContext(ctx)

	return err
}
