package rego

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/server/authz/engine/rego/source"
	authrpc "go.flipt.io/flipt/rpc/flipt/auth"
	"go.uber.org/zap/zaptest"
)

func TestEngine_NewEngine(t *testing.T) {
	ctx := context.Background()

	policy, err := os.ReadFile("../testdata/rbac.rego")
	require.NoError(t, err)

	data, err := os.ReadFile("../testdata/rbac.json")
	require.NoError(t, err)

	engine, err := newEngine(ctx, zaptest.NewLogger(t), withPolicySource(policySource(string(policy))), withDataSource(dataSource(string(data)), 5*time.Second))
	require.NoError(t, err)
	require.NotNil(t, engine)
}

func TestEngine_IsAllowed(t *testing.T) {
	policy, err := os.ReadFile("../testdata/rbac.rego")
	require.NoError(t, err)

	data, err := os.ReadFile("../testdata/rbac.json")
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	engine, err := newEngine(ctx, zaptest.NewLogger(t), withPolicySource(policySource(string(policy))), withDataSource(dataSource(string(data)), 5*time.Second))
	require.NoError(t, err)

	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name: "admin is allowed to create",
			input: `{
                "authentication": {
                    "method": 5,
                    "metadata": {
                        "io.flipt.auth.role": "admin"
                    }
                },
                "request": {
                    "action": "create",
                    "resource": "flag"
                }
            }`,
			expected: true,
		},
		{
			name: "admin is allowed to read",
			input: `{
                "authentication": {
                    "method": 5,
                    "metadata": {
                        "io.flipt.auth.role": "admin"
                    }
                },
                "request": {
                    "action": "read",
                    "resource": "flag"
                }
            }`,
			expected: true,
		},
		{
			name: "editor is allowed to create flags",
			input: `{
                "authentication": {
                    "method": 5,
                    "metadata": {
                        "io.flipt.auth.role": "editor"
                    }
                },
                "request": {
                    "action": "create",
                    "resource": "flag"
                }
            }`,
			expected: true,
		},
		{
			name: "editor is allowed to read",
			input: `{
                "authentication": {
                    "method": 5,
                    "metadata": {
                        "io.flipt.auth.role": "editor"
                    }
                },
                "request": {
                    "action": "read",
                    "resource": "flag"
                }
            }`,
			expected: true,
		},
		{
			name: "editor is not allowed to create namespaces",
			input: `{
                "authentication": {
                    "method": 5,
                    "metadata": {
                        "io.flipt.auth.role": "editor"
                    }
                },
                "request": {
                    "action": "create",
                    "resource": "namespace"
                }
            }`,
			expected: false,
		},
		{
			name: "viewer is allowed to read",
			input: `{
                "authentication": {
                    "method": 5,
                    "metadata": {
                        "io.flipt.auth.role": "viewer"
                    }
                },
                "request": {
                    "action": "read",
                    "resource": "segment"
                }
            }`,
			expected: true,
		},
		{
			name: "viewer is not allowed to create",
			input: `{
                "authentication": {
                    "method": 5,
                    "metadata": {
                        "io.flipt.auth.role": "viewer"
                    }
                },
                "request": {
                    "action": "create",
                    "resource": "flag"
                }
            }`,
			expected: false,
		},
		{
			name: "namespaced_viewer is allowed to read in namespace",
			input: `{
                "authentication": {
                    "method": 5,
                    "metadata": {
                        "io.flipt.auth.role": "namespaced_viewer"
                    }
                },
                "request": {
                    "action": "read",
                    "resource": "flag",
                    "namespace": "foo"
                }
            }`,
			expected: true,
		},
		{
			name: "namespaced_viewer is not allowed to read in unexpected namespace",
			input: `{
                "authentication": {
                    "method": 5,
                    "metadata": {
                        "io.flipt.auth.role": "namespaced_viewer"
                    }
                },
                "request": {
                    "action": "read",
                    "resource": "flag",
                    "namespace": "bar"
                }
            }`,
			expected: false,
		},
		{
			name: "namespaced_viewer is not allowed to read in without namespace scope",
			input: `{
                "authentication": {
                    "method": 5,
                    "metadata": {
                        "io.flipt.auth.role": "namespaced_viewer"
                    }
                },
                "request": {
                    "action": "read",
                    "resource": "flag"
                }
            }`,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var input map[string]interface{}

			err = json.Unmarshal([]byte(tt.input), &input)
			require.NoError(t, err)

			allowed, err := engine.IsAllowed(ctx, input)
			require.NoError(t, err)
			require.Equal(t, tt.expected, allowed)
		})
	}

	t.Run("viewable namespaces without definition", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		t.Cleanup(cancel)
		namespaces, err := engine.Namespaces(ctx, map[string]any{})
		require.Error(t, err)
		require.Nil(t, namespaces)
	})
}

func TestEngine_IsAuthMethod(t *testing.T) {
	tests := []struct {
		name     string
		input    authrpc.Method
		expected bool
	}{
		{name: "token", input: authrpc.Method_METHOD_TOKEN, expected: true},
		{name: "oidc", input: authrpc.Method_METHOD_OIDC, expected: true},
		{name: "k8s", input: authrpc.Method_METHOD_KUBERNETES, expected: true},
		{name: "kubernetes", input: authrpc.Method_METHOD_KUBERNETES, expected: true},
		{name: "github", input: authrpc.Method_METHOD_GITHUB, expected: true},
		{name: "jwt", input: authrpc.Method_METHOD_JWT, expected: true},
		{name: "none", input: authrpc.Method_METHOD_OIDC, expected: false},
	}
	data, err := os.ReadFile("../testdata/rbac.json")
	require.NoError(t, err)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			t.Cleanup(cancel)

			input := map[string]any{
				"authentication": authrpc.Authentication{Method: tt.input},
			}

			policy := fmt.Sprintf(`package flipt.authz.v1

            import rego.v1

            default allow := false

            allow if {
               flipt.is_auth_method(input, "%s")
            }
            `, tt.name)

			engine, err := newEngine(ctx, zaptest.NewLogger(t), withPolicySource(policySource(policy)), withDataSource(dataSource(string(data)), 5*time.Second))
			require.NoError(t, err)

			allowed, err := engine.IsAllowed(ctx, input)
			require.NoError(t, err)
			require.Equal(t, tt.expected, allowed)
		})
	}
}

func TestViewableNamespaces(t *testing.T) {
	policy, err := os.ReadFile("../testdata/viewable_namespaces.rego")
	require.NoError(t, err)

	data, err := os.ReadFile("../testdata/viewable_namespaces.json")
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	engine, err := newEngine(ctx, zaptest.NewLogger(t), withPolicySource(policySource(string(policy))), withDataSource(dataSource(string(data)), 5*time.Second))
	require.NoError(t, err)

	t.Cleanup(func() {
		assert.NoError(t, engine.Shutdown(ctx))
	})

	tt := []struct {
		name       string
		roles      []string
		namespaces []string
	}{
		{"empty", []string{}, []string{}},
		{"devs", []string{"devs"}, []string{"local", "staging"}},
		{"devsops", []string{"devs", "ops"}, []string{"local", "production", "staging"}},
	}
	for _, tt := range tt {
		t.Run(tt.name, func(t *testing.T) {
			namespaces, err := engine.Namespaces(ctx, map[string]any{"roles": tt.roles})
			require.NoError(t, err)
			require.Equal(t, tt.namespaces, namespaces)
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
