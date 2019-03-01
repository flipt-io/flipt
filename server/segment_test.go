package server

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/protobuf/ptypes/empty"
	flipt "github.com/markphelps/flipt/proto"
	"github.com/markphelps/flipt/storage"
	"github.com/stretchr/testify/assert"
)

var _ storage.SegmentStore = &segmentStoreMock{}

type segmentStoreMock struct {
	getSegmentFn       func(context.Context, *flipt.GetSegmentRequest) (*flipt.Segment, error)
	listSegmentsFn     func(context.Context, *flipt.ListSegmentRequest) ([]*flipt.Segment, error)
	createSegmentFn    func(context.Context, *flipt.CreateSegmentRequest) (*flipt.Segment, error)
	updateSegmentFn    func(context.Context, *flipt.UpdateSegmentRequest) (*flipt.Segment, error)
	deleteSegmentFn    func(context.Context, *flipt.DeleteSegmentRequest) error
	createConstraintFn func(context.Context, *flipt.CreateConstraintRequest) (*flipt.Constraint, error)
	updateConstraintFn func(context.Context, *flipt.UpdateConstraintRequest) (*flipt.Constraint, error)
	deleteConstraintFn func(context.Context, *flipt.DeleteConstraintRequest) error
}

func (m *segmentStoreMock) GetSegment(ctx context.Context, r *flipt.GetSegmentRequest) (*flipt.Segment, error) {
	return m.getSegmentFn(ctx, r)
}

func (m *segmentStoreMock) ListSegments(ctx context.Context, r *flipt.ListSegmentRequest) ([]*flipt.Segment, error) {
	return m.listSegmentsFn(ctx, r)
}

func (m *segmentStoreMock) CreateSegment(ctx context.Context, r *flipt.CreateSegmentRequest) (*flipt.Segment, error) {
	return m.createSegmentFn(ctx, r)
}

func (m *segmentStoreMock) UpdateSegment(ctx context.Context, r *flipt.UpdateSegmentRequest) (*flipt.Segment, error) {
	return m.updateSegmentFn(ctx, r)
}

func (m *segmentStoreMock) DeleteSegment(ctx context.Context, r *flipt.DeleteSegmentRequest) error {
	return m.deleteSegmentFn(ctx, r)
}

func (m *segmentStoreMock) CreateConstraint(ctx context.Context, r *flipt.CreateConstraintRequest) (*flipt.Constraint, error) {
	return m.createConstraintFn(ctx, r)
}

func (m *segmentStoreMock) UpdateConstraint(ctx context.Context, r *flipt.UpdateConstraintRequest) (*flipt.Constraint, error) {
	return m.updateConstraintFn(ctx, r)
}

func (m *segmentStoreMock) DeleteConstraint(ctx context.Context, r *flipt.DeleteConstraintRequest) error {
	return m.deleteConstraintFn(ctx, r)
}

func TestGetSegment(t *testing.T) {
	tests := []struct {
		name    string
		req     *flipt.GetSegmentRequest
		f       func(context.Context, *flipt.GetSegmentRequest) (*flipt.Segment, error)
		segment *flipt.Segment
		e       error
	}{
		{
			name: "ok",
			req:  &flipt.GetSegmentRequest{Key: "key"},
			f: func(_ context.Context, r *flipt.GetSegmentRequest) (*flipt.Segment, error) {
				assert.NotNil(t, r)
				assert.Equal(t, "key", r.Key)

				return &flipt.Segment{
					Key: r.Key,
				}, nil
			},
			segment: &flipt.Segment{
				Key: "key",
			},
			e: nil,
		},
		{
			name: "emptyKey",
			req:  &flipt.GetSegmentRequest{Key: ""},
			f: func(_ context.Context, r *flipt.GetSegmentRequest) (*flipt.Segment, error) {
				assert.NotNil(t, r)
				assert.Equal(t, "", r.Key)

				return &flipt.Segment{
					Key: "",
				}, nil
			},
			segment: nil,
			e:       EmptyFieldError("key"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				SegmentStore: &segmentStoreMock{
					getSegmentFn: tt.f,
				},
			}

			segment, err := s.GetSegment(context.TODO(), tt.req)
			assert.Equal(t, tt.e, err)
			assert.Equal(t, tt.segment, segment)
		})
	}
}

