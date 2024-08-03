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
			httpClient = opts.HTTPClient(t)
			protocol   = opts.Protocol()
		)

		if protocol == integration.ProtocolGRPC {
			t.Skip("REST tests are not applicable for gRPC")
		}

		t.Run("Evaluation Data", func(t *testing.T) {
			for _, namespace := range integration.Namespaces {
				t.Run(fmt.Sprintf("namespace %q", namespace.Expected), func(t *testing.T) {
					t.Logf("Get snapshot for namespace.")

					resp, err := httpClient.Get(fmt.Sprintf("%s/internal/v1/evaluation/snapshot/namespace/%s", opts.URL, namespace))

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

					req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/internal/v1/evaluation/snapshot/namespace/%s", opts.URL, namespace), nil)
					req.Header.Set("If-None-Match", etag)
					require.NoError(t, err)

					resp, err = httpClient.Do(req)
					require.NoError(t, err)
					require.NotNil(t, resp)
					assert.Equal(t, http.StatusNotModified, resp.StatusCode)
					if resp.Body != nil {
						resp.Body.Close()
					}
				})
			}
		})
	})
}
