package kafka

import (
	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/zap"
)

type klogger struct {
	logger *zap.Logger
}

func (l *klogger) Level() kgo.LogLevel {
	switch l.logger.Level() {
	case zap.InfoLevel:
		return kgo.LogLevelInfo
	case zap.WarnLevel:
		return kgo.LogLevelWarn
	case zap.ErrorLevel:
		return kgo.LogLevelError
	default:
		return kgo.LogLevelNone
	}
}

func (l *klogger) Log(level kgo.LogLevel, msg string, keyvals ...any) {
	if level == kgo.LogLevelDebug {
		// debug has too much information about each operation
		return
	}
	var fields []zap.Field
	for i := 0; i < len(keyvals); i += 2 {
		if i+1 < len(keyvals) {
			key, ok := keyvals[i].(string)
			if !ok {
				key = "unknown"
			}
			fields = append(fields, zap.Any(key, keyvals[i+1]))
		} else {
			fields = append(fields, zap.Any("unknown", keyvals[i]))
		}
	}
	switch level {
	case kgo.LogLevelInfo:
		l.logger.Info(msg, fields...)
	case kgo.LogLevelWarn:
		l.logger.Warn(msg, fields...)
	case kgo.LogLevelError:
		l.logger.Error(msg, fields...)
	}
}
