package authz

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEngine_NewEngine(t *testing.T) {
	ctx := context.Background()
	engine, err := NewEngine(ctx)
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
				"role": "admin",
				"action": "create",
				"scope": "flag"
			}`,
			expected: true,
		},
		{
			name: "admin is allowed to read",
			input: `{
				"role": "admin",
				"action": "read",
				"scope": "flag"
			}`,
			expected: true,
		},
		{
			name: "editor is allowed to create flags",
			input: `{
				"role": "editor",
				"action": "create",
				"scope": "flag"
			}`,
			expected: true,
		},
		{
			name: "editor is allowed to read",
			input: `{
				"role": "editor",
				"action": "read",
				"scope": "flag"
			}`,
			expected: true,
		},
		{
			name: "editor is not allowed to create namespaces",
			input: `{
				"role": "editor",
				"action": "create",
				"scope": "namespace"
			}`,
			expected: false,
		},
		{
			name: "viewer is allowed to read",
			input: `{
				"role": "viewer",
				"action": "read",
				"scope": "flag"
			}`,
			expected: true,
		},
		{
			name: "viewer is not allowed to create",
			input: `{
				"role": "viewer",
				"action": "create",
				"scope": "flag"
			}`,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			engine, err := NewEngine(ctx)
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
