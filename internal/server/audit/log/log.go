package log

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/hashicorp/go-multierror"
	"go.flipt.io/flipt/internal/server/audit"
	"go.uber.org/zap"
)

// filesystem is an interface that abstracts the filesystem operations used by the Sink.
type filesystem interface {
	OpenFile(name string, flag int, perm os.FileMode) (file, error)
	Stat(name string) (os.FileInfo, error)
	MkdirAll(path string, perm os.FileMode) error
}

// file is an interface that abstracts the file operations used by the Sink.
type file interface {
	io.WriteCloser
	Name() string
}

// osFS implements fileSystem using the local disk.
type osFS struct{}

func (osFS) OpenFile(name string, flag int, perm os.FileMode) (file, error) {
	return os.OpenFile(name, flag, perm)
}

func (osFS) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}

func (osFS) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

const sinkType = "log"

// Sink is the structure in charge of sending audit events to a specified file location.
type Sink struct {
	logger *zap.Logger
	f      file
	mtx    sync.Mutex
	enc    *json.Encoder
}

// NewSink is the constructor for a Sink.
func NewSink(logger *zap.Logger, path string) (audit.Sink, error) {
	return newSink(logger, path, osFS{})
}

// newSink is the constructor for a Sink visible for testing.
func newSink(logger *zap.Logger, path string, fs filesystem) (audit.Sink, error) {
	var f file = os.Stdout

	if path != "" {
		// check if path exists, if not create it
		dir := filepath.Dir(path)
		if _, err := fs.Stat(dir); err != nil {
			if !os.IsNotExist(err) {
				return nil, fmt.Errorf("checking log directory: %w", err)
			}

			if err := fs.MkdirAll(dir, 0755); err != nil {
				return nil, fmt.Errorf("creating log directory: %w", err)
			}
		}

		var err error
		f, err = fs.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
		if err != nil {
			return nil, fmt.Errorf("opening log file: %w", err)
		}
	}

	return &Sink{
		logger: logger,
		f:      f,
		enc:    json.NewEncoder(f),
	}, nil
}

func (l *Sink) SendAudits(ctx context.Context, events []audit.Event) error {
	l.mtx.Lock()
	defer l.mtx.Unlock()
	var result error

	for _, e := range events {
		err := l.enc.Encode(e)
		if err != nil {
			l.logger.Error("failed to write audit event", zap.String("file", l.f.Name()), zap.Error(err))
			result = multierror.Append(result, err)
		}
	}

	return result
}

func (l *Sink) Close() error {
	l.mtx.Lock()
	defer l.mtx.Unlock()
	return l.f.Close()
}

func (l *Sink) String() string {
	return sinkType
}
