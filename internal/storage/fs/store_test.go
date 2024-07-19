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

	resource := storage.NewResource("", "foo")
	storeMock.On("GetFlag", mock.Anything, resource).Return(&flipt.Flag{}, nil)

	_, err := ss.GetFlag(context.TODO(), resource)
	require.NoError(t, err)
}

func TestListFlags(t *testing.T) {
	storeMock := newSnapshotStoreMock()
	ss := NewStore(storeMock)

	ns := storage.NewNamespace("")
	listByDefault := storage.ListWithOptions[storage.NamespaceRequest](ns)
	storeMock.On("ListFlags", mock.Anything, listByDefault).Return(storage.ResultSet[*flipt.Flag]{}, nil)

	_, err := ss.ListFlags(context.TODO(), listByDefault)
	require.NoError(t, err)
}

func TestCountFlags(t *testing.T) {
	storeMock := newSnapshotStoreMock()
	ss := NewStore(storeMock)

	defaultNS := storage.NewNamespace("")
	storeMock.On("CountFlags", mock.Anything, defaultNS).Return(uint64(0), nil)

	_, err := ss.CountFlags(context.TODO(), defaultNS)
	require.NoError(t, err)
}

func TestGetRule(t *testing.T) {
	storeMock := newSnapshotStoreMock()
	ss := NewStore(storeMock)

	defaultNS := storage.NewNamespace("")
	storeMock.On("GetRule", mock.Anything, defaultNS, "rule").Return(&flipt.Rule{}, nil)

	_, err := ss.GetRule(context.TODO(), defaultNS, "rule")
	require.NoError(t, err)
}

func TestListRules(t *testing.T) {
	storeMock := newSnapshotStoreMock()
	ss := NewStore(storeMock)

	listByDefault := storage.ListWithOptions(
		storage.NewResource("", "flag"),
	)
	storeMock.On("ListRules", mock.Anything, listByDefault).Return(storage.ResultSet[*flipt.Rule]{}, nil)

	_, err := ss.ListRules(context.TODO(), listByDefault)
	require.NoError(t, err)
}

func TestCountRules(t *testing.T) {
	storeMock := newSnapshotStoreMock()
	ss := NewStore(storeMock)

	flag := storage.NewResource("", "flag")
	storeMock.On("CountRules", mock.Anything, flag).Return(uint64(0), nil)

	_, err := ss.CountRules(context.TODO(), flag)
	require.NoError(t, err)
}

func TestGetSegment(t *testing.T) {
	storeMock := newSnapshotStoreMock()
	ss := NewStore(storeMock)

	segment := storage.NewResource("", "segment")
	storeMock.On("GetSegment", mock.Anything, segment).Return(&flipt.Segment{}, nil)

	_, err := ss.GetSegment(context.TODO(), segment)
	require.NoError(t, err)
}

func TestListSegments(t *testing.T) {
	storeMock := newSnapshotStoreMock()
	ss := NewStore(storeMock)

	listByDefault := storage.ListWithOptions(
		storage.NewNamespace(""),
	)
	storeMock.On("ListSegments", mock.Anything, listByDefault).Return(storage.ResultSet[*flipt.Segment]{}, nil)

	_, err := ss.ListSegments(context.TODO(), listByDefault)
	require.NoError(t, err)
}

func TestCountSegments(t *testing.T) {
	storeMock := newSnapshotStoreMock()
	ss := NewStore(storeMock)

	defaultNs := storage.NewNamespace("")
	storeMock.On("CountSegments", mock.Anything, defaultNs).Return(uint64(0), nil)

	_, err := ss.CountSegments(context.TODO(), defaultNs)
	require.NoError(t, err)
}

func TestGetRollout(t *testing.T) {
	storeMock := newSnapshotStoreMock()
	ss := NewStore(storeMock)

	defaultNs := storage.NewNamespace("")
	storeMock.On("GetRollout", mock.Anything, defaultNs, "rollout").Return(&flipt.Rollout{}, nil)

	_, err := ss.GetRollout(context.TODO(), defaultNs, "rollout")
	require.NoError(t, err)
}

