package azblob

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"time"

	"go.flipt.io/flipt/internal/storage/fs/blob"
	"go.uber.org/zap"

	gcblob "gocloud.dev/blob"
	_ "gocloud.dev/blob/azureblob"
)

// FS is only for accessing files in a single bucket. The directory
// entries are cached. It is specifically intended for use by a source
// that calls fs.WalkDir and does not fully implement all fs operations
type FS struct {
	logger *zap.Logger

	// configuration
	containerName string
	urlstr        string

	// cached entries
	dirEntry *blob.Dir
}

// ensure FS implements fs.FS aka Open
var _ fs.FS = &FS{}

// ensure FS implements fs.StatFS aka Stat
var _ fs.StatFS = &FS{}

// ensure FS implements fs.ReadDirFS aka ReadDir
var _ fs.ReadDirFS = &FS{}

// New creates a FS for the container
func NewFS(logger *zap.Logger, schema string, containerName string) (*FS, error) {
	return &FS{
		logger:        logger,
		containerName: containerName,
		urlstr:        fmt.Sprintf("%s://%s", schema, containerName),
	}, nil
}

// Open implements fs.FS. it fetches the object contents from azure blob.
func (f *FS) Open(name string) (fs.File, error) {
	if name == "." {
		return f.dirEntry, nil
	}
	pathError := &fs.PathError{
		Op:   "Open",
		Path: name,
		Err:  fs.ErrNotExist,
	}
	if !fs.ValidPath(name) {
		pathError.Err = fs.ErrInvalid
		return nil, pathError
	}

	ctx := context.TODO()
	bucket, err := gcblob.OpenBucket(ctx, f.urlstr)
	if err != nil {
		return nil, err
	}
	defer bucket.Close()
	bucket.SetIOFSCallback(func() (context.Context, *gcblob.ReaderOptions) {
		return ctx, nil
	})
	return bucket.Open(name)
}

// Stat implements fs.StatFS. For the  filesystem, this gets the
// objects in the container and stores them for later use.
func (f *FS) Stat(name string) (fs.FileInfo, error) {
	// Stat can only be called on the current directory
	if name != "." {
		return nil, &fs.PathError{
			Op:   "Stat",
			Path: name,
			Err:  fs.ErrInvalid,
		}
	}

	return blob.NewDir(blob.NewFileInfo(name, 0, time.Time{})), nil
}

// ReadDir implements fs.ReadDirFS. This can only be called on the
// current directory
func (f *FS) ReadDir(name string) ([]fs.DirEntry, error) {
	// ReadDir can only be called on the current directory, aka
	// "." or the bucket
	if name != "." && name != f.containerName {
		return nil, &fs.PathError{
			Op:   "ReadDir",
			Path: name,
			Err:  fs.ErrInvalid,
		}
	}

	// instead of caching the entries in Open, fetch them here so
	// if the list is large, they are not stored on the FS object.
	entries := []fs.DirEntry{}

	ctx := context.TODO()
	bucket, err := gcblob.OpenBucket(ctx, f.urlstr)
	if err != nil {
		return nil, err
	}
	defer bucket.Close()
	iterator := bucket.List(&gcblob.ListOptions{})
	for {
		item, err := iterator.Next(ctx)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}
		fi := blob.NewFileInfo(
			item.Key,
			item.Size,
			item.ModTime,
		)
		fi.SetDir(item.IsDir)
		entries = append(entries, fi)

	}
	return entries, nil
}
