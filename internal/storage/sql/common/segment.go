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
func (s *Store) GetSegment(ctx context.Context, key string) (*flipt.Segment, error) {
	var (
		createdAt fliptsql.Timestamp
		updatedAt fliptsql.Timestamp

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

	query := s.builder.Select("id, segment_key, type, property, operator, value, created_at, updated_at").
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
			&constraint.SegmentKey,
			&constraint.Type,
			&constraint.Property,
			&constraint.Operator,
			&constraint.Value,
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
	Id         sql.NullString
	SegmentKey sql.NullString
	Type       sql.NullInt32
	Property   sql.NullString
	Operator   sql.NullString
	Value      sql.NullString
	CreatedAt  fliptsql.NullableTimestamp
	UpdatedAt  fliptsql.NullableTimestamp
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

		query = s.builder.Select("s.key, s.name, s.description, s.match_type, s.created_at, s.updated_at, c.id, c.segment_key, c.type, c.property, c.operator, c.value, c.created_at, c.updated_at").
			From("segments s").
			LeftJoin("constraints c ON s.key = c.segment_key").
			OrderBy(fmt.Sprintf("s.created_at %s", params.Order))
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
	uniqueSegments := make(map[string][]*flipt.Constraint)

	for rows.Next() {
		var (
			segment = &flipt.Segment{}
			c       = &optionalConstraint{}

			sCreatedAt fliptsql.Timestamp
			sUpdatedAt fliptsql.Timestamp
		)

		if err := rows.Scan(
			&segment.Key,
			&segment.Name,
			&segment.Description,
			&segment.MatchType,
			&sCreatedAt,
			&sUpdatedAt,
			&c.Id,
			&c.SegmentKey,
			&c.Type,
			&c.Property,
			&c.Operator,
			&c.Value,
			&c.CreatedAt,
			&c.UpdatedAt); err != nil {
			return results, err
		}

		segment.CreatedAt = sCreatedAt.Timestamp
		segment.UpdatedAt = sUpdatedAt.Timestamp

		// append segment to output results if we haven't seen it yet, to maintain order
		if _, ok := uniqueSegments[segment.Key]; !ok {
			segments = append(segments, segment)
		}

		// append constraint to segment if it exists (not null)
		if c.Id.Valid {
			constraint := &flipt.Constraint{
				Id: c.Id.String,
			}
			if c.SegmentKey.Valid {
				constraint.SegmentKey = c.SegmentKey.String
			}
			if c.Type.Valid {
				constraint.Type = flipt.ComparisonType(c.Type.Int32)
			}
			if c.Property.Valid {
				constraint.Property = c.Property.String
			}
			if c.Operator.Valid {
				constraint.Operator = c.Operator.String
			}
			if c.Value.Valid {
				constraint.Value = c.Value.String
			}
			if c.CreatedAt.IsValid() {
				constraint.CreatedAt = c.CreatedAt.Timestamp
			}
			if c.UpdatedAt.IsValid() {
				constraint.UpdatedAt = c.UpdatedAt.Timestamp
			}

			uniqueSegments[segment.Key] = append(uniqueSegments[segment.Key], constraint)
		}
	}

	if err := rows.Err(); err != nil {
		return results, err
	}

	if err := rows.Close(); err != nil {
		return results, err
	}

	// set constraints on segments before returning results
	for _, s := range segments {
		s.Constraints = uniqueSegments[s.Key]
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
		Values(
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
	query := s.builder.Update("segments").
		Set("name", r.Name).
		Set("description", r.Description).
		Set("match_type", r.MatchType).
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
		Values(
			c.Id,
			c.SegmentKey,
			c.Type,
			c.Property,
			c.Operator,
			c.Value,
			&fliptsql.Timestamp{Timestamp: c.CreatedAt},
			&fliptsql.Timestamp{Timestamp: c.UpdatedAt}).
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
		Set("updated_at", &fliptsql.Timestamp{Timestamp: timestamppb.Now()}).
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
		createdAt fliptsql.Timestamp
		updatedAt fliptsql.Timestamp

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
