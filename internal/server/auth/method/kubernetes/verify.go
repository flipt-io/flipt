package kubernetes

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"go.flipt.io/flipt/internal/config"
	"go.uber.org/zap"
)

type kubernetesOIDCVerifier struct {
	logger *zap.Logger
	config config.AuthenticationMethodKubernetesConfig
	client *http.Client
}

func newKubernetesOIDCVerifier(logger *zap.Logger, config config.AuthenticationMethodKubernetesConfig) (*kubernetesOIDCVerifier, error) {
	caCert, err := os.ReadFile(config.CAPath)
	if err != nil {
		logger.Error("reading CA certificate", zap.Error(err))

		return nil, fmt.Errorf("building OIDC client: %w", err)
	}

	rootCAs := x509.NewCertPool()
	if !rootCAs.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("failed to append cert from path: %q", config.CAPath)
	}

	transport := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		TLSClientConfig:       &tls.Config{RootCAs: rootCAs},
	}

	return &kubernetesOIDCVerifier{
		logger: logger,
		config: config,
		client: &http.Client{
			Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
				// Re-evaluate token from disk per request.
				// this client will be used by the OIDC code to periodically fetch
				// OIDC configuration and PKI key chain from k8s.
				// It handles caching that result.
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
		},
	}, nil
}

func (k *kubernetesOIDCVerifier) verify(ctx context.Context, jwt string) (c claims, err error) {
	provider, err := oidc.NewProvider(
		oidc.ClientContext(ctx, k.client),
		k.config.IssuerURL,
	)
	if err != nil {
		return c, err
	}

	token, err := provider.Verifier(&oidc.Config{
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
