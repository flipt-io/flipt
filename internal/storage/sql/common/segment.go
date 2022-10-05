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
	flipt "go.flipt.io/flipt/rpc/flipt"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// GetSegment gets a segment
func (s *Store) GetSegment(ctx context.Context, key string) (*flipt.Segment, error) {
	var (
		createdAt timestamp
		updatedAt timestamp

		segment = &flipt.Segment{}

		err = s.builder.Select("\"key\", name, description, match_type, created_at, updated_at").
			From("segments").
			Where(sq.Eq{"\"key\"": key}).
			QueryRowContext(ctx).Scan(
			&segment.Key,
			&segment.Name,
			&segment.Description,
			&segment.MatchType,
			&createdAt,
			&updatedAt)
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errs.ErrNotFoundf("segment %q", key)
		}

		return nil, err
	}

	segment.CreatedAt = createdAt.Timestamp
	segment.UpdatedAt = updatedAt.Timestamp

	if err := s.constraints(ctx, segment); err != nil {
		return segment, err
	}

	return segment, nil
}

// ListSegments lists all segments
func (s *Store) ListSegments(ctx context.Context, opts ...storage.QueryOption) (storage.ResultSet[*flipt.Segment], error) {
	params := &storage.QueryParams{}

	for _, opt := range opts {
		opt(params)
	}

	var (
		segments []*flipt.Segment
		results  = storage.ResultSet[*flipt.Segment]{}

		query = s.builder.Select("\"key\", name, description, match_type, created_at, updated_at").
			From("segments").
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
			segment   = &flipt.Segment{}
			createdAt timestamp
			updatedAt timestamp
		)

		if err := rows.Scan(
			&segment.Key,
			&segment.Name,
			&segment.Description,
			&segment.MatchType,
			&createdAt,
			&updatedAt); err != nil {
			return results, err
		}

		segment.CreatedAt = createdAt.Timestamp
		segment.UpdatedAt = updatedAt.Timestamp

		if err := s.constraints(ctx, segment); err != nil {
			return results, err
		}

		segments = append(segments, segment)
	}

	if err := rows.Err(); err != nil {
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

	return results, rows.Err()
}

// CountSegments counts all segments
func (s *Store) CountSegments(ctx context.Context) (uint64, error) {
	var count uint64

	if err := s.builder.Select("COUNT(*)").From("segments").QueryRowContext(ctx).Scan(&count); err != nil {
		return 0, err
	}

	return count, nil
}

// CreateSegment creates a segment
func (s *Store) CreateSegment(ctx context.Context, r *flipt.CreateSegmentRequest) (*flipt.Segment, error) {
	var (
		now     = timestamppb.Now()
		segment = &flipt.Segment{
			Key:         r.Key,
			Name:        r.Name,
			Description: r.Description,
			MatchType:   r.MatchType,
			CreatedAt:   now,
			UpdatedAt:   now,
		}
	)

	if _, err := s.builder.Insert("segments").
		Columns("\"key\"", "name", "description", "match_type", "created_at", "updated_at").
		Values(segment.Key, segment.Name, segment.Description, segment.MatchType, &timestamp{segment.CreatedAt}, &timestamp{segment.UpdatedAt}).
		ExecContext(ctx); err != nil {
		return nil, err
	}

	return segment, nil
}

// UpdateSegment updates an existing segment
func (s *Store) UpdateSegment(ctx context.Context, r *flipt.UpdateSegmentRequest) (*flipt.Segment, error) {
	query := s.builder.Update("segments").
		Set("name", r.Name).
		Set("description", r.Description).
		Set("match_type", r.MatchType).
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
		return nil, errs.ErrNotFoundf("segment %q", r.Key)
	}

	return s.GetSegment(ctx, r.Key)
}

// DeleteSegment deletes a segment
func (s *Store) DeleteSegment(ctx context.Context, r *flipt.DeleteSegmentRequest) error {
	_, err := s.builder.Delete("segments").
		Where(sq.Eq{"\"key\"": r.Key}).
		ExecContext(ctx)

	return err
}

