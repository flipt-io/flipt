package common

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	sq "github.com/Masterminds/squirrel"
	"github.com/gofrs/uuid"

	errs "go.flipt.io/flipt/errors"
	"go.flipt.io/flipt/internal/storage"
	fliptsql "go.flipt.io/flipt/internal/storage/sql"
	flipt "go.flipt.io/flipt/rpc/flipt"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// GetSegment gets a segment
func (s *Store) GetSegment(ctx context.Context, namespaceKey, key string) (*flipt.Segment, error) {
	if namespaceKey == "" {
		namespaceKey = storage.DefaultNamespace
	}

	var (
		createdAt fliptsql.Timestamp
		updatedAt fliptsql.Timestamp

		segment = &flipt.Segment{}

		err = s.builder.Select("namespace_key, \"key\", name, description, match_type, created_at, updated_at").
			From("segments").
			Where(sq.And{sq.Eq{"namespace_key": namespaceKey}, sq.Eq{"\"key\"": key}}).
			QueryRowContext(ctx).
			Scan(
				&segment.NamespaceKey,
				&segment.Key,
				&segment.Name,
				&segment.Description,
				&segment.MatchType,
				&createdAt,
				&updatedAt)
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errs.ErrNotFoundf(`segment "%s/%s"`, namespaceKey, key)
		}

		return nil, err
	}

	segment.CreatedAt = createdAt.Timestamp
	segment.UpdatedAt = updatedAt.Timestamp

	query := s.builder.Select("id, namespace_key, segment_key, type, property, operator, value, description, created_at, updated_at").
		From("constraints").
		Where(sq.Eq{"segment_key": segment.Key}).
		OrderBy("created_at ASC")

	rows, err := query.QueryContext(ctx)
	if err != nil {
		return segment, err
	}

	defer func() {
		if cerr := rows.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	for rows.Next() {
		var (
			constraint           flipt.Constraint
			createdAt, updatedAt fliptsql.Timestamp
		)

		if err := rows.Scan(
			&constraint.Id,
			&constraint.NamespaceKey,
			&constraint.SegmentKey,
			&constraint.Type,
			&constraint.Property,
			&constraint.Operator,
			&constraint.Value,
			&constraint.Description,
			&createdAt,
			&updatedAt); err != nil {
			return segment, err
		}

		constraint.CreatedAt = createdAt.Timestamp
		constraint.UpdatedAt = updatedAt.Timestamp
		segment.Constraints = append(segment.Constraints, &constraint)
	}

	return segment, rows.Err()
}

type optionalConstraint struct {
	Id           sql.NullString
	NamespaceKey sql.NullString
	SegmentKey   sql.NullString
	Type         sql.NullInt32
	Property     sql.NullString
	Operator     sql.NullString
	Value        sql.NullString
	Description  sql.NullString
	CreatedAt    fliptsql.NullableTimestamp
	UpdatedAt    fliptsql.NullableTimestamp
}

