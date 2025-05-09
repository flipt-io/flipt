package snapshot

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http"
	"testing"

	"bytes"
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/build/testing/integration"
	"go.flipt.io/flipt/rpc/v2/evaluation"
	"google.golang.org/protobuf/encoding/protojson"
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
						resp, err := httpClient.Get(url)

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

						resp, err = httpClient.Do(req)
						require.NoError(t, err)
						require.NotNil(t, resp)
						assert.Equal(t, http.StatusNotModified, resp.StatusCode)
						if resp.Body != nil {
							resp.Body.Close()
						}
					})
				}
			}
		})
	})
}

func TestEvaluationSnapshotNamespaceStream(t *testing.T) {
	integration.Harness(t, func(t *testing.T, opts integration.TestOpts) {
		if opts.Protocol() == integration.ProtocolGRPC {
			t.Skip("Streaming tests are not applicable for gRPC")
		}

		t.Run("Streaming Evaluation Data", func(t *testing.T) {
			streamingTransport := &http.Transport{
				DialContext: (&net.Dialer{
					Timeout:   10 * time.Second,
					KeepAlive: 5 * time.Second, // More aggressive TCP keepalive
				}).DialContext,
				IdleConnTimeout:       60 * time.Second,
				TLSHandshakeTimeout:   10 * time.Second,
				ResponseHeaderTimeout: 10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
				ForceAttemptHTTP2:     true,
				MaxIdleConns:          100,
				MaxIdleConnsPerHost:   10,
				// Enable HTTP/2 ping frames and other settings
				Proxy:           http.ProxyFromEnvironment,
				ReadBufferSize:  32 * 1024,
				WriteBufferSize: 32 * 1024,
				// HTTP/2 specific settings
				MaxResponseHeaderBytes: 4 * 1024, // 4KB
				DisableCompression:     false,    // Enable compression
			}

			httpClient := opts.HTTPClient(t, integration.WithTransport(streamingTransport))

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			newNamespaceKey := "test-namespace-stream" + uuid.New().String()
			nsReqBody := map[string]any{
				"environmentKey": "default",
				"key":            newNamespaceKey,
				"name":           "Test Namespace Stream",
			}
			nsBodyBytes, _ := json.Marshal(nsReqBody)
			nsReq, err := http.NewRequest(http.MethodPost, opts.URL.String()+"/api/v2/environments/default/namespaces", bytes.NewReader(nsBodyBytes))
			require.NoError(t, err)
			nsReq.Header.Set("Content-Type", "application/json")
			nsResp, err := httpClient.Do(nsReq)
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, nsResp.StatusCode)
			nsResp.Body.Close()

			var (
				snapshots = make(chan *evaluation.EvaluationNamespaceSnapshot, 1)
				errs      = make(chan error, 1)
			)

			go func() {
				url := opts.URL.String() + "/client/v2/environments/default/namespaces/" + newNamespaceKey + "/stream"
				req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
				require.NoError(t, err)

				resp, err := httpClient.Do(req)
				require.NoError(t, err)
				require.NotNil(t, resp)

				reader := bufio.NewReaderSize(resp.Body, 32*1024)

				type wrappedSnapshot struct {
					Result json.RawMessage `json:"result"`
				}

				defer func() {
					resp.Body.Close()
					close(snapshots)
					close(errs)
				}()

				for {
					line, err := reader.ReadBytes('\n')
					if err != nil {
						if err == io.EOF {
							break
						}
						errs <- err
						return
					}
					t.Logf("Received line: %s", string(line))
					var wrapper wrappedSnapshot
					if err := json.Unmarshal(line, &wrapper); err != nil {
						errs <- err
						return
					}
					var snap evaluation.EvaluationNamespaceSnapshot
					if err := protojson.Unmarshal(wrapper.Result, &snap); err != nil {
						errs <- err
						return
					}
					snapshots <- &snap
				}
			}()

			// Give the stream a moment to subscribe
			time.Sleep(500 * time.Millisecond)

			// Step 2: Mutate the environment via REST API: add a new flag resource to the new namespace
			flagPayload := map[string]any{
				"@type":   "flipt.core.Flag",
				"key":     "test-flag",
				"name":    "Test Flag",
				"enabled": true,
			}
			reqBody := map[string]any{
				"environmentKey": "default",
				"namespaceKey":   newNamespaceKey,
				"key":            "test-flag",
				"payload":        flagPayload,
			}
			bodyBytes, _ := json.Marshal(reqBody)
			req, err := http.NewRequest(http.MethodPost, opts.URL.String()+"/api/v2/environments/default/namespaces/"+newNamespaceKey+"/resources", bytes.NewReader(bodyBytes))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			resp, err := httpClient.Do(req)
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, resp.StatusCode)
			resp.Body.Close()

			count := 0
			// Wait for a snapshot or error
			select {
			case snap, ok := <-snapshots:
				if !ok {
					t.Fatal("Channel closed")
				}
				count++
				t.Logf("Received snapshot (%d): %+v", count, snap)
				require.NotNil(t, snap)

				assert.Equal(t, newNamespaceKey, snap.Namespace.Key)
				assert.NotEmpty(t, snap.Digest)
			case err := <-errs:
				t.Fatalf("Error from stream: %v", err)
			case <-time.After(5 * time.Second):
				t.Fatal("Timed out waiting for snapshot")
			}
		})
	})
}
