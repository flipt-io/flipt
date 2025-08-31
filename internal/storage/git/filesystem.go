package git

import (
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

	"github.com/go-git/go-git/v6/plumbing"
	"github.com/go-git/go-git/v6/plumbing/filemode"
	"github.com/go-git/go-git/v6/plumbing/object"
	gitstorage "github.com/go-git/go-git/v6/storage"
	"go.flipt.io/flipt/internal/containers"
	"go.flipt.io/flipt/internal/server/authn"
	envfs "go.flipt.io/flipt/internal/storage/environments/fs"
	"go.flipt.io/flipt/internal/storage/git/signing"
	"go.uber.org/zap"
)

var _ envfs.Filesystem = (*filesystem)(nil)

// filesystem implements the Filesystem interface for a particular git tree object.
type filesystem struct {
	logger  *zap.Logger
	base    *object.Commit
	tree    *object.Tree
	storage gitstorage.Storer

	sigName  string
	sigEmail string
	signer   signing.Signer
}

type filesystemOption struct {
	hash              plumbing.Hash
	sigName, sigEmail string
	signer            signing.Signer
}

func withBaseCommit(hash plumbing.Hash) containers.Option[filesystemOption] {
	return func(o *filesystemOption) {
		o.hash = hash
	}
}

func withSignature(name, email string) containers.Option[filesystemOption] {
	return func(o *filesystemOption) {
		o.sigName = name
		o.sigEmail = email
	}
}

func withSigner(signer signing.Signer) containers.Option[filesystemOption] {
	return func(o *filesystemOption) {
		o.signer = signer
	}
}

// emptyTreeObj is used to construct a tree object with no entries.
type emptyTreeObj struct {
	plumbing.EncodedObject
}

func (e emptyTreeObj) Type() plumbing.ObjectType {
	return plumbing.TreeObject
}

func (e emptyTreeObj) Hash() plumbing.Hash {
	return plumbing.ZeroHash
}

func (e emptyTreeObj) Size() int64 {
	return 0
}

func newFilesystem(logger *zap.Logger, storer gitstorage.Storer, opts ...containers.Option[filesystemOption]) (_ *filesystem, err error) {
	var (
		fopts = filesystemOption{
			sigName:  "flipt",
			sigEmail: "dev@flipt.io",
		}
		commit *object.Commit
	)

	containers.ApplyAll(&fopts, opts...)

	tree, err := object.DecodeTree(storer, emptyTreeObj{})
	if err != nil {
		return nil, err
	}

	// zero hash assumes we're building from an emptry repository
	// the caller needs to validate whether this is true or not
	// before calling newFilesystem with zero hash
	if fopts.hash != plumbing.ZeroHash {
		commit, err = object.GetCommit(storer, fopts.hash)
		if err != nil {
			return nil, fmt.Errorf("getting branch commit: %w", err)
		}

		tree, err = commit.Tree()
		if err != nil {
			return nil, err
		}
	}

	return &filesystem{
		logger:   logger,
		base:     commit,
		tree:     tree,
		storage:  storer,
		sigName:  fopts.sigName,
		sigEmail: fopts.sigEmail,
		signer:   fopts.signer,
	}, nil
}

