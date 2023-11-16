package gitfs

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/filemode"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage"
	"go.flipt.io/flipt/internal/containers"
	"go.uber.org/zap"
)

// FS is a filesystem implementation that decorates a git storage
// implementation and instance of a git tree.
// FS implements fs.FS and is therefore read-only by design.
type FS struct {
	logger *zap.Logger

	storage storage.Storer
	tree    *object.Tree
}

// New instantiates and instance of storage which retrieves files from
// the provided storage.Storer implementation and instance of *object.Tree.
func New(logger *zap.Logger, storage storage.Storer, tree *object.Tree) FS {
	return FS{logger, storage, tree}
}

// Options configures call to NewFromRepo.
type Options struct {
	ref plumbing.ReferenceName
}

// WithReference overrides the default reference to main.
func WithReference(ref plumbing.ReferenceName) containers.Option[Options] {
	return func(o *Options) {
		o.ref = ref
	}
}

// NewFromRepo is a convenience utility which constructs an instance of FS
// from the provided git repository.
// By default the returned FS serves the content from the root tree
// for the commit at reference HEAD.
func NewFromRepo(logger *zap.Logger, repo *git.Repository, opts ...containers.Option[Options]) (FS, error) {
	o := Options{ref: plumbing.HEAD}
	containers.ApplyAll(&o, opts...)

	ref, err := repo.Reference(o.ref, true)
	if err != nil {
		return FS{}, fmt.Errorf("resolving reference (%q): %w", o.ref, err)
	}

	return NewFromRepoHash(logger, repo, ref.Hash())
}

// NewFromRepoHash is a convenience utility which constructs an instance of FS
// from the provided git repository and hash string.
func NewFromRepoHashString(logger *zap.Logger, repo *git.Repository, hash string) (FS, error) {
	return NewFromRepoHash(logger, repo, plumbing.NewHash(hash))
}

// NewFromRepoHash is a convenience utility which constructs an instance of FS
// from the provided git repository and hash.
func NewFromRepoHash(logger *zap.Logger, repo *git.Repository, hash plumbing.Hash) (FS, error) {
	commit, err := repo.CommitObject(hash)
	if err != nil {
		return FS{}, fmt.Errorf("fetching commit (%q): %w", hash, err)
	}

	tree, err := repo.TreeObject(commit.TreeHash)
	if err != nil {
		return FS{}, fmt.Errorf("retrieving root tree (%q): %w", commit.TreeHash, err)
	}

	return New(logger.With(zap.Stringer("SHA", hash)), repo.Storer, tree), nil
}

// Open opens the named file.
// When Open returns an error, it should be of type *PathError
// with the Op field set to "open", the Path field set to name,
// and the Err field describing the problem.
//
// Open should reject attempts to open names that do not satisfy
// ValidPath(name), returning a *PathError with Err set to
// ErrInvalid or ErrNotExist.
func (f FS) Open(name string) (_ fs.File, err error) {
	defer func() {
		if err != nil {
			err = opPathError("Open", name)(err)
		}
	}()

	f.logger.Debug("open", zap.String("path", name))

	if !fs.ValidPath(name) {
		return nil, fs.ErrInvalid
	}

	entry := &object.TreeEntry{
		Name: ".",
		Mode: filemode.Dir,
		Hash: f.tree.Hash,
	}

	if name != "." {
		entry, err = f.tree.FindEntry(name)
		if err != nil {
			// adapt from plumbing specific error to fs compatible error.
			if errors.Is(err, object.ErrEntryNotFound) {
				return nil, fs.ErrNotExist
			}

			return nil, fmt.Errorf("finding entry %q: %w", name, err)
		}
	}

	info, err := fileInfoFromEntry(entry)
	if err != nil {
		return nil, fmt.Errorf("deriving info %q: %w", name, err)
	}

	if entry.Mode == filemode.Dir {
		tree, err := object.GetTree(f.storage, entry.Hash)
		if err != nil {
			return nil, fmt.Errorf("getting tree %q: %w", name, err)
		}

		var entries []fs.DirEntry
		for _, entry := range tree.Entries {
			entry := entry
			mode, err := entry.Mode.ToOSFileMode()
			if err != nil {
				return nil, err
			}

			dEntry := DirEntry{
				entry: entry,
				mode:  mode,
			}

			if entry.Mode != filemode.Dir && entry.Mode != filemode.Submodule {
				dEntry.fi, err = f.tree.TreeEntryFile(&entry)
				if err != nil {
					return nil, fmt.Errorf("opening entry file %q: %w", name, err)
				}
			}

			entries = append(entries, dEntry)
		}

		return &Dir{info: info, entries: entries}, nil
	}

	fi, err := f.tree.TreeEntryFile(entry)
	if err != nil {
		return nil, fmt.Errorf("opening entry file %q: %w", name, err)
	}

	rd, err := fi.Reader()
	if err != nil {
		return nil, err
	}

	info.size = fi.Blob.Size

	return &File{
		ReadCloser: rd,
		info:       info,
	}, nil
}

