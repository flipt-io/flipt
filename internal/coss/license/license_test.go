package license

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/keygen-sh/keygen-go/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/internal/product"
	"go.uber.org/zap/zaptest"
)

func TestManager_validateAndSet_NoLicenseKey(t *testing.T) {
	logger := zaptest.NewLogger(t)
	manager := &ManagerImpl{
		logger:  logger,
		config:  &config.LicenseConfig{Key: ""}, // Empty license key
		product: product.Pro,
		cache:   &licenseCache{},
	}

	ctx := t.Context()
	manager.validateAndSet(ctx)

	// Should fallback to OSS when no license key is provided
	assert.Equal(t, product.OSS, manager.product)
}

func TestManager_validateAndSet_FingerprintError(t *testing.T) {
	logger := zaptest.NewLogger(t)
	manager := &ManagerImpl{
		logger:        logger,
		config:        &config.LicenseConfig{Key: "test-key"},
		licenseType:   LicenseTypeOnline,
		fingerprinter: func(string) (string, error) { return "", errors.New("fingerprint error") },
		product:       product.Pro,
		cache:         &licenseCache{},
	}

	ctx := t.Context()
	manager.validateAndSet(ctx)

	// Should fallback to OSS when fingerprint fails
	assert.Equal(t, product.OSS, manager.product)
}

func TestManager_Shutdown_OnlineLicense(t *testing.T) {
	logger := zaptest.NewLogger(t)

	expiry := time.Now().Add(24 * time.Hour)
	license := &keygen.License{
		Expiry: &expiry,
	}

	ctx, cancel := context.WithCancel(t.Context())

	manager := &ManagerImpl{
		logger:        logger,
		license:       license,
		licenseType:   LicenseTypeOnline,
		fingerprinter: func(string) (string, error) { return "test-fingerprint", nil },
		productID:     "test-product",
		done:          make(chan struct{}),
		cancel:        cancel,
	}

	go func() {
		manager.periodicRevalidate(ctx)
	}()

	err := manager.Shutdown(t.Context())
	require.NoError(t, err)
}

func TestManager_Shutdown_OfflineLicense(t *testing.T) {
	logger := zaptest.NewLogger(t)

	expiry := time.Now().Add(24 * time.Hour)
	license := &keygen.License{
		Expiry: &expiry,
	}

	ctx, cancel := context.WithCancel(t.Context())

	manager := &ManagerImpl{
		logger:        logger,
		license:       license,
		licenseType:   LicenseTypeOffline, // Offline license should not deactivate
		fingerprinter: func(string) (string, error) { return "test-fingerprint", nil },
		productID:     "test-product",
		done:          make(chan struct{}),
		cancel:        cancel,
	}

	go func() {
		manager.periodicRevalidate(ctx)
	}()

	err := manager.Shutdown(t.Context())
	require.NoError(t, err)
}

func TestManager_validateOffline_Success(t *testing.T) {
	logger := zaptest.NewLogger(t)

	// Create a temporary license file for testing
	tempDir := t.TempDir()
	licenseFile := filepath.Join(tempDir, "license.cert")

	// Write a mock license certificate
	mockCert := "-----BEGIN LICENSE-----\nMOCK_CERTIFICATE_DATA\n-----END LICENSE-----"
	err := os.WriteFile(licenseFile, []byte(mockCert), 0o600)
	require.NoError(t, err)

	manager := &ManagerImpl{
		logger: logger,
		config: &config.LicenseConfig{
			File: licenseFile,
			Key:  "mock-decryption-key",
		},
	}

	ctx := t.Context()

	// This will fail because we don't have a real license file, but we can test the file reading logic
	_, err = manager.validateOffline(ctx)
	require.Error(t, err) // Expected to fail with mock data
}

