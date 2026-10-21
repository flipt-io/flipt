package kubernetes

import (
	"crypto/rand"
	"crypto/rsa"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/hashicorp/cap/oidc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/config"
	"go.uber.org/zap/zaptest"
)

func TestVerifierAudience(t *testing.T) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	tp := oidc.StartTestProvider(t, oidc.WithTestDefaults(&oidc.TestProviderDefaults{
		SigningKey: &oidc.TestSigningKey{PrivKey: priv, PubKey: priv.Public(), Alg: oidc.RS256},
	}))
	defer tp.Stop()

	token := func(aud ...string) string {
		claims := map[string]any{
			"exp": time.Now().Add(time.Hour).Unix(),
			"iss": tp.Addr(),
			"kubernetes.io": map[string]any{
				"namespace":      "applications",
				"serviceaccount": map[string]any{"name": "booking"},
			},
		}
		if len(aud) > 0 {
			claims["aud"] = aud
		}
		return oidc.TestSignJWT(t, priv, string(oidc.RS256), claims, nil)
	}

	writeFile := func(t *testing.T, name, contents string) string {
		t.Helper()
		path := filepath.Join(t.TempDir(), name)
		require.NoError(t, os.WriteFile(path, []byte(contents), 0o600))
		return path
	}

	tests := []struct {
		name      string
		audiences []string
		token     string
		wantErr   bool
	}{
		{name: "no audiences configured accepts any audience", token: token("vault")},
		{name: "configured audience accepted", audiences: []string{"flipt"}, token: token("flipt")},
		{name: "configured audience among several on the token", audiences: []string{"flipt"}, token: token("vault", "flipt")},
		{name: "any of several configured audiences accepted", audiences: []string{"flipt", "flipt-next"}, token: token("flipt-next")},
		{name: "token for another service rejected", audiences: []string{"flipt"}, token: token("vault"), wantErr: true},
		{name: "audience match is case-sensitive", audiences: []string{"flipt"}, token: token("Flipt"), wantErr: true},
		{name: "token without audience rejected", audiences: []string{"flipt"}, token: token(), wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			verifier, err := newKubernetesOIDCVerifier(zaptest.NewLogger(t), config.AuthenticationMethodKubernetesConfig{
				DiscoveryURL:            tp.Addr(),
				CAPath:                  writeFile(t, "ca.cert", tp.CACert()),
				ServiceAccountTokenPath: writeFile(t, "token", token("https://kubernetes.default.svc.cluster.local")),
				Audiences:               tt.audiences,
			})
			require.NoError(t, err)

			c, err := verifier.verify(t.Context(), tt.token)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "audience")
				return
			}

			require.NoError(t, err)
			assert.Equal(t, "applications", c.Identity.Namespace)
			assert.Equal(t, "booking", c.Identity.ServiceAccount.Name)
		})
	}
}
