package environments

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/server/authz"
	rpcenvironments "go.flipt.io/flipt/rpc/v2/environments"
	"go.uber.org/zap/zaptest"
)

func TestServer_ListEnvironments_FailsClosedWithoutScope(t *testing.T) {
	newStore := func(t *testing.T) *EnvironmentStore {
		t.Helper()
		newEnvironment := func(key string) *MockEnvironment {
			env := NewMockEnvironment(t)
			env.On("Key").Return(key)
			env.On("Default").Return(key == "default")
			env.On("Configuration").Return(nil).Maybe()
			return env
		}

		store, err := NewEnvironmentStore(zaptest.NewLogger(t), newEnvironment("default"), newEnvironment("staging"))
		require.NoError(t, err)
		return store
	}

	tests := []struct {
		name      string
		scope     []string
		withScope bool
		wantKeys  []string
	}{
		{name: "missing scope", wantKeys: []string{}},
		{name: "empty scope", withScope: true, scope: []string{}, wantKeys: []string{}},
		{name: "partial scope", withScope: true, scope: []string{"staging"}, wantKeys: []string{"staging"}},
		{name: "wildcard scope", withScope: true, scope: []string{"*"}, wantKeys: []string{"default", "staging"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			if tt.withScope {
				ctx = context.WithValue(ctx, authz.EnvironmentsKey, tt.scope)
			}
			server, err := NewServer(zaptest.NewLogger(t), newStore(t))
			require.NoError(t, err)

			resp, err := server.ListEnvironments(ctx, &rpcenvironments.ListEnvironmentsRequest{})
			require.NoError(t, err)
			keys := make([]string, 0, len(resp.Environments))
			for _, env := range resp.Environments {
				keys = append(keys, env.Key)
			}
			assert.ElementsMatch(t, tt.wantKeys, keys)
		})
	}
}

func TestServer_ListNamespaces_FailsClosedWithoutScope(t *testing.T) {
	newServer := func(t *testing.T) *Server {
		t.Helper()
		env := NewMockEnvironment(t)
		env.On("Key").Return("production")
		env.On("Default").Return(true)
		env.On("ListNamespaces", mock.Anything).Return(&rpcenvironments.ListNamespacesResponse{
			Items: []*rpcenvironments.Namespace{{Key: "analytics"}, {Key: "reporting"}},
		}, nil).Maybe()

		store, err := NewEnvironmentStore(zaptest.NewLogger(t), env)
		require.NoError(t, err)
		server, err := NewServer(zaptest.NewLogger(t), store)
		require.NoError(t, err)
		return server
	}

	tests := []struct {
		name      string
		scope     []string
		withScope bool
		wantKeys  []string
	}{
		{name: "missing scope", wantKeys: []string{}},
		{name: "empty scope", withScope: true, scope: []string{}, wantKeys: []string{}},
		{name: "partial scope", withScope: true, scope: []string{"reporting"}, wantKeys: []string{"reporting"}},
		{name: "wildcard scope", withScope: true, scope: []string{"*"}, wantKeys: []string{"analytics", "reporting"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			if tt.withScope {
				ctx = context.WithValue(ctx, authz.NamespacesKey, tt.scope)
			}

			resp, err := newServer(t).ListNamespaces(ctx, &rpcenvironments.ListNamespacesRequest{EnvironmentKey: "production"})
			require.NoError(t, err)
			keys := make([]string, 0, len(resp.Items))
			for _, namespace := range resp.Items {
				keys = append(keys, namespace.Key)
			}
			assert.ElementsMatch(t, tt.wantKeys, keys)
		})
	}
}
