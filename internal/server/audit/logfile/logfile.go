package logfile

import (
	"encoding/json"
	"os"
	"sync"

	"github.com/hashicorp/go-multierror"
	"go.flipt.io/flipt/internal/server/audit"
	"go.uber.org/zap"
)

const sinkType = "logfile"

// LogFileSink is the structure in charge of sending Audits to a specified file location.
type LogFileSink struct {
	logger *zap.Logger
	file   *os.File
	mtx    sync.Mutex
	enc    *json.Encoder
}

// NewSink is the constructor for a LogFileSink.
func NewSink(logger *zap.Logger, path string) (audit.Sink, error) {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}

	return &LogFileSink{
		logger: logger,
		file:   file,
		enc:    json.NewEncoder(file),
	}, nil
}

func (l *LogFileSink) SendAudits(events []audit.Event) error {
	l.logger.Debug("performing batched sending of audits", zap.Stringer("sink", l))

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

func (l *LogFileSink) Close() error {
	l.mtx.Lock()
	defer l.mtx.Unlock()
	return l.file.Close()
}

func (l *LogFileSink) String() string {
	return sinkType
}
