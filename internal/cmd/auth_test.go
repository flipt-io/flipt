package cmd

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/config"
	"go.uber.org/zap/zaptest"
)

func TestAuthenticationGRPC(t *testing.T) {
	var (
		logger = zaptest.NewLogger(t)
		ctx    = context.Background()
	)

	t.Run("auth disabled", func(t *testing.T) {
		cfg := config.Default()

		registers, interceptor, _, err := authenticationGRPC(ctx, logger, cfg, false, false)
		require.NoError(t, err)
		require.NotNil(t, registers)
		assert.Nil(t, interceptor)
	})

	t.Run("auth enabled", func(t *testing.T) {
		cfg := config.Default()
		cfg.Authentication.Required = true
		registers, interceptor, _, err := authenticationGRPC(ctx, logger, cfg, false, false)
		require.NoError(t, err)
		require.NotNil(t, interceptor)
		require.NotNil(t, registers)
	})
}
