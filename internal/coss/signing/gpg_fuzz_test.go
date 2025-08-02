//go:build go1.18
// +build go1.18

package signing

import (
	"context"
	"testing"

	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/internal/secrets"
	"go.uber.org/zap"
)

// FuzzMockSecretsManager for fuzz testing
type FuzzMockSecretsManager struct {
	data map[string][]byte
}

func (m *FuzzMockSecretsManager) GetSecretValue(ctx context.Context, ref secrets.Reference) ([]byte, error) {
	key := ref.Provider + ":" + ref.Path + ":" + ref.Key
	if data, exists := m.data[key]; exists {
		return data, nil
	}
	return []byte("mock-secret-value"), nil
}

func (m *FuzzMockSecretsManager) RegisterProvider(name string, provider secrets.Provider) error {
	return nil
}

func (m *FuzzMockSecretsManager) GetProvider(name string) (secrets.Provider, error) {
	return nil, nil
}

func (m *FuzzMockSecretsManager) GetSecret(ctx context.Context, providerName, path string) (*secrets.Secret, error) {
	return nil, nil
}

func (m *FuzzMockSecretsManager) ListSecrets(ctx context.Context, providerName, pathPrefix string) ([]string, error) {
	return nil, nil
}

func (m *FuzzMockSecretsManager) ListProviders() []string {
	return nil
}

func (m *FuzzMockSecretsManager) Close() error {
	return nil
}

func FuzzGPGKeyParsing(f *testing.F) {
	// Add seed corpus with various PGP key formats
	seeds := []string{
		// Valid PGP private key (truncated for brevity)
		`-----BEGIN PGP PRIVATE KEY BLOCK-----

mQENBF...truncated...
-----END PGP PRIVATE KEY BLOCK-----`,

		// Valid PGP public key
		`-----BEGIN PGP PUBLIC KEY BLOCK-----

mQENBF...truncated...
-----END PGP PUBLIC KEY BLOCK-----`,

		// Invalid armor headers
		`-----BEGIN INVALID BLOCK-----
content
-----END INVALID BLOCK-----`,

		// Malformed armor
		`-----BEGIN PGP PRIVATE KEY BLOCK-----
invalid content without proper encoding
-----END PGP PRIVATE KEY BLOCK-----`,

		// Empty blocks
		`-----BEGIN PGP PRIVATE KEY BLOCK-----
-----END PGP PRIVATE KEY BLOCK-----`,
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	// Add edge cases
	f.Add("")
	f.Add("-----BEGIN")
	f.Add("-----END")
	f.Add("-----BEGIN PGP PRIVATE KEY BLOCK-----")
	f.Add("-----END PGP PRIVATE KEY BLOCK-----")
	f.Add("not-a-pgp-key")

	f.Fuzz(func(t *testing.T, pgpKey string) {
		// Recover from panics
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("GPG key parsing panicked with input length %d: %v", len(pgpKey), r)
			}
		}()

		// Create mock secrets manager with the fuzzed key
		mockManager := &FuzzMockSecretsManager{
			data: map[string][]byte{
				"vault:signing:private_key": []byte(pgpKey),
			},
		}

		// Create GPG signer with mock
		secretRef := config.SecretReference{
			Provider: "vault",
			Path:     "signing",
			Key:      "private_key",
		}

		signer, err := NewGPGSigner(secretRef, "test-key", mockManager, zap.NewNop())
		if err != nil {
			// Skip if signer creation fails, we only care about panics
			return
		}

		// Try to load entity - this is where PGP parsing happens
		ctx := context.Background()
		_ = signer.loadEntity(ctx)

		// Try to get public key
		_, _ = signer.GetPublicKey(ctx)
	})
}

func FuzzGPGArmorParsing(f *testing.F) {
	// Add seed corpus with various armor formats
	seeds := []string{
		`-----BEGIN PGP SIGNATURE-----

iQEcBAABCgAGBQJhQ...
-----END PGP SIGNATURE-----`,

		`-----BEGIN PGP MESSAGE-----

hQEMA...
-----END PGP MESSAGE-----`,

		// Different line endings
		"-----BEGIN PGP SIGNATURE-----\r\niQEcBAABCgAGBQJhQ\r\n-----END PGP SIGNATURE-----",

		// With headers
		`-----BEGIN PGP SIGNATURE-----
Hash: SHA256

iQEcBAABCgAGBQJhQ...
-----END PGP SIGNATURE-----`,
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	// Edge cases
	f.Add("-----BEGIN PGP SIGNATURE-----\n-----END PGP SIGNATURE-----")
	f.Add("-----BEGIN\n-----END")
	f.Add("BEGIN PGP SIGNATURE")
	f.Add("-----BEGIN PGP SIGNATURE-----\ninvalid base64\n-----END PGP SIGNATURE-----")

	f.Fuzz(func(t *testing.T, armoredData string) {
		// Recover from panics
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Armor parsing panicked with input length %d: %v", len(armoredData), r)
			}
		}()

		// Test armor parsing by trying to decode it
		// This tests the armor decoding logic similar to what's used in GPG operations
		// We don't need to import the armor package directly since it's tested indirectly
		// through the GPG signer's operations
		mockManager := &FuzzMockSecretsManager{
			data: map[string][]byte{
				"vault:signing:private_key": []byte(armoredData),
			},
		}

		secretRef := config.SecretReference{
			Provider: "vault",
			Path:     "signing",
			Key:      "private_key",
		}

		signer, err := NewGPGSigner(secretRef, "test-key", mockManager, zap.NewNop())
		if err != nil {
			return
		}

		ctx := context.Background()
		_ = signer.loadEntity(ctx)
	})
}
