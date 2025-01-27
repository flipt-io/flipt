package fs

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/common"
	"go.flipt.io/flipt/internal/storage"
	"go.flipt.io/flipt/rpc/flipt/core"
)

func TestGetFlag(t *testing.T) {
	storeMock := newSnapshotStoreMock()
	ss := NewStore(storeMock)

	resource := storage.NewResource("", "foo")
	storeMock.On("GetFlag", mock.Anything, resource).Return(&core.Flag{}, nil)

	_, err := ss.GetFlag(context.TODO(), resource)
	require.NoError(t, err)
}

func TestListFlags(t *testing.T) {
	storeMock := newSnapshotStoreMock()
	ss := NewStore(storeMock)

	ns := storage.NewNamespace("")
	listByDefault := storage.ListWithOptions(ns)
	storeMock.On("ListFlags", mock.Anything, listByDefault).Return(storage.ResultSet[*core.Flag]{}, nil)

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

	_, err := ss.GetEvaluationDistributions(context.TODO(), storage.NewResource("", "flag"), id)
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
