package license

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/product"
	"go.uber.org/zap/zaptest"
)

// mockLicense implements the License interface for testing
type mockLicense struct {
	expiry           *time.Time
	activateErr      error
	deactivateErr    error
	activateCalled   bool
	deactivateCalled bool
}

func (m *mockLicense) Activate(ctx context.Context, fingerprint string) (License, error) {
	m.activateCalled = true
	if m.activateErr != nil {
		return nil, m.activateErr
	}
	return m, nil
}

func (m *mockLicense) Deactivate(ctx context.Context, fingerprint string) error {
	m.deactivateCalled = true
	return m.deactivateErr
}

func (m *mockLicense) GetExpiry() *time.Time {
	return m.expiry
}

func TestManager_validateAndSet_WithValidLicense(t *testing.T) {
	logger := zaptest.NewLogger(t)

	// Setup mock license with valid expiry
	future := time.Now().Add(24 * time.Hour)
	mockLic := &mockLicense{
		expiry: &future,
	}

	// Override the validator for testing
	originalValidator := licenseValidator
	defer func() { licenseValidator = originalValidator }()

	licenseValidator = func(ctx context.Context, fingerprints ...string) (License, error) {
		return mockLic, nil
	}

	manager := &ManagerImpl{
		logger:        logger,
		licenseKey:    "test-key",
		fingerprinter: func(string) (string, error) { return "test-fingerprint", nil },
		product:       product.OSS,
	}

	ctx := context.Background()
	manager.validateAndSet(ctx)

	// Assert that the license was set and product was upgraded to Pro
	assert.Equal(t, product.Pro, manager.product)
	assert.Equal(t, mockLic, manager.license)
}

func TestManager_validateAndSet_WithInvalidLicense(t *testing.T) {
	logger := zaptest.NewLogger(t)

	// Override the validator for testing
	originalValidator := licenseValidator
	defer func() { licenseValidator = originalValidator }()

	licenseValidator = func(ctx context.Context, fingerprints ...string) (License, error) {
		return nil, errors.New("invalid license")
	}

	manager := &ManagerImpl{
		logger:        logger,
		licenseKey:    "test-key",
		fingerprinter: func(string) (string, error) { return "test-fingerprint", nil },
		product:       product.OSS,
	}

	ctx := context.Background()
	manager.validateAndSet(ctx)

	// Assert that the product remains OSS due to invalid license
	assert.Equal(t, product.OSS, manager.product)
	assert.Nil(t, manager.license)
}

func TestManager_validateAndSet_WithNilExpiry(t *testing.T) {
	logger := zaptest.NewLogger(t)

	// Setup mock license with nil expiry
	mockLic := &mockLicense{
		expiry: nil, // This should cause validation to fail
	}

	// Override the validator for testing
	originalValidator := licenseValidator
	defer func() { licenseValidator = originalValidator }()

	licenseValidator = func(ctx context.Context, fingerprints ...string) (License, error) {
		return mockLic, nil
	}

	manager := &ManagerImpl{
		logger:        logger,
		licenseKey:    "test-key",
		fingerprinter: func(string) (string, error) { return "test-fingerprint", nil },
		product:       product.OSS,
	}

	ctx := context.Background()
	manager.validateAndSet(ctx)

	// Assert that the product remains OSS due to nil expiry
	assert.Equal(t, product.OSS, manager.product)
	assert.Nil(t, manager.license)
}

func TestManager_Shutdown_CallsDeactivate(t *testing.T) {
	logger := zaptest.NewLogger(t)

	mockLic := &mockLicense{}

	manager := &ManagerImpl{
		logger:        logger,
		license:       mockLic,
		fingerprinter: func(string) (string, error) { return "test-fingerprint", nil },
		productID:     "test-product",
		done:          make(chan struct{}),
		cancel:        func() {},
	}

	ctx := context.Background()
	err := manager.Shutdown(ctx)

	require.NoError(t, err)
	assert.True(t, mockLic.deactivateCalled)
}
