package environments

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	errs "go.flipt.io/flipt/errors"
	"go.flipt.io/flipt/rpc/v2/environments"
	"go.uber.org/zap/zaptest"
)

func TestServer_ProtectedEnvironmentRejectsMutations(t *testing.T) {
	logger := zaptest.NewLogger(t)

	mockEnv := NewMockEnvironment(t)
	mockEnv.On("Key").Return("production")
	mockEnv.On("Default").Return(true)
	mockEnv.On("Configuration").Return(&environments.EnvironmentConfiguration{})
	mockEnv.On("Protected").Return(true)

	store, err := NewEnvironmentStore(logger, mockEnv)
	require.NoError(t, err)

	server, err := NewServer(logger, store)
	require.NoError(t, err)

	t.Run("create namespace", func(t *testing.T) {
		_, err := server.CreateNamespace(t.Context(), &environments.UpdateNamespaceRequest{
			EnvironmentKey: "production",
			Key:            "default",
			Name:           "Default",
		})
		require.Error(t, err)
		assert.True(t, errs.AsMatch[errs.ErrInvalid](err))
		assert.Contains(t, err.Error(), `environment "production" is readonly`)
	})

	t.Run("update namespace", func(t *testing.T) {
		_, err := server.UpdateNamespace(t.Context(), &environments.UpdateNamespaceRequest{
			EnvironmentKey: "production",
			Key:            "default",
			Name:           "Default",
		})
		require.Error(t, err)
		assert.True(t, errs.AsMatch[errs.ErrInvalid](err))
		assert.Contains(t, err.Error(), `environment "production" is readonly`)
	})

	t.Run("delete namespace", func(t *testing.T) {
		_, err := server.DeleteNamespace(t.Context(), &environments.DeleteNamespaceRequest{
			EnvironmentKey: "production",
			Key:            "default",
		})
		require.Error(t, err)
		assert.True(t, errs.AsMatch[errs.ErrInvalid](err))
		assert.Contains(t, err.Error(), `environment "production" is readonly`)
	})

	t.Run("create resource", func(t *testing.T) {
		_, err := server.CreateResource(t.Context(), &environments.UpdateResourceRequest{
			EnvironmentKey: "production",
			NamespaceKey:   "default",
			Key:            "my-flag",
		})
		require.Error(t, err)
		assert.True(t, errs.AsMatch[errs.ErrInvalid](err))
		assert.Contains(t, err.Error(), `environment "production" is readonly`)
	})

	t.Run("update resource", func(t *testing.T) {
		_, err := server.UpdateResource(t.Context(), &environments.UpdateResourceRequest{
			EnvironmentKey: "production",
			NamespaceKey:   "default",
			Key:            "my-flag",
		})
		require.Error(t, err)
		assert.True(t, errs.AsMatch[errs.ErrInvalid](err))
		assert.Contains(t, err.Error(), `environment "production" is readonly`)
	})

	t.Run("delete resource", func(t *testing.T) {
		_, err := server.DeleteResource(t.Context(), &environments.DeleteResourceRequest{
			EnvironmentKey: "production",
			NamespaceKey:   "default",
			Key:            "my-flag",
			TypeUrl:        "flipt.core.Flag",
		})
		require.Error(t, err)
		assert.True(t, errs.AsMatch[errs.ErrInvalid](err))
		assert.Contains(t, err.Error(), `environment "production" is readonly`)
	})

	mockEnv.AssertNotCalled(t, "GetNamespace")
	mockEnv.AssertNotCalled(t, "CreateNamespace")
	mockEnv.AssertNotCalled(t, "UpdateNamespace")
	mockEnv.AssertNotCalled(t, "DeleteNamespace")
	mockEnv.AssertNotCalled(t, "Update")
}

func TestServer_BranchEnvironmentIsModifiable(t *testing.T) {
	logger := zaptest.NewLogger(t)

	base := "production"
	mockEnv := NewMockEnvironment(t)
	mockEnv.On("Key").Return("feature")
	mockEnv.On("Default").Return(true)
	mockEnv.On("Configuration").Return(&environments.EnvironmentConfiguration{
		Base: &base,
	})
	mockEnv.On("GetNamespace", mock.Anything, "default").
		Return(nil, errs.ErrNotFoundf("namespace %q", "default")).
		Once()
	mockEnv.On("CreateNamespace", mock.Anything, "", mock.MatchedBy(func(ns *environments.Namespace) bool {
		return ns.GetKey() == "default" && ns.GetName() == "Default"
	})).Return("rev-1", nil).Once()

	store, err := NewEnvironmentStore(logger, mockEnv)
	require.NoError(t, err)

	server, err := NewServer(logger, store)
	require.NoError(t, err)

	resp, err := server.CreateNamespace(t.Context(), &environments.UpdateNamespaceRequest{
		EnvironmentKey: "feature",
		Key:            "default",
		Name:           "Default",
	})
	require.NoError(t, err)
	assert.Equal(t, "rev-1", resp.Revision)
}

func TestServer_ListEnvironments_ProtectedFieldForBranches(t *testing.T) {
	logger := zaptest.NewLogger(t)

	t.Run("non-branch uses configured protected value", func(t *testing.T) {
		mockEnv := NewMockEnvironment(t)
		mockEnv.On("Key").Return("production")
		mockEnv.On("Default").Return(true)
		mockEnv.On("Protected").Return(true)
		mockEnv.On("Configuration").Return(&environments.EnvironmentConfiguration{})

		store, err := NewEnvironmentStore(logger, mockEnv)
		require.NoError(t, err)

		server, err := NewServer(logger, store)
		require.NoError(t, err)

		resp, err := server.ListEnvironments(t.Context(), &environments.ListEnvironmentsRequest{})
		require.NoError(t, err)
		require.Len(t, resp.Environments, 1)
		assert.True(t, resp.Environments[0].Protected)
	})

	t.Run("branch always reports not protected", func(t *testing.T) {
		base := "production"
		mockEnv := NewMockEnvironment(t)
		mockEnv.On("Key").Return("feature")
		mockEnv.On("Default").Return(true)
		mockEnv.On("Configuration").Return(&environments.EnvironmentConfiguration{
			Base: &base,
		})

		store, err := NewEnvironmentStore(logger, mockEnv)
		require.NoError(t, err)

		server, err := NewServer(logger, store)
		require.NoError(t, err)

		resp, err := server.ListEnvironments(t.Context(), &environments.ListEnvironmentsRequest{})
		require.NoError(t, err)
		require.Len(t, resp.Environments, 1)
		assert.False(t, resp.Environments[0].Protected)
	})
}

func TestServer_BranchEnvironmentResponseIsNotProtected(t *testing.T) {
	logger := zaptest.NewLogger(t)

	branched := NewMockEnvironment(t)
	branched.On("Key").Return("feature")
	branched.On("Default").Return(false)
	branched.On("Configuration").Return(&environments.EnvironmentConfiguration{})

	baseEnv := NewMockEnvironment(t)
	baseEnv.On("Key").Return("production")
	baseEnv.On("Default").Return(true)
	baseEnv.On("Branch", mock.Anything, "feature").Return(branched, nil).Once()

	store, err := NewEnvironmentStore(logger, baseEnv)
	require.NoError(t, err)

	server, err := NewServer(logger, store)
	require.NoError(t, err)

	resp, err := server.BranchEnvironment(t.Context(), &environments.BranchEnvironmentRequest{
		EnvironmentKey: "production",
		Key:            "feature",
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.False(t, resp.Protected)
	assert.Equal(t, "feature", resp.Key)
	assert.Equal(t, "feature", resp.Name)
}
