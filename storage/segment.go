package storage

import (
	"context"
	"database/sql"
	"strings"

	sq "github.com/Masterminds/squirrel"
	uuid "github.com/satori/go.uuid"

	proto "github.com/golang/protobuf/ptypes"
	flipt "github.com/markphelps/flipt/proto"
	sqlite3 "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
)

var _ SegmentStore = &SegmentStorage{}

// SegmentStorage is a SQL SegmentStore
type SegmentStorage struct {
	logger  logrus.FieldLogger
	builder sq.StatementBuilderType
}

// NewSegmentStorage creates a SegmentStorage
func NewSegmentStorage(logger logrus.FieldLogger, builder sq.StatementBuilderType) *SegmentStorage {
	return &SegmentStorage{
		logger:  logger.WithField("storage", "segment"),
		builder: builder,
	}
}

// GetSegment gets a segment
func (s *SegmentStorage) GetSegment(ctx context.Context, r *flipt.GetSegmentRequest) (*flipt.Segment, error) {
	s.logger.WithField("request", r).Debug("get segment")
	segment, err := s.segment(ctx, r.Key)
	s.logger.WithField("response", segment).Debug("get segment")
	return segment, err
}

func (s SegmentStorage) segment(ctx context.Context, key string) (*flipt.Segment, error) {
	var (
		createdAt timestamp
		updatedAt timestamp

		segment = &flipt.Segment{}

		err = s.builder.Select("key, name, description, created_at, updated_at").
			From("segments").
			Where(sq.Eq{"key": key}).
			QueryRowContext(ctx).Scan(
			&segment.Key,
			&segment.Name,
			&segment.Description,
			&createdAt,
			&updatedAt)
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFoundf("segment %q", key)
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
func (s *SegmentStorage) ListSegments(ctx context.Context, r *flipt.ListSegmentRequest) ([]*flipt.Segment, error) {
	s.logger.WithField("request", r).Debug("list segments")

	var (
		segments []*flipt.Segment

		query = s.builder.Select("key, name, description, created_at, updated_at").
			From("segments").
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
		return segments, err
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
			&createdAt,
			&updatedAt); err != nil {
			return segments, err
		}

		segment.CreatedAt = createdAt.Timestamp
		segment.UpdatedAt = updatedAt.Timestamp

		if err := s.constraints(ctx, segment); err != nil {
			return segments, err
		}

		segments = append(segments, segment)
	}

	s.logger.WithField("response", segments).Debug("list segments")
	return segments, rows.Err()
}

// CreateSegment creates a segment
func (s *SegmentStorage) CreateSegment(ctx context.Context, r *flipt.CreateSegmentRequest) (*flipt.Segment, error) {
	s.logger.WithField("request", r).Debug("create segment")

	var (
		now     = proto.TimestampNow()
		segment = &flipt.Segment{
			Key:         r.Key,
			Name:        r.Name,
			Description: r.Description,
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		query = s.builder.Insert("segments").
			Columns("key", "name", "description", "created_at", "updated_at").
			Values(segment.Key, segment.Name, segment.Description, &timestamp{segment.CreatedAt}, &timestamp{segment.UpdatedAt})
	)

	if _, err := query.ExecContext(ctx); err != nil {
		if sqliteErr, ok := err.(sqlite3.Error); ok {
			if sqliteErr.Code == sqlite3.ErrConstraint {
				return nil, ErrInvalidf("segment %q is not unique", r.Key)
			}
		}
		return nil, err
	}

	s.logger.WithField("response", segment).Debug("create segment")
	return segment, nil
}

// UpdateSegment updates an existing segment
func (s *SegmentStorage) UpdateSegment(ctx context.Context, r *flipt.UpdateSegmentRequest) (*flipt.Segment, error) {
	s.logger.WithField("request", r).Debug("update segment")

	query := s.builder.Update("segments").
		Set("name", r.Name).
		Set("description", r.Description).
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
		return nil, ErrNotFoundf("segment %q", r.Key)
	}

	segment, err := s.segment(ctx, r.Key)
	s.logger.WithField("response", segment).Debug("update segment")
	return segment, err
}

// DeleteSegment deletes a segment
func (s *SegmentStorage) DeleteSegment(ctx context.Context, r *flipt.DeleteSegmentRequest) error {
	s.logger.WithField("request", r).Debug("delete segment")

	_, err := s.builder.Delete("segments").
		Where(sq.Eq{"key": r.Key}).
		ExecContext(ctx)

	return err
}

// CreateConstraint creates a constraint
func (s *SegmentStorage) CreateConstraint(ctx context.Context, r *flipt.CreateConstraintRequest) (*flipt.Constraint, error) {
	s.logger.WithField("request", r).Debug("create constraint")

	var (
		operator = strings.ToLower(r.Operator)
		now      = proto.TimestampNow()
		c        = &flipt.Constraint{
			Id:         uuid.NewV4().String(),
			SegmentKey: r.SegmentKey,
			Type:       r.Type,
			Property:   r.Property,
			Operator:   operator,
			Value:      r.Value,
			CreatedAt:  now,
			UpdatedAt:  now,
		}
	)

	// validate operator works for this constraint type
	switch c.Type {
	case flipt.ComparisonType_STRING_COMPARISON_TYPE:
		if _, ok := stringOperators[c.Operator]; !ok {
			return nil, ErrInvalidf("constraint operator %q is not valid for type string", r.Operator)
		}
	case flipt.ComparisonType_NUMBER_COMPARISON_TYPE:
		if _, ok := numberOperators[c.Operator]; !ok {
			return nil, ErrInvalidf("constraint operator %q is not valid for type number", r.Operator)
		}
	case flipt.ComparisonType_BOOLEAN_COMPARISON_TYPE:
		if _, ok := booleanOperators[c.Operator]; !ok {
			return nil, ErrInvalidf("constraint operator %q is not valid for type boolean", r.Operator)
		}
	default:
		return nil, ErrInvalidf("invalid constraint type: %q", c.Type.String())
	}

	// unset value if operator does not require it
	if _, ok := noValueOperators[c.Operator]; ok {
		c.Value = ""
	}

	query := s.builder.Insert("constraints").
		Columns("id", "segment_key", "type", "property", "operator", "value", "created_at", "updated_at").
		Values(c.Id, c.SegmentKey, c.Type, c.Property, c.Operator, c.Value, &timestamp{c.CreatedAt}, &timestamp{c.UpdatedAt})

	if _, err := query.ExecContext(ctx); err != nil {
		if sqliteErr, ok := err.(sqlite3.Error); ok {
			if sqliteErr.ExtendedCode == sqlite3.ErrConstraintForeignKey {
				return nil, ErrNotFoundf("segment %q", r.SegmentKey)
			}
		}
		return nil, err
	}

	s.logger.WithField("response", c).Debug("create constraint")
	return c, nil
}

// UpdateConstraint updates an existing constraint
func (s *SegmentStorage) UpdateConstraint(ctx context.Context, r *flipt.UpdateConstraintRequest) (*flipt.Constraint, error) {
	s.logger.WithField("request", r).Debug("update constraint")

	operator := strings.ToLower(r.Operator)
	// validate operator works for this constraint type
	switch r.Type {
	case flipt.ComparisonType_STRING_COMPARISON_TYPE:
		if _, ok := stringOperators[operator]; !ok {
			return nil, ErrInvalidf("constraint operator %q is not valid for type string", r.Operator)
		}
	case flipt.ComparisonType_NUMBER_COMPARISON_TYPE:
		if _, ok := numberOperators[operator]; !ok {
			return nil, ErrInvalidf("constraint operator %q is not valid for type number", r.Operator)
		}
	case flipt.ComparisonType_BOOLEAN_COMPARISON_TYPE:
		if _, ok := booleanOperators[operator]; !ok {
			return nil, ErrInvalidf("constraint operator %q is not valid for type boolean", r.Operator)
		}
	default:
		return nil, ErrInvalidf("invalid constraint type: %q", r.Type.String())
	}

	// unset value if operator does not require it
	if _, ok := noValueOperators[operator]; ok {
		r.Value = ""
	}

	res, err := s.builder.Update("constraints").
		Set("type", r.Type).
		Set("property", r.Property).
		Set("operator", operator).
		Set("value", r.Value).
		Set("updated_at", &timestamp{proto.TimestampNow()}).
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
		return nil, ErrNotFoundf("constraint %q", r.Id)
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

	s.logger.WithField("response", c).Debug("update constraint")
	return c, nil
}

// DeleteConstraint deletes a constraint
func (s *SegmentStorage) DeleteConstraint(ctx context.Context, r *flipt.DeleteConstraintRequest) error {
	s.logger.WithField("request", r).Debug("delete constraint")

	_, err := s.builder.Delete("constraints").
		Where(sq.And{sq.Eq{"id": r.Id}, sq.Eq{"segment_key": r.SegmentKey}}).
		ExecContext(ctx)

	return err
}

func (s *SegmentStorage) constraints(ctx context.Context, segment *flipt.Segment) error {
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
