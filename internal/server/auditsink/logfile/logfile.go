package logfile

import (
	"encoding/json"
	"os"
	"sync"

	"go.flipt.io/flipt/internal/server/auditsink"
	"go.uber.org/zap"
)

const sinkType = "logfile"

// LogFileSink is the structure in charge of sending Audits to a specified file location.
type LogFileSink struct {
	logger *zap.Logger
	file   *os.File
	mtx    sync.Mutex
}

// NewLogFileSink is the constructor for a LogFileSink.
func NewLogFileSink(logger *zap.Logger, filePath string) (*LogFileSink, error) {
	file, err := os.Create(filePath)
	if err != nil {
		return nil, err
	}

	lfs := &LogFileSink{
		logger: logger,
		file:   file,
	}

	return lfs, nil
}

func (l *LogFileSink) SendAudits(auditEvents []auditsink.AuditEvent) error {
	l.logger.Debug("performing batched sending of audits", zap.Stringer("sink", l))
	l.mtx.Lock()
	defer l.mtx.Unlock()
	for _, ae := range auditEvents {
		b, err := json.Marshal(ae)
		if err != nil {
			l.logger.Error("failed to encode audit event", zap.Stringer("sink", l))
			continue
		}
		_, err = l.file.Write(append(b, '\n'))
		if err != nil {
			l.logger.Error("failed to write audit event to file", zap.String("file", l.file.Name()))
			continue
		}
	}
	return nil
}

func (l *LogFileSink) String() string {
	return sinkType
}
