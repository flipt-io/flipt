package blob

import (
	"io"
	"io/fs"
)

type Dir struct {
	*FileInfo
}

// ensure Dir implements fs.FileInfo
var _ fs.FileInfo = &Dir{}

func (d *Dir) Stat() (fs.FileInfo, error) {
	return d.FileInfo, nil
}

func (d *Dir) Read(p []byte) (n int, err error) {
	return 0, io.EOF
}

func (d *Dir) Close() error {
	return nil
}

func (d *Dir) IsDir() bool {
	return true
}

func (d *Dir) Mode() fs.FileMode {
	return fs.ModeDir
}

func NewDir(fileInfo *FileInfo) *Dir {
	return &Dir{fileInfo}
}
