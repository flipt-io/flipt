package cmd

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/internal/info"
	"go.uber.org/zap/zaptest"
)

func TestNewGRPCServer(t *testing.T) {
	tmp := t.TempDir()
	cfg := &config.Config{}
	cfg.Database.URL = fmt.Sprintf("file:%s", filepath.Join(tmp, "flipt.db"))
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	s, err := NewGRPCServer(ctx, zaptest.NewLogger(t), cfg, info.Flipt{}, false)
	require.NoError(t, err)
	t.Cleanup(func() {
		err := s.Shutdown(ctx)
		assert.NoError(t, err)
	})
	assert.NotEmpty(t, s.Server.GetServiceInfo())
}
