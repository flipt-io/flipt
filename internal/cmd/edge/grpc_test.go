package edge

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/internal/info"
	"go.uber.org/zap/zaptest"
)

func TestNewGRPCServer(t *testing.T) {
	cfg := &config.Config{
		Storage: config.StorageConfig{
			Type: config.LocalStorageType,
			Local: &config.StorageLocalConfig{
				Path: ".",
			},
		},
	}
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	s, err := NewGRPCServer(ctx, zaptest.NewLogger(t), cfg, info.Flipt{})
	require.NoError(t, err)
	t.Cleanup(func() {
		err := s.Shutdown(ctx)
		assert.NoError(t, err)
	})
	assert.NotEmpty(t, s.Server.GetServiceInfo())
}
