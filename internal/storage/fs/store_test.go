package fs

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/common"
	"go.flipt.io/flipt/internal/storage"
	"go.flipt.io/flipt/rpc/flipt"
)

func TestGetFlag(t *testing.T) {
	storeMock := newSnapshotStoreMock()
	ss := NewStore(storeMock)

	storeMock.On("GetFlag", mock.Anything, flipt.DefaultNamespace, "foo").Return(&flipt.Flag{}, nil)

	_, err := ss.GetFlag(context.TODO(), "", "foo")
	require.NoError(t, err)
}

func TestListFlags(t *testing.T) {
	storeMock := newSnapshotStoreMock()
	ss := NewStore(storeMock)

	storeMock.On("ListFlags", mock.Anything, flipt.DefaultNamespace, mock.Anything).Return(storage.ResultSet[*flipt.Flag]{}, nil)

	_, err := ss.ListFlags(context.TODO(), "")
	require.NoError(t, err)
}

func TestCountFlags(t *testing.T) {
	storeMock := newSnapshotStoreMock()
	ss := NewStore(storeMock)

	storeMock.On("CountFlags", mock.Anything, flipt.DefaultNamespace).Return(uint64(0), nil)

	_, err := ss.CountFlags(context.TODO(), "")
	require.NoError(t, err)
}

func TestGetRule(t *testing.T) {
	storeMock := newSnapshotStoreMock()
	ss := NewStore(storeMock)

	storeMock.On("GetRule", mock.Anything, flipt.DefaultNamespace, "").Return(&flipt.Rule{}, nil)

	_, err := ss.GetRule(context.TODO(), "", "")
	require.NoError(t, err)
}

func TestListRules(t *testing.T) {
	storeMock := newSnapshotStoreMock()
	ss := NewStore(storeMock)

	storeMock.On("ListRules", mock.Anything, flipt.DefaultNamespace, "", mock.Anything).Return(storage.ResultSet[*flipt.Rule]{}, nil)

	_, err := ss.ListRules(context.TODO(), "", "")
	require.NoError(t, err)
}

func TestCountRules(t *testing.T) {
	storeMock := newSnapshotStoreMock()
	ss := NewStore(storeMock)

	storeMock.On("CountRules", mock.Anything, flipt.DefaultNamespace, "").Return(uint64(0), nil)

	_, err := ss.CountRules(context.TODO(), "", "")
	require.NoError(t, err)
}

func TestGetSegment(t *testing.T) {
	storeMock := newSnapshotStoreMock()
	ss := NewStore(storeMock)

	storeMock.On("GetSegment", mock.Anything, flipt.DefaultNamespace, "").Return(&flipt.Segment{}, nil)

	_, err := ss.GetSegment(context.TODO(), "", "")
	require.NoError(t, err)
}

func TestListSegments(t *testing.T) {
	storeMock := newSnapshotStoreMock()
	ss := NewStore(storeMock)

	storeMock.On("ListSegments", mock.Anything, flipt.DefaultNamespace, mock.Anything).Return(storage.ResultSet[*flipt.Segment]{}, nil)

	_, err := ss.ListSegments(context.TODO(), "")
	require.NoError(t, err)
}

func TestCountSegments(t *testing.T) {
	storeMock := newSnapshotStoreMock()
	ss := NewStore(storeMock)

	storeMock.On("CountSegments", mock.Anything, flipt.DefaultNamespace).Return(uint64(0), nil)

	_, err := ss.CountSegments(context.TODO(), "")
	require.NoError(t, err)
}

func TestGetRollout(t *testing.T) {
	storeMock := newSnapshotStoreMock()
	ss := NewStore(storeMock)

	storeMock.On("GetRollout", mock.Anything, flipt.DefaultNamespace, "").Return(&flipt.Rollout{}, nil)

	_, err := ss.GetRollout(context.TODO(), "", "")
	require.NoError(t, err)
}

func TestListRollouts(t *testing.T) {
	storeMock := newSnapshotStoreMock()
	ss := NewStore(storeMock)

	storeMock.On("ListRollouts", mock.Anything, flipt.DefaultNamespace, "", mock.Anything).Return(storage.ResultSet[*flipt.Rollout]{}, nil)

	_, err := ss.ListRollouts(context.TODO(), "", "")
	require.NoError(t, err)
}

func TestCountRollouts(t *testing.T) {
	storeMock := newSnapshotStoreMock()
	ss := NewStore(storeMock)

	storeMock.On("CountRollouts", mock.Anything, flipt.DefaultNamespace, "").Return(uint64(0), nil)

	_, err := ss.CountRollouts(context.TODO(), "", "")
	require.NoError(t, err)
}

func TestGetNamespace(t *testing.T) {
	storeMock := newSnapshotStoreMock()
	ss := NewStore(storeMock)

	storeMock.On("GetNamespace", mock.Anything, flipt.DefaultNamespace).Return(&flipt.Namespace{}, nil)

	_, err := ss.GetNamespace(context.TODO(), "")
	require.NoError(t, err)
}

func TestListNamespaces(t *testing.T) {
	storeMock := newSnapshotStoreMock()
	ss := NewStore(storeMock)

	storeMock.On("ListNamespaces", mock.Anything, mock.Anything).Return(storage.ResultSet[*flipt.Namespace]{}, nil)

	_, err := ss.ListNamespaces(context.TODO())
	require.NoError(t, err)
}

func TestCountNamespaces(t *testing.T) {
	storeMock := newSnapshotStoreMock()
	ss := NewStore(storeMock)

	storeMock.On("CountNamespaces", mock.Anything).Return(uint64(0), nil)

	_, err := ss.CountNamespaces(context.TODO())
	require.NoError(t, err)
}

func TestGetEvaluationRules(t *testing.T) {
	storeMock := newSnapshotStoreMock()
	ss := NewStore(storeMock)

	storeMock.On("GetEvaluationRules", mock.Anything, flipt.DefaultNamespace, "").Return([]*storage.EvaluationRule{}, nil)

	_, err := ss.GetEvaluationRules(context.TODO(), "", "")
	require.NoError(t, err)
}

func TestGetEvaluationDistributions(t *testing.T) {
	storeMock := newSnapshotStoreMock()
	ss := NewStore(storeMock)

	storeMock.On("GetEvaluationDistributions", mock.Anything, "").Return([]*storage.EvaluationDistribution{}, nil)

	_, err := ss.GetEvaluationDistributions(context.TODO(), "")
	require.NoError(t, err)
}

func TestGetEvaluationRollouts(t *testing.T) {
	storeMock := newSnapshotStoreMock()
	ss := NewStore(storeMock)

	storeMock.On("GetEvaluationRollouts", mock.Anything, flipt.DefaultNamespace, "").Return([]*storage.EvaluationRollout{}, nil)

	_, err := ss.GetEvaluationRollouts(context.TODO(), "", "")
	require.NoError(t, err)
}

type snapshotStoreMock struct {
	*common.StoreMock
}

func newSnapshotStoreMock() snapshotStoreMock {
	return snapshotStoreMock{
		StoreMock: &common.StoreMock{},
	}
}

// View accepts a function which takes a *StoreSnapshot.
// The SnapshotStore will supply a snapshot which is valid
// for the lifetime of the provided function call.
func (s snapshotStoreMock) View(fn func(storage.ReadOnlyStore) error) error {
	return fn(s.StoreMock)
}

func (s snapshotStoreMock) String() string {
	return "mock"
}