func TestListSegments(t *testing.T) {
	tests := []struct {
		name     string
		req      *flipt.ListSegmentRequest
		f        func(context.Context, *flipt.ListSegmentRequest) ([]*flipt.Segment, error)
		segments *flipt.SegmentList
		e        error
	}{
		{
			name: "ok",
			req:  &flipt.ListSegmentRequest{},
			f: func(context.Context, *flipt.ListSegmentRequest) ([]*flipt.Segment, error) {
				return []*flipt.Segment{
					{Key: "1"},
				}, nil
			},
			segments: &flipt.SegmentList{
				Segments: []*flipt.Segment{
					{
						Key: "1",
					},
				},
			},
			e: nil,
		},
		{
			name: "error test",
			req:  &flipt.ListSegmentRequest{},
			f: func(context.Context, *flipt.ListSegmentRequest) ([]*flipt.Segment, error) {
				return nil, errors.New("error test")
			},
			segments: nil,
			e:        errors.New("error test"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				SegmentStore: &segmentStoreMock{
					listSegmentsFn: tt.f,
				},
			}

			segments, err := s.ListSegments(context.TODO(), tt.req)
			assert.Equal(t, tt.e, err)
			assert.Equal(t, tt.segments, segments)
		})
	}
}

func TestCreateSegment(t *testing.T) {
	tests := []struct {
		name    string
		req     *flipt.CreateSegmentRequest
		f       func(context.Context, *flipt.CreateSegmentRequest) (*flipt.Segment, error)
		segment *flipt.Segment
		e       error
	}{
		{
			name: "ok",
			req: &flipt.CreateSegmentRequest{
				Key:         "key",
				Name:        "name",
				Description: "desc",
			},
			f: func(_ context.Context, r *flipt.CreateSegmentRequest) (*flipt.Segment, error) {
				assert.NotNil(t, r)
				assert.Equal(t, "key", r.Key)
				assert.Equal(t, "name", r.Name)
				assert.Equal(t, "desc", r.Description)

				return &flipt.Segment{
					Key:         r.Key,
					Name:        r.Name,
					Description: r.Description,
				}, nil
			},
			segment: &flipt.Segment{
				Key:         "key",
				Name:        "name",
				Description: "desc",
			},
			e: nil,
		},
		{
			name: "emptyKey",
			req: &flipt.CreateSegmentRequest{
				Key:         "",
				Name:        "name",
				Description: "desc",
			},
			f: func(_ context.Context, r *flipt.CreateSegmentRequest) (*flipt.Segment, error) {
				assert.NotNil(t, r)
				assert.Equal(t, "", r.Key)
				assert.Equal(t, "name", r.Name)
				assert.Equal(t, "desc", r.Description)

				return &flipt.Segment{
					Key:         "",
					Name:        r.Name,
					Description: r.Description,
				}, nil
			},
			segment: nil,
			e:       EmptyFieldError("key"),
		},
		{
			name: "emptyName",
			req: &flipt.CreateSegmentRequest{
				Key:         "key",
				Name:        "",
				Description: "desc",
			},
			f: func(_ context.Context, r *flipt.CreateSegmentRequest) (*flipt.Segment, error) {
				assert.NotNil(t, r)
				assert.Equal(t, "key", r.Key)
				assert.Equal(t, "", r.Name)
				assert.Equal(t, "desc", r.Description)

				return &flipt.Segment{
					Key:         r.Key,
					Name:        "",
					Description: r.Description,
				}, nil
			},
			segment: nil,
			e:       EmptyFieldError("name"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				SegmentStore: &segmentStoreMock{
					createSegmentFn: tt.f,
				},
			}

			segment, err := s.CreateSegment(context.TODO(), tt.req)
			assert.Equal(t, tt.e, err)
			assert.Equal(t, tt.segment, segment)
		})
	}
}