func opPathError(op, path string) func(error) error {
	return func(err error) error {
		return &fs.PathError{
			Op:   op,
			Path: path,
			Err:  err,
		}
	}
}

// File is a representation of a file which can be read.
type File struct {
	io.ReadCloser

	info FileInfo
}

// Seek attempts to seek the embedded read-closer.
// If the embedded read closer implements seek, then it delegates
// to that instances implementation. Alternatively, it returns
// an error signifying that the File cannot be seeked.
func (f *File) Seek(offset int64, whence int) (int64, error) {
	if seek, ok := f.ReadCloser.(io.Seeker); ok {
		return seek.Seek(offset, whence)
	}

	return 0, errors.New("seeker cannot seek")
}

func (f *File) Stat() (fs.FileInfo, error) {
	return f.info, nil
}

// Dir is a representation of a directory found within a filesystem.
// The contents of the directory can be listed and accessed via
// the ReadDir method.
// It is an implementation of fs.File.
type Dir struct {
	info    fs.FileInfo
	entries []fs.DirEntry
	idx     int
}

func (d *Dir) Stat() (fs.FileInfo, error) {
	return d.info, nil
}

func (d *Dir) Read(p []byte) (n int, err error) {
	return 0, io.EOF
}

func (d *Dir) Close() error {
	return nil
}

// ReadDir reads the contents of the directory and returns
// a slice of up to n DirEntry values in directory order.
// Subsequent calls on the same file will yield further DirEntry values.
//
// If n > 0, ReadDir returns at most n DirEntry structures.
// In this case, if ReadDir returns an empty slice, it will return
// a non-nil error explaining why.
// At the end of a directory, the error is io.EOF.
// (ReadDir must return io.EOF itself, not an error wrapping io.EOF.)
//
// If n <= 0, ReadDir returns all the DirEntry values from the directory
// in a single slice. In this case, if ReadDir succeeds (reads all the way
// to the end of the directory), it returns the slice and a nil error.
// If it encounters an error before the end of the directory,
// ReadDir returns the DirEntry list read until that point and a non-nil error.
func (d *Dir) ReadDir(n int) (dst []fs.DirEntry, err error) {
	if n <= 0 {
		return d.entries, nil
	}

	l := len(d.entries[d.idx:])
	if l <= 0 {
		return nil, io.EOF
	}

	if l >= n {
		l = n
	}

	start := d.idx

	d.idx += l

	return d.entries[start:d.idx], nil
}

// FileInfo contains metadata about a file including its
// name, size, mode and last modified timestamp.
type FileInfo struct {
	name string
	size int64
	mode fs.FileMode
	mod  time.Time
}

func fileInfoFromEntry(entry *object.TreeEntry) (FileInfo, error) {
	mode, err := entry.Mode.ToOSFileMode()
	if err != nil {
		return FileInfo{}, err
	}

	return FileInfo{
		name: entry.Name,
		mode: mode,
	}, nil
}

func (f FileInfo) Name() string {
	return f.name
}

func (f FileInfo) Size() int64 {
	return f.size
}

func (f FileInfo) Mode() fs.FileMode {
	return f.mode
}

func (f FileInfo) ModTime() time.Time {
	return f.mod
}

func (f FileInfo) IsDir() bool {
	return f.mode.IsDir()
}

func (f FileInfo) Sys() any {
	return nil
}

// DirEntry represents an entry within a file system directory.
// The entry could itself be a directory or another type of file.
type DirEntry struct {
	entry object.TreeEntry
	fi    *object.File
	mode  fs.FileMode
}

// Name returns the name of the file (or subdirectory) described by the entry.
// This name is only the final element of the path (the base name), not the entire path.
// For example, Name would return "hello.go" not "home/gopher/hello.go".
func (d DirEntry) Name() string {
	return d.entry.Name
}

// IsDir reports whether the entry describes a directory.
func (d DirEntry) IsDir() bool {
	return d.entry.Mode == filemode.Dir
}

// Type returns the type bits for the entry.
// The type bits are a subset of the usual FileMode bits, those returned by the FileMode.Type method.
func (d DirEntry) Type() fs.FileMode {
	return d.mode
}

// Info returns the FileInfo for the file or subdirectory described by the entry.
// The returned FileInfo may be from the time of the original directory read
// or from the time of the call to Info. If the file has been removed or renamed
// since the directory read, Info may return an error satisfying errors.Is(err, ErrNotExist).
// If the entry denotes a symbolic link, Info reports the information about the link itself,
// not the link's target.
func (d DirEntry) Info() (fs.FileInfo, error) {
	info := FileInfo{
		name: d.entry.Name,
		mode: d.mode,
	}

	if d.entry.Mode != filemode.Dir && d.fi != nil {
		info.size = d.fi.Blob.Size
	}

	return info, nil
}