// CreateConstraint creates a constraint
func (s *Store) CreateConstraint(ctx context.Context, r *flipt.CreateConstraintRequest) (*flipt.Constraint, error) {
	var (
		operator = strings.ToLower(r.Operator)
		now      = timestamppb.Now()
		c        = &flipt.Constraint{
			Id:         uuid.Must(uuid.NewV4()).String(),
			SegmentKey: r.SegmentKey,
			Type:       r.Type,
			Property:   r.Property,
			Operator:   operator,
			Value:      r.Value,
			CreatedAt:  now,
			UpdatedAt:  now,
		}
	)

	// unset value if operator does not require it
	if _, ok := flipt.NoValueOperators[c.Operator]; ok {
		c.Value = ""
	}

	if _, err := s.builder.Insert("constraints").
		Columns("id", "segment_key", "type", "property", "operator", "value", "created_at", "updated_at").
		Values(c.Id, c.SegmentKey, c.Type, c.Property, c.Operator, c.Value, &timestamp{c.CreatedAt}, &timestamp{c.UpdatedAt}).
		ExecContext(ctx); err != nil {
		return nil, err
	}

	return c, nil
}

// UpdateConstraint updates an existing constraint
func (s *Store) UpdateConstraint(ctx context.Context, r *flipt.UpdateConstraintRequest) (*flipt.Constraint, error) {
	operator := strings.ToLower(r.Operator)

	// unset value if operator does not require it
	if _, ok := flipt.NoValueOperators[operator]; ok {
		r.Value = ""
	}

	res, err := s.builder.Update("constraints").
		Set("type", r.Type).
		Set("property", r.Property).
		Set("operator", operator).
		Set("value", r.Value).
		Set("updated_at", &timestamp{timestamppb.Now()}).
		Where(sq.And{sq.Eq{"id": r.Id}, sq.Eq{"segment_key": r.SegmentKey}}).
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
		createdAt timestamp
		updatedAt timestamp

		c = &flipt.Constraint{}
	)

	if err := s.builder.Select("id, segment_key, type, property, operator, value, created_at, updated_at").
		From("constraints").
		Where(sq.And{sq.Eq{"id": r.Id}, sq.Eq{"segment_key": r.SegmentKey}}).
		QueryRowContext(ctx).
		Scan(&c.Id, &c.SegmentKey, &c.Type, &c.Property, &c.Operator, &c.Value, &createdAt, &updatedAt); err != nil {
		return nil, err
	}

	c.CreatedAt = createdAt.Timestamp
	c.UpdatedAt = updatedAt.Timestamp

	return c, nil
}

// DeleteConstraint deletes a constraint
func (s *Store) DeleteConstraint(ctx context.Context, r *flipt.DeleteConstraintRequest) error {
	_, err := s.builder.Delete("constraints").
		Where(sq.And{sq.Eq{"id": r.Id}, sq.Eq{"segment_key": r.SegmentKey}}).
		ExecContext(ctx)

	return err
}

func (s *Store) constraints(ctx context.Context, segment *flipt.Segment) error {
	query := s.builder.Select("id, segment_key, type, property, operator, value, created_at, updated_at").
		From("constraints").
		Where(sq.Eq{"segment_key": segment.Key}).
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
			constraint           flipt.Constraint
			createdAt, updatedAt timestamp
		)

		if err := rows.Scan(
			&constraint.Id,
			&constraint.SegmentKey,
			&constraint.Type,
			&constraint.Property,
			&constraint.Operator,
			&constraint.Value,
			&createdAt,
			&updatedAt); err != nil {
			return err
		}

		constraint.CreatedAt = createdAt.Timestamp
		constraint.UpdatedAt = updatedAt.Timestamp
		segment.Constraints = append(segment.Constraints, &constraint)
	}

	return rows.Err()
}