func TestManager_validateOffline_FileNotFound(t *testing.T) {
	logger := zaptest.NewLogger(t)

	manager := &ManagerImpl{
		logger: logger,
		config: &config.LicenseConfig{
			File: "/nonexistent/license.cert",
			Key:  "mock-decryption-key",
		},
	}

	ctx := t.Context()
	_, err := manager.validateOffline(ctx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no such file or directory")
}

func TestManager_LicenseTypeDetermination(t *testing.T) {
	tests := []struct {
		name         string
		config       *config.LicenseConfig
		expectedType LicenseType
	}{
		{
			name: "online license when no file specified",
			config: &config.LicenseConfig{
				Key:  "online-license-key",
				File: "",
			},
			expectedType: LicenseTypeOnline,
		},
		{
			name: "offline license when file specified",
			config: &config.LicenseConfig{
				Key:  "offline-license-key",
				File: "/path/to/license.cert",
			},
			expectedType: LicenseTypeOffline,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zaptest.NewLogger(t)

			manager := &ManagerImpl{
				logger:      logger,
				config:      tt.config,
				licenseType: LicenseTypeOnline, // Default to online
			}

			// Simulate the license type determination logic from NewManager
			if tt.config.File != "" {
				manager.licenseType = LicenseTypeOffline
			}

			assert.Equal(t, tt.expectedType, manager.licenseType)
		})
	}
}

func TestManager_validateAndSet_OfflineLicense(t *testing.T) {
	logger := zaptest.NewLogger(t)

	// Create a temporary license file for testing
	tempDir := t.TempDir()
	licenseFile := filepath.Join(tempDir, "license.cert")

	// Write a mock license certificate
	mockCert := "-----BEGIN LICENSE-----\nMOCK_CERTIFICATE_DATA\n-----END LICENSE-----"
	err := os.WriteFile(licenseFile, []byte(mockCert), 0o600)
	require.NoError(t, err)

	manager := &ManagerImpl{
		logger: logger,
		config: &config.LicenseConfig{
			Key:  "mock-decryption-key",
			File: licenseFile,
		},
		licenseType: LicenseTypeOffline,
		product:     product.Pro,
		cache:       &licenseCache{},
	}

	ctx := t.Context()
	manager.validateAndSet(ctx)

	// Should fallback to OSS when offline license validation fails (with mock data)
	assert.Equal(t, product.OSS, manager.product)
}

func TestManager_validateAndSet_LicenseWithoutExpiry(t *testing.T) {
	logger := zaptest.NewLogger(t)

	manager := &ManagerImpl{
		logger:      logger,
		config:      &config.LicenseConfig{Key: "test-key"},
		licenseType: LicenseTypeOnline,
		product:     product.Pro,
		cache:       &licenseCache{},
	}

	// cache a license without an expiry date so validateAndSet short circuits
	// before any API call
	manager.cache.set(&keygen.License{Key: "test-key"}, cacheDuration)

	manager.validateAndSet(t.Context())

	assert.Equal(t, product.OSS, manager.Product())
	assert.Nil(t, manager.license)
}

func TestManager_Shutdown_FingerprintError(t *testing.T) {
	logger := zaptest.NewLogger(t)

	expiry := time.Now().Add(24 * time.Hour)
	license := &keygen.License{
		Expiry: &expiry,
	}

	ctx, cancel := context.WithCancel(t.Context())

	manager := &ManagerImpl{
		logger:        logger,
		license:       license,
		licenseType:   LicenseTypeOnline,
		fingerprinter: func(string) (string, error) { return "", errors.New("fingerprint error") },
		productID:     "test-product",
		done:          make(chan struct{}),
		cancel:        cancel,
	}

	go func() {
		manager.periodicRevalidate(ctx)
	}()

	err := manager.Shutdown(t.Context())
	require.Error(t, err) // Should return error when fingerprint fails
	assert.Contains(t, err.Error(), "fingerprint error")
}

