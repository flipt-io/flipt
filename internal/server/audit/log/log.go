package log

import (
	"context"

	"go.flipt.io/flipt/internal/server/audit"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const sinkType = "log"

// Sink is the structure in charge of sending audit events to a specified file location.
type Sink struct {
	logger *zap.Logger
}

type logOptions struct {
	Path string
}

type Option func(*logOptions)

func WithPath(path string) Option {
	return func(o *logOptions) {
		o.Path = path
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

// newSink is the constructor for a Sink visible for testing.
func newSink(opts logOptions) (audit.Sink, error) {
	encoding := zapcore.EncoderConfig{
		TimeKey:        "T",
		LevelKey:       "",
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

	cfg := zap.Config{
		Level:            zap.NewAtomicLevelAt(zapcore.InfoLevel),
		Development:      false,
		Encoding:         "json",
		EncoderConfig:    encoding,
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}

	if opts.Path != "" {
		cfg.OutputPaths = []string{opts.Path}
	}

	return &Sink{
		logger: zap.Must(cfg.Build()),
	}, nil
}

func (l *Sink) SendAudits(_ context.Context, events []audit.Event) error {
	for _, e := range events {
		l.logger.Info("audit", zap.Any("event", e))
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
