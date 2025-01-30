package git

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/filemode"
	"github.com/go-git/go-git/v5/plumbing/object"
	gitstorage "github.com/go-git/go-git/v5/storage"
	"go.flipt.io/flipt/internal/server/authn"
	envfs "go.flipt.io/flipt/internal/storage/environments/fs"
	"go.uber.org/zap"
)

var _ envfs.Filesystem = (*filesystem)(nil)

type filesystem struct {
	logger  *zap.Logger
	base    *object.Commit
	tree    *object.Tree
	storage gitstorage.Storer

	sigName  string
	sigEmail string
}

// ReadDir reads the directory named by dirname and returns a list of
// directory entries sorted by filename.
func (f *filesystem) ReadDir(path string) (infos []os.FileInfo, err error) {
	f.logger.Debug("ReadDir", zap.String("path", path))

	subtree := f.tree
	if path != "." {
		subtree, err = f.tree.Tree(path)
		if errorIsNotFound(err) {
			return nil, fmt.Errorf("path %q: %w", path, os.ErrNotExist)
		}
	}

	for _, entry := range subtree.Entries {
		entry := entry
		info, err := entryToFileInfo(&entry)
		if err != nil {
			return nil, err
		}

		info.size, _ = f.tree.Size(filepath.Join(path, entry.Name))
		infos = append(infos, info)
	}

	return
}

// MkdirAll creates a directory named path, along with any necessary
// parents, and returns nil, or else returns an error. The permission bits
// perm are used for all directories that MkdirAll creates. If path is/
// already a directory, MkdirAll does nothing and returns nil.
func (f *filesystem) MkdirAll(filename string, perm os.FileMode) error {
	logger := f.logger.With(zap.String("path", filename))
	logger.Debug("MkdirAll Started", zap.Stringer("tree", f.tree.Hash))
	defer func() {
		f.logger.Debug("MkdirAll Finished", zap.Stringer("tree", f.tree.Hash))
	}()

	entry, err := f.tree.FindEntry(filename)
	if err == nil {
		if entry.Mode.IsFile() {
			return fmt.Errorf("path %q: %w and is not a directory", filename, fs.ErrExist)
		}

		// directory already exists
		return nil
	}

	if !errorIsNotFound(err) {
		return fmt.Errorf("path %q: %w", filename, err)
	}

	return updatePath(
		logger,
		f.storage,
		f.tree,
		append(strings.Split(filename, "/"), ".gitkeep"),
		true,
		&plumbing.ZeroHash,
	)
}

// OpenFile is the generalized open call; most users will use Open or Create
// instead. It opens the named file with specified flag (O_RDONLY etc.) and
// perm, (0666 etc.) if applicable. If successful, methods on the returned
// File can be used for I/O.
func (f *filesystem) OpenFile(filename string, flag int, perm os.FileMode) (envfs.File, error) {
	f.logger.Debug("OpenFile",
		zap.String("path", filename),
		zap.Bool("create", flag&os.O_CREATE == os.O_CREATE))

	fi, err := f.tree.File(filename)
	if flag&os.O_CREATE == 0 {
		if err != nil {
			if errorIsNotFound(err) {
				return nil, fmt.Errorf("path %q: %w", filename, os.ErrNotExist)
			}

			return nil, err
		}

		rd, err := fi.Reader()
		if err != nil {
			return nil, err
		}

		return &file{
			ReadCloser: rd,
			info: &fileInfo{
				name: filename,
				mode: os.ModePerm,
			},
		}, nil
	}

	if (flag&os.O_RDWR > 0 || flag&os.O_WRONLY > 0) && flag&os.O_TRUNC == 0 {
		return nil, errors.New("truncation currently required when writing to files")
	}

	file := &file{
		info: &fileInfo{
			name: filename,
			mode: os.ModePerm,
		},
		logger:  f.logger,
		tree:    f.tree,
		storage: f.storage,
		obj:     f.storage.NewEncodedObject(),
	}

	file.obj.SetType(plumbing.BlobObject)

	if err != nil {
		if !errorIsNotFound(err) {
			return nil, fmt.Errorf("opening file %q for writing: %w", filename, err)
		}

		file.ReadCloser = io.NopCloser(&bytes.Buffer{})
	} else {
		rd, err := fi.Reader()
		if err != nil {
			return nil, fmt.Errorf("opening reader for file %q: %w", filename, err)
		}

		file.ReadCloser = rd
	}

	return file, nil
}

