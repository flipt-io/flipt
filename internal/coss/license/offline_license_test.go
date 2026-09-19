package license

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/keygen-sh/keygen-go/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/internal/product"
	"go.uber.org/zap"
)

// makeOfflineLicenseFixture builds a deterministic, non-production Keygen
// AES-256-GCM + Ed25519 license-file certificate. It keeps the tests hermetic
// while exercising the same parsing, signature verification, and decryption
// paths as a checked-out Keygen license file.
func makeOfflineLicenseFixture(t *testing.T, privateKey ed25519.PrivateKey, licenseKey string, licenseExpiry, fileExpiry time.Time) string {
	t.Helper()

	issued := time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC)
	payload := map[string]any{
		"data": map[string]any{
			"type": "licenses",
			"id":   "test-license",
			"attributes": map[string]any{
				"name":             "Flipt Pro test license",
				"key":              licenseKey,
				"expiry":           licenseExpiry,
				"scheme":           "ED25519_SIGN",
				"requireHeartbeat": false,
				"created":          issued,
				"updated":          issued,
				"metadata":         map[string]any{},
			},
		},
		"meta": map[string]any{
			"issued": issued,
			"expiry": fileExpiry,
			"ttl":    int(time.Hour.Seconds()),
		},
	}
	plaintext, err := json.Marshal(payload)
	require.NoError(t, err)

	key := sha256.Sum256([]byte(licenseKey))
	block, err := aes.NewCipher(key[:])
	require.NoError(t, err)
	gcm, err := cipher.NewGCM(block)
	require.NoError(t, err)

	// A deterministic nonce is safe here because these are test-only keys and
	// certificates. It keeps failures reproducible and no fixture is reused in
	// production.
	nonceHash := sha256.Sum256([]byte(licenseKey + licenseExpiry.String() + fileExpiry.String()))
	nonce := nonceHash[:gcm.NonceSize()]
	sealed := gcm.Seal(nil, nonce, plaintext, nil)
	ciphertext, tag := sealed[:len(sealed)-gcm.Overhead()], sealed[len(sealed)-gcm.Overhead():]
	enc := strings.Join([]string{
		base64.StdEncoding.EncodeToString(ciphertext),
		base64.StdEncoding.EncodeToString(nonce),
		base64.StdEncoding.EncodeToString(tag),
	}, ".")

	certificate := map[string]string{
		"enc": enc,
		"sig": base64.StdEncoding.EncodeToString(ed25519.Sign(privateKey, []byte("license/"+enc))),
		"alg": "aes-256-gcm+ed25519",
	}
	encoded, err := json.Marshal(certificate)
	require.NoError(t, err)

	return "-----BEGIN LICENSE FILE-----\n" + base64.StdEncoding.EncodeToString(encoded) + "\n-----END LICENSE FILE-----\n"
}

func writeOfflineLicenseFixture(t *testing.T, dir, name, certificate string) string {
	t.Helper()

	path := filepath.Join(dir, name)
	require.NoError(t, os.WriteFile(path, []byte(certificate), 0o600))
	return path
}

func corruptOfflineLicenseSignature(t *testing.T, certificate string) string {
	t.Helper()

	payload := strings.TrimSpace(certificate)
	payload = strings.TrimPrefix(payload, "-----BEGIN LICENSE FILE-----")
	payload = strings.TrimSuffix(payload, "-----END LICENSE FILE-----")
	decoded, err := base64.StdEncoding.DecodeString(strings.TrimSpace(payload))
	require.NoError(t, err)

	var contents map[string]string
	require.NoError(t, json.Unmarshal(decoded, &contents))
	contents["sig"] = base64.StdEncoding.EncodeToString(make([]byte, ed25519.SignatureSize))
	decoded, err = json.Marshal(contents)
	require.NoError(t, err)

	return "-----BEGIN LICENSE FILE-----\n" + base64.StdEncoding.EncodeToString(decoded) + "\n-----END LICENSE FILE-----\n"
}

