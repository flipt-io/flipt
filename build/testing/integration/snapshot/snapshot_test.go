package snapshot

import (
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/build/testing/integration"
)

// Snapshot tests the client evaluation snapshot API.
func TestSnapshot(t *testing.T) {
	integration.Harness(t, func(t *testing.T, opts integration.TestOpts) {
		var (
			protocol = opts.Protocol()
		)

		if protocol == integration.ProtocolGRPC {
			t.Skip("REST tests are not applicable for gRPC")
		}

		t.Run("Evaluation Data", func(t *testing.T) {
			httpClient := opts.HTTPClient(t)
			// for testing backwards compatibility with v1
			for _, path := range []string{"/client/v2/environments/default/namespaces/%s/snapshot", "/internal/v1/evaluation/snapshot/namespace/%s"} {
				for _, namespace := range integration.Namespaces {
					url := opts.URL.String() + fmt.Sprintf(path, namespace.Expected)
					t.Run(fmt.Sprintf("namespace %q", namespace.Expected), func(t *testing.T) {
						testSnapshotForNamespace(t, httpClient, url, namespace.Expected)
					})
				}
			}
		})
	})
}

func testSnapshotForNamespace(t *testing.T, client *http.Client, url, namespace string) {
	t.Logf("Get snapshot %q.", url)

	resp, err := client.Get(url)

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

	t.Logf("Get snapshot %q with etag/if-none-match.", url)

	req, err := http.NewRequest(http.MethodGet, url, nil)
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
