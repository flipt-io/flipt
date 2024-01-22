package sql

import (
	"io/fs"
	"time"
)

type fileInfo string

func (f fileInfo) Name() string {
	return string(f)
}

func (f fileInfo) Size() int64 {
	return 0
}

func (f fileInfo) Mode() fs.FileMode {
	return fs.ModePerm
}

func (f fileInfo) ModTime() time.Time {
	return time.Now()
}

func (f fileInfo) IsDir() bool {
	return false
}

func (f fileInfo) Sys() any {
	return nil
}
