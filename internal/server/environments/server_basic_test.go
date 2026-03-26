package environments

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	rpcenvironments "go.flipt.io/flipt/rpc/v2/environments"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func TestNewServerAndRegisterGRPC(t *testing.T) {
	s, err := NewServer(zap.NewNop(), nil, &EnvironmentStore{})
	require.NoError(t, err)
	require.NotNil(t, s)

	grpcServer := grpc.NewServer()
	assert.NotPanics(t, func() {
		s.RegisterGRPC(grpcServer)
	})
}

func TestGetNamespace(t *testing.T) {
	ctx := t.Context()
	env := NewMockEnvironment(t)

	env.EXPECT().
		GetNamespace(ctx, "default").
		Return(&rpcenvironments.NamespaceResponse{
			Namespace: &rpcenvironments.Namespace{
				Key: "default",
			},
			Revision: "rev-1",
		}, nil).
		Once()

	s := &Server{
		logger: zap.NewNop(),
		envs: &EnvironmentStore{
			byKey: map[string]Environment{
				"env-a": env,
			},
		},
	}

	resp, err := s.GetNamespace(ctx, &rpcenvironments.GetNamespaceRequest{
		EnvironmentKey: "env-a",
		Key:            "default",
	})
	require.NoError(t, err)
	assert.Equal(t, "default", resp.Namespace.Key)
	assert.Equal(t, "rev-1", resp.Revision)
}
