package server

import (
	"context"
	"testing"

	"github.com/markphelps/flipt/errors"

	"github.com/golang/protobuf/ptypes/empty"
	flipt "github.com/markphelps/flipt/rpc"
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
	if m.getSegmentFn == nil {
		return &flipt.Segment{}, nil
	}
	return m.getSegmentFn(ctx, r)
}

func (m *segmentStoreMock) ListSegments(ctx context.Context, r *flipt.ListSegmentRequest) ([]*flipt.Segment, error) {
	if m.listSegmentsFn == nil {
		return []*flipt.Segment{}, nil
	}
	return m.listSegmentsFn(ctx, r)
}

func (m *segmentStoreMock) CreateSegment(ctx context.Context, r *flipt.CreateSegmentRequest) (*flipt.Segment, error) {
	if m.createSegmentFn == nil {
		return &flipt.Segment{}, nil
	}
	return m.createSegmentFn(ctx, r)
}

func (m *segmentStoreMock) UpdateSegment(ctx context.Context, r *flipt.UpdateSegmentRequest) (*flipt.Segment, error) {
	if m.updateSegmentFn == nil {
		return &flipt.Segment{}, nil
	}
	return m.updateSegmentFn(ctx, r)
}

func (m *segmentStoreMock) DeleteSegment(ctx context.Context, r *flipt.DeleteSegmentRequest) error {
	if m.deleteSegmentFn == nil {
		return nil
	}
	return m.deleteSegmentFn(ctx, r)
}

func (m *segmentStoreMock) CreateConstraint(ctx context.Context, r *flipt.CreateConstraintRequest) (*flipt.Constraint, error) {
	if m.createConstraintFn == nil {
		return &flipt.Constraint{}, nil
	}
	return m.createConstraintFn(ctx, r)
}

func (m *segmentStoreMock) UpdateConstraint(ctx context.Context, r *flipt.UpdateConstraintRequest) (*flipt.Constraint, error) {
	if m.updateConstraintFn == nil {
		return &flipt.Constraint{}, nil
	}
	return m.updateConstraintFn(ctx, r)
}

func (m *segmentStoreMock) DeleteConstraint(ctx context.Context, r *flipt.DeleteConstraintRequest) error {
	if m.deleteConstraintFn == nil {
		return nil
	}
	return m.deleteConstraintFn(ctx, r)
}

func TestGetSegment(t *testing.T) {
	tests := []struct {
		name    string
		req     *flipt.GetSegmentRequest
		f       func(context.Context, *flipt.GetSegmentRequest) (*flipt.Segment, error)
		segment *flipt.Segment
		wantErr error
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
		},
	}

	for _, tt := range tests {
		var (
			f       = tt.f
			req     = tt.req
			segment = tt.segment
			wantErr = tt.wantErr
		)

		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				SegmentStore: &segmentStoreMock{
					getSegmentFn: f,
				},
			}

			got, err := s.GetSegment(context.TODO(), req)
			assert.Equal(t, wantErr, err)
			assert.Equal(t, segment, got)
		})
	}
}

func TestListSegments(t *testing.T) {
	tests := []struct {
		name     string
		req      *flipt.ListSegmentRequest
		f        func(context.Context, *flipt.ListSegmentRequest) ([]*flipt.Segment, error)
		segments *flipt.SegmentList
		wantErr  error
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
		},
		{
			name: "error",
			req:  &flipt.ListSegmentRequest{},
			f: func(context.Context, *flipt.ListSegmentRequest) ([]*flipt.Segment, error) {
				return nil, errors.New("error test")
			},
			wantErr: errors.New("error test"),
		},
	}

	for _, tt := range tests {
		var (
			f        = tt.f
			req      = tt.req
			segments = tt.segments
			wantErr  = tt.wantErr
		)

		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				SegmentStore: &segmentStoreMock{
					listSegmentsFn: f,
				},
			}

			got, err := s.ListSegments(context.TODO(), req)
			assert.Equal(t, wantErr, err)
			assert.Equal(t, segments, got)
		})
	}
}

func TestCreateSegment(t *testing.T) {
	tests := []struct {
		name    string
		req     *flipt.CreateSegmentRequest
		f       func(context.Context, *flipt.CreateSegmentRequest) (*flipt.Segment, error)
		segment *flipt.Segment
		wantErr error
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
		},
	}

	for _, tt := range tests {
		var (
			f       = tt.f
			req     = tt.req
			segment = tt.segment
			wantErr = tt.wantErr
		)

		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				SegmentStore: &segmentStoreMock{
					createSegmentFn: f,
				},
			}

			got, err := s.CreateSegment(context.TODO(), req)
			assert.Equal(t, wantErr, err)
			assert.Equal(t, segment, got)
		})
	}
}