func TestUpdateSegment(t *testing.T) {
	tests := []struct {
		name    string
		req     *flipt.UpdateSegmentRequest
		f       func(context.Context, *flipt.UpdateSegmentRequest) (*flipt.Segment, error)
		segment *flipt.Segment
		e       error
	}{
		{
			name: "ok",
			req: &flipt.UpdateSegmentRequest{
				Key:         "key",
				Name:        "name",
				Description: "desc",
			},
			f: func(_ context.Context, r *flipt.UpdateSegmentRequest) (*flipt.Segment, error) {
				assert.NotNil(t, r)
				assert.Equal(t, "key", r.Key)
				assert.Equal(t, "name", r.Name)
				assert.Equal(t, "desc", r.Description)

				return &flipt.Segment{
					Key:         r.Key,
					Name:        r.Name,
					Description: r.Description,
				}, nil
			},
			segment: &flipt.Segment{
				Key:         "key",
				Name:        "name",
				Description: "desc",
			},
			e: nil,
		},
		{
			name: "emptyKey",
			req: &flipt.UpdateSegmentRequest{
				Key:         "",
				Name:        "name",
				Description: "desc",
			},
			f: func(_ context.Context, r *flipt.UpdateSegmentRequest) (*flipt.Segment, error) {
				assert.NotNil(t, r)
				assert.Equal(t, "", r.Key)
				assert.Equal(t, "name", r.Name)
				assert.Equal(t, "desc", r.Description)

				return &flipt.Segment{
					Key:         "",
					Name:        r.Name,
					Description: r.Description,
				}, nil
			},
			segment: nil,
			e:       EmptyFieldError("key"),
		},
		{
			name: "emptyName",
			req: &flipt.UpdateSegmentRequest{
				Key:         "key",
				Name:        "",
				Description: "desc",
			},
			f: func(_ context.Context, r *flipt.UpdateSegmentRequest) (*flipt.Segment, error) {
				assert.NotNil(t, r)
				assert.Equal(t, "key", r.Key)
				assert.Equal(t, "name", r.Name)
				assert.Equal(t, "desc", r.Description)

				return &flipt.Segment{
					Key:         r.Key,
					Name:        "",
					Description: r.Description,
				}, nil
			},
			segment: nil,
			e:       EmptyFieldError("name"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				SegmentStore: &segmentStoreMock{
					updateSegmentFn: tt.f,
				},
			}

			segment, err := s.UpdateSegment(context.TODO(), tt.req)
			assert.Equal(t, tt.e, err)
			assert.Equal(t, tt.segment, segment)
		})
	}
}

func TestDeleteSegment(t *testing.T) {
	tests := []struct {
		name  string
		req   *flipt.DeleteSegmentRequest
		f     func(context.Context, *flipt.DeleteSegmentRequest) error
		empty *empty.Empty
		e     error
	}{
		{
			name: "ok",
			req:  &flipt.DeleteSegmentRequest{Key: "key"},
			f: func(_ context.Context, r *flipt.DeleteSegmentRequest) error {
				assert.NotNil(t, r)
				assert.Equal(t, "key", r.Key)
				return nil
			},
			empty: &empty.Empty{},
			e:     nil,
		},
		{
			name: "emptyKey",
			req:  &flipt.DeleteSegmentRequest{Key: ""},
			f: func(_ context.Context, r *flipt.DeleteSegmentRequest) error {
				assert.NotNil(t, r)
				assert.Equal(t, "", r.Key)
				return nil
			},
			empty: nil,
			e:     EmptyFieldError("key"),
		},
		{
			name: "error test",
			req:  &flipt.DeleteSegmentRequest{Key: "key"},
			f: func(_ context.Context, r *flipt.DeleteSegmentRequest) error {
				assert.NotNil(t, r)
				assert.Equal(t, "key", r.Key)
				return errors.New("error test")
			},
			empty: nil,
			e:     errors.New("error test"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				SegmentStore: &segmentStoreMock{
					deleteSegmentFn: tt.f,
				},
			}

			resp, err := s.DeleteSegment(context.TODO(), tt.req)
			assert.Equal(t, tt.e, err)
			assert.Equal(t, tt.empty, resp)
		})
	}
}

