package redis

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	goredis "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/config"
)

func TestTLSInsecure(t *testing.T) {
	tests := []struct {
		name     string
		input    bool
		expected bool
		mode     config.RedisCacheMode
	}{
		{"single: insecure disabled", false, false, config.RedisCacheModeSingle},
		{"single: insecure enabled", true, true, config.RedisCacheModeSingle},
		{"cluster: insecure disabled", false, false, config.RedisCacheModeCluster},
		{"cluster: insecure enabled", true, true, config.RedisCacheModeCluster},
	}
	for _, tt := range tests {
		cfg := config.Default()
		cfg.Authentication.Session.Storage.Redis.RequireTLS = true
		cfg.Authentication.Session.Storage.Redis.InsecureSkipTLS = tt.input
		cfg.Authentication.Session.Storage.Redis.Mode = tt.mode

		client, err := NewClient(*cfg)
		require.NoError(t, err)

		switch c := client.(type) {
		case *goredis.Client:
			require.Equal(t, tt.expected, c.Options().TLSConfig.InsecureSkipVerify)
		case *goredis.ClusterClient:
			require.Equal(t, tt.expected, c.Options().TLSConfig.InsecureSkipVerify)
		default:
			t.Fatalf("unexpected client type: %T", client)
		}
	}
}

func TestTLSCABundle(t *testing.T) {
	ca := generateCA(t)

	tests := []struct {
		name string
		mode config.RedisCacheMode
	}{
		{"single", config.RedisCacheModeSingle},
		{"cluster", config.RedisCacheModeCluster},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.Default()
			cfg.Authentication.Session.Storage.Redis.RequireTLS = true
			cfg.Authentication.Session.Storage.Redis.CaCertBytes = ca
			cfg.Authentication.Session.Storage.Redis.Mode = tt.mode

			client, err := NewClient(*cfg)
			require.NoError(t, err)

			switch c := client.(type) {
			case *goredis.Client:
				require.NotNil(t, c.Options().TLSConfig.RootCAs)
			case *goredis.ClusterClient:
				require.NotNil(t, c.Options().TLSConfig.RootCAs)
			default:
				t.Fatalf("unexpected client type: %T", client)
			}
		})

		t.Run("load CA not provided", func(t *testing.T) {
			cfg := config.Default()
			cfg.Authentication.Session.Storage.Redis.RequireTLS = true
			cfg.Authentication.Session.Storage.Redis.Mode = tt.mode

			client, err := NewClient(*cfg)

			require.NoError(t, err)

			switch c := client.(type) {
			case *goredis.Client:
				require.Nil(t, c.Options().TLSConfig.RootCAs)
			case *goredis.ClusterClient:
				require.Nil(t, c.Options().TLSConfig.RootCAs)
			default:
				t.Fatalf("unexpected client type: %T", client)
			}
		})

		t.Run("load CABytes successful", func(t *testing.T) {
			cfg := config.Default()
			cfg.Authentication.Session.Storage.Redis.RequireTLS = true
			cfg.Authentication.Session.Storage.Redis.CaCertBytes = ca
			cfg.Authentication.Session.Storage.Redis.Mode = tt.mode

			client, err := NewClient(*cfg)

			require.NoError(t, err)

			switch c := client.(type) {
			case *goredis.Client:
				require.NotNil(t, c.Options().TLSConfig.RootCAs)
			case *goredis.ClusterClient:
				require.NotNil(t, c.Options().TLSConfig.RootCAs)
			}
		})

		t.Run("load CAPath successful", func(t *testing.T) {
			var (
				dir    = t.TempDir()
				cafile = filepath.Join(dir, "cafile.pem")
				err    = os.WriteFile(cafile, []byte(ca), 0600)
			)

			require.NoError(t, err)

			cfg := config.Default()
			cfg.Authentication.Session.Storage.Redis.RequireTLS = true
			cfg.Authentication.Session.Storage.Redis.CaCertPath = cafile
			cfg.Authentication.Session.Storage.Redis.Mode = tt.mode

			client, err := NewClient(*cfg)

			require.NoError(t, err)

			switch c := client.(type) {
			case *goredis.Client:
				require.NotNil(t, c.Options().TLSConfig.RootCAs)
			case *goredis.ClusterClient:
				require.NotNil(t, c.Options().TLSConfig.RootCAs)
			}
		})

		t.Run("load CAPath failure", func(t *testing.T) {
			var (
				dir    = t.TempDir()
				cafile = filepath.Join(dir, "cafile.pem")
			)

			cfg := config.Default()
			cfg.Authentication.Session.Storage.Redis.RequireTLS = true
			cfg.Authentication.Session.Storage.Redis.CaCertPath = cafile
			cfg.Authentication.Session.Storage.Redis.Mode = tt.mode

			_, err := NewClient(*cfg)

			require.Error(t, err)
		})
	}
}

func generateCA(t *testing.T) string {
	t.Helper()
	ca := &x509.Certificate{
		SerialNumber: big.NewInt(10002),
		IsCA:         true,
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(5 * time.Minute),
	}

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	cabytes, err := x509.CreateCertificate(rand.Reader, ca, ca, &key.PublicKey, key)
	require.NoError(t, err)

	buf := new(bytes.Buffer)
	err = pem.Encode(buf, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cabytes,
	})
	require.NoError(t, err)

	return buf.String()
}
