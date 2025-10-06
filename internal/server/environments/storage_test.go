package environments

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/common"
	"go.flipt.io/flipt/rpc/flipt"
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
