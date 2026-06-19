package client

import (
	"context"
	"errors"
	"net"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	errs "go.flipt.io/flipt/errors"
	"go.flipt.io/flipt/internal/server/common"
	"go.flipt.io/flipt/internal/server/environments"
	"go.flipt.io/flipt/internal/server/evaluation"
	rpcevaluation "go.flipt.io/flipt/rpc/v2/evaluation"
	"go.uber.org/zap/zaptest"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
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
	assert.True(t, s.SkipsAuthorization(t.Context()))
}

type testInprocAddr struct{}

func (testInprocAddr) Network() string { return "inproc" }
func (testInprocAddr) String() string  { return "0" }

func TestServer_SkipsAuthentication(t *testing.T) {
	tests := []struct {
		name      string
		skipOFREP bool
		md        metadata.MD
		peer      *peer.Peer
		expected  bool
	}{
		{
			name:      "ofrep exclusion disabled",
			skipOFREP: false,
			expected:  false,
		},
		{
			name:      "no ofrep stream marker",
			skipOFREP: true,
			expected:  false,
		},
		{
			name:      "marker present but not inproc peer",
			skipOFREP: true,
			md:        metadata.Pairs(common.HeaderFliptOFREPStream, "true"),
			peer:      &peer.Peer{Addr: &net.TCPAddr{IP: net.IPv4(1, 2, 3, 4), Port: 9000}},
			expected:  false,
		},
		{
			name:      "marker present and inproc peer",
			skipOFREP: true,
			md:        metadata.Pairs(common.HeaderFliptOFREPStream, "true"),
			peer:      &peer.Peer{Addr: testInprocAddr{}},
			expected:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zaptest.NewLogger(t)
			s := NewServer(logger, nil, WithSkipOFREPAuthn(tt.skipOFREP))

			ctx := t.Context()
			if tt.md != nil {
				ctx = metadata.NewIncomingContext(ctx, tt.md)
			}
			if tt.peer != nil {
				ctx = peer.NewContext(ctx, tt.peer)
			}

			assert.Equal(t, tt.expected, s.SkipsAuthentication(ctx))
		})
	}
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
	resp, err := s.EvaluationSnapshotNamespace(t.Context(), req)
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
	envStore.On("GetFromContext", mock.Anything).Return(mockEnv, nil)

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
	resp, err := s.EvaluationSnapshotNamespace(t.Context(), req)
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
	resp, err := s.EvaluationSnapshotNamespace(t.Context(), req)
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
		ctx = metadata.NewIncomingContext(t.Context(), metadata.Pairs("GrpcGateway-If-None-Match", digest))
	)

	resp, err := s.EvaluationSnapshotNamespace(ctx, req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not modified")
	assert.Empty(t, resp.Digest)
}

type fakeCloser struct{}

func (f *fakeCloser) Close() error { return nil }

