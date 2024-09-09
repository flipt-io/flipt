package snapshot_test

import (
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/build/testing/integration"
)

// Snapshot tests the evaluation data snapshot API.
func TestSnapshot(t *testing.T) {
	integration.Harness(t, func(t *testing.T, opts integration.TestOpts) {
		var (
			protocol = opts.Protocol()
		)

		if protocol == integration.ProtocolGRPC {
			t.Skip("REST tests are not applicable for gRPC")
		}

		t.Run("Evaluation Data", func(t *testing.T) {
			for _, role := range []string{"admin", "editor", "viewer"} {
				t.Run(fmt.Sprintf("role %q", role), func(t *testing.T) {
					httpClient := opts.HTTPClient(t, integration.WithRole(role))
					for _, namespace := range integration.Namespaces {
						t.Run(fmt.Sprintf("namespace %q", namespace.Expected), func(t *testing.T) {
							testSnapshotForNamespace(t, httpClient, opts.URL.String(), namespace.Expected)
						})
					}
				})
			}

			t.Run("With Namespace Scoped Token", func(t *testing.T) {
				for _, namespace := range integration.Namespaces {
					t.Run(fmt.Sprintf("namespace %q", namespace.Expected), func(t *testing.T) {
						clientOpts := []integration.ClientOpt{integration.WithNamespace(namespace.Expected), integration.WithRole("admin")}
						httpClient := opts.HTTPClient(t, clientOpts...)
						testSnapshotForNamespace(t, httpClient, opts.URL.String(), namespace.Expected)
					})
				}
			})
		})
	})
}

func testSnapshotForNamespace(t *testing.T, client *http.Client, url, namespace string) {
	t.Logf("Get snapshot for namespace.")

	resp, err := client.Get(fmt.Sprintf("%s/internal/v1/evaluation/snapshot/namespace/%s", url, namespace))

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// get etag from response
	etag := resp.Header.Get("ETag")
	assert.NotEmpty(t, etag)

	// read body
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	resp.Body.Close()

	assert.NotEmpty(t, body)

	t.Logf("Get snapshot for namespace with etag/if-none-match.")

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/internal/v1/evaluation/snapshot/namespace/%s", url, namespace), nil)
	req.Header.Set("If-None-Match", etag)
	require.NoError(t, err)

	resp, err = client.Do(req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, http.StatusNotModified, resp.StatusCode)
	if resp.Body != nil {
		resp.Body.Close()
	}
}
