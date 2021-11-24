package cache

import (
	"context"
	"testing"

	"github.com/markphelps/flipt/errors"
	flipt "github.com/markphelps/flipt/rpc/gen/proto/go/flipt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetSegment(t *testing.T) {
	var (
		store   = &storeMock{}
		cacher  = &cacherSpy{}
		subject = NewStore(logger, cacher, store)
	)

	store.On("GetSegment", mock.Anything, mock.Anything).Return(&flipt.Segment{Key: "foo"}, nil)
	cacher.On("Get", mock.Anything).Return(&flipt.Segment{}, false).Once()
	cacher.On("Set", mock.Anything, mock.Anything)

	got, err := subject.GetSegment(context.TODO(), "foo")
	require.NoError(t, err)
	assert.NotNil(t, got)

	// shouldnt exist in the cache so it should be added
	cacher.AssertCalled(t, "Set", "segment:foo", mock.Anything)
	cacher.AssertCalled(t, "Get", "segment:foo")

	cacher.On("Get", mock.Anything).Return(&flipt.Segment{Key: "foo"}, true)

	got, err = subject.GetSegment(context.TODO(), "foo")
	require.NoError(t, err)
	assert.NotNil(t, got)

	// should already exist in the cache so it should NOT be added
	cacher.AssertNumberOfCalls(t, "Set", 1)
	cacher.AssertNumberOfCalls(t, "Get", 2)
}

func TestGetSegmentNotFound(t *testing.T) {
	var (
		store   = &storeMock{}
		cacher  = &cacherSpy{}
		subject = NewStore(logger, cacher, store)
	)

	store.On("GetSegment", mock.Anything, mock.Anything).Return(&flipt.Segment{}, errors.ErrNotFound("foo"))
	cacher.On("Get", mock.Anything).Return(&flipt.Segment{}, false).Once()

	_, err := subject.GetSegment(context.TODO(), "foo")
	require.Error(t, err)

	// doesnt exists so it should not be added
	cacher.AssertNotCalled(t, "Set")
	cacher.AssertCalled(t, "Get", "segment:foo")
}

func TestListSegments(t *testing.T) {
	var (
		store   = &storeMock{}
		cacher  = &cacherSpy{}
		subject = NewStore(logger, cacher, store)
	)

	ret := []*flipt.Segment{
		{Key: "foo"},
		{Key: "bar"},
	}

	store.On("ListSegments", mock.Anything, mock.Anything).Return(ret, nil)

	got, err := subject.ListSegments(context.TODO())
	require.NoError(t, err)
	assert.NotEmpty(t, got)
	assert.Len(t, got, 2)

	cacher.AssertNotCalled(t, "Set")
	cacher.AssertNotCalled(t, "Get")
}

func TestCreateSegment(t *testing.T) {
	var (
		store   = &storeMock{}
		cacher  = &cacherSpy{}
		subject = NewStore(logger, cacher, store)
	)

	store.On("CreateSegment", mock.Anything, mock.Anything).Return(&flipt.Segment{Key: "foo"}, nil)
	cacher.On("Flush")

	segment, err := subject.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{Key: "foo"})
	require.NoError(t, err)
	assert.NotNil(t, segment)

	// should not be added
	cacher.AssertNotCalled(t, "Set")
	// should flush cache
	cacher.AssertCalled(t, "Flush")
}

func TestUpdateSegment(t *testing.T) {
	var (
		store   = &storeMock{}
		cacher  = &cacherSpy{}
		subject = NewStore(logger, cacher, store)
	)

	store.On("UpdateSegment", mock.Anything, mock.Anything).Return(&flipt.Segment{Key: "foo"}, nil)
	cacher.On("Flush")

	_, err := subject.UpdateSegment(context.TODO(), &flipt.UpdateSegmentRequest{Key: "foo"})
	require.NoError(t, err)

	// should not be added
	cacher.AssertNotCalled(t, "Set")
	// should flush cache
	cacher.AssertCalled(t, "Flush")
}

func TestDeleteSegment(t *testing.T) {
	var (
		store   = &storeMock{}
		cacher  = &cacherSpy{}
		subject = NewStore(logger, cacher, store)
	)

	store.On("DeleteSegment", mock.Anything, mock.Anything).Return(nil)
	cacher.On("Flush")

	err := subject.DeleteSegment(context.TODO(), &flipt.DeleteSegmentRequest{Key: "foo"})
	require.NoError(t, err)

	// should not be added
	cacher.AssertNotCalled(t, "Set")
	// should flush cache
	cacher.AssertCalled(t, "Flush")
}

func TestCreateConstraint(t *testing.T) {
	var (
		store   = &storeMock{}
		cacher  = &cacherSpy{}
		subject = NewStore(logger, cacher, store)
	)

	store.On("CreateConstraint", mock.Anything, mock.Anything).Return(&flipt.Constraint{SegmentKey: "foo"}, nil)
	cacher.On("Flush")

	_, err := subject.CreateConstraint(context.TODO(), &flipt.CreateConstraintRequest{SegmentKey: "foo"})
	require.NoError(t, err)

	// should not be added
	cacher.AssertNotCalled(t, "Set")
	// should flush cache
	cacher.AssertCalled(t, "Flush")
}

func TestUpdateConstraint(t *testing.T) {
	var (
		store   = &storeMock{}
		cacher  = &cacherSpy{}
		subject = NewStore(logger, cacher, store)
	)

	store.On("UpdateConstraint", mock.Anything, mock.Anything).Return(&flipt.Constraint{SegmentKey: "foo"}, nil)
	cacher.On("Flush")

	_, err := subject.UpdateConstraint(context.TODO(), &flipt.UpdateConstraintRequest{SegmentKey: "foo"})
	require.NoError(t, err)

	// should not be added
	cacher.AssertNotCalled(t, "Set")
	// should flush cache
	cacher.AssertCalled(t, "Flush")
}

func TestDeleteConstraint(t *testing.T) {
	var (
		store   = &storeMock{}
		cacher  = &cacherSpy{}
		subject = NewStore(logger, cacher, store)
	)

	store.On("DeleteConstraint", mock.Anything, mock.Anything).Return(nil)
	cacher.On("Flush")

	err := subject.DeleteConstraint(context.TODO(), &flipt.DeleteConstraintRequest{SegmentKey: "foo"})
	require.NoError(t, err)

	// should not be added
	cacher.AssertNotCalled(t, "Set")
	// should flush cache
	cacher.AssertCalled(t, "Flush")
}
