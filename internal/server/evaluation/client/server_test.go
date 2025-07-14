package client

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/server/environments"
	"go.flipt.io/flipt/internal/server/evaluation"
	rpcevaluation "go.flipt.io/flipt/rpc/v2/evaluation"
	"go.uber.org/zap/zaptest"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type mockStream struct {
	mock.Mock
	ctx  context.Context
	sent []*rpcevaluation.EvaluationNamespaceSnapshot
	grpc.ServerStream
}

func (m *mockStream) Context() context.Context {
	return m.ctx
}

func (m *mockStream) Send(snap *rpcevaluation.EvaluationNamespaceSnapshot) error {
	m.sent = append(m.sent, snap)
	return m.Called(snap).Error(0)
}

func TestNewServerAndSkipsAuthorization(t *testing.T) {
	var (
		logger = zaptest.NewLogger(t)
		store  = &environments.EnvironmentStore{}
		s      = NewServer(logger, store)
	)

	assert.NotNil(t, s)
	assert.True(t, s.SkipsAuthorization(context.Background()))
}

func TestServer_EvaluationSnapshotNamespace_Success(t *testing.T) {
	var (
		logger   = zaptest.NewLogger(t)
		mockEnv  = environments.NewMockEnvironment(t)
		envStore = evaluation.NewMockEnvironmentStore(t)
	)

	mockEnv.On("Key").Return("env-key")
	envStore.On("Get", mock.Anything, "env-key").Return(mockEnv, nil)

	expectedSnap := &rpcevaluation.EvaluationNamespaceSnapshot{
		Digest:    "digest",
		Namespace: &rpcevaluation.EvaluationNamespace{Key: "ns"},
		Flags:     []*rpcevaluation.EvaluationFlag{},
	}
	mockEnv.On("EvaluationNamespaceSnapshot", mock.Anything, "ns-key", mock.Anything).Return(expectedSnap, nil)

	s := NewServer(logger, envStore)
	req := &rpcevaluation.EvaluationNamespaceSnapshotRequest{
		EnvironmentKey: "env-key",
		Key:            "ns-key",
	}
	resp, err := s.EvaluationSnapshotNamespace(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, expectedSnap, resp)
}

func TestServer_EvaluationSnapshotNamespace_EnvNotFound(t *testing.T) {
	var (
		logger   = zaptest.NewLogger(t)
		mockEnv  = environments.NewMockEnvironment(t)
		envStore = evaluation.NewMockEnvironmentStore(t)
	)

	mockEnv.On("Key").Return("env-key")
	envStore.On("Get", mock.Anything, "env-key").Return(nil, errors.New("not found"))
	envStore.On("GetFromContext", mock.Anything).Return(mockEnv)

	expectedSnap := &rpcevaluation.EvaluationNamespaceSnapshot{
		Digest:    "digest",
		Namespace: &rpcevaluation.EvaluationNamespace{Key: "ns"},
		Flags:     []*rpcevaluation.EvaluationFlag{},
	}
	mockEnv.On("EvaluationNamespaceSnapshot", mock.Anything, "ns-key", mock.Anything).Return(expectedSnap, nil)

	s := NewServer(logger, envStore)
	req := &rpcevaluation.EvaluationNamespaceSnapshotRequest{
		EnvironmentKey: "env-key",
		Key:            "ns-key",
	}
	resp, err := s.EvaluationSnapshotNamespace(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, expectedSnap, resp)
}

func TestServer_EvaluationSnapshotNamespace_SnapshotError(t *testing.T) {
	var (
		logger   = zaptest.NewLogger(t)
		mockEnv  = environments.NewMockEnvironment(t)
		envStore = evaluation.NewMockEnvironmentStore(t)
	)

	mockEnv.On("Key").Return("env-key")
	envStore.On("Get", mock.Anything, "env-key").Return(mockEnv, nil)

	mockEnv.On("EvaluationNamespaceSnapshot", mock.Anything, "ns-key", mock.Anything).Return(nil, errors.New("snap error"))

	s := NewServer(logger, envStore)
	req := &rpcevaluation.EvaluationNamespaceSnapshotRequest{
		EnvironmentKey: "env-key",
		Key:            "ns-key",
	}
	resp, err := s.EvaluationSnapshotNamespace(context.Background(), req)
	assert.Nil(t, resp)
	assert.Error(t, err)
}