func TestServer_EvaluationSnapshotNamespaceStream(t *testing.T) {
	const (
		d1 = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
		d2 = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	)

	t.Run("invalid request digest length", func(t *testing.T) {
		logger := zaptest.NewLogger(t)
		stream := &mockStream{ctx: t.Context()}
		s := NewServer(logger, nil)
		req := &rpcevaluation.EvaluationNamespaceSnapshotStreamRequest{
			EnvironmentKey: "env-key",
			Key:            "ns-key",
			Digest:         new("too-short"),
		}

		err := s.EvaluationSnapshotNamespaceStream(req, stream)
		var target errs.ErrValidation
		require.ErrorAs(t, err, &target)
		assert.Equal(t, "invalid field digest: must be a 40-character string", target.Error())
	})

	t.Run("success", func(t *testing.T) {
		var (
			logger   = zaptest.NewLogger(t)
			mockEnv  = environments.NewMockEnvironment(t)
			envStore = evaluation.NewMockEnvironmentStore(t)
		)

		mockEnv.On("Key").Return("env-key")
		envStore.On("Get", mock.Anything, "env-key").Return(mockEnv, nil)

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

		stream := &mockStream{ctx: t.Context()}
		stream.On("Send", mock.Anything).Return(nil)
		s := NewServer(logger, envStore)
		req := &rpcevaluation.EvaluationNamespaceSnapshotStreamRequest{EnvironmentKey: "env-key", Key: "ns-key"}

		err := s.EvaluationSnapshotNamespaceStream(req, stream)
		require.NoError(t, err)
		assert.Len(t, stream.sent, 2)
		assert.Equal(t, "d1", stream.sent[0].Digest)
		assert.Equal(t, "d2", stream.sent[1].Digest)
	})

	t.Run("subscribe error", func(t *testing.T) {
		var (
			logger   = zaptest.NewLogger(t)
			mockEnv  = environments.NewMockEnvironment(t)
			envStore = evaluation.NewMockEnvironmentStore(t)
		)

		mockEnv.On("Key").Return("env-key")
		envStore.On("Get", mock.Anything, "env-key").Return(mockEnv, nil)

		mockEnv.On("EvaluationNamespaceSnapshotSubscribe", mock.Anything, "ns-key", mock.Anything).Return(nil, errors.New("subscribe error"))

		stream := &mockStream{ctx: t.Context()}
		stream.On("Send", mock.Anything).Return(nil)
		s := NewServer(logger, envStore)
		req := &rpcevaluation.EvaluationNamespaceSnapshotStreamRequest{EnvironmentKey: "env-key", Key: "ns-key"}

		err := s.EvaluationSnapshotNamespaceStream(req, stream)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "subscribe error")
	})

	t.Run("context cancel", func(t *testing.T) {
		var (
			logger   = zaptest.NewLogger(t)
			mockEnv  = environments.NewMockEnvironment(t)
			envStore = evaluation.NewMockEnvironmentStore(t)
		)

		mockEnv.On("Key").Return("env-key")
		envStore.On("Get", mock.Anything, "env-key").Return(mockEnv, nil)

		var (
			wait = make(chan struct{})
			once sync.Once
		)

		closeWait := func() { once.Do(func() { close(wait) }) }

		mockEnv.On("EvaluationNamespaceSnapshotSubscribe", mock.Anything, "ns-key", mock.Anything).Return(
			&fakeCloser{}, nil,
		).Run(func(args mock.Arguments) {
			ch := args.Get(2).(chan<- *rpcevaluation.EvaluationNamespaceSnapshot)
			go func() {
				ch <- &rpcevaluation.EvaluationNamespaceSnapshot{Digest: "d1"}
				<-wait
				close(ch)
			}()
		})

		ctx, cancel := context.WithCancel(t.Context())
		t.Cleanup(func() { closeWait(); cancel() })
		stream := &mockStream{ctx: ctx}
		stream.On("Send", mock.Anything).Return(nil).Run(func(_ mock.Arguments) {
			closeWait()
			cancel()
		})
		s := NewServer(logger, envStore)
		req := &rpcevaluation.EvaluationNamespaceSnapshotStreamRequest{EnvironmentKey: "env-key", Key: "ns-key"}

		err := s.EvaluationSnapshotNamespaceStream(req, stream)
		require.NoError(t, err)
		require.GreaterOrEqual(t, len(stream.sent), 1)
		assert.Equal(t, "d1", stream.sent[0].Digest)
	})

	t.Run("nil snapshot", func(t *testing.T) {
		var (
			logger   = zaptest.NewLogger(t)
			mockEnv  = environments.NewMockEnvironment(t)
			envStore = evaluation.NewMockEnvironmentStore(t)
		)

		mockEnv.On("Key").Return("env-key")
		envStore.On("Get", mock.Anything, "env-key").Return(mockEnv, nil)

		mockEnv.On("EvaluationNamespaceSnapshotSubscribe", mock.Anything, "ns-key", mock.Anything).Return(
			&fakeCloser{}, nil,
		).Run(func(args mock.Arguments) {
			ch := args.Get(2).(chan<- *rpcevaluation.EvaluationNamespaceSnapshot)
			go func() {
				ch <- nil
				ch <- &rpcevaluation.EvaluationNamespaceSnapshot{Digest: "d1"}
				close(ch)
			}()
		})

		stream := &mockStream{ctx: t.Context()}
		stream.On("Send", mock.Anything).Return(nil)
		s := NewServer(logger, envStore)
		req := &rpcevaluation.EvaluationNamespaceSnapshotStreamRequest{EnvironmentKey: "env-key", Key: "ns-key"}

		err := s.EvaluationSnapshotNamespaceStream(req, stream)
		require.NoError(t, err)
		require.Len(t, stream.sent, 1)
		assert.Equal(t, "d1", stream.sent[0].Digest)
	})

	t.Run("request digest match", func(t *testing.T) {
		var (
			logger   = zaptest.NewLogger(t)
			mockEnv  = environments.NewMockEnvironment(t)
			envStore = evaluation.NewMockEnvironmentStore(t)
		)

		mockEnv.On("Key").Return("env-key")
		envStore.On("Get", mock.Anything, "env-key").Return(mockEnv, nil)

		mockEnv.On("EvaluationNamespaceSnapshotSubscribe", mock.Anything, "ns-key", mock.Anything).Return(
			&fakeCloser{}, nil,
		).Run(func(args mock.Arguments) {
			ch := args.Get(2).(chan<- *rpcevaluation.EvaluationNamespaceSnapshot)
			go func() {
				ch <- &rpcevaluation.EvaluationNamespaceSnapshot{Digest: d1}
				close(ch)
			}()
		})

		stream := &mockStream{ctx: t.Context()}
		stream.On("Send", mock.Anything).Return(nil)
		s := NewServer(logger, envStore)
		req := &rpcevaluation.EvaluationNamespaceSnapshotStreamRequest{
			EnvironmentKey: "env-key",
			Key:            "ns-key",
			Digest:         new(d1),
		}

		err := s.EvaluationSnapshotNamespaceStream(req, stream)
		require.NoError(t, err)
		require.Empty(t, stream.sent)
	})

	t.Run("same digest dedup", func(t *testing.T) {
		var (
			logger   = zaptest.NewLogger(t)
			mockEnv  = environments.NewMockEnvironment(t)
			envStore = evaluation.NewMockEnvironmentStore(t)
		)

		mockEnv.On("Key").Return("env-key")
		envStore.On("Get", mock.Anything, "env-key").Return(mockEnv, nil)

		mockEnv.On("EvaluationNamespaceSnapshotSubscribe", mock.Anything, "ns-key", mock.Anything).Return(
			&fakeCloser{}, nil,
		).Run(func(args mock.Arguments) {
			ch := args.Get(2).(chan<- *rpcevaluation.EvaluationNamespaceSnapshot)
			go func() {
				ch <- &rpcevaluation.EvaluationNamespaceSnapshot{Digest: "d1"}
				ch <- &rpcevaluation.EvaluationNamespaceSnapshot{Digest: "d1"}
				close(ch)
			}()
		})

		stream := &mockStream{ctx: t.Context()}
		stream.On("Send", mock.Anything).Return(nil)
		s := NewServer(logger, envStore)
		req := &rpcevaluation.EvaluationNamespaceSnapshotStreamRequest{EnvironmentKey: "env-key", Key: "ns-key"}

		err := s.EvaluationSnapshotNamespaceStream(req, stream)
		require.NoError(t, err)
		require.Len(t, stream.sent, 1)
		assert.Equal(t, "d1", stream.sent[0].Digest)
	})

	t.Run("request digest revert", func(t *testing.T) {
		// Regression test: after the stream sends d2 to the client (who started with d1),
		// a snapshot reverting to d1 must still be sent. The initial-request digest
		// should only suppress the very first snapshot, not subsequent ones.
		var (
			logger   = zaptest.NewLogger(t)
			mockEnv  = environments.NewMockEnvironment(t)
			envStore = evaluation.NewMockEnvironmentStore(t)
		)

		mockEnv.On("Key").Return("env-key")
		envStore.On("Get", mock.Anything, "env-key").Return(mockEnv, nil)

		mockEnv.On("EvaluationNamespaceSnapshotSubscribe", mock.Anything, "ns-key", mock.Anything).Return(
			&fakeCloser{}, nil,
		).Run(func(args mock.Arguments) {
			ch := args.Get(2).(chan<- *rpcevaluation.EvaluationNamespaceSnapshot)
			go func() {
				ch <- &rpcevaluation.EvaluationNamespaceSnapshot{Digest: d1} // skipped: client already has d1
				ch <- &rpcevaluation.EvaluationNamespaceSnapshot{Digest: d2} // sent: new state
				ch <- &rpcevaluation.EvaluationNamespaceSnapshot{Digest: d1} // must be sent: client has d2, not d1
				close(ch)
			}()
		})

		stream := &mockStream{ctx: t.Context()}
		stream.On("Send", mock.Anything).Return(nil)
		s := NewServer(logger, envStore)
		req := &rpcevaluation.EvaluationNamespaceSnapshotStreamRequest{
			EnvironmentKey: "env-key",
			Key:            "ns-key",
			Digest:         new(d1),
		}

		err := s.EvaluationSnapshotNamespaceStream(req, stream)
		require.NoError(t, err)
		require.Len(t, stream.sent, 2)
		assert.Equal(t, d2, stream.sent[0].Digest)
		assert.Equal(t, d1, stream.sent[1].Digest)
	})

	t.Run("channel closed", func(t *testing.T) {
		var (
			logger   = zaptest.NewLogger(t)
			mockEnv  = environments.NewMockEnvironment(t)
			envStore = evaluation.NewMockEnvironmentStore(t)
		)

		mockEnv.On("Key").Return("env-key")
		envStore.On("Get", mock.Anything, "env-key").Return(mockEnv, nil)

		mockEnv.On("EvaluationNamespaceSnapshotSubscribe", mock.Anything, "ns-key", mock.Anything).Return(
			&fakeCloser{}, nil,
		).Run(func(args mock.Arguments) {
			ch := args.Get(2).(chan<- *rpcevaluation.EvaluationNamespaceSnapshot)
			go func() {
				close(ch)
			}()
		})

		stream := &mockStream{ctx: t.Context()}
		stream.On("Send", mock.Anything).Return(nil)
		s := NewServer(logger, envStore)
		req := &rpcevaluation.EvaluationNamespaceSnapshotStreamRequest{EnvironmentKey: "env-key", Key: "ns-key"}

		err := s.EvaluationSnapshotNamespaceStream(req, stream)
		require.NoError(t, err)
		require.Empty(t, stream.sent)
	})

	t.Run("send error", func(t *testing.T) {
		var (
			logger   = zaptest.NewLogger(t)
			mockEnv  = environments.NewMockEnvironment(t)
			envStore = evaluation.NewMockEnvironmentStore(t)
		)

		mockEnv.On("Key").Return("env-key")
		envStore.On("Get", mock.Anything, "env-key").Return(mockEnv, nil)

		mockEnv.On("EvaluationNamespaceSnapshotSubscribe", mock.Anything, "ns-key", mock.Anything).Return(
			&fakeCloser{}, nil,
		).Run(func(args mock.Arguments) {
			ch := args.Get(2).(chan<- *rpcevaluation.EvaluationNamespaceSnapshot)
			go func() {
				ch <- &rpcevaluation.EvaluationNamespaceSnapshot{Digest: "d1"}
				close(ch)
			}()
		})

		stream := &mockStream{ctx: t.Context()}
		stream.On("Send", mock.Anything).Return(errors.New("send error"))
		s := NewServer(logger, envStore)
		req := &rpcevaluation.EvaluationNamespaceSnapshotStreamRequest{EnvironmentKey: "env-key", Key: "ns-key"}

		err := s.EvaluationSnapshotNamespaceStream(req, stream)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "send error")
	})

	t.Run("env not found", func(t *testing.T) {
		var (
			logger   = zaptest.NewLogger(t)
			envStore = evaluation.NewMockEnvironmentStore(t)
		)

		envStore.On("Get", mock.Anything, "env-key").Return(nil, errors.New("not found"))
		envStore.On("GetFromContext", mock.Anything).Return(nil, errors.New("not found from context"))

		stream := &mockStream{ctx: t.Context()}
		s := NewServer(logger, envStore)
		req := &rpcevaluation.EvaluationNamespaceSnapshotStreamRequest{EnvironmentKey: "env-key", Key: "ns-key"}

		err := s.EvaluationSnapshotNamespaceStream(req, stream)
		require.Error(t, err)
	})
}
