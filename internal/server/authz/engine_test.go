package authz

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestEngine_NewEngine(t *testing.T) {
	ctx := context.Background()
	engine, err := NewEngine(ctx, zaptest.NewLogger(t), testRBACPolicy, WithDataSource(testRoleDefinitions, 5*time.Second))
	require.NoError(t, err)
	require.NotNil(t, engine)
}

func TestEngine_IsAllowed(t *testing.T) {
	var tests = []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name: "admin is allowed to create",
			input: `{
                "authentication": {
                    "method": "METHOD_JWT",
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
                    "method": "METHOD_JWT",
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
                    "method": "METHOD_JWT",
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
                    "method": "METHOD_JWT",
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
                    "method": "METHOD_JWT",
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
                    "method": "METHOD_JWT",
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
                    "method": "METHOD_JWT",
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
                    "method": "METHOD_JWT",
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
                    "method": "METHOD_JWT",
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
                    "method": "METHOD_JWT",
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
			ctx := context.Background()
			engine, err := NewEngine(ctx, zaptest.NewLogger(t), testRBACPolicy, WithDataSource(testRoleDefinitions, 5*time.Second))
			require.NoError(t, err)

			var input map[string]interface{}

			err = json.Unmarshal([]byte(tt.input), &input)
			require.NoError(t, err)

			allowed, err := engine.IsAllowed(ctx, input)
			require.NoError(t, err)
			require.Equal(t, tt.expected, allowed)
		})
	}
}

type policySource string

func (p policySource) Get(context.Context, []byte) ([]byte, []byte, error) {
	return []byte(p), nil, nil
}

type dataSource string

func (d dataSource) Get(context.Context, []byte) (data map[string]any, _ []byte, _ error) {
	return data, nil, json.Unmarshal([]byte(d), &data)
}

var (
	testRBACPolicy = policySource(`package flipt.authz.v1

import data
import rego.v1

default allow = false

allow if {
	some rule in has_rules

	permit_string(rule.resource, input.request.resource)
	permit_slice(rule.actions, input.request.action)
	permit_string(rule.namespace, input.request.namespace)
}

allow if {
	some rule in has_rules

	permit_string(rule.resource, input.request.resource)
	permit_slice(rule.actions, input.request.action)
	not rule.namespace
}

has_rules contains rules if {
	some role in data.roles
	role.name == input.authentication.metadata["io.flipt.auth.role"]
	rules := role.rules[_]
}

permit_string(allowed, _) if {
	allowed == "*"
}

permit_string(allowed, requested) if {
	allowed == requested
}

permit_slice(allowed, _) if {
	allowed[_] = "*"
}

permit_slice(allowed, requested) if {
	allowed[_] = requested
}`)
	testRoleDefinitions = dataSource(`{
    "version": "0.1.0",
    "roles": [
        {
            "name": "admin",
            "rules": [
                {
                    "resource": "*",
                    "actions": [
                        "*"
                    ]
                }
            ]
        },
        {
            "name": "editor",
            "rules": [
                {
                    "resource": "namespace",
                    "actions": [
                        "read"
                    ]
                },
                {
                    "resource": "authentication",
                    "actions": [
                        "read"
                    ]
                },
                {
                    "resource": "flag",
                    "actions": [
                        "create",
                        "read",
                        "update",
                        "delete"
                    ]
                },
                {
                    "resource": "segment",
                    "actions": [
                        "create",
                        "read",
                        "update",
                        "delete"
                    ]
                }
            ]
        },
        {
            "name": "viewer",
            "rules": [
                {
                    "resource": "*",
                    "actions": [
                        "read"
                    ]
                }
            ]
        },
        {
            "name": "namespaced_viewer",
            "rules": [
                {
                    "resource": "*",
                    "actions": [
                        "read"
                    ],
                    "namespace": "foo"
                }
            ]
        }
    ]
}`)
)
