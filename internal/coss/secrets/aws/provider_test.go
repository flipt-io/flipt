package aws

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/internal/secrets"
	"go.uber.org/zap"
)

func TestFactoryRegistration(t *testing.T) {
	factory, exists := secrets.GetProviderFactory("aws")
	require.True(t, exists, "aws provider factory should be registered")
	assert.NotNil(t, factory)
}

func TestFactory_MissingConfig(t *testing.T) {
	factory, exists := secrets.GetProviderFactory("aws")
	require.True(t, exists, "aws provider factory should be registered")

	cfg := &config.Config{}
	_, err := factory(cfg, zap.NewNop())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "aws provider configuration not found")
}

func TestProvider_ImplementsInterface(t *testing.T) {
	var _ secrets.Provider = (*Provider)(nil)
}
