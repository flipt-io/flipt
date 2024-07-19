package object

import (
	"io/fs"
	"time"
)

// ensure FileInfo implements fs.FileInfo
var _ fs.FileInfo = &FileInfo{}

// ensure FileInfo implements fs.DirEntry
var _ fs.DirEntry = &FileInfo{}

type FileInfo struct {
	name    string
	size    int64
	modTime time.Time
	isDir   bool
	etag    string
}

func (fi *FileInfo) Name() string {
	return fi.name
}

func (fi *FileInfo) Size() int64 {
	return fi.size
}

func (fi *FileInfo) Type() fs.FileMode {
	return 0
}

func (fi *FileInfo) Mode() fs.FileMode {
	return fs.ModePerm
}

func (fi *FileInfo) ModTime() time.Time {
	return fi.modTime
}

func (fi *FileInfo) IsDir() bool {
	return fi.isDir
}
func (fi *FileInfo) SetDir(v bool) {
	fi.isDir = v
}

func (fi *FileInfo) Sys() any {
	return nil
}

func (fi *FileInfo) Info() (fs.FileInfo, error) {
	return fi, nil
}

func (fi *FileInfo) Etag() string {
	return fi.etag
}