func TestUpdateSegment(t *testing.T) {
	tests := []struct {
		name    string
		req     *flipt.UpdateSegmentRequest
		f       func(context.Context, *flipt.UpdateSegmentRequest) (*flipt.Segment, error)
		segment *flipt.Segment
		wantErr error
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
		},
	}

	for _, tt := range tests {
		var (
			f       = tt.f
			req     = tt.req
			segment = tt.segment
			wantErr = tt.wantErr
		)

		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				SegmentStore: &segmentStoreMock{
					updateSegmentFn: f,
				},
			}

			got, err := s.UpdateSegment(context.TODO(), req)
			assert.Equal(t, wantErr, err)
			assert.Equal(t, segment, got)
		})
	}
}

func TestDeleteSegment(t *testing.T) {
	tests := []struct {
		name    string
		req     *flipt.DeleteSegmentRequest
		f       func(context.Context, *flipt.DeleteSegmentRequest) error
		empty   *empty.Empty
		wantErr error
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
		},
		{
			name: "error",
			req:  &flipt.DeleteSegmentRequest{Key: "key"},
			f: func(_ context.Context, r *flipt.DeleteSegmentRequest) error {
				assert.NotNil(t, r)
				assert.Equal(t, "key", r.Key)
				return errors.New("error test")
			},
			wantErr: errors.New("error test"),
		},
	}

	for _, tt := range tests {
		var (
			f       = tt.f
			req     = tt.req
			empty   = tt.empty
			wantErr = tt.wantErr
		)

		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				SegmentStore: &segmentStoreMock{
					deleteSegmentFn: f,
				},
			}

			got, err := s.DeleteSegment(context.TODO(), req)
			assert.Equal(t, wantErr, err)
			assert.Equal(t, empty, got)
		})
	}
}

func TestCreateConstraint(t *testing.T) {
	tests := []struct {
		name       string
		req        *flipt.CreateConstraintRequest
		f          func(context.Context, *flipt.CreateConstraintRequest) (*flipt.Constraint, error)
		constraint *flipt.Constraint
		wantErr    error
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
		},
	}

	for _, tt := range tests {
		var (
			f          = tt.f
			req        = tt.req
			constraint = tt.constraint
			wantErr    = tt.wantErr
		)

		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				SegmentStore: &segmentStoreMock{
					createConstraintFn: f,
				},
			}

			got, err := s.CreateConstraint(context.TODO(), req)
			assert.Equal(t, wantErr, err)
			assert.Equal(t, constraint, got)
		})
	}
}

func TestUpdateConstraint(t *testing.T) {
	tests := []struct {
		name       string
		req        *flipt.UpdateConstraintRequest
		f          func(context.Context, *flipt.UpdateConstraintRequest) (*flipt.Constraint, error)
		constraint *flipt.Constraint
		wantErr    error
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
		},
	}

	for _, tt := range tests {
		var (
			f          = tt.f
			req        = tt.req
			constraint = tt.constraint
			wantErr    = tt.wantErr
		)

		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				SegmentStore: &segmentStoreMock{
					updateConstraintFn: f,
				},
			}

			got, err := s.UpdateConstraint(context.TODO(), req)
			assert.Equal(t, wantErr, err)
			assert.Equal(t, constraint, got)
		})
	}
}

func TestDeleteConstraint(t *testing.T) {
	tests := []struct {
		name    string
		req     *flipt.DeleteConstraintRequest
		f       func(context.Context, *flipt.DeleteConstraintRequest) error
		empty   *empty.Empty
		wantErr error
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
		},
		{
			name: "error",
			req:  &flipt.DeleteConstraintRequest{Id: "id", SegmentKey: "segmentKey"},
			f: func(_ context.Context, r *flipt.DeleteConstraintRequest) error {
				assert.NotNil(t, r)
				assert.Equal(t, "id", r.Id)
				assert.Equal(t, "segmentKey", r.SegmentKey)
				return errors.New("error test")
			},
			wantErr: errors.New("error test"),
		},
	}

	for _, tt := range tests {
		var (
			f       = tt.f
			req     = tt.req
			empty   = tt.empty
			wantErr = tt.wantErr
		)

		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				SegmentStore: &segmentStoreMock{
					deleteConstraintFn: f,
				},
			}

			got, err := s.DeleteConstraint(context.TODO(), req)
			assert.Equal(t, wantErr, err)
			assert.Equal(t, empty, got)
		})
	}
}
