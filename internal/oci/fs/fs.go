package fs

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"time"

	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	storagefs "go.flipt.io/flipt/internal/storage/fs"
	"gopkg.in/yaml.v3"
	"oras.land/oras-go/v2/content"
)

type FS struct {
	fetcher  content.Fetcher
	manifest v1.Manifest
	layers   map[string]v1.Descriptor
}

func New(fetcher content.Fetcher, manifest v1.Manifest) *FS {
	fs := &FS{
		fetcher:  fetcher,
		manifest: manifest,
		layers:   map[string]v1.Descriptor{},
	}

	for _, layer := range manifest.Layers {
		fs.layers[fmt.Sprintf("%s.json", layer.Digest.Hex())] = layer
	}

	return fs
}

// Open opens the named file.
// When Open returns an error, it should be of type *PathError
// with the Op field set to "open", the Path field set to name,
// and the Err field describing the problem.
//
// Open should reject attempts to open names that do not satisfy
// ValidPath(name), returning a *PathError with Err set to
// ErrInvalid or ErrNotExist.
func (f *FS) Open(name string) (_ fs.File, err error) {
	defer func() {
		if err != nil {
			err = opPathError("Open", name)(err)
		}
	}()

	if !fs.ValidPath(name) {
		return nil, fs.ErrInvalid
	}

	created, err := time.Parse(time.RFC3339, f.manifest.Annotations[v1.AnnotationCreated])
	if err != nil {
		return nil, err
	}

	switch name {
	case ".":
		var entries []fs.DirEntry
		for _, layer := range f.layers {
			entries = append(entries, &DirEntry{
				desc: layer,
				mod:  created,
			})
		}

		return &Dir{
			info: FileInfo{
				name: ".",
				mode: fs.ModeDir,
				mod:  created,
			},
			entries: entries,
		}, nil
	case ".flipt.yml":
		index := storagefs.FliptIndex{
			Version: "1.0",
		}

		for layer := range f.layers {
			index.Include = append(index.Include, layer)
		}

		buf := &bytes.Buffer{}
		if err := yaml.NewEncoder(buf).Encode(&index); err != nil {
			return nil, err
		}

		return &File{
			ReadCloser: io.NopCloser(buf),
			info: FileInfo{
				name: ".flipt.yml",
				mod:  created,
				mode: fs.ModePerm,
				size: int64(buf.Len()),
			},
		}, nil
	}

	layer, ok := f.layers[name]
	if !ok {
		return nil, fs.ErrNotExist
	}

	rc, err := f.fetcher.Fetch(context.Background(), layer)
	if err != nil {
		return nil, err
	}

	return &File{
		ReadCloser: rc,
		info: FileInfo{
			name: name,
			mod:  created,
			mode: fs.ModePerm,
			size: layer.Size,
		},
	}, nil
}

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
	return &f.info, nil
}

type Dir struct {
	info    FileInfo
	entries []fs.DirEntry
	idx     int
}

func (d *Dir) Stat() (fs.FileInfo, error) {
	return d.info, nil
}

func (d *Dir) Read(_ []byte) (int, error) {
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

type FileInfo struct {
	name string
	size int64
	mode fs.FileMode
	mod  time.Time
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

type DirEntry struct {
	desc v1.Descriptor
	mod  time.Time
}

// Name returns the name of the file (or subdirectory) described by the entry.
// This name is only the final element of the path (the base name), not the entire path.
// For example, Name would return "hello.go" not "home/gopher/hello.go".
func (d *DirEntry) Name() string {
	return d.desc.Digest.Hex() + ".json"
}

// IsDir reports whether the entry describes a directory.
func (d *DirEntry) IsDir() bool {
	return false
}

// Type returns the type bits for the entry.
// The type bits are a subset of the usual FileMode bits, those returned by the FileMode.Type method.
func (d *DirEntry) Type() fs.FileMode {
	return fs.ModePerm
}

// Info returns the FileInfo for the file or subdirectory described by the entry.
// The returned FileInfo may be from the time of the original directory read
// or from the time of the call to Info. If the file has been removed or renamed
// since the directory read, Info may return an error satisfying errors.Is(err, ErrNotExist).
// If the entry denotes a symbolic link, Info reports the information about the link itself,
// not the link's target.
func (d *DirEntry) Info() (fs.FileInfo, error) {
	return FileInfo{
		name: d.Name(),
		mode: fs.ModePerm,
		size: d.desc.Size,
		mod:  d.mod,
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
