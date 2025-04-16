package v2

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
	"golang.org/x/oauth2"
)

type kubernetesOIDCVerifier struct {
	logger   *zap.Logger
	verifier *oidc.IDTokenVerifier
}

func newKubernetesOIDCVerifier(logger *zap.Logger, config config.AuthenticationMethodKubernetesConfig) (*kubernetesOIDCVerifier, error) {
	// Read CA cert if provided
	var rootCAs *x509.CertPool
	if config.CAPath != "" {
		rootCAs = x509.NewCertPool()
		b, err := os.ReadFile(config.CAPath)
		if err != nil {
			return nil, fmt.Errorf("reading CA cert: %w", err)
		}
		if !rootCAs.AppendCertsFromPEM(b) {
			return nil, fmt.Errorf("appending CA cert to pool")
		}
	}

	// Configure HTTP client with custom transport
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
		Timeout: 30 * time.Second,
	}

	// Configure OIDC provider with custom client
	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, client)
	provider, err := oidc.NewProvider(ctx, config.DiscoveryURL)
	if err != nil {
		return nil, fmt.Errorf("creating OIDC provider: %w", err)
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
