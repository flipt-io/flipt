package blob

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"strings"
	"time"

	"go.uber.org/zap"
	gcblob "gocloud.dev/blob"
)

// FS is only for accessing files in a single bucket. The directory
// entries are cached. It is specifically intended for use by a source
// that calls fs.WalkDir and does not fully implement all fs operations
type FS struct {
	logger *zap.Logger

	// configuration
	bucket string
	prefix string
	urlstr string

	// cached entries
	dirEntry *Dir
}

// ensure FS implements fs.FS aka Open
var _ fs.FS = &FS{}

// ensure FS implements fs.StatFS aka Stat
var _ fs.StatFS = &FS{}

// ensure FS implements fs.ReadDirFS aka ReadDir
var _ fs.ReadDirFS = &FS{}

func StrUrl(schema, bucket string) string {
	return fmt.Sprintf("%s://%s", schema, bucket)
}

// New creates a FS for the container
func NewFS(logger *zap.Logger, urlstr string, bucket string, prefix string) (*FS, error) {
	if prefix != "" {
		prefix = strings.Trim(prefix, "/") + "/" // to match "a/subfolder/"
	}

	return &FS{
		logger: logger,
		bucket: bucket,
		urlstr: urlstr,
		prefix: prefix,
	}, nil
}

func (f *FS) openBucket(ctx context.Context) (*gcblob.Bucket, error) {
	bucket, err := gcblob.OpenBucket(ctx, f.urlstr)
	if err != nil {
		return nil, err
	}
	if f.prefix != "" {
		bucket = gcblob.PrefixedBucket(bucket, f.prefix)
	}
	bucket.SetIOFSCallback(func() (context.Context, *gcblob.ReaderOptions) {
		return ctx, nil
	})
	return bucket, err
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

	ctx := context.Background()
	bucket, err := f.openBucket(ctx)
	if err != nil {
		return nil, err
	}
	defer bucket.Close()
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

	return NewDir(NewFileInfo(name, 0, time.Time{})), nil
}

// ReadDir implements fs.ReadDirFS. This can only be called on the
// current directory
func (f *FS) ReadDir(name string) ([]fs.DirEntry, error) {
	// ReadDir can only be called on the current directory, aka
	// "." or the bucket
	if name != "." && name != f.bucket {
		return nil, &fs.PathError{
			Op:   "ReadDir",
			Path: name,
			Err:  fs.ErrInvalid,
		}
	}

	// instead of caching the entries in Open, fetch them here so
	// if the list is large, they are not stored on the FS object.
	entries := []fs.DirEntry{}

	ctx := context.Background()
	bucket, err := f.openBucket(ctx)
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
		fi := NewFileInfo(
			item.Key,
			item.Size,
			item.ModTime,
		)
		fi.SetDir(item.IsDir)
		entries = append(entries, fi)

	}
	return entries, nil
}
