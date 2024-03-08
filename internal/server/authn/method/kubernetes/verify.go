package kubernetes

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"go.flipt.io/flipt/internal/config"
	"go.uber.org/zap"
)

// kubernetesOIDCVerifier uses the go-oidc library to obtain the OIDC configuration
// and JWKS key material, in order to verify Kubernetes issued service account tokens.
// It is configured to leverage the local systems own service account token and the clusters
// CA certificate in order to obtain this information from the local cluster.
// The material returned from the OIDC configuration endpoints is trusted as the signing party
// in order to validate presented service account tokens.
type kubernetesOIDCVerifier struct {
	logger   *zap.Logger
	config   config.AuthenticationMethodKubernetesConfig
	provider *oidc.Provider
}

func newKubernetesOIDCVerifier(logger *zap.Logger, config config.AuthenticationMethodKubernetesConfig) (*kubernetesOIDCVerifier, error) {
	ctx := context.Background()
	caCert, err := os.ReadFile(config.CAPath)
	if err != nil {
		logger.Error("reading CA certificate", zap.Error(err))

		return nil, fmt.Errorf("building OIDC client: %w", err)
	}

	rootCAs := x509.NewCertPool()
	if !rootCAs.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("failed to append cert from path: %q", config.CAPath)
	}

	// adapted from the Go net/http.DefaultTransport
	// This transport only uses the configured CA certificate
	// PEM found at the configured path on the filesystem.
	transport := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		TLSClientConfig: &tls.Config{
			RootCAs:    rootCAs,
			MinVersion: tls.VersionTLS12,
		},
	}

	client := &http.Client{
		Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
			// Re-evaluate token from disk per request.
			// This client will be used by the OIDC code to periodically fetch
			// OIDC configuration and JWKS key chain from the k8s api-server.
			// The OIDC wrapper handles caching that result.
			// Each time it needs to request again we should re-read the SA token
			// as it may have been refreshed by kubernetes.
			token, err := os.ReadFile(config.ServiceAccountTokenPath)
			if err != nil {
				logger.Error("reading service account token", zap.Error(err))

				return nil, fmt.Errorf("authentication OIDC client: %w", err)
			}

			if _, ok := r.Header["Authorization"]; !ok {
				r.Header["Authorization"] = []string{fmt.Sprintf("Bearer %s", token)}
			}

			return transport.RoundTrip(r)
		}),
	}

	// Kubernetes is not an OIDC / OAuth provider in the traditional sense
	// and they go off-specification. The Issuer returned by the "well-known" endpoint
	// does not match the supplied discovery URL.
	// To ensure this isn't a problem we skip the OIDC libraries requirement
	// that both URLs should match.
	// We also instruct the library to use the issuer retrieved from the discovery
	// endpoint when we verify service account ID tokens.
	issuer, err := resolveTokenIssuer(ctx, client, config.DiscoveryURL)
	if err != nil {
		return nil, err
	}

	provider, err := oidc.NewProvider(
		// skip issuer verification when NewProvider requests the discovery document.
		oidc.InsecureIssuerURLContext(
			oidc.ClientContext(ctx, client),
			// override the issuer to match the discovery endpoint response.
			issuer,
		),
		config.DiscoveryURL,
	)
	if err != nil {
		return nil, err
	}

	return &kubernetesOIDCVerifier{
		logger:   logger,
		config:   config,
		provider: provider,
	}, nil
}

func (k *kubernetesOIDCVerifier) verify(ctx context.Context, jwt string) (c claims, err error) {
	token, err := k.provider.Verifier(&oidc.Config{
		// we're not interested in the OIDC client ID
		// as we're not doing a real OAuth 2 flow.
		SkipClientIDCheck: true,
	}).Verify(ctx, jwt)
	if err != nil {
		return c, err
	}

	c.Expiration = token.Expiry.Unix()

	err = token.Claims(&c)

	return
}

type transportFunc func(*http.Request) (*http.Response, error)

func (fn transportFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return fn(r)
}

func resolveTokenIssuer(ctx context.Context, client *http.Client, discoveryURL string) (string, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		"GET",
		strings.TrimSuffix(discoveryURL, "/")+"/.well-known/openid-configuration",
		nil)
	if err != nil {
		return "", err
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetching OIDC configuration: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading OIDC configuration: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("OIDC configuration response status %q: %w", resp.Status, err)
	}

	var config struct {
		Issuer string `json:"issuer"`
	}

	if err = json.Unmarshal(body, &config); err != nil {
		return "", fmt.Errorf("OIDC configuration unmarshal: %w", err)
	}

	return config.Issuer, nil
}