func TestCreateConstraint(t *testing.T) {
	tests := []struct {
		name       string
		req        *flipt.CreateConstraintRequest
		f          func(context.Context, *flipt.CreateConstraintRequest) (*flipt.Constraint, error)
		constraint *flipt.Constraint
		e          error
	}{
		{
			name: "ok",
			req: &flipt.CreateConstraintRequest{
				SegmentKey: "segmentKey",
				Type:       flipt.ComparisonType_BOOLEAN_COMPARISON_TYPE,
				Property:   "foo",
				Operator:   "EQ",
				Value:      "bar",
			},
			f: func(_ context.Context, r *flipt.CreateConstraintRequest) (*flipt.Constraint, error) {
				assert.NotNil(t, r)
				assert.Equal(t, "segmentKey", r.SegmentKey)
				assert.Equal(t, flipt.ComparisonType_BOOLEAN_COMPARISON_TYPE, r.Type)
				assert.Equal(t, "foo", r.Property)
				assert.Equal(t, "EQ", r.Operator)
				assert.Equal(t, "bar", r.Value)

				return &flipt.Constraint{
					SegmentKey: r.SegmentKey,
					Type:       r.Type,
					Property:   r.Property,
					Operator:   r.Operator,
					Value:      r.Value,
				}, nil
			},
			constraint: &flipt.Constraint{
				SegmentKey: "segmentKey",
				Type:       flipt.ComparisonType_BOOLEAN_COMPARISON_TYPE,
				Property:   "foo",
				Operator:   "EQ",
				Value:      "bar",
			},
			e: nil,
		},
		{
			name: "emptySegmentKey",
			req: &flipt.CreateConstraintRequest{
				SegmentKey: "",
				Type:       flipt.ComparisonType_BOOLEAN_COMPARISON_TYPE,
				Property:   "foo",
				Operator:   "EQ",
				Value:      "bar",
			},
			f: func(_ context.Context, r *flipt.CreateConstraintRequest) (*flipt.Constraint, error) {
				assert.NotNil(t, r)
				assert.Equal(t, "", r.SegmentKey)
				assert.Equal(t, flipt.ComparisonType_BOOLEAN_COMPARISON_TYPE, r.Type)
				assert.Equal(t, "foo", r.Property)
				assert.Equal(t, "EQ", r.Operator)
				assert.Equal(t, "bar", r.Value)

				return &flipt.Constraint{
					SegmentKey: "",
					Type:       r.Type,
					Property:   r.Property,
					Operator:   r.Operator,
					Value:      r.Value,
				}, nil
			},
			constraint: nil,
			e:          EmptyFieldError("segmentKey"),
		},
		{
			name: "emptyProperty",
			req: &flipt.CreateConstraintRequest{
				SegmentKey: "segmentKey",
				Type:       flipt.ComparisonType_BOOLEAN_COMPARISON_TYPE,
				Property:   "",
				Operator:   "EQ",
				Value:      "bar",
			},
			f: func(_ context.Context, r *flipt.CreateConstraintRequest) (*flipt.Constraint, error) {
				assert.NotNil(t, r)
				assert.Equal(t, "segmentKey", r.SegmentKey)
				assert.Equal(t, flipt.ComparisonType_BOOLEAN_COMPARISON_TYPE, r.Type)
				assert.Equal(t, "", r.Property)
				assert.Equal(t, "EQ", r.Operator)
				assert.Equal(t, "bar", r.Value)

				return &flipt.Constraint{
					SegmentKey: r.SegmentKey,
					Type:       r.Type,
					Property:   "",
					Operator:   r.Operator,
					Value:      r.Value,
				}, nil
			},
			constraint: nil,
			e:          EmptyFieldError("property"),
		},
		{
			name: "emptyOperator",
			req: &flipt.CreateConstraintRequest{
				SegmentKey: "segmentKey",
				Type:       flipt.ComparisonType_BOOLEAN_COMPARISON_TYPE,
				Property:   "foo",
				Operator:   "",
				Value:      "bar",
			},
			f: func(_ context.Context, r *flipt.CreateConstraintRequest) (*flipt.Constraint, error) {
				assert.NotNil(t, r)
				assert.Equal(t, "segmentKey", r.SegmentKey)
				assert.Equal(t, flipt.ComparisonType_BOOLEAN_COMPARISON_TYPE, r.Type)
				assert.Equal(t, "foo", r.Property)
				assert.Equal(t, "", r.Operator)
				assert.Equal(t, "bar", r.Value)

				return &flipt.Constraint{
					SegmentKey: r.SegmentKey,
					Type:       r.Type,
					Property:   r.Property,
					Operator:   "",
					Value:      r.Value,
				}, nil
			},
			constraint: nil,
			e:          EmptyFieldError("operator"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				SegmentStore: &segmentStoreMock{
					createConstraintFn: tt.f,
				},
			}

			constraint, err := s.CreateConstraint(context.TODO(), tt.req)
			assert.Equal(t, tt.e, err)
			assert.Equal(t, tt.constraint, constraint)
		})
	}
}

