package kubernetes_test

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/hashicorp/cap/oidc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/config"
	k8stesting "go.flipt.io/flipt/internal/server/authn/method/kubernetes/testing"
	"go.flipt.io/flipt/rpc/flipt/auth"
	"go.uber.org/zap/zaptest"
	"google.golang.org/protobuf/encoding/protojson"
)

func Test_Server(t *testing.T) {
	var (
		router = chi.NewRouter()
		// httpServer is the test server used for hosting
		// Flipts oidc authorize and callback handles
		httpServer = httptest.NewServer(router)
		// rewriting http server to use localhost as it is a domain and
		// the <=go1.18 implementation will propagate cookies on it.
		// From go1.19+ cookiejar support IP addresses as cookie domains.
		clientAddress = strings.Replace(httpServer.URL, "127.0.0.1", "localhost", 1)

		logger = zaptest.NewLogger(t)
		ctx    = context.Background()
	)

	priv, err := rsa.GenerateKey(rand.Reader, 4096)
	require.NoError(t, err)

	tp := oidc.StartTestProvider(
		t,
		oidc.WithTestDefaults(&oidc.TestProviderDefaults{
			SigningKey: &oidc.TestSigningKey{
				PrivKey: priv,
				PubKey:  priv.Public(),
				Alg:     oidc.RS256,
			},
		}),
	)
	defer tp.Stop()

	// write CA certification to temporary file
	caPath := writeStringToTemp(t, "ca-*.cert", tp.CACert())

	t.Log("CA Path", caPath)

	fliptServiceAccountToken := oidc.TestSignJWT(t, priv, string(oidc.RS256),
		map[string]any{
			"exp": time.Now().Add(24 * time.Hour).Unix(),
			"iss": tp.Addr(),
			"kubernetes.io": map[string]any{
				"namespace": "flipt",
				"pod": map[string]any{
					"name": "flipt-7d26f049-kdurb",
					"uid":  "bd8299f9-c50f-4b76-af33-9d8e3ef2b850",
				},
				"serviceaccount": map[string]any{
					"name": "flipt",
					"uid":  "4f18914e-f276-44b2-aebd-27db1d8f8def",
				},
			},
		}, nil)
	fliptServiceAccountTokenPath := writeStringToTemp(t, "token-*", fliptServiceAccountToken)

	t.Log("SA Path", fliptServiceAccountTokenPath)

	var (
		authConfig = config.AuthenticationConfig{
			Methods: config.AuthenticationMethods{
				Kubernetes: config.AuthenticationMethod[config.AuthenticationMethodKubernetesConfig]{
					Enabled: true,
					Method: config.AuthenticationMethodKubernetesConfig{
						DiscoveryURL:            tp.Addr(),
						CAPath:                  caPath,
						ServiceAccountTokenPath: fliptServiceAccountTokenPath,
					},
				},
			},
		}
		server = k8stesting.StartHTTPServer(t, ctx, logger, authConfig, router)
	)

	t.Cleanup(func() { _ = server.Stop() })

	client := &http.Client{}

	t.Run("Valid service account token", func(t *testing.T) {
		var (
			verifyURL = clientAddress + "/auth/v1/method/kubernetes/serviceaccount"
			// generate service account JWT using the configured private key
			payload = auth.VerifyServiceAccountRequest{
				ServiceAccountToken: oidc.TestSignJWT(
					t,
					priv,
					string(oidc.RS256),
					map[string]any{
						"exp": time.Now().Add(24 * time.Hour).Unix(),
						"iss": tp.Addr(),
						"kubernetes.io": map[string]any{
							"namespace": "applications",
							"pod": map[string]any{
								"name": "booking-7d26f049-kdurb",
								"uid":  "bd8299f9-c50f-4b76-af33-9d8e3ef2b850",
							},
							"serviceaccount": map[string]any{
								"name": "booking",
								"uid":  "4f18914e-f276-44b2-aebd-27db1d8f8def",
							},
						},
					},
					nil,
				),
			}
		)

		data, err := protojson.Marshal(&payload)
		require.NoError(t, err)

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, verifyURL, bytes.NewReader(data))
		require.NoError(t, err)

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		respData, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		var response auth.CallbackResponse
		if !assert.NoError(t, protojson.Unmarshal(respData, &response)) {
			t.Log("Unexpected response", string(respData))
			t.FailNow()
		}

		assert.NotEmpty(t, response.ClientToken)
		assert.Equal(t, auth.Method_METHOD_KUBERNETES, response.Authentication.Method)
		assert.Equal(t, map[string]string{
			"io.flipt.auth.k8s.namespace":           "applications",
			"io.flipt.auth.k8s.pod.name":            "booking-7d26f049-kdurb",
			"io.flipt.auth.k8s.pod.uid":             "bd8299f9-c50f-4b76-af33-9d8e3ef2b850",
			"io.flipt.auth.k8s.serviceaccount.name": "booking",
			"io.flipt.auth.k8s.serviceaccount.uid":  "4f18914e-f276-44b2-aebd-27db1d8f8def",
		}, response.Authentication.Metadata)

		// ensure expiry is set
		assert.NotNil(t, response.Authentication.ExpiresAt)
	})
}

func writeStringToTemp(t *testing.T, pattern, contents string) string {
	fi, err := os.CreateTemp("", pattern)
	require.NoError(t, err)

	_, err = io.Copy(fi, strings.NewReader(contents))
	require.NoError(t, err)

	err = fi.Close()
	require.NoError(t, err)

	return fi.Name()
}
