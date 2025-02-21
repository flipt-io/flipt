package osfs

import (
	"io/fs"
	"os"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/osfs"
	envsfs "go.flipt.io/flipt/internal/storage/environments/fs"
)

// Filesystem is a wrapper around the billy osfs implementation which
// conforms to our own subset interface
type Filesystem struct {
	billy.Filesystem
}

func New(baseDir string) *Filesystem {
	return &Filesystem{osfs.New(baseDir)}
}

// Open opens the named file for reading. If successful, methods on the
// returned file can be used for reading; the associated file descriptor has
// mode O_RDONLY.
func (f *Filesystem) Open(filename string) (envsfs.File, error) {
	return f.OpenFile(filename, os.O_RDONLY, os.ModePerm)
}

// OpenFile is the generalized open call; most users will use Open or Create
// instead. It opens the named file with specified flag (O_RDONLY etc.) and
// perm, (0666 etc.) if applicable. If successful, methods on the returned
// File can be used for I/O.
func (f *Filesystem) OpenFile(filename string, flag int, perm os.FileMode) (envsfs.File, error) {
	fi, err := f.Filesystem.OpenFile(filename, flag, perm)
	if err != nil {
		return nil, err
	}

	stat, err := f.Stat(filename)
	if err != nil {
		return nil, err
	}

	return file{fi, stat}, nil
}

type file struct {
	billy.File
	info fs.FileInfo
}

func (f file) Stat() (fs.FileInfo, error) {
	return f.info, nil
}

// Stat returns a FileInfo describing the named file.
func (f *Filesystem) Stat(filename string) (os.FileInfo, error) {
	return f.Filesystem.Stat(filename)
}

// Remove removes the named file or directory.
func (f *Filesystem) Remove(filename string) error {
	return f.Filesystem.Remove(filename)
}

// ReadDir reads the directory named by dirname and returns a list of
// directory entries sorted by filename.
func (f *Filesystem) ReadDir(path string) ([]os.FileInfo, error) {
	return f.Filesystem.ReadDir(path)
}

// MkdirAll creates a directory named path, along with any necessary
// parents, and returns nil, or else returns an error. The permission bits
// perm are used for all directories that MkdirAll creates. If path is/
// already a directory, MkdirAll does nothing and returns nil.
func (f *Filesystem) MkdirAll(filename string, perm os.FileMode) error {
	return f.Filesystem.MkdirAll(filename, perm)
}
