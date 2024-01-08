//nolint:goconst
package server

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/common"
	"go.flipt.io/flipt/internal/storage"
	flipt "go.flipt.io/flipt/rpc/flipt"
	"go.uber.org/zap/zaptest"
)

func TestGetSegment(t *testing.T) {
	var (
		store  = &common.StoreMock{}
		logger = zaptest.NewLogger(t)
		s      = &Server{
			logger: logger,
			store:  store,
		}
		req = &flipt.GetSegmentRequest{Key: "foo"}
	)

	store.On("GetSegment", mock.Anything, storage.NewResource("", "foo")).Return(&flipt.Segment{
		Key: req.Key,
	}, nil)

	got, err := s.GetSegment(context.TODO(), req)
	require.NoError(t, err)

	assert.NotNil(t, got)
}

func TestListSegments_PaginationOffset(t *testing.T) {
	var (
		store  = &common.StoreMock{}
		logger = zaptest.NewLogger(t)
		s      = &Server{
			logger: logger,
			store:  store,
		}
	)

	defer store.AssertExpectations(t)

	store.On("ListSegments", mock.Anything, storage.ListWithOptions(storage.NewNamespace(""),
		storage.ListWithQueryParamOptions[storage.NamespaceRequest](
			storage.WithOffset(10),
		),
	)).Return(
		storage.ResultSet[*flipt.Segment]{
			Results: []*flipt.Segment{
				{
					Key: "foo",
				},
			},
			NextPageToken: "YmFy",
		}, nil)

	store.On("CountSegments", mock.Anything, storage.NewNamespace("")).Return(uint64(1), nil)

	got, err := s.ListSegments(context.TODO(), &flipt.ListSegmentRequest{
		Offset: 10,
	})

	require.NoError(t, err)

	assert.NotEmpty(t, got.Segments)
	assert.Equal(t, "YmFy", got.NextPageToken)
	assert.Equal(t, int32(1), got.TotalCount)
}

func TestListSegments_PaginationPageToken(t *testing.T) {
	var (
		store  = &common.StoreMock{}
		logger = zaptest.NewLogger(t)
		s      = &Server{
			logger: logger,
			store:  store,
		}
	)

	defer store.AssertExpectations(t)

	store.On("ListSegments", mock.Anything, storage.ListWithOptions(storage.NewNamespace(""),
		storage.ListWithQueryParamOptions[storage.NamespaceRequest](
			storage.WithOffset(10),
			storage.WithPageToken("Zm9v"),
		),
	)).Return(
		storage.ResultSet[*flipt.Segment]{
			Results: []*flipt.Segment{
				{
					Key: "foo",
				},
			},
			NextPageToken: "YmFy",
		}, nil)

	store.On("CountSegments", mock.Anything, storage.NewNamespace("")).Return(uint64(1), nil)

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
		store  = &common.StoreMock{}
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
		store  = &common.StoreMock{}
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
		store  = &common.StoreMock{}
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
		store  = &common.StoreMock{}
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
		store  = &common.StoreMock{}
		logger = zaptest.NewLogger(t)
		s      = &Server{
			logger: logger,
			store:  store,
		}
		req = &flipt.UpdateConstraintRequest{
			Id:          "1",
			SegmentKey:  "segmentKey",
			Type:        flipt.ComparisonType_STRING_COMPARISON_TYPE,
			Property:    "property",
			Operator:    flipt.OpEQ,
			Value:       "value",
			Description: "desc",
		}
	)

	store.On("UpdateConstraint", mock.Anything, req).Return(&flipt.Constraint{
		Id:          req.Id,
		SegmentKey:  req.SegmentKey,
		Type:        req.Type,
		Property:    req.Property,
		Operator:    req.Operator,
		Value:       req.Value,
		Description: req.Description,
	}, nil)

	got, err := s.UpdateConstraint(context.TODO(), req)
	require.NoError(t, err)

	assert.NotNil(t, got)
}

func TestDeleteConstraint(t *testing.T) {
	var (
		store  = &common.StoreMock{}
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
