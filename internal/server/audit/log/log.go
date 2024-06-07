package log

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/internal/server/audit"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	sinkType = "log"
	auditKey = "AUDIT"
)

// Sink is the structure in charge of sending audit events to a specified file location.
type Sink struct {
	logger *zap.Logger
}

type logOptions struct {
	path     string
	encoding config.LogEncoding
}

type Option func(*logOptions)

func WithPath(path string) Option {
	return func(o *logOptions) {
		o.path = path
	}
}

func WithEncoding(encoding config.LogEncoding) Option {
	return func(o *logOptions) {
		o.encoding = encoding
	}
}

// NewSink is the constructor for a Sink.
func NewSink(opts ...Option) (audit.Sink, error) {
	oo := &logOptions{}

	for _, o := range opts {
		o(oo)
	}

	return newSink(*oo)
}

var (
	consoleEncoderConfig = zapcore.EncoderConfig{
		TimeKey:        "T",
		LevelKey:       zapcore.OmitKey,
		NameKey:        "N",
		CallerKey:      zapcore.OmitKey,
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "M",
		StacktraceKey:  zapcore.OmitKey,
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalColorLevelEncoder,
		EncodeTime:     zapcore.RFC3339TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	fileEncoderConfig = zapcore.EncoderConfig{
		TimeKey:        zapcore.OmitKey,
		LevelKey:       zapcore.OmitKey,
		NameKey:        zapcore.OmitKey,
		CallerKey:      zapcore.OmitKey,
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     zapcore.OmitKey,
		StacktraceKey:  zapcore.OmitKey,
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalColorLevelEncoder,
		EncodeTime:     zapcore.RFC3339TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
)

// newSink is the constructor for a Sink visible for testing.
func newSink(opts logOptions) (audit.Sink, error) {
	cfg := zap.Config{
		Level:            zap.NewAtomicLevelAt(zapcore.InfoLevel),
		Development:      false,
		Encoding:         "console",
		EncoderConfig:    consoleEncoderConfig,
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}

	if opts.encoding != "" {
		cfg.Encoding = string(opts.encoding)
	}

	if opts.path != "" {
		// check path dir exists, if not create it
		dir := filepath.Dir(opts.path)
		if _, err := os.Stat(dir); err != nil {
			if !os.IsNotExist(err) {
				return nil, fmt.Errorf("checking log file path: %w", err)
			}

			if err := os.MkdirAll(dir, 0755); err != nil {
				return nil, fmt.Errorf("creating log file path: %w", err)
			}
		}

		cfg.OutputPaths = []string{opts.path}
		cfg.EncoderConfig = fileEncoderConfig
	}

	return &Sink{
		logger: zap.Must(cfg.Build()),
	}, nil
}

func (l *Sink) SendAudits(_ context.Context, events []audit.Event) error {
	for _, e := range events {
		l.logger.Info(auditKey, zap.Inline(e))
	}

	return nil
}

func (l *Sink) Close() error {
	// see: https://github.com/uber-go/zap/issues/991
	//nolint:errcheck
	_ = l.logger.Sync()
	return nil
}

func (l *Sink) String() string {
	return sinkType
}
