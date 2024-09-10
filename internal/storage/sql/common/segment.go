package common

import (
	"context"
	"database/sql"
	"encoding/base64"
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
)

// GetSegment gets a segment
func (s *Store) GetSegment(ctx context.Context, req storage.ResourceRequest) (*flipt.Segment, error) {
	var (
		createdAt fliptsql.Timestamp
		updatedAt fliptsql.Timestamp

		segment = &flipt.Segment{}

		err = s.builder.Select("namespace_key, \"key\", name, description, match_type, created_at, updated_at").
			From("segments").
			Where(sq.Eq{"namespace_key": req.Namespace(), "\"key\"": req.Key}).
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
			return nil, errs.ErrNotFoundf("segment %q", req)
		}

		return nil, err
	}

	segment.CreatedAt = createdAt.Timestamp
	segment.UpdatedAt = updatedAt.Timestamp

	query := s.builder.Select("id, namespace_key, segment_key, type, property, operator, value, description, created_at, updated_at").
		From("constraints").
		Where(sq.And{sq.Eq{"namespace_key": segment.NamespaceKey}, sq.Eq{"segment_key": segment.Key}}).
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
			description          sql.NullString
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
			&description,
			&createdAt,
			&updatedAt); err != nil {
			return segment, err
		}

		constraint.CreatedAt = createdAt.Timestamp
		constraint.UpdatedAt = updatedAt.Timestamp
		constraint.Description = description.String
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
func (s *Store) ListSegments(ctx context.Context, req *storage.ListRequest[storage.NamespaceRequest]) (storage.ResultSet[*flipt.Segment], error) {
	var (
		segments []*flipt.Segment
		results  = storage.ResultSet[*flipt.Segment]{}

		query = s.builder.Select("namespace_key, \"key\", name, description, match_type, created_at, updated_at").
			From("segments").
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

	if err := s.setConstraints(ctx, req.Predicate.Namespace(), segmentsByKey); err != nil {
		return results, err
	}

	var next *flipt.Segment

	if len(segments) > int(req.QueryParams.Limit) && req.QueryParams.Limit > 0 {
		next = segments[len(segments)-1]
		segments = segments[:req.QueryParams.Limit]
	}

	results.Results = segments

	if next != nil {
		out, err := json.Marshal(PageToken{Key: next.Key, Offset: offset + uint64(len(segments))})
		if err != nil {
			return results, fmt.Errorf("encoding page token %w", err)
		}
		results.NextPageToken = base64.StdEncoding.EncodeToString(out)
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
func (s *Store) CountSegments(ctx context.Context, ns storage.NamespaceRequest) (uint64, error) {
	var count uint64

	if err := s.builder.Select("COUNT(*)").
		From("segments").
		Where(sq.Eq{"namespace_key": ns.Namespace()}).
		QueryRowContext(ctx).
		Scan(&count); err != nil {
		return 0, err
	}

	return count, nil
}

// CreateSegment creates a segment
func (s *Store) CreateSegment(ctx context.Context, r *flipt.CreateSegmentRequest) (_ *flipt.Segment, err error) {
	defer func() {
		if err == nil {
			err = s.setVersion(ctx, r.NamespaceKey)
		}
	}()

	if r.NamespaceKey == "" {
		r.NamespaceKey = storage.DefaultNamespace
	}

	var (
		now     = flipt.Now()
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
			int32(segment.MatchType),
			&fliptsql.Timestamp{Timestamp: segment.CreatedAt},
			&fliptsql.Timestamp{Timestamp: segment.UpdatedAt}).
		ExecContext(ctx); err != nil {
		return nil, err
	}

	return segment, nil
}

// UpdateSegment updates an existing segment
func (s *Store) UpdateSegment(ctx context.Context, r *flipt.UpdateSegmentRequest) (_ *flipt.Segment, err error) {
	defer func() {
		if err == nil {
			err = s.setVersion(ctx, r.NamespaceKey)
		}
	}()

	if r.NamespaceKey == "" {
		r.NamespaceKey = storage.DefaultNamespace
	}

	query := s.builder.Update("segments").
		Set("name", r.Name).
		Set("description", r.Description).
		Set("match_type", r.MatchType).
		Set("updated_at", &fliptsql.Timestamp{Timestamp: flipt.Now()}).
		Where(sq.Eq{"namespace_key": r.NamespaceKey, "\"key\"": r.Key})

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
		return nil, errs.ErrNotFoundf("segment %q", p)
	}

	return s.GetSegment(ctx, p)
}

// DeleteSegment deletes a segment
func (s *Store) DeleteSegment(ctx context.Context, r *flipt.DeleteSegmentRequest) (err error) {
	defer func() {
		if err == nil {
			err = s.setVersion(ctx, r.NamespaceKey)
		}
	}()

	if r.NamespaceKey == "" {
		r.NamespaceKey = storage.DefaultNamespace
	}

	_, err = s.builder.Delete("segments").
		Where(sq.And{sq.Eq{"namespace_key": r.NamespaceKey}, sq.Eq{"\"key\"": r.Key}}).
		ExecContext(ctx)

	return err
}

// CreateConstraint creates a constraint
func (s *Store) CreateConstraint(ctx context.Context, r *flipt.CreateConstraintRequest) (_ *flipt.Constraint, err error) {
	defer func() {
		if err == nil {
			err = s.setVersion(ctx, r.NamespaceKey)
		}
	}()

	if r.NamespaceKey == "" {
		r.NamespaceKey = storage.DefaultNamespace
	}

	var (
		operator = strings.ToLower(r.Operator)
		now      = flipt.Now()
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
			int32(c.Type),
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
func (s *Store) UpdateConstraint(ctx context.Context, r *flipt.UpdateConstraintRequest) (_ *flipt.Constraint, err error) {
	defer func() {
		if err == nil {
			err = s.setVersion(ctx, r.NamespaceKey)
		}
	}()

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
		Set("updated_at", &fliptsql.Timestamp{Timestamp: flipt.Now()}).
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
func (s *Store) DeleteConstraint(ctx context.Context, r *flipt.DeleteConstraintRequest) (err error) {
	defer func() {
		if err == nil {
			err = s.setVersion(ctx, r.NamespaceKey)
		}
	}()

	if r.NamespaceKey == "" {
		r.NamespaceKey = storage.DefaultNamespace
	}

	_, err = s.builder.Delete("constraints").
		Where(sq.And{sq.Eq{"id": r.Id}, sq.Eq{"segment_key": r.SegmentKey}, sq.Eq{"namespace_key": r.NamespaceKey}}).
		ExecContext(ctx)

	return err
}