// ListSegments lists all segments
func (s *Store) ListSegments(ctx context.Context, namespaceKey string, opts ...storage.QueryOption) (storage.ResultSet[*flipt.Segment], error) {
	params := &storage.QueryParams{}

	if namespaceKey == "" {
		namespaceKey = storage.DefaultNamespace
	}

	for _, opt := range opts {
		opt(params)
	}

	var (
		segments []*flipt.Segment
		results  = storage.ResultSet[*flipt.Segment]{}

		query = s.builder.Select("namespace_key, \"key\", name, description, match_type, created_at, updated_at").
			From("segments").
			Where(sq.Eq{"namespace_key": namespaceKey}).
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

	// keep track of segments we've seen so we don't append duplicates because of the join
	segmentsByKey := make(map[string]*flipt.Segment)

	for rows.Next() {
		var (
			segment    = &flipt.Segment{}
			sCreatedAt fliptsql.Timestamp
			sUpdatedAt fliptsql.Timestamp
		)

		if err := rows.Scan(
			&segment.NamespaceKey,
			&segment.Key,
			&segment.Name,
			&segment.Description,
			&segment.MatchType,
			&sCreatedAt,
			&sUpdatedAt); err != nil {
			return results, err
		}

		segment.CreatedAt = sCreatedAt.Timestamp
		segment.UpdatedAt = sUpdatedAt.Timestamp

		segments = append(segments, segment)
		segmentsByKey[segment.Key] = segment
	}

	if err := rows.Err(); err != nil {
		return results, err
	}

	if err := rows.Close(); err != nil {
		return results, err
	}

	if err := s.setConstraints(ctx, namespaceKey, segmentsByKey); err != nil {
		return results, err
	}

	var next *flipt.Segment

	if len(segments) > int(params.Limit) && params.Limit > 0 {
		next = segments[len(segments)-1]
		segments = segments[:params.Limit]
	}

	results.Results = segments

	if next != nil {
		out, err := json.Marshal(PageToken{Key: next.Key, Offset: offset + uint64(len(segments))})
		if err != nil {
			return results, fmt.Errorf("encoding page token %w", err)
		}
		results.NextPageToken = string(out)
	}

	return results, nil
}

func (s *Store) setConstraints(ctx context.Context, namespaceKey string, segmentsByKey map[string]*flipt.Segment) error {
	allSegmentKeys := make([]string, 0, len(segmentsByKey))
	for k := range segmentsByKey {
		allSegmentKeys = append(allSegmentKeys, k)
	}

	query := s.builder.Select("id, namespace_key, segment_key, type, property, operator, value, description, created_at, updated_at").
		From("constraints").
		Where(sq.Eq{"namespace_key": namespaceKey, "segment_key": allSegmentKeys}).
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
			constraint optionalConstraint
			cCreatedAt fliptsql.NullableTimestamp
			cUpdatedAt fliptsql.NullableTimestamp
		)

		if err := rows.Scan(
			&constraint.Id,
			&constraint.NamespaceKey,
			&constraint.SegmentKey,
			&constraint.Type,
			&constraint.Property,
			&constraint.Operator,
			&constraint.Value,
			&constraint.Description,
			&cCreatedAt,
			&cUpdatedAt); err != nil {
			return err
		}

		if segment, ok := segmentsByKey[constraint.SegmentKey.String]; ok {
			segment.Constraints = append(segment.Constraints, &flipt.Constraint{
				Id:           constraint.Id.String,
				NamespaceKey: constraint.NamespaceKey.String,
				SegmentKey:   constraint.SegmentKey.String,
				Type:         flipt.ComparisonType(constraint.Type.Int32),
				Property:     constraint.Property.String,
				Operator:     constraint.Operator.String,
				Value:        constraint.Value.String,
				Description:  constraint.Description.String,
				CreatedAt:    cCreatedAt.Timestamp,
				UpdatedAt:    cUpdatedAt.Timestamp,
			})
		}
	}

	if err := rows.Err(); err != nil {
		return err
	}

	return rows.Close()
}

// CountSegments counts all segments
func (s *Store) CountSegments(ctx context.Context, namespaceKey string) (uint64, error) {
	var count uint64

	if namespaceKey == "" {
		namespaceKey = storage.DefaultNamespace
	}

	if err := s.builder.Select("COUNT(*)").
		From("segments").
		Where(sq.Eq{"namespace_key": namespaceKey}).
		QueryRowContext(ctx).
		Scan(&count); err != nil {
		return 0, err
	}

	return count, nil
}

// CreateSegment creates a segment
func (s *Store) CreateSegment(ctx context.Context, r *flipt.CreateSegmentRequest) (*flipt.Segment, error) {
	if r.NamespaceKey == "" {
		r.NamespaceKey = storage.DefaultNamespace
	}

	var (
		now     = timestamppb.Now()
		segment = &flipt.Segment{
			NamespaceKey: r.NamespaceKey,
			Key:          r.Key,
			Name:         r.Name,
			Description:  r.Description,
			MatchType:    r.MatchType,
			CreatedAt:    now,
			UpdatedAt:    now,
		}
	)

	if _, err := s.builder.Insert("segments").
		Columns("namespace_key", "\"key\"", "name", "description", "match_type", "created_at", "updated_at").
		Values(
			segment.NamespaceKey,
			segment.Key,
			segment.Name,
			segment.Description,
			segment.MatchType,
			&fliptsql.Timestamp{Timestamp: segment.CreatedAt},
			&fliptsql.Timestamp{Timestamp: segment.UpdatedAt}).
		ExecContext(ctx); err != nil {
		return nil, err
	}

	return segment, nil
}

// UpdateSegment updates an existing segment
func (s *Store) UpdateSegment(ctx context.Context, r *flipt.UpdateSegmentRequest) (*flipt.Segment, error) {
	if r.NamespaceKey == "" {
		r.NamespaceKey = storage.DefaultNamespace
	}

	query := s.builder.Update("segments").
		Set("name", r.Name).
		Set("description", r.Description).
		Set("match_type", r.MatchType).
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
		return nil, errs.ErrNotFoundf(`segment "%s/%s"`, r.NamespaceKey, r.Key)
	}

	return s.GetSegment(ctx, r.NamespaceKey, r.Key)
}