func newOfflineManager(path, licenseKey, publicKey string) *ManagerImpl {
	return &ManagerImpl{
		logger:      zap.NewNop(),
		config:      &config.LicenseConfig{File: path, Key: licenseKey},
		licenseType: LicenseTypeOffline,
		verifyKey:   publicKey,
		product:     product.OSS,
		cache:       &licenseCache{},
	}
}

func TestManager_validateAndSet_OfflineLicensePairs(t *testing.T) {
	privateKey := ed25519.NewKeyFromSeed(bytes.Repeat([]byte{7}, ed25519.SeedSize))
	publicKey := hex.EncodeToString(privateKey.Public().(ed25519.PublicKey))
	dir := t.TempDir()
	const future = "2099-01-01T00:00:00Z"
	futureExpiry, err := time.Parse(time.RFC3339, future)
	require.NoError(t, err)

	oldKey := "old-license-key"
	bridgeKey := "bridge-license-key"
	oldPath := writeOfflineLicenseFixture(t, dir, "old.lic", makeOfflineLicenseFixture(t, privateKey, oldKey, futureExpiry, futureExpiry))
	bridgePath := writeOfflineLicenseFixture(t, dir, "bridge.lic", makeOfflineLicenseFixture(t, privateKey, bridgeKey, futureExpiry, futureExpiry))

	tests := []struct {
		name            string
		path            string
		key             string
		expectedProduct product.Product
	}{
		// Issuing a bridge license must not invalidate the old certificate.
		{name: "old pair remains valid after bridge is issued", path: oldPath, key: oldKey, expectedProduct: product.Pro},
		// A restart with both new configuration values accepts the bridge.
		{name: "replacing the configured file and key together enables Pro", path: bridgePath, key: bridgeKey, expectedProduct: product.Pro},
		{name: "bridge file with old key disables Pro", path: bridgePath, key: oldKey, expectedProduct: product.OSS},
		{name: "old file with bridge key disables Pro", path: oldPath, key: bridgeKey, expectedProduct: product.OSS},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := newOfflineManager(tt.path, tt.key, publicKey)
			manager.validateAndSet(t.Context())

			assert.Equal(t, tt.expectedProduct, manager.Product())
			if tt.expectedProduct == product.Pro {
				require.NotNil(t, manager.license)
				assert.Equal(t, tt.key, manager.license.Key)
			} else {
				assert.Nil(t, manager.license)
			}
		})
	}
}

func TestManager_ProductRejectsLicenseAfterExpiry(t *testing.T) {
	expiredAt := time.Now().Add(-time.Second)
	manager := &ManagerImpl{
		license: &keygen.License{Expiry: &expiredAt},
		product: product.Pro,
	}

	assert.Equal(t, product.OSS, manager.Product())
}

func TestManager_validateAndSet_OfflineLicenseRejectsInvalidFiles(t *testing.T) {
	privateKey := ed25519.NewKeyFromSeed(bytes.Repeat([]byte{9}, ed25519.SeedSize))
	publicKey := hex.EncodeToString(privateKey.Public().(ed25519.PublicKey))
	dir := t.TempDir()
	const licenseKey = "test-license-key"
	futureExpiry := time.Date(2099, time.January, 1, 0, 0, 0, 0, time.UTC)
	pastExpiry := time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC)

	valid := makeOfflineLicenseFixture(t, privateKey, licenseKey, futureExpiry, futureExpiry)
	badSignature := corruptOfflineLicenseSignature(t, valid)

	tests := []struct {
		name        string
		certificate string
	}{
		{name: "malformed", certificate: "not a license certificate"},
		{name: "wrong signature", certificate: badSignature},
		{name: "expired file TTL", certificate: makeOfflineLicenseFixture(t, privateKey, licenseKey, futureExpiry, pastExpiry)},
		{name: "expired embedded license", certificate: makeOfflineLicenseFixture(t, privateKey, licenseKey, pastExpiry, futureExpiry)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := writeOfflineLicenseFixture(t, dir, strings.ReplaceAll(tt.name, " ", "-")+".lic", tt.certificate)
			manager := newOfflineManager(path, licenseKey, publicKey)
			manager.validateAndSet(t.Context())

			assert.Equal(t, product.OSS, manager.Product())
			assert.Nil(t, manager.license)
		})
	}
}
