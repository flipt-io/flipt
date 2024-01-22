package object

import (
	"io"
	"io/fs"
	"time"
)

type File struct {
	key          string
	length       int64
	body         io.ReadCloser
	lastModified time.Time
}

// ensure File implements the fs.File interface
var _ fs.File = &File{}

func (f *File) Stat() (fs.FileInfo, error) {
	return &FileInfo{
		name:    f.key,
		size:    f.length,
		modTime: f.lastModified,
	}, nil
}

func (f *File) Read(p []byte) (int, error) {
	return f.body.Read(p)
}

func (f *File) Close() error {
	return f.body.Close()
}

func NewFile(key string, length int64, body io.ReadCloser, lastModified time.Time) *File {
	return &File{
		key:          key,
		length:       length,
		body:         body,
		lastModified: lastModified,
	}
}
