package server

import (
	"context"
	"testing"

	"github.com/golang/protobuf/ptypes/empty"
	flipt "github.com/markphelps/flipt/proto"
	"github.com/markphelps/flipt/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		name string
		req  *flipt.GetSegmentRequest
		f    func(context.Context, *flipt.GetSegmentRequest) (*flipt.Segment, error)
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
			require.NoError(t, err)
			assert.NotNil(t, segment)
		})
	}
}

func TestListSegments(t *testing.T) {
	tests := []struct {
		name string
		req  *flipt.ListSegmentRequest
		f    func(context.Context, *flipt.ListSegmentRequest) ([]*flipt.Segment, error)
	}{
		{
			name: "ok",
			req:  &flipt.ListSegmentRequest{},
			f: func(context.Context, *flipt.ListSegmentRequest) ([]*flipt.Segment, error) {
				return []*flipt.Segment{
					{Key: "1"},
				}, nil
			},
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
			require.NoError(t, err)
			assert.NotEmpty(t, segments)
		})
	}
}

func TestCreateSegment(t *testing.T) {
	tests := []struct {
		name string
		req  *flipt.CreateSegmentRequest
		f    func(context.Context, *flipt.CreateSegmentRequest) (*flipt.Segment, error)
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
			require.NoError(t, err)
			assert.NotNil(t, segment)
		})
	}
}

func TestUpdateSegment(t *testing.T) {
	tests := []struct {
		name string
		req  *flipt.UpdateSegmentRequest
		f    func(context.Context, *flipt.UpdateSegmentRequest) (*flipt.Segment, error)
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
			require.NoError(t, err)
			assert.NotNil(t, segment)
		})
	}
}

func TestDeleteSegment(t *testing.T) {
	tests := []struct {
		name string
		req  *flipt.DeleteSegmentRequest
		f    func(context.Context, *flipt.DeleteSegmentRequest) error
	}{
		{
			name: "ok",
			req:  &flipt.DeleteSegmentRequest{Key: "key"},
			f: func(_ context.Context, r *flipt.DeleteSegmentRequest) error {
				assert.NotNil(t, r)
				assert.Equal(t, "key", r.Key)
				return nil
			},
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
			require.NoError(t, err)
			assert.Equal(t, &empty.Empty{}, resp)
		})
	}
}

func TestCreateConstraint(t *testing.T) {
	tests := []struct {
		name string
		req  *flipt.CreateConstraintRequest
		f    func(context.Context, *flipt.CreateConstraintRequest) (*flipt.Constraint, error)
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
			require.NoError(t, err)
			assert.NotNil(t, constraint)
		})
	}
}

func TestUpdateConstraint(t *testing.T) {
	tests := []struct {
		name string
		req  *flipt.UpdateConstraintRequest
		f    func(context.Context, *flipt.UpdateConstraintRequest) (*flipt.Constraint, error)
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
			require.NoError(t, err)
			assert.NotNil(t, constraint)
		})
	}
}

func TestDeleteConstraint(t *testing.T) {
	tests := []struct {
		name string
		req  *flipt.DeleteConstraintRequest
		f    func(context.Context, *flipt.DeleteConstraintRequest) error
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
			require.NoError(t, err)
			assert.Equal(t, &empty.Empty{}, resp)
		})
	}
}