func TestServer_EvaluationSnapshotNamespace_EtagMatch(t *testing.T) {
	var (
		logger   = zaptest.NewLogger(t)
		mockEnv  = environments.NewMockEnvironment(t)
		envStore = evaluation.NewMockEnvironmentStore(t)
	)

	mockEnv.On("Key").Return("env-key")
	envStore.On("Get", mock.Anything, "env-key").Return(mockEnv, nil)

	digest := "digest"
	expectedSnap := &rpcevaluation.EvaluationNamespaceSnapshot{
		Digest:    digest,
		Namespace: &rpcevaluation.EvaluationNamespace{Key: "ns"},
		Flags:     []*rpcevaluation.EvaluationFlag{},
	}
	mockEnv.On("EvaluationNamespaceSnapshot", mock.Anything, "ns-key", mock.Anything).Return(expectedSnap, nil)

	var (
		s   = NewServer(logger, envStore)
		req = &rpcevaluation.EvaluationNamespaceSnapshotRequest{
			EnvironmentKey: "env-key",
			Key:            "ns-key",
		}
		ctx = metadata.NewIncomingContext(context.Background(), metadata.Pairs("GrpcGateway-If-None-Match", digest))
	)

	resp, err := s.EvaluationSnapshotNamespace(ctx, req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not modified")
	assert.Empty(t, resp.Digest)
}

type fakeCloser struct{}

func (f *fakeCloser) Close() error { return nil }

func TestServer_EvaluationSnapshotNamespaceStream_Success(t *testing.T) {
	var (
		logger   = zaptest.NewLogger(t)
		mockEnv  = environments.NewMockEnvironment(t)
		envStore = evaluation.NewMockEnvironmentStore(t)
	)

	mockEnv.On("Key").Return("env-key")
	envStore.On("GetFromContext", mock.Anything).Return(mockEnv)

	mockEnv.On("EvaluationNamespaceSnapshotSubscribe", mock.Anything, "ns-key", mock.Anything).Return(
		&fakeCloser{}, nil,
	).Run(func(args mock.Arguments) {
		ch := args.Get(2).(chan<- *rpcevaluation.EvaluationNamespaceSnapshot)
		go func() {
			ch <- &rpcevaluation.EvaluationNamespaceSnapshot{Digest: "d1"}
			ch <- &rpcevaluation.EvaluationNamespaceSnapshot{Digest: "d2"}
			close(ch)
		}()
	})

	stream := &mockStream{ctx: context.Background()}
	stream.On("Send", mock.Anything).Return(nil)
	s := NewServer(logger, envStore)
	req := &rpcevaluation.EvaluationNamespaceSnapshotStreamRequest{Key: "ns-key"}

	err := s.EvaluationSnapshotNamespaceStream(req, stream)
	require.NoError(t, err)
	assert.Len(t, stream.sent, 2)
	assert.Equal(t, "d1", stream.sent[0].Digest)
	assert.Equal(t, "d2", stream.sent[1].Digest)
}

func TestServer_EvaluationSnapshotNamespaceStream_SubscribeError(t *testing.T) {
	var (
		logger   = zaptest.NewLogger(t)
		mockEnv  = environments.NewMockEnvironment(t)
		envStore = evaluation.NewMockEnvironmentStore(t)
	)

	mockEnv.On("Key").Return("env-key")
	envStore.On("GetFromContext", mock.Anything).Return(mockEnv)

	mockEnv.On("EvaluationNamespaceSnapshotSubscribe", mock.Anything, "ns-key", mock.Anything).Return(nil, errors.New("subscribe error"))

	stream := &mockStream{ctx: context.Background()}
	stream.On("Send", mock.Anything).Return(nil)
	s := NewServer(logger, envStore)
	req := &rpcevaluation.EvaluationNamespaceSnapshotStreamRequest{Key: "ns-key"}

	err := s.EvaluationSnapshotNamespaceStream(req, stream)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "subscribe error")
}

func TestServer_EvaluationSnapshotNamespaceStream_ContextCancel(t *testing.T) {
	var (
		logger   = zaptest.NewLogger(t)
		mockEnv  = environments.NewMockEnvironment(t)
		envStore = evaluation.NewMockEnvironmentStore(t)
	)

	mockEnv.On("Key").Return("env-key")
	envStore.On("GetFromContext", mock.Anything).Return(mockEnv)
	wait := make(chan struct{})
	t.Cleanup(func() { close(wait) })
	mockEnv.On("EvaluationNamespaceSnapshotSubscribe", mock.Anything, "ns-key", mock.Anything).Return(
		&fakeCloser{}, nil,
	).Run(func(args mock.Arguments) {
		ch := args.Get(2).(chan<- *rpcevaluation.EvaluationNamespaceSnapshot)
		go func() {
			ch <- &rpcevaluation.EvaluationNamespaceSnapshot{Digest: "d1"}
			// simulate context cancel before next send
			wait <- struct{}{}
			close(ch)
		}()
	})

	ctx, cancel := context.WithCancel(t.Context())
	stream := &mockStream{ctx: ctx}
	stream.On("Send", mock.Anything).Return(nil)
	s := NewServer(logger, envStore)
	req := &rpcevaluation.EvaluationNamespaceSnapshotStreamRequest{Key: "ns-key"}

	go func() {
		<-wait
		cancel()
	}()

	err := s.EvaluationSnapshotNamespaceStream(req, stream)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(stream.sent), 1)
	assert.Equal(t, "d1", stream.sent[0].Digest)
}
