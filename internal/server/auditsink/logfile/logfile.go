package logfile

import (
	"os"

	"go.flipt.io/flipt/internal/server/auditsink"
)

const sinkType = "logfile"

// LogFileSink is the structure in charge of sending Audits
// to a specified file location.
type LogFileSink struct {
	File *os.File
}

// NewLogFileSink is the constructor for a LogFileSink.
func NewLogFileSink(filePath string) (*LogFileSink, error) {
	file, err := os.Create(filePath)
	if err != nil {
		return nil, err
	}

	return &LogFileSink{
		File: file,
	}, nil
}

func (l *LogFileSink) SendAudits(auditEvents []*auditsink.AuditEvent) error {
	return nil
}

func (l *LogFileSink) String() string {
	return sinkType
}
