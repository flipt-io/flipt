package kafka

import (
	"bytes"
	"testing"

	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"
	"gotest.tools/v3/assert"
)

func TestLogLevel(t *testing.T) {
	tests := []struct {
		input    zapcore.Level
		expected kgo.LogLevel
	}{
		{zapcore.DebugLevel, kgo.LogLevelNone},
		{zapcore.InfoLevel, kgo.LogLevelInfo},
		{zapcore.WarnLevel, kgo.LogLevelWarn},
		{zapcore.ErrorLevel, kgo.LogLevelError},
	}
	for _, tt := range tests {
		t.Run(tt.input.String(), func(t *testing.T) {
			l := &klogger{
				logger: zaptest.NewLogger(t, zaptest.Level(tt.input)),
			}
			assert.Equal(t, tt.expected, l.Level())
		})
	}
}

func TestLog(t *testing.T) {
	b := bytes.Buffer{}
	enc := zap.NewProductionEncoderConfig()
	enc.TimeKey = zapcore.OmitKey
	c := zapcore.NewCore(
		zapcore.NewConsoleEncoder(enc),
		zap.CombineWriteSyncers(zapcore.AddSync(&b)),
		zapcore.DebugLevel,
	)
	l := &klogger{
		logger: zaptest.NewLogger(t, zaptest.Level(zapcore.DebugLevel), zaptest.WrapOptions(
			zap.WrapCore(func(zapcore.Core) zapcore.Core {
				return c
			}))),
	}
	l.Log(kgo.LogLevelDebug, "this is a debug", "level", "debug")
	l.Log(kgo.LogLevelInfo, "this is a info")
	l.Log(kgo.LogLevelWarn, "this is a warn", "level", "warn")
	l.Log(kgo.LogLevelError, "this is a error", "level", "error", "some")
	expected := `info	this is a info
warn	this is a warn	{"level": "warn"}
error	this is a error	{"level": "error", "unknown": "some"}
`
	assert.Equal(t, expected, b.String())
}
