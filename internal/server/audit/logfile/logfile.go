package logfile

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/hashicorp/go-multierror"
	"go.flipt.io/flipt/internal/server/audit"
	"go.uber.org/zap"
)

const sinkType = "logfile"

// Sink is the structure in charge of sending Audits to a specified file location.
type Sink struct {
	logger *zap.Logger
	file   *os.File
	mtx    sync.Mutex
	enc    *json.Encoder
}

// NewSink is the constructor for a Sink.
func NewSink(logger *zap.Logger, path string) (audit.Sink, error) {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		return nil, fmt.Errorf("opening log file: %w", err)
	}

	return &Sink{
		logger: logger,
		file:   file,
		enc:    json.NewEncoder(file),
	}, nil
}

func (l *Sink) SendAudits(ctx context.Context, events []audit.Event) error {
	l.mtx.Lock()
	defer l.mtx.Unlock()
	var result error

	for _, e := range events {
		err := l.enc.Encode(e)
		if err != nil {
			l.logger.Error("failed to write audit event to file", zap.String("file", l.file.Name()), zap.Error(err))
			result = multierror.Append(result, err)
		}
	}

	return result
}

func (l *Sink) Close() error {
	l.mtx.Lock()
	defer l.mtx.Unlock()
	return l.file.Close()
}

func (l *Sink) String() string {
	return sinkType
}
