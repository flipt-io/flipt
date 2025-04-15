package rego

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/server/authz/engine/rego/source"
	"go.flipt.io/flipt/rpc/flipt"
	"go.uber.org/zap/zaptest"
)

func TestEngine_NewEngine(t *testing.T) {
	ctx := context.Background()

	policy, err := os.ReadFile("../testdata/rbac_v2.rego")
	require.NoError(t, err)

	data, err := os.ReadFile("../testdata/rbac_v2.json")
	require.NoError(t, err)

	engine, err := newEngine(ctx, zaptest.NewLogger(t),
		withPolicySource(policySource(string(policy))),
		withDataSource(dataSource(string(data)), 5*time.Second))
	require.NoError(t, err)
	require.NotNil(t, engine)
}

func TestEngine_IsAllowed(t *testing.T) {
	policy, err := os.ReadFile("../testdata/rbac_v2.rego")
	require.NoError(t, err)

	data, err := os.ReadFile("../testdata/rbac_v2.json")
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	engine, err := newEngine(ctx, zaptest.NewLogger(t),
		withPolicySource(policySource(string(policy))),
		withDataSource(dataSource(string(data)), 5*time.Second))
	require.NoError(t, err)

	tests := []struct {
		name     string
		input    map[string]interface{}
		expected bool
	}{
		{
			name: "admin can create namespace in default environment",
			input: map[string]interface{}{
				"authentication": map[string]interface{}{
					"metadata": map[string]interface{}{
						"io.flipt.auth.user": "admin@company.com",
					},
				},
				"request": flipt.Request{
					Scope:       flipt.ScopeNamespace,
					Environment: ptr("default"),
					Resource:    flipt.ResourceNamespace,
					Action:      flipt.ActionCreate,
				},
			},
			expected: true,
		},
		{
			name: "namespace_admin can create namespace in development environment",
			input: map[string]interface{}{
				"authentication": map[string]interface{}{
					"metadata": map[string]interface{}{
						"io.flipt.auth.groups": []string{"platform-team"},
					},
				},
				"request": flipt.Request{
					Scope:       flipt.ScopeNamespace,
					Environment: ptr("development"),
					Resource:    flipt.ResourceNamespace,
					Action:      flipt.ActionCreate,
				},
			},
			expected: true,
		},
		{
			name: "developer can create resource in frontend namespace",
			input: map[string]interface{}{
				"authentication": map[string]interface{}{
					"metadata": map[string]interface{}{
						"io.flipt.auth.groups": []string{"dev-team"},
					},
				},
				"request": flipt.Request{
					Scope:       flipt.ScopeResource,
					Environment: ptr("development"),
					Namespace:   ptr("frontend"),
					Resource:    flipt.ResourceAny,
					Action:      flipt.ActionCreate,
				},
			},
			expected: true,
		},
		{
			name: "readonly can only read in analytics namespace",
			input: map[string]interface{}{
				"authentication": map[string]interface{}{
					"metadata": map[string]interface{}{
						"io.flipt.auth.user": "analyst@company.com",
					},
				},
				"request": flipt.Request{
					Scope:       flipt.ScopeResource,
					Environment: ptr("production"),
					Namespace:   ptr("analytics"),
					Resource:    flipt.ResourceAny,
					Action:      flipt.ActionRead,
				},
			},
			expected: true,
		},
		{
			name: "readonly cannot create in analytics namespace",
			input: map[string]interface{}{
				"authentication": map[string]interface{}{
					"metadata": map[string]interface{}{
						"io.flipt.auth.user": "analyst@company.com",
					},
				},
				"request": flipt.Request{
					Scope:       flipt.ScopeResource,
					Environment: ptr("production"),
					Namespace:   ptr("analytics"),
					Resource:    flipt.ResourceAny,
					Action:      flipt.ActionCreate,
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			allowed, err := engine.IsAllowed(ctx, tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, allowed)
		})
	}
}

func TestEngine_ViewableEnvironments(t *testing.T) {
	policy, err := os.ReadFile("../testdata/rbac_v2.rego")
	require.NoError(t, err)

	data, err := os.ReadFile("../testdata/rbac_v2.json")
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	engine, err := newEngine(ctx, zaptest.NewLogger(t),
		withPolicySource(policySource(string(policy))),
		withDataSource(dataSource(string(data)), 5*time.Second))
	require.NoError(t, err)

	tests := []struct {
		name        string
		input       map[string]interface{}
		expected    []string
		shouldError bool
	}{
		{
			name: "admin can see all environments",
			input: map[string]interface{}{
				"authentication": map[string]interface{}{
					"metadata": map[string]interface{}{
						"io.flipt.auth.user": "admin@company.com",
					},
				},
			},
			expected: []string{"*"},
		},
		{
			name: "namespace_admin can see development and staging",
			input: map[string]interface{}{
				"authentication": map[string]interface{}{
					"metadata": map[string]interface{}{
						"io.flipt.auth.groups": []string{"platform-team"},
					},
				},
			},
			expected: []string{"development", "staging"},
		},
		{
			name: "developer can see development and staging",
			input: map[string]interface{}{
				"authentication": map[string]interface{}{
					"metadata": map[string]interface{}{
						"io.flipt.auth.groups": []string{"dev-team"},
					},
				},
			},
			expected: []string{"development", "staging"},
		},
		{
			name: "readonly can see production",
			input: map[string]interface{}{
				"authentication": map[string]interface{}{
					"metadata": map[string]interface{}{
						"io.flipt.auth.user": "analyst@company.com",
					},
				},
			},
			expected: []string{"production"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			environments, err := engine.ViewableEnvironments(ctx, tt.input)
			if tt.shouldError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.ElementsMatch(t, tt.expected, environments)
		})
	}
}

func TestEngine_ViewableNamespaces(t *testing.T) {
	policy, err := os.ReadFile("../testdata/rbac_v2.rego")
	require.NoError(t, err)

	data, err := os.ReadFile("../testdata/rbac_v2.json")
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	engine, err := newEngine(ctx, zaptest.NewLogger(t),
		withPolicySource(policySource(string(policy))),
		withDataSource(dataSource(string(data)), 5*time.Second))
	require.NoError(t, err)

	tests := []struct {
		name        string
		env         string
		input       map[string]interface{}
		expected    []string
		shouldError bool
	}{
		{
			name: "admin can see all namespaces in production",
			env:  "production",
			input: map[string]interface{}{
				"authentication": map[string]interface{}{
					"metadata": map[string]interface{}{
						"io.flipt.auth.user": "admin@company.com",
					},
				},
			},
			expected: []string{"*"},
		},
		{
			name: "namespace_admin can see all namespaces in development",
			env:  "development",
			input: map[string]interface{}{
				"authentication": map[string]interface{}{
					"metadata": map[string]interface{}{
						"io.flipt.auth.groups": []string{"platform-team"},
					},
				},
			},
			expected: []string{"*"},
		},
		{
			name: "developer can see frontend and backend in development",
			env:  "development",
			input: map[string]interface{}{
				"authentication": map[string]interface{}{
					"metadata": map[string]interface{}{
						"io.flipt.auth.groups": []string{"dev-team"},
					},
				},
			},
			expected: []string{"frontend", "backend"},
		},
		{
			name: "readonly can see analytics and reporting in production",
			env:  "production",
			input: map[string]interface{}{
				"authentication": map[string]interface{}{
					"metadata": map[string]interface{}{
						"io.flipt.auth.user": "analyst@company.com",
					},
				},
			},
			expected: []string{"analytics", "reporting"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			namespaces, err := engine.ViewableNamespaces(ctx, tt.env, tt.input)
			if tt.shouldError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.ElementsMatch(t, tt.expected, namespaces)
		})
	}
}

type policySource string

func (p policySource) Get(context.Context, source.Hash) ([]byte, source.Hash, error) {
	return []byte(p), nil, nil
}

type dataSource string

func (d dataSource) Get(context.Context, source.Hash) (data map[string]any, _ source.Hash, _ error) {
	return data, nil, json.Unmarshal([]byte(d), &data)
}

func ptr[T any](v T) *T {
	return &v
}
