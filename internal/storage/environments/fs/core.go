package fs

import (
	"context"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"go.flipt.io/flipt/errors"
	"go.flipt.io/flipt/internal/server/environments"
	rpcenvs "go.flipt.io/flipt/rpc/v2/environments"
	"go.uber.org/zap"
)

// ResourceStorage is an interface that exposes get, list, put and delete
// operations for configuration resources stored on filesystem mediums
// It should handle reading and writing resources through the Filesystem abstraction
type ResourceStorage interface {
	ResourceType() environments.ResourceType

	GetResource(_ context.Context, _ Filesystem, namespace, key string) (*rpcenvs.Resource, error)
	ListResources(_ context.Context, _ Filesystem, namespace string) ([]*rpcenvs.Resource, error)
	PutResource(context.Context, Filesystem, *rpcenvs.Resource) error
	DeleteResource(_ context.Context, _ Filesystem, namespace, key string) error
}

// Filesystem is a subset of a filesystem API surface area
// Our storage engine requires the ability to open regular files
// and directories for reading and writing both metadata and contents
type Filesystem interface {
	// OpenFile is the generalized open call; most users will use Open or Create
	// instead. It opens the named file with specified flag (O_RDONLY etc.) and
	// perm, (0666 etc.) if applicable. If successful, methods on the returned
	// File can be used for I/O.
	OpenFile(filename string, flag int, perm os.FileMode) (File, error)
	// Stat returns a FileInfo describing the named file.
	Stat(filename string) (os.FileInfo, error)
	// Remove removes the named file or directory.
	Remove(filename string) error

	// ReadDir reads the directory named by dirname and returns a list of
	// directory entries sorted by filename.
	ReadDir(path string) ([]os.FileInfo, error)
	// MkdirAll creates a directory named path, along with any necessary
	// parents, and returns nil, or else returns an error. The permission bits
	// perm are used for all directories that MkdirAll creates. If path is/
	// already a directory, MkdirAll does nothing and returns nil.
	MkdirAll(filename string, perm os.FileMode) error
}

type File interface {
	Stat() (fs.FileInfo, error)
	io.Writer
	io.Reader
	io.Closer
}

func ToFS(fs Filesystem) fs.FS {
	return fsAdaptor{fs}
}

type fsAdaptor struct {
	Filesystem
}

func (i fsAdaptor) Open(name string) (fs.File, error) {
	return i.Filesystem.OpenFile(name, os.O_RDONLY, os.ModePerm)
}

// Storage is a combined NamespaceStorage and sub-resource storage implementation
// ResourceFSStorage types can be retrieved by type name via the Resource method
type Storage struct {
	*NamespaceStorage
	resources map[environments.ResourceType]ResourceStorage
}

// NewStorage takes namespace storage and a set of type identified ResourceStorage implementations
// and returns an Storage implementation
func NewStorage(logger *zap.Logger, rs ...ResourceStorage) Storage {
	resources := map[environments.ResourceType]ResourceStorage{}
	for _, r := range rs {
		resources[r.ResourceType()] = r
	}

	return Storage{
		NamespaceStorage: NewNamespaceStorage(logger),
		resources:        resources,
	}
}

// Resource fetches the underlying ResourceFSStorage implementation identified by type name
func (fs Storage) Resource(typ environments.ResourceType) (ResourceStorage, error) {
	store, ok := fs.resources[typ]
	if !ok {
		return nil, errors.ErrNotFoundf("resource type %q", typ)
	}

	return store, nil
}

type subFilesystem struct {
	fs  Filesystem
	dir string
}

// SubFilesystem decorates the provided filesystem with one which scopes
// all provided paths with a parent directory before delegating.
func SubFilesystem(fs Filesystem, dir string) Filesystem {
	if dir == "" {
		return fs
	}

	return &subFilesystem{fs, dir}
}

// OpenFile is the generalized open call; most users will use Open or Create
// instead. It opens the named file with specified flag (O_RDONLY etc.) and
// perm, (0666 etc.) if applicable. If successful, methods on the returned
// File can be used for I/O.
func (s *subFilesystem) OpenFile(filename string, flag int, perm os.FileMode) (File, error) {
	return s.fs.OpenFile(filepath.Join(s.dir, filename), flag, perm)
}

// Stat returns a FileInfo describing the named file.
func (s *subFilesystem) Stat(filename string) (os.FileInfo, error) {
	return s.fs.Stat(filepath.Join(s.dir, filename))
}

// Remove removes the named file or directory.
func (s *subFilesystem) Remove(filename string) error {
	return s.fs.Remove(filepath.Join(s.dir, filename))
}

// ReadDir reads the directory named by dirname and returns a list of
// directory entries sorted by filename.
func (s *subFilesystem) ReadDir(path string) ([]os.FileInfo, error) {
	return s.fs.ReadDir(filepath.Join(s.dir, path))
}

// MkdirAll creates a directory named path, along with any necessary
// parents, and returns nil, or else returns an error. The permission bits
// perm are used for all directories that MkdirAll creates. If path is/
// already a directory, MkdirAll does nothing and returns nil.
func (s *subFilesystem) MkdirAll(filename string, perm os.FileMode) error {
	return s.fs.MkdirAll(filepath.Join(s.dir, filename), perm)
}