func TestListRollouts(t *testing.T) {
	storeMock := newSnapshotStoreMock()
	ss := NewStore(storeMock)

	listByDefault := storage.ListWithOptions(
		storage.NewResource("", "flag"),
	)
	storeMock.On("ListRollouts", mock.Anything, listByDefault).Return(storage.ResultSet[*flipt.Rollout]{}, nil)

	_, err := ss.ListRollouts(context.TODO(), listByDefault)
	require.NoError(t, err)
}

func TestCountRollouts(t *testing.T) {
	storeMock := newSnapshotStoreMock()
	ss := NewStore(storeMock)

	flag := storage.NewResource("", "flag")
	storeMock.On("CountRollouts", mock.Anything, flag).Return(uint64(0), nil)

	_, err := ss.CountRollouts(context.TODO(), flag)
	require.NoError(t, err)
}

func TestGetNamespace(t *testing.T) {
	storeMock := newSnapshotStoreMock()
	ss := NewStore(storeMock)

	ns := storage.NewNamespace("")
	storeMock.On("GetNamespace", mock.Anything, ns).Return(&flipt.Namespace{}, nil)

	_, err := ss.GetNamespace(context.TODO(), ns)
	require.NoError(t, err)
}

func TestListNamespaces(t *testing.T) {
	storeMock := newSnapshotStoreMock()
	ss := NewStore(storeMock)

	list := storage.ListWithOptions(
		storage.ReferenceRequest{},
	)
	storeMock.On("ListNamespaces", mock.Anything, list).Return(storage.ResultSet[*flipt.Namespace]{}, nil)

	_, err := ss.ListNamespaces(context.TODO(), list)
	require.NoError(t, err)
}

func TestCountNamespaces(t *testing.T) {
	storeMock := newSnapshotStoreMock()
	ss := NewStore(storeMock)

	ref := storage.ReferenceRequest{}
	storeMock.On("CountNamespaces", mock.Anything, ref).Return(uint64(0), nil)

	_, err := ss.CountNamespaces(context.TODO(), ref)
	require.NoError(t, err)
}

func TestGetEvaluationRules(t *testing.T) {
	storeMock := newSnapshotStoreMock()
	ss := NewStore(storeMock)

	flag := storage.NewResource("", "flag")
	storeMock.On("GetEvaluationRules", mock.Anything, flag).Return([]*storage.EvaluationRule{}, nil)

	_, err := ss.GetEvaluationRules(context.TODO(), flag)
	require.NoError(t, err)
}

func TestGetEvaluationDistributions(t *testing.T) {
	storeMock := newSnapshotStoreMock()
	ss := NewStore(storeMock)

	id := storage.NewID("id")
	storeMock.On("GetEvaluationDistributions", mock.Anything, id).Return([]*storage.EvaluationDistribution{}, nil)

	_, err := ss.GetEvaluationDistributions(context.TODO(), id)
	require.NoError(t, err)
}

func TestGetEvaluationRollouts(t *testing.T) {
	storeMock := newSnapshotStoreMock()
	ss := NewStore(storeMock)

	flag := storage.NewResource("", "flag")
	storeMock.On("GetEvaluationRollouts", mock.Anything, flag).Return([]*storage.EvaluationRollout{}, nil)

	_, err := ss.GetEvaluationRollouts(context.TODO(), flag)
	require.NoError(t, err)
}

func TestGetVersion(t *testing.T) {
	storeMock := newSnapshotStoreMock()
	ss := NewStore(storeMock)

	ns := storage.NewNamespace("default")
	storeMock.On("GetVersion", mock.Anything, ns).Return("x0-y1", nil)

	version, err := ss.GetVersion(context.TODO(), ns)
	require.NoError(t, err)
	require.Equal(t, "x0-y1", version)
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
func (s snapshotStoreMock) View(_ context.Context, _ storage.Reference, fn func(storage.ReadOnlyStore) error) error {
	return fn(s.StoreMock)
}

func (s snapshotStoreMock) String() string {
	return "mock"
}
