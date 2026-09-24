package environments

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/errors"
	"go.flipt.io/flipt/internal/common"
	"go.flipt.io/flipt/rpc/flipt"
	rpcenvironments "go.flipt.io/flipt/rpc/v2/environments"
	"go.uber.org/zap/zaptest"
)

func TestEnvironmentStore_GetFromContext_DefaultFallback(t *testing.T) {
	logger := zaptest.NewLogger(t)

	// Create a mock environment with default: true but key != "default"
	mockEnv := NewMockEnvironment(t)
	mockEnv.On("Key").Return("production")
	mockEnv.On("Default").Return(true)

	store, err := NewEnvironmentStore(logger, mockEnv)
	require.NoError(t, err)
	require.NotNil(t, store)

	tests := []struct {
		name          string
		setupContext  func() context.Context
		expectedKey   string
		expectError   bool
		errorContains string
	}{
		{
			name:         "no environment in context returns default environment",
			setupContext: t.Context,
			expectedKey:  "production",
			expectError:  false,
		},
		{
			name: "explicit 'default' in context falls back to default environment when no 'default' key exists",
			setupContext: func() context.Context {
				return common.WithFliptEnvironment(t.Context(), flipt.DefaultEnvironment)
			},
			expectedKey: "production",
			expectError: false,
		},
		{
			name: "explicit 'production' in context returns production environment",
			setupContext: func() context.Context {
				return common.WithFliptEnvironment(t.Context(), "production")
			},
			expectedKey: "production",
			expectError: false,
		},
		{
			name: "non-existent environment in context returns error",
			setupContext: func() context.Context {
				return common.WithFliptEnvironment(t.Context(), "non-existent")
			},
			expectError:   true,
			errorContains: "non-existent",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setupContext()
			env, err := store.GetFromContext(ctx)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
			} else {
				require.NoError(t, err)
				require.NotNil(t, env)
				assert.Equal(t, tt.expectedKey, env.Key())
			}
		})
	}
}

func TestEnvironmentStore_GetFromContext_WithActualDefaultEnvironment(t *testing.T) {
	logger := zaptest.NewLogger(t)

	// Create an environment with key "default"
	mockDefaultEnv := NewMockEnvironment(t)
	mockDefaultEnv.On("Key").Return("default")
	mockDefaultEnv.On("Default").Return(false)

	// Create another environment with default: true
	mockProdEnv := NewMockEnvironment(t)
	mockProdEnv.On("Key").Return("production")
	mockProdEnv.On("Default").Return(true)

	store, err := NewEnvironmentStore(logger, mockDefaultEnv, mockProdEnv)
	require.NoError(t, err)
	require.NotNil(t, store)

	tests := []struct {
		name        string
		contextEnv  string
		expectedKey string
	}{
		{
			name:        "explicit 'default' returns 'default' environment",
			contextEnv:  "default",
			expectedKey: "default",
		},
		{
			name:        "explicit 'production' returns 'production' environment",
			contextEnv:  "production",
			expectedKey: "production",
		},
		{
			name:        "no context environment returns default (production)",
			contextEnv:  "",
			expectedKey: "production",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			if tt.contextEnv != "" {
				ctx = common.WithFliptEnvironment(ctx, tt.contextEnv)
			}

			env, err := store.GetFromContext(ctx)
			require.NoError(t, err)
			require.NotNil(t, env)
			assert.Equal(t, tt.expectedKey, env.Key())
		})
	}
}

