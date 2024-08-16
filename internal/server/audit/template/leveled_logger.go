package template

import (
	"github.com/hashicorp/go-retryablehttp"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewLeveledLogger(logger *zap.Logger) retryablehttp.LeveledLogger {
	return &LeveledLogger{logger}
}

type LeveledLogger struct {
	logger *zap.Logger
}

func (l *LeveledLogger) Error(msg string, keyvals ...interface{}) {
	if l.logger.Core().Enabled(zapcore.ErrorLevel) {
		l.logger.Error(msg, l.fields(keyvals)...)
	}
}

func (l *LeveledLogger) Info(msg string, keyvals ...interface{}) {
	if l.logger.Core().Enabled(zapcore.InfoLevel) {
		l.logger.Info(msg, l.fields(keyvals)...)
	}
}
func (l *LeveledLogger) Debug(msg string, keyvals ...interface{}) {
	if l.logger.Core().Enabled(zapcore.DebugLevel) {
		l.logger.Debug(msg, l.fields(keyvals)...)
	}
}

func (l *LeveledLogger) Warn(msg string, keyvals ...any) {
	if l.logger.Core().Enabled(zapcore.WarnLevel) {
		l.logger.Warn(msg, l.fields(keyvals)...)
	}
}

func (l *LeveledLogger) fields(keyvals []any) []zap.Field {
	fields := make([]zap.Field, 0, len(keyvals)/2)
	for i := 0; i < len(keyvals); i += 2 {
		k, v := keyvals[i], keyvals[i+1]
		fields = append(fields, zap.Any(k.(string), v))
	}
	return fields
}
