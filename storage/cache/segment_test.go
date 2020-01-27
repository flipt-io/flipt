package cache

import (
	"context"
	"testing"

	"github.com/markphelps/flipt/errors"
	flipt "github.com/markphelps/flipt/rpc"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetSegment(t *testing.T) {
	var (
		logger, _ = test.NewNullLogger()
		store     = &segmentStoreMock{}
		spy       = &cacherSpy{}
		subject   = NewSegmentCache(logger, spy, store)
	)

	store.On("GetSegment", mock.Anything, mock.Anything).Return(&flipt.Segment{Key: "foo"}, nil)

	got, err := subject.GetSegment(context.TODO(), &flipt.GetSegmentRequest{Key: "foo"})
	require.NoError(t, err)
	assert.NotNil(t, got)

	// shouldnt exist in the cache so it should be added
	spy.AssertCalled(t, "Set")
	spy.AssertCalled(t, "Get")

	got, err = subject.GetSegment(context.TODO(), &flipt.GetSegmentRequest{Key: "foo"})
	require.NoError(t, err)
	assert.NotNil(t, got)

	// should already exist in the cache so it should NOT be added
	spy.AssertNumberOfCalls(t, "Set", 1)
	spy.AssertNumberOfCalls(t, "Get", 2)
}

func TestGetSegmentNotFound(t *testing.T) {
	var (
		logger, _ = test.NewNullLogger()
		store     = &segmentStoreMock{}
		spy       = &cacherSpy{}
		subject   = NewSegmentCache(logger, spy, store)
	)

	store.On("GetSegment", mock.Anything, mock.Anything).Return(&flipt.Segment{}, errors.ErrNotFound("foo"))

	_, err := subject.GetSegment(context.TODO(), &flipt.GetSegmentRequest{Key: "foo"})
	require.Error(t, err)

	// doesnt exists so it should not be added
	spy.AssertNotCalled(t, "Set")
	spy.AssertCalled(t, "Get")
}

func TestListSegments(t *testing.T) {
	var (
		logger, _ = test.NewNullLogger()
		store     = &segmentStoreMock{}
		spy       = &cacherSpy{}
		subject   = NewSegmentCache(logger, spy, store)
	)

	store.On("ListSegments", mock.Anything, mock.Anything).Return([]*flipt.Segment{
		{Key: "foo"},
		{Key: "bar"},
	}, nil)

	got, err := subject.ListSegments(context.TODO(), &flipt.ListSegmentRequest{})
	require.NoError(t, err)
	assert.NotEmpty(t, got)
	assert.Len(t, got, 2)

	// shouldnt exist in the cache so it should be added
	spy.AssertCalled(t, "Set")
	spy.AssertCalled(t, "Get")

	got, err = subject.ListSegments(context.TODO(), &flipt.ListSegmentRequest{})
	require.NoError(t, err)
	assert.NotEmpty(t, got)
	assert.Len(t, got, 2)

	// should already exist in the cache so it should NOT be added
	spy.AssertNumberOfCalls(t, "Set", 1)
	spy.AssertNumberOfCalls(t, "Get", 2)
}

func TestCreateSegment(t *testing.T) {
	var (
		logger, _ = test.NewNullLogger()
		store     = &segmentStoreMock{}
		spy       = &cacherSpy{}
		subject   = NewSegmentCache(logger, spy, store)
	)

	store.On("CreateSegment", mock.Anything, mock.Anything).Return(&flipt.Segment{Key: "foo"}, nil)

	segment, err := subject.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{Key: "foo"})
	require.NoError(t, err)
	assert.NotNil(t, segment)

	// should not be added
	spy.AssertNotCalled(t, "Set")
	// should flush cache
	spy.AssertCalled(t, "Flush")
}

func TestUpdateSegment(t *testing.T) {
	var (
		logger, _ = test.NewNullLogger()
		store     = &segmentStoreMock{}
		spy       = &cacherSpy{}
		subject   = NewSegmentCache(logger, spy, store)
	)

	store.On("UpdateSegment", mock.Anything, mock.Anything).Return(&flipt.Segment{Key: "foo"}, nil)

	_, err := subject.UpdateSegment(context.TODO(), &flipt.UpdateSegmentRequest{Key: "foo"})
	require.NoError(t, err)

	// should not be added
	spy.AssertNotCalled(t, "Set")
	// should flush cache
	spy.AssertCalled(t, "Flush")
}

func TestDeleteSegment(t *testing.T) {
	var (
		logger, _ = test.NewNullLogger()
		store     = &segmentStoreMock{}
		spy       = &cacherSpy{}
		subject   = NewSegmentCache(logger, spy, store)
	)

	store.On("DeleteSegment", mock.Anything, mock.Anything).Return(nil)

	err := subject.DeleteSegment(context.TODO(), &flipt.DeleteSegmentRequest{Key: "foo"})
	require.NoError(t, err)

	// should not be added
	spy.AssertNotCalled(t, "Set")
	// should flush cache
	spy.AssertCalled(t, "Flush")
}

func TestCreateConstraint(t *testing.T) {
	var (
		logger, _ = test.NewNullLogger()
		store     = &segmentStoreMock{}
		spy       = &cacherSpy{}
		subject   = NewSegmentCache(logger, spy, store)
	)

	store.On("CreateConstraint", mock.Anything, mock.Anything).Return(&flipt.Constraint{SegmentKey: "foo"}, nil)

	_, err := subject.CreateConstraint(context.TODO(), &flipt.CreateConstraintRequest{SegmentKey: "foo"})
	require.NoError(t, err)

	// should not be added
	spy.AssertNotCalled(t, "Set")
	// should flush cache
	spy.AssertCalled(t, "Flush")
}

func TestUpdateConstraint(t *testing.T) {
	var (
		logger, _ = test.NewNullLogger()
		store     = &segmentStoreMock{}
		spy       = &cacherSpy{}
		subject   = NewSegmentCache(logger, spy, store)
	)

	store.On("UpdateConstraint", mock.Anything, mock.Anything).Return(&flipt.Constraint{SegmentKey: "foo"}, nil)

	_, err := subject.UpdateConstraint(context.TODO(), &flipt.UpdateConstraintRequest{SegmentKey: "foo"})
	require.NoError(t, err)

	// should not be added
	spy.AssertNotCalled(t, "Set")
	// should flush cache
	spy.AssertCalled(t, "Flush")
}

func TestDeleteConstraint(t *testing.T) {
	var (
		logger, _ = test.NewNullLogger()
		store     = &segmentStoreMock{}
		spy       = &cacherSpy{}
		subject   = NewSegmentCache(logger, spy, store)
	)

	store.On("DeleteConstraint", mock.Anything, mock.Anything).Return(nil)

	err := subject.DeleteConstraint(context.TODO(), &flipt.DeleteConstraintRequest{SegmentKey: "foo"})
	require.NoError(t, err)

	// should not be added
	spy.AssertNotCalled(t, "Set")
	// should flush cache
	spy.AssertCalled(t, "Flush")
}
