package testing

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"go.flipt.io/flipt/internal/containers"
	"go.flipt.io/flipt/internal/storage/environments/fs"
)

type Filesystem struct {
	t    *testing.T
	data map[string]any
}

func NewFilesystem(t *testing.T, opts ...containers.Option[Filesystem]) *Filesystem {
	fs := &Filesystem{
		t:    t,
		data: map[string]any{},
	}

	containers.ApplyAll(fs, opts...)

	return fs
}

func WithFile(name, contents string) containers.Option[Filesystem] {
	return func(f *Filesystem) {
		base, name := path.Split(name)
		dir, err := f.mkdirAll(base)
		if err != nil {
			panic(err)
		}

		buf := []byte(contents)
		dir[name] = &File{
			buf:    buf,
			Reader: bytes.NewBuffer(buf),
			info: FileInfo{
				name: name,
				size: int64(len(contents)),
				mode: os.ModePerm,
			},
		}
	}
}

func WithDirectory(name string, opts ...containers.Option[Filesystem]) containers.Option[Filesystem] {
	return func(f *Filesystem) {
		dir, err := f.mkdirAll(name)
		if err != nil {
			panic(err)
		}

		// recursively apply options with a subtree rooted at our new directory
		containers.ApplyAll(&Filesystem{data: dir}, opts...)
	}
}

// OpenFile is the generalized open call; most users will use Open or Create
// instead. It opens the named file with specified flag (O_RDONLY etc.) and
// perm, (0666 etc.) if applicable. If successful, methods on the returned
// File can be used for I/O.
func (f *Filesystem) OpenFile(filename string, flag int, perm os.FileMode) (fs.File, error) {
	f.t.Logf("OpenFile(%q, %d, %s)", filename, flag, perm)

	base, name := path.Split(filename)
	dir, err := f.getDir(base)
	if err != nil {
		return nil, err
	}

	n, ok := dir[name]
	if !ok {
		if flag&os.O_CREATE == 0 {
			return nil, fmt.Errorf("file %q: %w", filename, os.ErrNotExist)
		}

		fi := &File{
			buf:    []byte{},
			Reader: &bytes.Buffer{},
			Writer: &bytes.Buffer{},
			info: FileInfo{
				name: filename,
				mode: perm,
			},
		}

		n = fi
		dir[name] = fi
	}

	switch fi := n.(type) {
	case *File:
		return fi, nil
	case map[string]any:
		return &File{
			info: FileInfo{
				name:  filename,
				mode:  os.ModeDir,
				isDir: true,
			},
		}, nil
	default:
		panic("unexpected file tree type")
	}
}

// Stat returns a FileInfo describing the named file.
func (f *Filesystem) Stat(filename string) (os.FileInfo, error) {
	f.t.Logf("Stat(%q)", filename)

	fi, err := f.OpenFile(filename, os.O_RDONLY, 0644)
	if err != nil {
		return nil, err
	}

	return &fi.(*File).info, nil
}

// Remove removes the named file or directory.
func (f *Filesystem) Remove(filename string) error {
	f.t.Logf("Remove(%q)", filename)

	base, name := path.Split(filename)
	dir, err := f.getDir(base)
	if err != nil {
		return err
	}

	delete(dir, name)

	return nil
}

// ReadDir reads the directory named by dirname and returns a list of
// directory entries sorted by filename.
func (f *Filesystem) ReadDir(path string) (entries []os.FileInfo, _ error) {
	f.t.Logf("ReadDir(%q)", path)

	dir, err := f.getDir(path)
	if err != nil {
		return nil, err
	}

	for _, name := range keys(dir) {
		switch ent := dir[name].(type) {
		case *File:
			entries = append(entries, &ent.info)
		case map[string]any:
			entries = append(entries, &FileInfo{
				name:  filepath.Join(path, name),
				mode:  os.ModeDir,
				isDir: true,
			})
		default:
			panic("unexpected file tree type")
		}
	}

	return
}

func keys(m map[string]any) (keys []string) {
	for k := range m {
		keys = append(keys, k)
	}

	sort.Strings(keys)
	return
}

// MkdirAll creates a directory named path, along with any necessary
// parents, and returns nil, or else returns an error. The permission bits
// perm are used for all directories that MkdirAll creates. If path is/
// already a directory, MkdirAll does nothing and returns nil.
func (f *Filesystem) MkdirAll(filename string, _ os.FileMode) error {
	f.t.Logf("MkdirAll(%q)", filename)

	_, err := f.mkdirAll(filename)
	return err
}

func (f *Filesystem) mkdirAll(filename string) (map[string]any, error) {
	parts := strings.Split(strings.TrimSpace(filename), string(os.PathSeparator))
	if len(parts) == 0 {
		return nil, nil
	}

	tree := f.data
	for i, part := range parts {
		if part == "" {
			continue
		}

		dir, ok := tree[part]
		if ok {
			if tree, ok = dir.(map[string]any); !ok {
				return nil, fmt.Errorf("path already exists and is not a directory: %q", strings.Join(parts[:i+1], string(os.PathSeparator)))
			}

			continue
		}

		{
			dir := map[string]any{}
			tree[part] = dir
			tree = dir
		}
	}

	return tree, nil
}

func (f *Filesystem) getDir(path string) (map[string]any, error) {
	if path == "." {
		return f.data, nil
	}

	parts := strings.Split(strings.TrimSpace(path), string(os.PathSeparator))
	if len(parts) == 0 {
		return f.data, nil
	}

	tree := f.data
	for i, p := range parts {
		if p == "" {
			continue
		}

		d, ok := tree[p]
		if !ok {
			return nil, fmt.Errorf("directory %q: %w", strings.Join(parts[:i+1], string(os.PathSeparator)), os.ErrNotExist)
		}

		tree, ok = d.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("expected directory found file %q: %w", strings.Join(parts[:i+1], string(os.PathSeparator)), os.ErrInvalid)
		}
	}

	return tree, nil
}

type File struct {
	buf []byte

	Reader *bytes.Buffer
	Writer *bytes.Buffer

	info FileInfo
}

func (f *File) Write(p []byte) (n int, err error) {
	if f.Writer == nil {
		f.Writer = &bytes.Buffer{}
	}
	n, err = f.Writer.Write(p)
	if err == nil {
		// Keep buf in sync with Writer
		f.buf = f.Writer.Bytes()
		// Keep Reader in sync with latest content
		f.Reader = bytes.NewBuffer(f.buf)
	}
	return n, err
}

func (f *File) Read(p []byte) (n int, err error) {
	if f.Reader == nil {
		f.Reader = bytes.NewBuffer(f.buf)
	}
	return f.Reader.Read(p)
}

func (f *File) Name() string {
	return f.info.name
}

func (f *File) Stat() (os.FileInfo, error) {
	return &f.info, nil
}

func (f *File) Close() error {
	// copy writer contents into reader and reset writer
	if f.Writer != nil {
		f.buf = f.Writer.Bytes()
		f.Writer = nil
	}

	f.Reader = bytes.NewBuffer(f.buf)

	return nil
}

type FileInfo struct {
	name  string
	size  int64
	mode  os.FileMode
	mod   time.Time
	isDir bool
}

func (f *FileInfo) Name() string {
	return f.name
}

func (f *FileInfo) Size() int64 {
	return f.size
}

func (f *FileInfo) Mode() os.FileMode {
	return f.mode
}

func (f *FileInfo) ModTime() time.Time {
	return f.mod
}

func (f *FileInfo) IsDir() bool {
	return f.isDir
}

func (f *FileInfo) Sys() any {
	return nil
}