func TestEnvironmentStore_EvaluationReady(t *testing.T) {
	logger := zaptest.NewLogger(t)

	newStaticMock := func(key string, isDefault, ready bool) *MockEnvironment {
		m := NewMockEnvironment(t)
		m.On("Key").Return(key)
		m.On("Default").Return(isDefault)
		// Maybe: map iteration order is random, so a short-circuiting
		// EvaluationReady may never poll every environment.
		m.On("Configuration").Return(&rpcenvironments.EnvironmentConfiguration{}).Maybe()
		m.On("HasSnapshot").Return(ready).Maybe()
		return m
	}

	t.Run("empty store is not ready", func(t *testing.T) {
		only := NewMockEnvironment(t)
		only.On("Key").Return("only")
		only.On("Default").Return(true)
		store, err := NewEnvironmentStore(logger, only)
		require.NoError(t, err)
		store.Remove("only")
		assert.False(t, store.EvaluationReady())
	})

	t.Run("all static environments ready", func(t *testing.T) {
		store, err := NewEnvironmentStore(logger,
			newStaticMock("production", true, true),
			newStaticMock("staging", false, true))
		require.NoError(t, err)
		assert.True(t, store.EvaluationReady())
	})

	t.Run("one static environment not ready", func(t *testing.T) {
		store, err := NewEnvironmentStore(logger,
			newStaticMock("production", true, true),
			newStaticMock("staging", false, false))
		require.NoError(t, err)
		assert.False(t, store.EvaluationReady())
	})

	t.Run("unready branched environments are excluded", func(t *testing.T) {
		base := "production"
		branch := NewMockEnvironment(t)
		branch.On("Key").Return("feature-1")
		branch.On("Default").Return(false)
		branch.On("Configuration").Return(&rpcenvironments.EnvironmentConfiguration{Base: &base})
		// NOTE: no HasSnapshot expectation — excluded branches must never be polled.

		store, err := NewEnvironmentStore(logger, newStaticMock("production", true, true), branch)
		require.NoError(t, err)
		assert.True(t, store.EvaluationReady())
	})
}

func TestEnvironmentStore_NewEnvironmentStore_DefaultSelection(t *testing.T) {
	logger := zaptest.NewLogger(t)

	tests := []struct {
		name            string
		environments    []Environment
		expectedDefault string
		expectError     bool
	}{
		{
			name: "environment with default: true is selected",
			environments: func() []Environment {
				mockEnv1 := NewMockEnvironment(t)
				mockEnv1.On("Key").Return("staging")
				mockEnv1.On("Default").Return(false)

				mockEnv2 := NewMockEnvironment(t)
				mockEnv2.On("Key").Return("production")
				mockEnv2.On("Default").Return(true)

				return []Environment{mockEnv1, mockEnv2}
			}(),
			expectedDefault: "production",
		},
		{
			name: "environment named 'default' is used when no default: true",
			environments: func() []Environment {
				mockEnv := NewMockEnvironment(t)
				mockEnv.On("Key").Return("default")
				mockEnv.On("Default").Return(false)

				return []Environment{mockEnv}
			}(),
			expectedDefault: "default",
		},
		{
			name: "single environment is used as default",
			environments: func() []Environment {
				mockEnv := NewMockEnvironment(t)
				mockEnv.On("Key").Return("production")
				mockEnv.On("Default").Return(false)

				return []Environment{mockEnv}
			}(),
			expectedDefault: "production",
		},
		{
			name: "error when multiple environments and no default",
			environments: func() []Environment {
				mockEnv1 := NewMockEnvironment(t)
				mockEnv1.On("Key").Return("staging")
				mockEnv1.On("Default").Return(false)

				mockEnv2 := NewMockEnvironment(t)
				mockEnv2.On("Key").Return("production")
				mockEnv2.On("Default").Return(false)

				return []Environment{mockEnv1, mockEnv2}
			}(),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, err := NewEnvironmentStore(logger, tt.environments...)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, store)
				assert.Equal(t, tt.expectedDefault, store.defaultEnv.Key())
			}
		})
	}
}

func TestEnvironmentStore_AddBranch(t *testing.T) {
	logger := zaptest.NewLogger(t)

	newBranchMock := func(key, base string) *MockEnvironment {
		m := NewMockEnvironment(t)
		m.On("Key").Return(key).Maybe()
		m.On("Configuration").Return(&rpcenvironments.EnvironmentConfiguration{Base: &base}).Maybe()
		return m
	}

	t.Run("does not replace a static environment", func(t *testing.T) {
		production := NewMockEnvironment(t)
		production.On("Key").Return("production")
		production.On("Default").Return(true)

		store, err := NewEnvironmentStore(logger, production)
		require.NoError(t, err)

		err = store.AddBranch(newBranchMock("Production", "staging"))
		assert.True(t, errors.AsMatch[errors.ErrAlreadyExists](err))

		got, err := store.Get(t.Context(), "production")
		require.NoError(t, err)
		assert.Same(t, production, got)
	})

	t.Run("does not replace a branch of another environment", func(t *testing.T) {
		production := NewMockEnvironment(t)
		production.On("Key").Return("production")
		production.On("Default").Return(true)

		store, err := NewEnvironmentStore(logger, production)
		require.NoError(t, err)

		stagingFoo := newBranchMock("foo", "staging")
		require.NoError(t, store.AddBranch(stagingFoo))
		assert.True(t, errors.AsMatch[errors.ErrAlreadyExists](store.AddBranch(newBranchMock("foo", "production"))))

		got, err := store.Get(t.Context(), "foo")
		require.NoError(t, err)
		assert.Same(t, stagingFoo, got)
	})

	t.Run("adding the same environment again is a no-op", func(t *testing.T) {
		production := NewMockEnvironment(t)
		production.On("Key").Return("production")
		production.On("Default").Return(true)

		store, err := NewEnvironmentStore(logger, production)
		require.NoError(t, err)

		foo := newBranchMock("foo", "production")
		require.NoError(t, store.AddBranch(foo))
		require.NoError(t, store.AddBranch(foo))

		got, err := store.Get(t.Context(), "foo")
		require.NoError(t, err)
		assert.Same(t, foo, got)
	})
}

