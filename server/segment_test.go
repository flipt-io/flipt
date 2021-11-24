package server

import (
	"context"
	"testing"

	flipt "github.com/markphelps/flipt/rpc/flipt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetSegment(t *testing.T) {
	var (
		store = &storeMock{}
		s     = &Server{
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

func TestListSegments(t *testing.T) {
	var (
		store = &storeMock{}
		s     = &Server{
			logger: logger,
			store:  store,
		}
	)

	store.On("ListSegments", mock.Anything, mock.Anything).Return(
		[]*flipt.Segment{
			{
				Key: "foo",
			},
		}, nil)

	got, err := s.ListSegments(context.TODO(), &flipt.ListSegmentRequest{})
	require.NoError(t, err)

	assert.NotEmpty(t, got.Segments)
}

func TestCreateSegment(t *testing.T) {
	var (
		store = &storeMock{}
		s     = &Server{
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
		store = &storeMock{}
		s     = &Server{
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
		store = &storeMock{}
		s     = &Server{
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
		store = &storeMock{}
		s     = &Server{
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
		store = &storeMock{}
		s     = &Server{
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
		store = &storeMock{}
		s     = &Server{
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
