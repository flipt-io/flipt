package redis

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestZapRedisLogger(t *testing.T) {
	core, recorded := observer.New(zapcore.DebugLevel)
	logger := zap.New(core)

	l := &zapRedisLogger{logger: logger}
	l.Printf(t.Context(), "hello %s", "world")

	require.Equal(t, 1, recorded.Len())
	entry := recorded.All()[0]
	assert.Equal(t, zapcore.DebugLevel, entry.Level)
	assert.Equal(t, "hello world", entry.Message)
}

func TestZapRedisLogger_WithStoreField(t *testing.T) {
	core, recorded := observer.New(zapcore.DebugLevel)
	logger := zap.New(core)

	l := &zapRedisLogger{logger: logger.With(zap.String("store", "redis"))}
	l.Printf(t.Context(), "test")

	require.Equal(t, 1, recorded.Len())
	entry := recorded.All()[0]
	assert.Equal(t, "redis", entry.ContextMap()["store"])
}

func TestSetGlobalRedisLogger(t *testing.T) {
	core, recorded := observer.New(zapcore.DebugLevel)
	logger := zap.New(core)

	require.NotPanics(t, func() { SetGlobalRedisLogger(logger) })

	assert.Empty(t, recorded.All())
}