func TestManager_Shutdown_NilLicense(t *testing.T) {
	logger := zaptest.NewLogger(t)

	ctx, cancel := context.WithCancel(t.Context())

	manager := &ManagerImpl{
		logger:        logger,
		license:       nil, // No license set
		licenseType:   LicenseTypeOnline,
		fingerprinter: func(string) (string, error) { return "test-fingerprint", nil },
		productID:     "test-product",
		done:          make(chan struct{}),
		cancel:        cancel,
	}

	go func() {
		manager.periodicRevalidate(ctx)
	}()

	err := manager.Shutdown(t.Context())
	require.NoError(t, err) // Should not error when license is nil
}

func TestManager_Product(t *testing.T) {
	logger := zaptest.NewLogger(t)

	manager := &ManagerImpl{
		logger:  logger,
		product: product.Pro,
	}

	result := manager.Product()
	assert.Equal(t, product.Pro, result)
}

func TestProtectedMachineID(t *testing.T) {
	// ProtectedMachineID should produce a deterministic, non-empty hex string
	result := ProtectedMachineID("my-app", "my-machine-id")
	assert.NotEmpty(t, result)

	// Same inputs should produce same output
	assert.Equal(t, result, ProtectedMachineID("my-app", "my-machine-id"))

	// Different inputs should produce different output
	assert.NotEqual(t, result, ProtectedMachineID("other-app", "my-machine-id"))
	assert.NotEqual(t, result, ProtectedMachineID("my-app", "other-machine-id"))
}

func TestManager_ConfiguredMachineID(t *testing.T) {
	logger := zaptest.NewLogger(t)

	const machineID = "my-custom-machine-id"

	manager := &ManagerImpl{
		logger:      logger,
		config:      &config.LicenseConfig{Key: "test-key", MachineID: machineID},
		licenseType: LicenseTypeOnline,
		fingerprinter: func(appID string) (string, error) {
			return ProtectedMachineID(appID, machineID), nil
		},
		product:   product.Pro,
		cache:     &licenseCache{},
		productID: "test-product",
	}

	// Verify the fingerprinter uses the configured machine ID
	fp, err := manager.fingerprinter("test-product")
	require.NoError(t, err)
	assert.Equal(t, ProtectedMachineID("test-product", machineID), fp)
}

func TestManager_validateAndSet_CachedLicenseExpiry(t *testing.T) {
	const licenseKey = "test-license-key"

	tests := []struct {
		name            string
		expiry          time.Time
		expectedProduct product.Product
		expectLicense   bool
	}{
		{
			name:            "valid cached license",
			expiry:          time.Now().Add(24 * time.Hour),
			expectedProduct: product.Pro,
			expectLicense:   true,
		},
		{
			name:            "expired cached license",
			expiry:          time.Now().Add(-time.Second),
			expectedProduct: product.OSS,
			expectLicense:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cached := &keygen.License{Key: licenseKey, Expiry: &tt.expiry}
			manager := &ManagerImpl{
				logger:      zaptest.NewLogger(t),
				config:      &config.LicenseConfig{Key: licenseKey},
				licenseType: LicenseTypeOnline,
				product:     product.Pro,
				cache:       &licenseCache{},
			}
			// cache the license so validateAndSet short circuits before any API call
			manager.cache.set(cached, cacheDuration)

			manager.validateAndSet(t.Context())

			assert.Equal(t, tt.expectedProduct, manager.Product())
			if tt.expectLicense {
				require.NotNil(t, manager.license)
				assert.Equal(t, licenseKey, manager.license.Key)
			} else {
				assert.Nil(t, manager.license)
			}
		})
	}
}

func TestHasValidExpiry(t *testing.T) {
	future := time.Now().Add(time.Hour)
	past := time.Now().Add(-time.Hour)

	assert.False(t, hasValidExpiry(nil), "nil license")
	assert.False(t, hasValidExpiry(&keygen.License{}), "license without expiry")
	assert.False(t, hasValidExpiry(&keygen.License{Expiry: &past}), "expired license")
	assert.True(t, hasValidExpiry(&keygen.License{Expiry: &future}), "unexpired license")
}
