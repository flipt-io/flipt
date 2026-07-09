package redis

import (
	"context"

	goredis "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// zapRedisLogger adapts go-redis internal logs to zap.
type zapRedisLogger struct {
	logger *zap.Logger
}

// Printf adapts a go-redis log line to zap.
// The context argument is not used.
func (l *zapRedisLogger) Printf(_ context.Context, format string, v ...any) {
	l.logger.Sugar().Debugf(format, v...)
}

// SetGlobalRedisLogger adapts go-redis internal logging to the provided zap logger.
func SetGlobalRedisLogger(logger *zap.Logger) {
	goredis.SetLogger(&zapRedisLogger{logger: logger.With(zap.String("store", "redis"))})
}
