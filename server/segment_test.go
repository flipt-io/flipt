//nolint:goconst
package server

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	flipt "go.flipt.io/flipt/rpc/flipt"
	"go.flipt.io/flipt/storage"
	"go.uber.org/zap/zaptest"
)

func TestGetSegment(t *testing.T) {
	var (
		store  = &storeMock{}
		logger = zaptest.NewLogger(t)
		s      = &Server{
			logger: logger,
			store:  store,
		}
		req = &flipt.GetSegmentRequest{Key: "foo"}
	)

	store.On("GetSegment", mock.Anything, "foo").Return(&flipt.Segment{
		Key: req.Key,
	}, nil)

	got, err := s.GetSegment(context.TODO(), req)
	require.NoError(t, err)

	assert.NotNil(t, got)
}

func TestListSegments_PaginationNextPageToken(t *testing.T) {
	var (
		store  = &storeMock{}
		logger = zaptest.NewLogger(t)
		s      = &Server{
			logger: logger,
			store:  store,
		}
	)

	defer store.AssertExpectations(t)

	params := storage.QueryParams{}
	store.On("ListSegments", mock.Anything, mock.MatchedBy(func(opts []storage.QueryOption) bool {
		for _, opt := range opts {
			opt(&params)
		}

		// assert page token is preferred over offset
		return params.PageToken == "foo" && params.Offset == 0
	})).Return(
		storage.ResultSet[*flipt.Segment]{
			Results: []*flipt.Segment{
				{
					Key: "foo",
				},
			},
			NextPageToken: "bar",
		}, nil)

	store.On("CountSegments", mock.Anything).Return(uint64(1), nil)

	got, err := s.ListSegments(context.TODO(), &flipt.ListSegmentRequest{
		PageToken: "Zm9v",
		Offset:    10,
	})

	require.NoError(t, err)

	assert.NotEmpty(t, got.Segments)
	assert.Equal(t, "YmFy", got.NextPageToken)
	assert.Equal(t, int32(1), got.TotalCount)
}

func TestCreateSegment(t *testing.T) {
	var (
		store  = &storeMock{}
		logger = zaptest.NewLogger(t)
		s      = &Server{
			logger: logger,
			store:  store,
		}
		req = &flipt.CreateSegmentRequest{
			Key:         "key",
			Name:        "name",
			Description: "desc",
		}
	)

	store.On("CreateSegment", mock.Anything, req).Return(&flipt.Segment{
		Key:         req.Key,
		Name:        req.Name,
		Description: req.Description,
	}, nil)

	got, err := s.CreateSegment(context.TODO(), req)
	require.NoError(t, err)

	assert.NotNil(t, got)
}

func TestUpdateSegment(t *testing.T) {
	var (
		store  = &storeMock{}
		logger = zaptest.NewLogger(t)
		s      = &Server{
			logger: logger,
			store:  store,
		}
		req = &flipt.UpdateSegmentRequest{
			Key:         "key",
			Name:        "name",
			Description: "desc",
		}
	)

	store.On("UpdateSegment", mock.Anything, req).Return(&flipt.Segment{
		Key:         req.Key,
		Name:        req.Name,
		Description: req.Description,
	}, nil)

	got, err := s.UpdateSegment(context.TODO(), req)
	require.NoError(t, err)

	assert.NotNil(t, got)
}

func TestDeleteSegment(t *testing.T) {
	var (
		store  = &storeMock{}
		logger = zaptest.NewLogger(t)
		s      = &Server{
			logger: logger,
			store:  store,
		}
		req = &flipt.DeleteSegmentRequest{
			Key: "key",
		}
	)

	store.On("DeleteSegment", mock.Anything, req).Return(nil)

	got, err := s.DeleteSegment(context.TODO(), req)
	require.NoError(t, err)

	assert.NotNil(t, got)
}

func TestCreateConstraint(t *testing.T) {
	var (
		store  = &storeMock{}
		logger = zaptest.NewLogger(t)
		s      = &Server{
			logger: logger,
			store:  store,
		}
		req = &flipt.CreateConstraintRequest{
			SegmentKey: "segmentKey",
			Type:       flipt.ComparisonType_STRING_COMPARISON_TYPE,
			Property:   "property",
			Operator:   flipt.OpEQ,
			Value:      "value",
		}
	)

	store.On("CreateConstraint", mock.Anything, req).Return(&flipt.Constraint{
		Id:         "1",
		SegmentKey: req.SegmentKey,
		Type:       req.Type,
		Property:   req.Property,
		Operator:   req.Operator,
		Value:      req.Value,
	}, nil)

	got, err := s.CreateConstraint(context.TODO(), req)
	require.NoError(t, err)

	assert.NotNil(t, got)
}

func TestUpdateConstraint(t *testing.T) {
	var (
		store  = &storeMock{}
		logger = zaptest.NewLogger(t)
		s      = &Server{
			logger: logger,
			store:  store,
		}
		req = &flipt.UpdateConstraintRequest{
			Id:         "1",
			SegmentKey: "segmentKey",
			Type:       flipt.ComparisonType_STRING_COMPARISON_TYPE,
			Property:   "property",
			Operator:   flipt.OpEQ,
			Value:      "value",
		}
	)

	store.On("UpdateConstraint", mock.Anything, req).Return(&flipt.Constraint{
		Id:         req.Id,
		SegmentKey: req.SegmentKey,
		Type:       req.Type,
		Property:   req.Property,
		Operator:   req.Operator,
		Value:      req.Value,
	}, nil)

	got, err := s.UpdateConstraint(context.TODO(), req)
	require.NoError(t, err)

	assert.NotNil(t, got)
}

func TestDeleteConstraint(t *testing.T) {
	var (
		store  = &storeMock{}
		logger = zaptest.NewLogger(t)
		s      = &Server{
			logger: logger,
			store:  store,
		}
		req = &flipt.DeleteConstraintRequest{
			Id: "1",
		}
	)

	store.On("DeleteConstraint", mock.Anything, req).Return(nil)

	got, err := s.DeleteConstraint(context.TODO(), req)
	require.NoError(t, err)

	assert.NotNil(t, got)
}
