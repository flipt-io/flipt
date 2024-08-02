package snapshot

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/build/testing/integration"
	"go.flipt.io/flipt/rpc/flipt"
	"golang.org/x/exp/rand"
)

// Snapshot tests the evaluation data snapshot API without using the SDK (other than for creating)
func Snapshot(t *testing.T, ctx context.Context, opts integration.TestOpts) {
	var (
		client     = opts.TokenClient(t)
		httpClient = opts.HTTPClient(t)
		protocol   = opts.Protocol()
	)

	if protocol == integration.ProtocolGRPC {
		t.Skip("REST tests are not applicable for gRPC")
	}

	t.Run("Evaluation Data", func(t *testing.T) {
		t.Log(`Create namespace.`)

		_, err := client.Flipt().CreateNamespace(ctx, &flipt.CreateNamespaceRequest{
			Key:  integration.ProductionNamespace,
			Name: "Production",
		})
		require.NoError(t, err)

		for _, namespace := range integration.Namespaces {
			flag := fmt.Sprintf("%x", rand.Int63())
			t.Run(fmt.Sprintf("namespace %q", namespace.Expected), func(t *testing.T) {
				// create some flags
				_, err = client.Flipt().CreateFlag(ctx, &flipt.CreateFlagRequest{
					NamespaceKey: namespace.Key,
					Key:          fmt.Sprintf("test_%s", flag),
					Name:         "Test",
					Description:  "This is a test flag",
					Enabled:      true,
				})
				require.NoError(t, err)

				t.Log("Create a new flag in a disabled state.")

				_, err = client.Flipt().CreateFlag(ctx, &flipt.CreateFlagRequest{
					NamespaceKey: namespace.Key,
					Key:          fmt.Sprintf("disabled_%s", flag),
					Name:         "Disabled",
					Description:  "This is a disabled test flag",
					Enabled:      false,
				})
				require.NoError(t, err)

				t.Log("Create a new enabled boolean flag with key \"boolean_enabled\".")

				_, err = client.Flipt().CreateFlag(ctx, &flipt.CreateFlagRequest{
					Type:         flipt.FlagType_BOOLEAN_FLAG_TYPE,
					NamespaceKey: namespace.Key,
					Key:          fmt.Sprintf("boolean_enabled_%s", flag),
					Name:         "Boolean Enabled",
					Description:  "This is an enabled boolean test flag",
					Enabled:      true,
				})
				require.NoError(t, err)

				t.Log("Create a new flag in a disabled state.")

				_, err = client.Flipt().CreateFlag(ctx, &flipt.CreateFlagRequest{
					Type:         flipt.FlagType_BOOLEAN_FLAG_TYPE,
					NamespaceKey: namespace.Key,
					Key:          fmt.Sprintf("boolean_disabled_%s", flag),
					Name:         "Boolean Disabled",
					Description:  "This is a disabled boolean test flag",
					Enabled:      false,
				})
				require.NoError(t, err)

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
}
