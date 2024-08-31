package bundle

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/open-policy-agent/contrib/logging/plugins/ozap"
	"github.com/open-policy-agent/opa/sdk"
	sdktest "github.com/open-policy-agent/opa/sdk/test"
	"github.com/open-policy-agent/opa/storage/inmem"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

func TestEngine_IsAllowed(t *testing.T) {
	ctx := context.Background()

	policy, err := os.ReadFile("../testdata/rbac.rego")
	require.NoError(t, err)

	data, err := os.ReadFile("../testdata/rbac.json")
	require.NoError(t, err)

	var (
		server = sdktest.MustNewServer(
			sdktest.MockBundle("/bundles/bundle.tar.gz", map[string]string{
				"main.rego": string(policy),
				"data.json": string(data),
			}),
		)
		config = fmt.Sprintf(`{
		"services": {
			"test": {
				"url": %q
			}
		},
		"bundles": {
			"test": {
				"resource": "/bundles/bundle.tar.gz"
			}
		},
	}`, server.URL())
	)

	t.Cleanup(server.Stop)

	opa, err := sdk.New(ctx, sdk.Options{
		Config: strings.NewReader(config),
		Store:  inmem.New(),
		Logger: ozap.Wrap(zaptest.NewLogger(t), &zap.AtomicLevel{}),
	})

	require.NoError(t, err)
	assert.NotNil(t, opa)

	engine := &Engine{
		opa:    opa,
		logger: zaptest.NewLogger(t),
	}

	var tests = []struct {
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

	assert.NoError(t, engine.Shutdown(ctx))
}