func TestUpdateConstraint(t *testing.T) {
	tests := []struct {
		name       string
		req        *flipt.UpdateConstraintRequest
		f          func(context.Context, *flipt.UpdateConstraintRequest) (*flipt.Constraint, error)
		constraint *flipt.Constraint
		e          error
	}{
		{
			name: "ok",
			req: &flipt.UpdateConstraintRequest{
				Id:         "1",
				SegmentKey: "segmentKey",
				Type:       flipt.ComparisonType_BOOLEAN_COMPARISON_TYPE,
				Property:   "foo",
				Operator:   "EQ",
				Value:      "bar",
			},
			f: func(_ context.Context, r *flipt.UpdateConstraintRequest) (*flipt.Constraint, error) {
				assert.NotNil(t, r)
				assert.Equal(t, "1", r.Id)
				assert.Equal(t, "segmentKey", r.SegmentKey)
				assert.Equal(t, flipt.ComparisonType_BOOLEAN_COMPARISON_TYPE, r.Type)
				assert.Equal(t, "foo", r.Property)
				assert.Equal(t, "EQ", r.Operator)
				assert.Equal(t, "bar", r.Value)

				return &flipt.Constraint{
					Id:         r.Id,
					SegmentKey: r.SegmentKey,
					Type:       r.Type,
					Property:   r.Property,
					Operator:   r.Operator,
					Value:      r.Value,
				}, nil
			},
			constraint: &flipt.Constraint{
				Id:         "1",
				SegmentKey: "segmentKey",
				Type:       flipt.ComparisonType_BOOLEAN_COMPARISON_TYPE,
				Property:   "foo",
				Operator:   "EQ",
				Value:      "bar",
			},
			e: nil,
		},
		{
			name: "emptyID",
			req: &flipt.UpdateConstraintRequest{
				Id:         "",
				SegmentKey: "segmentKey",
				Type:       flipt.ComparisonType_BOOLEAN_COMPARISON_TYPE,
				Property:   "foo",
				Operator:   "EQ",
				Value:      "bar",
			},
			f: func(_ context.Context, r *flipt.UpdateConstraintRequest) (*flipt.Constraint, error) {
				assert.NotNil(t, r)
				assert.Equal(t, "", r.Id)
				assert.Equal(t, "segmentKey", r.SegmentKey)
				assert.Equal(t, flipt.ComparisonType_BOOLEAN_COMPARISON_TYPE, r.Type)
				assert.Equal(t, "foo", r.Property)
				assert.Equal(t, "EQ", r.Operator)
				assert.Equal(t, "bar", r.Value)

				return &flipt.Constraint{
					Id:         "",
					SegmentKey: r.SegmentKey,
					Type:       r.Type,
					Property:   r.Property,
					Operator:   r.Operator,
					Value:      r.Value,
				}, nil
			},
			constraint: nil,
			e:          EmptyFieldError("id"),
		},
		{
			name: "emptySegmentKey",
			req: &flipt.UpdateConstraintRequest{
				Id:         "1",
				SegmentKey: "",
				Type:       flipt.ComparisonType_BOOLEAN_COMPARISON_TYPE,
				Property:   "foo",
				Operator:   "EQ",
				Value:      "bar",
			},
			f: func(_ context.Context, r *flipt.UpdateConstraintRequest) (*flipt.Constraint, error) {
				assert.NotNil(t, r)
				assert.Equal(t, "1", r.Id)
				assert.Equal(t, "", r.SegmentKey)
				assert.Equal(t, flipt.ComparisonType_BOOLEAN_COMPARISON_TYPE, r.Type)
				assert.Equal(t, "foo", r.Property)
				assert.Equal(t, "EQ", r.Operator)
				assert.Equal(t, "bar", r.Value)

				return &flipt.Constraint{
					Id:         r.Id,
					SegmentKey: "",
					Type:       r.Type,
					Property:   r.Property,
					Operator:   r.Operator,
					Value:      r.Value,
				}, nil
			},
			constraint: nil,
			e:          EmptyFieldError("segmentKey"),
		},
		{
			name: "emptyProperty",
			req: &flipt.UpdateConstraintRequest{
				Id:         "1",
				SegmentKey: "segmentKey",
				Type:       flipt.ComparisonType_BOOLEAN_COMPARISON_TYPE,
				Property:   "",
				Operator:   "EQ",
				Value:      "bar",
			},
			f: func(_ context.Context, r *flipt.UpdateConstraintRequest) (*flipt.Constraint, error) {
				assert.NotNil(t, r)
				assert.Equal(t, "1", r.Id)
				assert.Equal(t, "segmentKey", r.SegmentKey)
				assert.Equal(t, flipt.ComparisonType_BOOLEAN_COMPARISON_TYPE, r.Type)
				assert.Equal(t, "", r.Property)
				assert.Equal(t, "EQ", r.Operator)
				assert.Equal(t, "bar", r.Value)

				return &flipt.Constraint{
					Id:         r.Id,
					SegmentKey: r.SegmentKey,
					Type:       r.Type,
					Property:   "",
					Operator:   r.Operator,
					Value:      r.Value,
				}, nil
			},
			constraint: nil,
			e:          EmptyFieldError("property"),
		},
		{
			name: "emptyOperator",
			req: &flipt.UpdateConstraintRequest{
				Id:         "1",
				SegmentKey: "segmentKey",
				Type:       flipt.ComparisonType_BOOLEAN_COMPARISON_TYPE,
				Property:   "foo",
				Operator:   "",
				Value:      "bar",
			},
			f: func(_ context.Context, r *flipt.UpdateConstraintRequest) (*flipt.Constraint, error) {
				assert.NotNil(t, r)
				assert.Equal(t, "1", r.Id)
				assert.Equal(t, "segmentKey", r.SegmentKey)
				assert.Equal(t, flipt.ComparisonType_BOOLEAN_COMPARISON_TYPE, r.Type)
				assert.Equal(t, "foo", r.Property)
				assert.Equal(t, "", r.Operator)
				assert.Equal(t, "bar", r.Value)

				return &flipt.Constraint{
					Id:         r.Id,
					SegmentKey: r.SegmentKey,
					Type:       r.Type,
					Property:   r.Property,
					Operator:   "",
					Value:      r.Value,
				}, nil
			},
			constraint: nil,
			e:          EmptyFieldError("operator"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				SegmentStore: &segmentStoreMock{
					updateConstraintFn: tt.f,
				},
			}

			constraint, err := s.UpdateConstraint(context.TODO(), tt.req)
			assert.Equal(t, tt.e, err)
			assert.Equal(t, tt.constraint, constraint)
		})
	}
}