// Stat returns a FileInfo describing the named file.
func (f *filesystem) Stat(filename string) (_ os.FileInfo, err error) {
	entry, err := f.tree.FindEntry(filename)
	if err != nil {
		if errorIsNotFound(err) {
			return nil, fmt.Errorf("path %q: %w", filename, os.ErrNotExist)
		}

		return nil, fmt.Errorf("path %q: %w", filename, err)
	}

	info, err := entryToFileInfo(entry)
	if err != nil {
		return nil, fmt.Errorf("gathering info: %w", err)
	}

	info.size, _ = f.tree.Size(filename)

	return info, nil
}

// Remove removes the named file or directory.
func (f *filesystem) Remove(filename string) error {
	entry, err := f.tree.FindEntry(filename)
	if err != nil {
		if errorIsNotFound(err) {
			return nil
		}

		return fmt.Errorf("removing path %q: %w", filename, err)
	}

	var hash *plumbing.Hash
	if entry.Mode.IsFile() {
		hash = &entry.Hash
	}

	return updatePath(
		f.logger,
		f.storage,
		f.tree,
		strings.Split(filename, "/"),
		false,
		hash,
	)
}

func entryToFileInfo(entry *object.TreeEntry) (*fileInfo, error) {
	mode, err := entry.Mode.ToOSFileMode()
	if err != nil {
		return nil, err
	}

	return &fileInfo{
		name:  entry.Name,
		mode:  mode,
		isDir: !entry.Mode.IsFile(),
	}, nil
}

type file struct {
	io.ReadCloser

	logger  *zap.Logger
	info    *fileInfo
	tree    *object.Tree
	storage gitstorage.Storer
	obj     plumbing.EncodedObject
}

func (f *file) Close() error {
	if err := f.ReadCloser.Close(); err != nil {
		return fmt.Errorf("closing %q: %w", f.info.name, err)
	}

	if f.obj != nil {
		hash, err := f.storage.SetEncodedObject(f.obj)
		if err != nil {
			return err
		}

		return updatePath(
			f.logger,
			f.storage,
			f.tree,
			strings.Split(f.info.name, "/"),
			true,
			&hash,
		)
	}

	return nil
}

func (f *file) Stat() (fs.FileInfo, error) {
	return f.info, nil
}

func (f *file) Write(p []byte) (n int, err error) {
	wr, err := f.obj.Writer()
	if err != nil {
		return 0, err
	}

	return wr.Write(p)
}

type fileInfo struct {
	name  string
	size  int64
	mode  fs.FileMode
	mod   time.Time
	isDir bool
}

func (f *fileInfo) Name() string {
	return f.name
}

func (f *fileInfo) Size() int64 {
	return f.size
}

func (f *fileInfo) Mode() fs.FileMode {
	return f.mode
}

func (f *fileInfo) ModTime() time.Time {
	return f.mod
}

func (f *fileInfo) IsDir() bool {
	return f.isDir
}

func (f *fileInfo) Sys() any {
	return nil
}

