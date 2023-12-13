package azblob

import (
	"context"
	"io/fs"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/bloberror"
	"go.flipt.io/flipt/internal/s3fs"
	"go.uber.org/zap"
)

// FS is only for accessing files in a single bucket. The directory
// entries are cached. It is specifically intended for use by a source
// that calls fs.WalkDir and does not fully implement all fs operations
type FS struct {
	logger *zap.Logger
	client ClientAPI

	// configuration
	containerName string

	// cached entries
	dirEntry *s3fs.Dir
}

// ensure FS implements fs.FS aka Open
var _ fs.FS = &FS{}

// ensure FS implements fs.StatFS aka Stat
var _ fs.StatFS = &FS{}

// ensure FS implements fs.ReadDirFS aka ReadDir
var _ fs.ReadDirFS = &FS{}

// New creates a FS for the container
func NewFS(logger *zap.Logger, client ClientAPI, containerName string) (*FS, error) {
	return &FS{
		logger:        logger,
		client:        client,
		containerName: containerName,
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

	output, err := f.client.DownloadStream(context.Background(), f.containerName, name, nil)
	if err != nil {
		if bloberror.HasCode(err,
			bloberror.ContainerNotFound, bloberror.BlobNotFound, bloberror.ResourceNotFound,
		) {
			return nil, pathError
		}
		pathError.Err = err
		return nil, pathError
	}

	return s3fs.NewFile(
		f.containerName,
		name,
		*output.ContentLength,
		output.Body,
		*output.LastModified,
	), nil
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

	return s3fs.NewDir(s3fs.NewFileInfo(name, 0, time.Time{})), nil
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

	pager := f.client.NewListBlobsFlatPager(f.containerName, nil)

	for pager.More() {
		resp, err := pager.NextPage(context.TODO())
		if err != nil {
			return nil, err
		}
		for _, blob := range resp.Segment.BlobItems {
			fi := s3fs.NewFileInfo(
				*blob.Name,
				*blob.Properties.ContentLength,
				*blob.Properties.LastModified,
			)
			entries = append(entries, fi)
		}
	}
	return entries, nil
}