// DeleteSegment deletes a segment
func (s *Store) DeleteSegment(ctx context.Context, r *flipt.DeleteSegmentRequest) error {
	if r.NamespaceKey == "" {
		r.NamespaceKey = storage.DefaultNamespace
	}

	_, err := s.builder.Delete("segments").
		Where(sq.And{sq.Eq{"namespace_key": r.NamespaceKey}, sq.Eq{"\"key\"": r.Key}}).
		ExecContext(ctx)

	return err
}

// CreateConstraint creates a constraint
func (s *Store) CreateConstraint(ctx context.Context, r *flipt.CreateConstraintRequest) (*flipt.Constraint, error) {
	if r.NamespaceKey == "" {
		r.NamespaceKey = storage.DefaultNamespace
	}

	var (
		operator = strings.ToLower(r.Operator)
		now      = timestamppb.Now()
		c        = &flipt.Constraint{
			Id:           uuid.Must(uuid.NewV4()).String(),
			NamespaceKey: r.NamespaceKey,
			SegmentKey:   r.SegmentKey,
			Type:         r.Type,
			Property:     r.Property,
			Operator:     operator,
			Value:        r.Value,
			CreatedAt:    now,
			UpdatedAt:    now,
			Description:  r.Description,
		}
	)

	// unset value if operator does not require it
	if _, ok := flipt.NoValueOperators[c.Operator]; ok {
		c.Value = ""
	}

	if _, err := s.builder.Insert("constraints").
		Columns("id", "namespace_key", "segment_key", "type", "property", "operator", "value", "description", "created_at", "updated_at").
		Values(
			c.Id,
			c.NamespaceKey,
			c.SegmentKey,
			c.Type,
			c.Property,
			c.Operator,
			c.Value,
			c.Description,
			&fliptsql.Timestamp{Timestamp: c.CreatedAt},
			&fliptsql.Timestamp{Timestamp: c.UpdatedAt}).
		ExecContext(ctx); err != nil {
		return nil, err
	}

	return c, nil
}

// UpdateConstraint updates an existing constraint
func (s *Store) UpdateConstraint(ctx context.Context, r *flipt.UpdateConstraintRequest) (*flipt.Constraint, error) {
	if r.NamespaceKey == "" {
		r.NamespaceKey = storage.DefaultNamespace
	}

	var (
		whereClause = sq.And{sq.Eq{"id": r.Id}, sq.Eq{"segment_key": r.SegmentKey}, sq.Eq{"namespace_key": r.NamespaceKey}}
		operator    = strings.ToLower(r.Operator)
	)

	// unset value if operator does not require it
	if _, ok := flipt.NoValueOperators[operator]; ok {
		r.Value = ""
	}

	res, err := s.builder.Update("constraints").
		Set("type", r.Type).
		Set("property", r.Property).
		Set("operator", operator).
		Set("value", r.Value).
		Set("description", r.Description).
		Set("updated_at", &fliptsql.Timestamp{Timestamp: timestamppb.Now()}).
		Where(whereClause).
		ExecContext(ctx)
	if err != nil {
		return nil, err
	}

	count, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}

	if count != 1 {
		return nil, errs.ErrNotFoundf("constraint %q", r.Id)
	}

	var (
		createdAt fliptsql.Timestamp
		updatedAt fliptsql.Timestamp

		c = &flipt.Constraint{}
	)

	if err := s.builder.Select("id, namespace_key, segment_key, type, property, operator, value, description, created_at, updated_at").
		From("constraints").
		Where(whereClause).
		QueryRowContext(ctx).
		Scan(&c.Id, &c.NamespaceKey, &c.SegmentKey, &c.Type, &c.Property, &c.Operator, &c.Value, &c.Description, &createdAt, &updatedAt); err != nil {
		return nil, err
	}

	c.CreatedAt = createdAt.Timestamp
	c.UpdatedAt = updatedAt.Timestamp

	return c, nil
}

// DeleteConstraint deletes a constraint
func (s *Store) DeleteConstraint(ctx context.Context, r *flipt.DeleteConstraintRequest) error {
	if r.NamespaceKey == "" {
		r.NamespaceKey = storage.DefaultNamespace
	}

	_, err := s.builder.Delete("constraints").
		Where(sq.And{sq.Eq{"id": r.Id}, sq.Eq{"segment_key": r.SegmentKey}, sq.Eq{"namespace_key": r.NamespaceKey}}).
		ExecContext(ctx)

	return err
}