// updatePath recursively descends into the provided tree node and updates
// the entries signified by the provided path
// given the provided blob hash is zero, then it deletes the leaf and rewrites the path
// otherwise, it creates the path and inserts the blob
func updatePath(logger *zap.Logger, storage gitstorage.Storer, node *object.Tree, parts []string, insert bool, blob *plumbing.Hash) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("updating path %q: %w", parts, err)
		}
	}()

	if len(parts) == 0 {
		return nil
	}

	if parts[0] == "" {
		parts = parts[1:]
		if len(parts) == 0 {
			return nil
		}
	}

	toSort := object.TreeEntrySorter(node.Entries)
	if !sort.IsSorted(toSort) {
		sort.Sort(toSort)
	}

	// build a target for search / insertion
	// only the last entry in parts is considered a regular file
	target := object.TreeEntry{Name: parts[0], Mode: filemode.Dir}
	if len(parts) == 1 && blob != nil {
		target.Mode = filemode.Regular
	}

	// the comparison function here matches the less function for object.TreeEntrySorter
	// the difference being it has been implemented here using the newer generic form
	i, ok := slices.BinarySearchFunc(node.Entries, target, func(te1, te2 object.TreeEntry) int {
		name1 := te1.Name
		name2 := te2.Name
		if te1.Mode == filemode.Dir {
			name1 += "/"
		}
		if te2.Mode == filemode.Dir {
			name2 += "/"
		}

		return strings.Compare(name1, name2)
	})

	// first we check whether we're doing an insert or a delete
	if !insert {
		// performing a delete
		if !ok {
			// existing entry not present and therefore node goes unchanged
			return nil
		}

		if target.Mode.IsFile() {
			// adding a blob
			node.Entries = append(node.Entries[:i], node.Entries[i+1:]...)
		} else {
			// adding a tree
			tree, err := object.GetTree(storage, node.Entries[i].Hash)
			if err != nil {
				return err
			}

			if err := updatePath(logger, storage, tree, parts[1:], insert, blob); err != nil {
				return err
			}

			node.Entries[i].Hash = tree.Hash
		}
	} else {
		// performing an insert or update
		if target.Mode.IsFile() {
			// adding blob (assumes blob hash has been inserted)
			target.Hash = *blob

			if len(node.Entries) == 1 && node.Entries[0].Name == ".gitkeep" {
				// replace .gitkeep when inserting a blob
				// into a directory with currently no contents
				ok = true
				i = 0
				node.Entries[0].Name = target.Name
			}
		} else {
			// inserting or updating tree
			child := &object.Tree{}
			if ok {
				// updating existing tree
				child, err = object.GetTree(storage, node.Entries[i].Hash)
				if err != nil {
					return fmt.Errorf("getting tree %q (%s): %w", target.Name, node.Entries[i].Hash, err)
				}
			}

			// descend into tree with rest of path
			if err := updatePath(logger, storage, child, parts[1:], insert, blob); err != nil {
				return err
			}

			target.Hash = child.Hash
		}

		if ok {
			// has existing entry in parent
			node.Entries[i].Hash = target.Hash
		} else {
			// needs inserting in parent
			node.Entries = slices.Insert(node.Entries, i, target)
		}
	}

	obj := storage.NewEncodedObject()
	if err := node.Encode(obj); err != nil {
		return err
	}

	node.Hash, err = storage.SetEncodedObject(obj)
	if err != nil {
		return err
	}

	logger.Debug("Updating tree",
		zap.Strings("path", parts),
		zap.Stringer("tree_hash", node.Hash),
		zap.Stringer("blob_hash", blob))

	// decode back into node to reset state
	return node.Decode(obj)
}

func (f *filesystem) commit(ctx context.Context, msg string) (*object.Commit, error) {
	signature := object.Signature{
		Name:  f.sigName,
		Email: f.sigEmail,
		When:  time.Now().UTC(),
	}

	if actor := authn.ActorFromContext(ctx); actor != nil {
		signature.Name = actor.Name
		signature.Email = actor.Email
	}

	var hashes []plumbing.Hash
	if f.base != nil {
		hashes = []plumbing.Hash{f.base.Hash}
	}

	commit := &object.Commit{
		Author:       signature,
		Committer:    signature,
		Message:      msg,
		TreeHash:     f.tree.Hash,
		ParentHashes: hashes,
	}

	obj := f.storage.NewEncodedObject()
	err := commit.Encode(obj)
	if err != nil {
		return nil, fmt.Errorf("encoding commit: %w", err)
	}

	commit.Hash, err = f.storage.SetEncodedObject(obj)
	if err != nil {
		return nil, fmt.Errorf("storing commit object: %w", err)
	}

	return commit, nil
}

func errorIsNotFound(err error) bool {
	return errors.Is(err, object.ErrEntryNotFound) ||
		errors.Is(err, object.ErrDirectoryNotFound) ||
		errors.Is(err, object.ErrFileNotFound) ||
		errors.Is(err, plumbing.ErrObjectNotFound)
}
