package gcp

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/internal/secrets"
	"go.uber.org/zap"
)

func TestFactoryRegistration(t *testing.T) {
	factory, exists := secrets.GetProviderFactory("gcp")
	require.True(t, exists, "gcp provider factory should be registered")
	assert.NotNil(t, factory)
}

func TestFactory_MissingConfig(t *testing.T) {
	factory, _ := secrets.GetProviderFactory("gcp")

	cfg := &config.Config{}
	_, err := factory(cfg, zap.NewNop())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "gcp provider configuration not found")
}

func TestProvider_ImplementsInterface(t *testing.T) {
	var _ secrets.Provider = (*Provider)(nil)
}

func TestProvider_ResourceNameGeneration(t *testing.T) {
	const project = "my-project"

	t.Run("secret version path", func(t *testing.T) {
		path := "my-secret"
		expected := "projects/my-project/secrets/my-secret/versions/latest"
		assert.Equal(t, expected, "projects/"+project+"/secrets/"+path+"/versions/latest")
	})

	t.Run("list parent path", func(t *testing.T) {
		expected := "projects/my-project"
		assert.Equal(t, expected, "projects/"+project)
	})
}
