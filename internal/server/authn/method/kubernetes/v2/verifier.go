package v2

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

type kubernetesOIDCVerifier struct {
	logger   *zap.Logger
	verifier *oidc.IDTokenVerifier
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
	client := &http.Client{
		Transport: &http.Transport{
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
		},
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
	// Configure verifier
	verifier := provider.Verifier(&oidc.Config{
		SkipClientIDCheck: true,
	})

	return &kubernetesOIDCVerifier{
		logger:   logger,
		verifier: verifier,
	}, nil
}

func (k *kubernetesOIDCVerifier) verify(ctx context.Context, jwt string) (claims, error) {
	// Verify token
	token, err := k.verifier.Verify(ctx, jwt)
	if err != nil {
		return claims{}, fmt.Errorf("verifying token: %w", err)
	}

	// Parse claims
	var c claims
	if err := token.Claims(&c); err != nil {
		return claims{}, fmt.Errorf("parsing claims: %w", err)
	}

	return c, nil
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