// ReadDir reads the directory named by dirname and returns a list of
// directory entries sorted by filename.
func (f *filesystem) ReadDir(path string) (infos []os.FileInfo, err error) {
	f.logger.Debug("readDir", zap.String("path", path))

	subtree := f.tree
	if path != "." {
		subtree, err = f.tree.Tree(path)
		if errorIsNotFound(err) {
			return nil, fmt.Errorf("path %q: %w", path, os.ErrNotExist)
		}
	}

	for _, entry := range subtree.Entries {
		info, err := f.entryToFileInfo(&entry)
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
	logger.Debug("mkdirAll Started", zap.Stringer("tree", f.tree.Hash))
	defer func() {
		f.logger.Debug("mkdirAll Finished", zap.Stringer("tree", f.tree.Hash))
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

	// Create an empty blob for .gitkeep
	obj := f.storage.NewEncodedObject()
	obj.SetType(plumbing.BlobObject)
	obj.SetSize(0)

	gitkeepHash, err := f.storage.SetEncodedObject(obj)
	if err != nil {
		return fmt.Errorf("creating .gitkeep blob: %w", err)
	}

	return updatePath(
		logger,
		f.storage,
		f.tree,
		append(strings.Split(filename, "/"), ".gitkeep"),
		true,
		&gitkeepHash,
	)
}

// OpenFile is the generalized open call; most users will use Open or Create
// instead. It opens the named file with specified flag (O_RDONLY etc.) and
// perm, (0666 etc.) if applicable. If successful, methods on the returned
// File can be used for I/O.
func (f *filesystem) OpenFile(filename string, flag int, perm os.FileMode) (envfs.File, error) {
	f.logger.Debug("openFile",
		zap.String("path", filename),
		zap.Bool("create", flag&os.O_CREATE == os.O_CREATE))

	var mod time.Time
	if f.base != nil {
		mod = f.base.Committer.When
	}

	var (
		entry *object.TreeEntry
		err   error
	)

	if filename == "." {
		entry = &object.TreeEntry{
			Name: "/",
			Hash: f.tree.Hash,
			Mode: filemode.Dir,
		}
	} else {
		entry, err = f.tree.FindEntry(filename)
	}

	if err != nil {
		if !errorIsNotFound(err) {
			return nil, err
		}

		if flag&os.O_CREATE == 0 {
			return nil, fmt.Errorf("path %q: %w", filename, os.ErrNotExist)
		}
	}

	obj := f.storage.NewEncodedObject()
	obj.SetType(plumbing.BlobObject)
	if entry != nil {
		if !entry.Mode.IsFile() {
			// handle directories
			tree := f.tree
			if filename != "." {
				tree, err = f.tree.Tree(filename)
				if err != nil {
					return nil, err
				}
			}

			dir := dir{
				stat: &fileInfo{
					name:  filename,
					mode:  os.ModeDir,
					mod:   mod,
					isDir: true,
				},
			}

			for _, entry := range tree.Entries {
				info, err := f.entryToFileInfo(&entry)
				if err != nil {
					return nil, err
				}

				dir.entries = append(dir.entries, dirEntry{info})
			}

			return dir, nil
		}

		if flag&os.O_TRUNC == 0 {
			obj, err = f.storage.EncodedObject(plumbing.BlobObject, entry.Hash)
			if err != nil {
				return nil, err
			}
		}
	}

	rd, err := obj.Reader()
	if err != nil {
		return nil, err
	}

	file := &file{
		info: &fileInfo{
			name: filename,
			mode: os.ModePerm,
			mod:  mod,
		},
		logger:   f.logger,
		tree:     f.tree,
		storage:  f.storage,
		obj:      obj,
		writable: flag&os.O_WRONLY > 0 || flag&os.O_RDWR > 0,
	}

	if flag&os.O_APPEND > 0 {
		wr, err := file.obj.Writer()
		if err != nil {
			return nil, err
		}

		if _, err := io.Copy(wr, rd); err != nil {
			return nil, err
		}
	}

	file.ReadCloser = rd

	return file, nil
}

// Stat returns a FileInfo describing the named file.
func (f *filesystem) Stat(filename string) (_ os.FileInfo, err error) {
	entry := &object.TreeEntry{
		Name: filename,
		Mode: filemode.Dir,
		Hash: f.tree.Hash,
	}

	if filename != "." {
		entry, err = f.tree.FindEntry(filename)
		if err != nil {
			if errorIsNotFound(err) {
				return nil, fmt.Errorf("path %q: %w", filename, os.ErrNotExist)
			}

			return nil, fmt.Errorf("path %q: %w", filename, err)
		}
	}

	info, err := f.entryToFileInfo(entry)
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

func (f *filesystem) entryToFileInfo(entry *object.TreeEntry) (*fileInfo, error) {
	mode, err := entry.Mode.ToOSFileMode()
	if err != nil {
		return nil, err
	}

	var mod time.Time
	if f.base != nil {
		mod = f.base.Committer.When
	}

	return &fileInfo{
		name:  entry.Name,
		mode:  mode,
		mod:   mod,
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

	writable bool
	written  int
}

func (f *file) Close() error {
	if err := f.ReadCloser.Close(); err != nil {
		return fmt.Errorf("closing %q: %w", f.info.name, err)
	}

	if f.written > 0 {
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
	if !f.writable {
		return 0, fmt.Errorf("writing to read-only file")
	}

	defer func() {
		if err == nil {
			f.written += n
		}
	}()

	wr, err := f.obj.Writer()
	if err != nil {
		return 0, err
	}

	return wr.Write(p)
}

type dir struct {
	stat    os.FileInfo
	entries []fs.DirEntry
}

func (d dir) Read(p []byte) (n int, err error) {
	return 0, io.EOF
}

func (d dir) Write(p []byte) (n int, err error) {
	return 0, fmt.Errorf("writing to directory")
}

func (d dir) Close() error {
	return nil
}

func (d dir) Stat() (fs.FileInfo, error) {
	return d.stat, nil
}

func (d dir) ReadDir(n int) ([]fs.DirEntry, error) {
	return d.entries, nil
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

type dirEntry struct {
	info *fileInfo
}

func (d dirEntry) Name() string {
	return d.info.name
}

func (d dirEntry) IsDir() bool {
	return d.info.isDir
}

func (d dirEntry) Type() fs.FileMode {
	return d.info.mode
}

func (d dirEntry) Info() (fs.FileInfo, error) {
	return d.info, nil
}

// validateTree ensures a tree object is valid and can be safely stored
func validateTree(tree *object.Tree) error {
	// Check for duplicate entries
	seen := make(map[string]bool)
	for _, entry := range tree.Entries {
		if seen[entry.Name] {
			return fmt.Errorf("duplicate tree entry: %s", entry.Name)
		}
		seen[entry.Name] = true

		// Validate entry name doesn't contain invalid characters
		if strings.Contains(entry.Name, "/") {
			return fmt.Errorf("tree entry name contains slash: %s", entry.Name)
		}

		// Ensure hash is not zero (except for gitmodules which can be)
		if entry.Hash.IsZero() && entry.Mode != filemode.Submodule {
			return fmt.Errorf("tree entry has zero hash: %s", entry.Name)
		}
	}

	// Ensure tree entries are properly sorted
	toSort := object.TreeEntrySorter(tree.Entries)
	if !sort.IsSorted(toSort) {
		return fmt.Errorf("tree entries are not properly sorted")
	}

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
			// deleting a blob
			node.Entries = slices.Delete(node.Entries, i, i+1)
		} else {
			// deleting from a tree
			tree, err := object.GetTree(storage, node.Entries[i].Hash)
			if err != nil {
				return err
			}

			if err := updatePath(logger, storage, tree, parts[1:], insert, blob); err != nil {
				return err
			}

			// Check if the tree is now empty or only contains .gitkeep
			// If so, remove the directory entry entirely
			if len(tree.Entries) == 0 || (len(tree.Entries) == 1 && tree.Entries[0].Name == ".gitkeep") {
				// Remove empty directory from parent
				node.Entries = slices.Delete(node.Entries, i, i+1)
				logger.Debug("removed empty directory after deletion",
					zap.String("directory", parts[0]))
			} else {
				// Update the tree hash for non-empty directory
				node.Entries[i].Hash = tree.Hash
			}
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

	// Skip validation and encoding for empty trees - they have a well-known hash
	if len(node.Entries) == 0 {
		// Set the well-known empty tree hash
		node.Hash = plumbing.NewHash("4b825dc642cb6eb9a060e54bf8d69288fbee4904")
		logger.Debug("updating tree to empty",
			zap.Strings("path", parts),
			zap.Stringer("tree_hash", node.Hash))
		return nil
	}

	// Validate tree before encoding
	if err := validateTree(node); err != nil {
		return fmt.Errorf("invalid tree state: %w", err)
	}

	obj := storage.NewEncodedObject()
	if err := node.Encode(obj); err != nil {
		return err
	}

	node.Hash, err = storage.SetEncodedObject(obj)
	if err != nil {
		return err
	}

	logger.Debug("updating tree",
		zap.Strings("path", parts),
		zap.Stringer("tree_hash", node.Hash),
		zap.Stringer("blob_hash", blob),
		zap.Int("entries", len(node.Entries)))

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

	// Ensure we have a valid tree hash and the tree is in storage
	treeHash := f.tree.Hash
	if treeHash.IsZero() || len(f.tree.Entries) == 0 {
		// Create and store an empty tree
		emptyTree := &object.Tree{}
		obj := f.storage.NewEncodedObject()
		if err := emptyTree.Encode(obj); err != nil {
			return nil, fmt.Errorf("encoding empty tree: %w", err)
		}
		var err error
		treeHash, err = f.storage.SetEncodedObject(obj)
		if err != nil {
			return nil, fmt.Errorf("storing empty tree: %w", err)
		}
		f.tree.Hash = treeHash
	} else {
		// Ensure the tree is in storage
		if _, err := f.storage.EncodedObject(plumbing.TreeObject, treeHash); err != nil {
			// Tree not in storage, store it
			obj := f.storage.NewEncodedObject()
			if err := f.tree.Encode(obj); err != nil {
				return nil, fmt.Errorf("encoding tree: %w", err)
			}
			if _, err := f.storage.SetEncodedObject(obj); err != nil {
				return nil, fmt.Errorf("storing tree: %w", err)
			}
		}
	}

	commit := &object.Commit{
		Author:       signature,
		Committer:    signature,
		Message:      msg,
		TreeHash:     treeHash,
		ParentHashes: hashes,
	}

	// Attempt to sign the commit if a signer is available.
	// If no signer is available, the commit will not be signed.
	// If signing fails, an error will be returned.
	if f.signer != nil {
		pgpSig, err := f.signer.SignCommit(ctx, commit)
		if err != nil {
			return nil, fmt.Errorf("signing commit: %w", err)
		}
		commit.PGPSignature = pgpSig

		f.logger.Debug("signed commit",
			zap.String("tree_hash", commit.TreeHash.String()),
			zap.String("message", commit.Message))
	}

	obj := f.storage.NewEncodedObject()
	err := commit.Encode(obj)
	if err != nil {
		return nil, fmt.Errorf("encoding commit: %w", err)
	}

	hash, err := f.storage.SetEncodedObject(obj)
	if err != nil {
		return nil, fmt.Errorf("storing commit object: %w", err)
	}
	commit.Hash = hash

	// Return a commit object that can access storage
	// This is needed so that commit.Tree() can retrieve the tree from storage
	storedCommit, err := object.GetCommit(f.storage, hash)
	if err != nil {
		return commit, nil // Fall back to returning the original commit
	}

	// Preserve the original PGP signature if it exists
	// (GetCommit might add trailing whitespace when decoding)
	if commit.PGPSignature != "" {
		storedCommit.PGPSignature = commit.PGPSignature
	}

	return storedCommit, nil
}

func errorIsNotFound(err error) bool {
	return errors.Is(err, object.ErrEntryNotFound) ||
		errors.Is(err, object.ErrDirectoryNotFound) ||
		errors.Is(err, object.ErrFileNotFound) ||
		errors.Is(err, plumbing.ErrObjectNotFound)
}
