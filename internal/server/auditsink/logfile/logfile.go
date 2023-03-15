package logfile

import (
	"os"

	"go.flipt.io/flipt/internal/server/auditsink"
	"go.uber.org/zap"
)

const sinkType = "logfile"

// LogFileSink is the structure in charge of sending Audits
// to a specified file location.
type LogFileSink struct {
	logger *zap.Logger
	file   *os.File
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
	return nil
}

func (l *LogFileSink) String() string {
	return sinkType
}