func TestEnvironmentStore_RemoveBranch(t *testing.T) {
	logger := zaptest.NewLogger(t)

	production := NewMockEnvironment(t)
	production.On("Key").Return("production")
	production.On("Default").Return(true)
	production.On("Configuration").Return(&rpcenvironments.EnvironmentConfiguration{})

	const base = "staging"
	foo := NewMockEnvironment(t)
	foo.On("Key").Return("foo")
	foo.On("Configuration").Return(&rpcenvironments.EnvironmentConfiguration{Base: new(base)})

	store, err := NewEnvironmentStore(logger, production)
	require.NoError(t, err)
	require.NoError(t, store.AddBranch(foo))

	// a deleted branch named after a static environment must not remove it
	store.RemoveBranch("staging", "production")
	_, err = store.Get(t.Context(), "production")
	require.NoError(t, err)

	// a branch is only removed on behalf of its own base environment
	store.RemoveBranch("production", "foo")
	_, err = store.Get(t.Context(), "foo")
	require.NoError(t, err)

	store.RemoveBranch("staging", "foo")
	_, err = store.Get(t.Context(), "foo")
	assert.True(t, errors.AsMatch[errors.ErrNotFound](err))
}

func TestEnvironmentStore_DeleteBranch_KeepsStaticEnvironment(t *testing.T) {
	logger := zaptest.NewLogger(t)

	production := NewMockEnvironment(t)
	production.On("Key").Return("production")
	production.On("Default").Return(true)
	production.On("Configuration").Return(&rpcenvironments.EnvironmentConfiguration{})

	staging := NewMockEnvironment(t)
	staging.On("Key").Return("staging")
	staging.On("Default").Return(false)
	staging.On("DeleteBranch", mock.Anything, "production").Return(nil)

	store, err := NewEnvironmentStore(logger, production, staging)
	require.NoError(t, err)

	require.NoError(t, store.DeleteBranch(t.Context(), "staging", "production"))

	got, err := store.Get(t.Context(), "production")
	require.NoError(t, err)
	assert.Same(t, production, got)
	staging.AssertExpectations(t)
}

func TestEnvironmentStore_BranchConcurrentWithAddRemove(t *testing.T) {
	logger := zaptest.NewLogger(t)

	const base = "default"
	defaultEnv := NewMockEnvironment(t)
	defaultEnv.On("Key").Return(base)
	defaultEnv.On("Default").Return(true)
	defaultEnv.On("Branch", mock.Anything, "new-branch").Return(nil, errors.ErrNotImplemented("branch"))

	store, err := NewEnvironmentStore(logger, defaultEnv)
	require.NoError(t, err)

	branches := make([]*MockEnvironment, 16)
	for i := range branches {
		m := NewMockEnvironment(t)
		m.On("Key").Return(fmt.Sprintf("branch-%d", i))
		m.On("Configuration").Return(&rpcenvironments.EnvironmentConfiguration{Base: new(base)})
		branches[i] = m
	}

	const iterations = 500
	var wg sync.WaitGroup

	// simulates the repository subscriber adding and removing discovered branches
	wg.Go(func() {
		for i := range iterations {
			b := branches[i%len(branches)]
			if err := store.AddBranch(b); err != nil {
				t.Error(err)
			}
			store.RemoveBranch(base, b.Key())
		}
	})

	// simulates concurrent BranchEnvironment API calls
	wg.Go(func() {
		for range iterations {
			_, err := store.Branch(t.Context(), base, "new-branch")
			assert.Error(t, err)
		}
	})

	wg.Wait()
}
