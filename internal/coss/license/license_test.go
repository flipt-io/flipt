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

func TestManager_setOSSProduct(t *testing.T) {
	logger := zaptest.NewLogger(t)
	manager := &ManagerImpl{
		logger:  logger,
		product: product.Pro, // Start with Pro
	}

	manager.setOSSProduct()

	assert.Equal(t, product.OSS, manager.product)
}

func TestManager_validateAndSet_NoLicenseKey(t *testing.T) {
	logger := zaptest.NewLogger(t)
	manager := &ManagerImpl{
		logger:     logger,
		licenseKey: "", // Empty license key
		product:    product.Pro,
	}

	ctx := context.Background()
	manager.validateAndSet(ctx)

	// Should fallback to OSS when no license key is provided
	assert.Equal(t, product.OSS, manager.product)
}

func TestManager_validateAndSet_FingerprintError(t *testing.T) {
	logger := zaptest.NewLogger(t)
	manager := &ManagerImpl{
		logger:        logger,
		licenseKey:    "test-key",
		fingerprinter: func(string) (string, error) { return "", errors.New("fingerprint error") },
		product:       product.Pro,
	}

	ctx := context.Background()
	manager.validateAndSet(ctx)

	// Should fallback to OSS when fingerprint fails
	assert.Equal(t, product.OSS, manager.product)
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

func TestManager_Shutdown_HandlesDeactivateError(t *testing.T) {
	logger := zaptest.NewLogger(t)

	mockLic := &mockLicense{
		deactivateErr: errors.New("deactivate error"),
	}

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

	require.NoError(t, err) // Shutdown should not return error even if deactivate fails
	assert.True(t, mockLic.deactivateCalled)
}

func TestKeygenLicenseWrapper_NilLicenseHandling(t *testing.T) {
	wrapper := &keygenLicenseWrapper{License: nil}

	ctx := context.Background()

	// Test Activate with nil license
	_, err := wrapper.Activate(ctx, "fingerprint")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "license is nil")

	// Test Deactivate with nil license
	err = wrapper.Deactivate(ctx, "fingerprint")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "license is nil")

	// Test GetExpiry with nil license
	expiry := wrapper.GetExpiry()
	assert.Nil(t, expiry)
}