package fs

import (
	"net/http"
	"os"
	"time"
)

// ModTimeFS is an http.FileSystem wrapper that modifies
// underlying fs such that all of its file mod times are set to zero.
type ModTimeFS struct {
	fs http.FileSystem
}

// NewModTimeFS creates a ModTimeFS
func NewModTimeFS(fs http.FileSystem) ModTimeFS {
	return ModTimeFS{fs: fs}
}

func (m ModTimeFS) Open(name string) (http.File, error) {
	f, err := m.fs.Open(name)
	if err != nil {
		return nil, err
	}
	return modTimeFile{f}, nil
}

type modTimeFile struct {
	http.File
}

func (m modTimeFile) Stat() (os.FileInfo, error) {
	fi, err := m.File.Stat()
	if err != nil {
		return nil, err
	}
	return modTimeFileInfo{fi}, nil
}

type modTimeFileInfo struct {
	os.FileInfo
}

func (modTimeFileInfo) ModTime() time.Time {
	return time.Time{}
}
