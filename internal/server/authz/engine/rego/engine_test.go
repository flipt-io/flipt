package rego

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/open-policy-agent/opa/v1/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/server/authz/engine/rego/source"
	"go.flipt.io/flipt/rpc/flipt"
	"go.uber.org/zap/zaptest"
)

func TestEngine_NewEngine(t *testing.T) {
	ctx := t.Context()

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

	ctx, cancel := context.WithCancel(t.Context())
	t.Cleanup(cancel)
	engine, err := newEngine(ctx, zaptest.NewLogger(t),
		withPolicySource(policySource(string(policy))),
		withDataSource(dataSource(string(data)), 5*time.Second))
	require.NoError(t, err)

	tests := []struct {
		name     string
		input    map[string]any
		expected bool
	}{
		{
			name: "admin can create namespace in default environment",
			input: map[string]any{
				"authentication": map[string]any{
					"metadata": map[string]any{
						"io.flipt.auth.user": "admin@company.com",
					},
				},
				"request": flipt.Request{
					Scope:       flipt.ScopeEnvironment,
					Environment: new("default"),
					Action:      flipt.ActionCreate,
				},
			},
			expected: true,
		},
		{
			name: "namespace_admin can create namespace in development environment",
			input: map[string]any{
				"authentication": map[string]any{
					"metadata": map[string]any{
						"io.flipt.auth.groups": []string{"platform-team"},
					},
				},
				"request": flipt.Request{
					Scope:       flipt.ScopeEnvironment,
					Environment: new("development"),
					Action:      flipt.ActionCreate,
				},
			},
			expected: true,
		},
		{
			name: "developer can create resource in frontend namespace",
			input: map[string]any{
				"authentication": map[string]any{
					"metadata": map[string]any{
						"io.flipt.auth.groups": []string{"dev-team"},
					},
				},
				"request": flipt.Request{
					Scope:       flipt.ScopeNamespace,
					Environment: new("development"),
					Namespace:   new("frontend"),
					Action:      flipt.ActionCreate,
				},
			},
			expected: true,
		},
		{
			name: "readonly can only read in analytics namespace",
			input: map[string]any{
				"authentication": map[string]any{
					"metadata": map[string]any{
						"io.flipt.auth.user": "analyst@company.com",
					},
				},
				"request": flipt.Request{
					Scope:       flipt.ScopeNamespace,
					Environment: new("production"),
					Namespace:   new("analytics"),
					Action:      flipt.ActionRead,
				},
			},
			expected: true,
		},
		{
			name: "readonly cannot create in analytics namespace",
			input: map[string]any{
				"authentication": map[string]any{
					"metadata": map[string]any{
						"io.flipt.auth.user": "analyst@company.com",
					},
				},
				"request": flipt.Request{
					Scope:       flipt.ScopeNamespace,
					Environment: new("production"),
					Namespace:   new("analytics"),
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

	ctx, cancel := context.WithCancel(t.Context())
	t.Cleanup(cancel)
	engine, err := newEngine(ctx, zaptest.NewLogger(t),
		withPolicySource(policySource(string(policy))),
		withDataSource(dataSource(string(data)), 5*time.Second))
	require.NoError(t, err)

	tests := []struct {
		name        string
		input       map[string]any
		expected    []string
		shouldError bool
	}{
		{
			name: "admin can see all environments",
			input: map[string]any{
				"authentication": map[string]any{
					"metadata": map[string]any{
						"io.flipt.auth.user": "admin@company.com",
					},
				},
			},
			expected: []string{"*"},
		},
		{
			name: "namespace_admin can see development and staging",
			input: map[string]any{
				"authentication": map[string]any{
					"metadata": map[string]any{
						"io.flipt.auth.groups": []string{"platform-team"},
					},
				},
			},
			expected: []string{"development", "staging"},
		},
		{
			name: "developer can see development and staging",
			input: map[string]any{
				"authentication": map[string]any{
					"metadata": map[string]any{
						"io.flipt.auth.groups": []string{"dev-team"},
					},
				},
			},
			expected: []string{"development", "staging"},
		},
		{
			name: "readonly can see production",
			input: map[string]any{
				"authentication": map[string]any{
					"metadata": map[string]any{
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
			assert.NotNil(t, environments)
		})
	}
}

func TestEngine_ViewableEnvironmentsFromData(t *testing.T) {
	tests := []struct {
		name        string
		data        string
		expected    []string
		shouldError bool
	}{
		{
			name:     "restricted scope",
			data:     `{"flipt":{"authz":{"v2":{"viewable_environments":["production"]}}}}`,
			expected: []string{"production"},
		},
		{
			name:     "empty scope",
			data:     `{"flipt":{"authz":{"v2":{"viewable_environments":[]}}}}`,
			expected: []string{},
		},
		{
			name:     "wildcard scope",
			data:     `{"flipt":{"authz":{"v2":{"viewable_environments":["*"]}}}}`,
			expected: []string{"*"},
		},
		{
			name: "undefined scope",
			data: `{}`,
		},
		{
			name:        "invalid scope type",
			data:        `{"flipt":{"authz":{"v2":{"viewable_environments":"production"}}}}`,
			shouldError: true,
		},
		{
			name:        "null scope",
			data:        `{"flipt":{"authz":{"v2":{"viewable_environments":null}}}}`,
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			engine, err := newEngine(ctx, zaptest.NewLogger(t),
				withPolicySource(policySource(policyWithoutViewableScopes)),
				withDataSource(dataSource(tt.data), time.Hour))
			require.NoError(t, err)

			environments, err := engine.ViewableEnvironments(ctx, nil)
			if tt.shouldError {
				require.ErrorContains(t, err, "unexpected result type")
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expected, environments)
		})
	}
}

func TestEngine_DataReloadUpdatesViewableEnvironments(t *testing.T) {
	ctx := t.Context()
	engine, err := newEngine(ctx, zaptest.NewLogger(t),
		withPolicySource(policySource(policyWithoutViewableScopes)),
		withDataSource(dataSource(`{}`), time.Hour))
	require.NoError(t, err)

	environments, err := engine.ViewableEnvironments(ctx, nil)
	require.NoError(t, err)
	assert.Nil(t, environments)

	for _, tt := range []struct {
		name     string
		data     string
		expected []string
	}{
		{
			name:     "add scope",
			data:     `{"flipt":{"authz":{"v2":{"viewable_environments":["production"]}}}}`,
			expected: []string{"production"},
		},
		{
			name:     "replace scope",
			data:     `{"flipt":{"authz":{"v2":{"viewable_environments":["staging"]}}}}`,
			expected: []string{"staging"},
		},
		{
			name:     "empty scope",
			data:     `{"flipt":{"authz":{"v2":{"viewable_environments":[]}}}}`,
			expected: []string{},
		},
		{
			name: "remove scope",
			data: `{}`,
		},
		{
			name:     "restore scope",
			data:     `{"flipt":{"authz":{"v2":{"viewable_environments":["production"]}}}}`,
			expected: []string{"production"},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			engine.mu.Lock()
			engine.dataSource = dataSource(tt.data)
			engine.mu.Unlock()
			require.NoError(t, engine.updateData(ctx, storage.ReplaceOp))

			environments, err := engine.ViewableEnvironments(ctx, nil)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, environments)
		})
	}
}

func TestEngine_PolicyReloadWithDataScope(t *testing.T) {
	for _, tt := range []struct {
		name   string
		policy string
	}{
		{name: "parse failure", policy: policyInvalid},
		{name: "compile failure", policy: policyUncompilable},
	} {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			policySource := &reloadablePolicySource{policy: policyWithoutViewableScopes}
			engine, err := newEngine(ctx, zaptest.NewLogger(t),
				withPolicySource(policySource),
				withDataSource(dataSource(`{"flipt":{"authz":{"v2":{"viewable_environments":["production"]}}}}`), time.Hour))
			require.NoError(t, err)

			environments, err := engine.ViewableEnvironments(ctx, nil)
			require.NoError(t, err)
			assert.Equal(t, []string{"production"}, environments)

			policySource.Set(tt.policy)
			require.Error(t, engine.updatePolicy(ctx))

			environments, err = engine.ViewableEnvironments(ctx, nil)
			require.NoError(t, err)
			assert.Equal(t, []string{}, environments)

			policySource.Set(policyWithoutViewableScopes)
			require.NoError(t, engine.updatePolicy(ctx))

			environments, err = engine.ViewableEnvironments(ctx, nil)
			require.NoError(t, err)
			assert.Equal(t, []string{"production"}, environments)
		})
	}
}

func TestEngine_PolicyReloadReplacesOptionalQueries(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	t.Cleanup(cancel)

	policySource := &reloadablePolicySource{policy: policyWithViewableScopes}
	engine, err := newEngine(ctx, zaptest.NewLogger(t), withPolicySource(policySource))
	require.NoError(t, err)

	input := map[string]any{"authentication": map[string]any{}}
	environments, err := engine.ViewableEnvironments(ctx, input)
	require.NoError(t, err)
	assert.Equal(t, []string{"production"}, environments)

	namespaces, err := engine.ViewableNamespaces(ctx, "production", input)
	require.NoError(t, err)
	assert.Equal(t, []string{"analytics"}, namespaces)

	policySource.Set(policyInvalid)
	require.Error(t, engine.updatePolicy(ctx))

	environments, err = engine.ViewableEnvironments(ctx, input)
	require.NoError(t, err)
	assert.Empty(t, environments)

	namespaces, err = engine.ViewableNamespaces(ctx, "production", input)
	require.NoError(t, err)
	assert.Empty(t, namespaces)

	policySource.Set(policyWithViewableScopes)
	require.NoError(t, engine.updatePolicy(ctx))

	environments, err = engine.ViewableEnvironments(ctx, input)
	require.NoError(t, err)
	assert.Equal(t, []string{"production"}, environments)

	namespaces, err = engine.ViewableNamespaces(ctx, "production", input)
	require.NoError(t, err)
	assert.Equal(t, []string{"analytics"}, namespaces)

	// A policy that parses but fails to compile must also fail closed rather
	// than keep serving the previous policy's queries.
	policySource.Set(policyUncompilable)
	require.ErrorContains(t, engine.updatePolicy(ctx), "preparing policy allow")

	allowed, err := engine.IsAllowed(ctx, input)
	require.NoError(t, err)
	assert.False(t, allowed)

	environments, err = engine.ViewableEnvironments(ctx, input)
	require.NoError(t, err)
	assert.Empty(t, environments)
	assert.NotNil(t, environments)

	namespaces, err = engine.ViewableNamespaces(ctx, "production", input)
	require.NoError(t, err)
	assert.Empty(t, namespaces)
	assert.NotNil(t, namespaces)

	policySource.Set(policyWithEmptyViewableScopes)
	require.NoError(t, engine.updatePolicy(ctx))

	environments, err = engine.ViewableEnvironments(ctx, input)
	require.NoError(t, err)
	assert.Empty(t, environments)
	assert.NotNil(t, environments)

	// A defined namespace rule with no result for the environment is an
	// empty scope, not an undefined one.
	namespaces, err = engine.ViewableNamespaces(ctx, "production", input)
	require.NoError(t, err)
	assert.Empty(t, namespaces)
	assert.NotNil(t, namespaces)

	policySource.Set(policyWithoutViewableScopes)
	require.NoError(t, engine.updatePolicy(ctx))

	environments, err = engine.ViewableEnvironments(ctx, input)
	require.NoError(t, err)
	assert.Nil(t, environments)

	namespaces, err = engine.ViewableNamespaces(ctx, "production", input)
	require.NoError(t, err)
	assert.Nil(t, namespaces)
}

func TestEngine_ViewableNamespaces(t *testing.T) {
	policy, err := os.ReadFile("../testdata/rbac_v2.rego")
	require.NoError(t, err)

	data, err := os.ReadFile("../testdata/rbac_v2.json")
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(t.Context())
	t.Cleanup(cancel)
	engine, err := newEngine(ctx, zaptest.NewLogger(t),
		withPolicySource(policySource(string(policy))),
		withDataSource(dataSource(string(data)), 5*time.Second))
	require.NoError(t, err)

	tests := []struct {
		name        string
		env         string
		input       map[string]any
		expected    []string
		shouldError bool
	}{
		{
			name: "admin can see all namespaces in production",
			env:  "production",
			input: map[string]any{
				"authentication": map[string]any{
					"metadata": map[string]any{
						"io.flipt.auth.user": "admin@company.com",
					},
				},
			},
			expected: []string{"*"},
		},
		{
			name: "namespace_admin can see all namespaces in development",
			env:  "development",
			input: map[string]any{
				"authentication": map[string]any{
					"metadata": map[string]any{
						"io.flipt.auth.groups": []string{"platform-team"},
					},
				},
			},
			expected: []string{"*"},
		},
		{
			name: "developer can see frontend and backend in development",
			env:  "development",
			input: map[string]any{
				"authentication": map[string]any{
					"metadata": map[string]any{
						"io.flipt.auth.groups": []string{"dev-team"},
					},
				},
			},
			expected: []string{"frontend", "backend"},
		},
		{
			name: "readonly can see analytics and reporting in production",
			env:  "production",
			input: map[string]any{
				"authentication": map[string]any{
					"metadata": map[string]any{
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

const (
	policyWithViewableScopes = `package flipt.authz.v2

import rego.v1

default allow := true

viewable_environments := ["production"]
viewable_namespaces(_) := ["analytics"]
`
	policyWithoutViewableScopes = `package flipt.authz.v2

import rego.v1

default allow := true
`
	policyWithEmptyViewableScopes = `package flipt.authz.v2

import rego.v1

default allow := true

viewable_environments := []
viewable_namespaces("staging") := ["analytics"]
`
	policyUncompilable = `package flipt.authz.v2

import rego.v1

viewable_environments := ["production"]
viewable_namespaces(_) := ["analytics"]

allow if undefined_function(input.request)
`
	policyInvalid = `package flipt.authz.v2

import rego.v1

default allow := true

allow :=
`
)

type policySource string

func (p policySource) Get(context.Context, source.Hash) ([]byte, source.Hash, error) {
	return []byte(p), nil, nil
}

type reloadablePolicySource struct {
	mu     sync.RWMutex
	policy string
}

func (p *reloadablePolicySource) Get(context.Context, source.Hash) ([]byte, source.Hash, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	hash := sha256.Sum256([]byte(p.policy))
	return []byte(p.policy), hash[:], nil
}

func (p *reloadablePolicySource) Set(policy string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.policy = policy
}

type dataSource string

func (d dataSource) Get(context.Context, source.Hash) (data map[string]any, _ source.Hash, _ error) {
	return data, nil, json.Unmarshal([]byte(d), &data)
}