func TestDeleteConstraint(t *testing.T) {
	tests := []struct {
		name  string
		req   *flipt.DeleteConstraintRequest
		f     func(context.Context, *flipt.DeleteConstraintRequest) error
		empty *empty.Empty
		e     error
	}{
		{
			name: "ok",
			req:  &flipt.DeleteConstraintRequest{Id: "1", SegmentKey: "segmentKey"},
			f: func(_ context.Context, r *flipt.DeleteConstraintRequest) error {
				assert.NotNil(t, r)
				assert.Equal(t, "1", r.Id)
				assert.Equal(t, "segmentKey", r.SegmentKey)
				return nil
			},
			empty: &empty.Empty{},
			e:     nil,
		},
		{
			name: "emptyID",
			req:  &flipt.DeleteConstraintRequest{Id: "", SegmentKey: "segmentKey"},
			f: func(_ context.Context, r *flipt.DeleteConstraintRequest) error {
				assert.NotNil(t, r)
				assert.Equal(t, "", r.Id)
				assert.Equal(t, "segmentKey", r.SegmentKey)
				return nil
			},
			empty: nil,
			e:     EmptyFieldError("id"),
		},
		{
			name: "emptySegmentKey",
			req:  &flipt.DeleteConstraintRequest{Id: "id", SegmentKey: ""},
			f: func(_ context.Context, r *flipt.DeleteConstraintRequest) error {
				assert.NotNil(t, r)
				assert.Equal(t, "id", r.Id)
				assert.Equal(t, "", r.SegmentKey)
				return nil
			},
			empty: nil,
			e:     EmptyFieldError("segmentKey"),
		},
		{
			name: "error test",
			req:  &flipt.DeleteConstraintRequest{Id: "id", SegmentKey: "segmentKey"},
			f: func(_ context.Context, r *flipt.DeleteConstraintRequest) error {
				assert.NotNil(t, r)
				assert.Equal(t, "id", r.Id)
				assert.Equal(t, "segmentKey", r.SegmentKey)
				return errors.New("error test")
			},
			empty: nil,
			e:     errors.New("error test"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				SegmentStore: &segmentStoreMock{
					deleteConstraintFn: tt.f,
				},
			}

			resp, err := s.DeleteConstraint(context.TODO(), tt.req)
			assert.Equal(t, tt.e, err)
			assert.Equal(t, tt.empty, resp)
		})
	}
}
