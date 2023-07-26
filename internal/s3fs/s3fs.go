package s3fs

import (
	"context"
	"io"
	"io/fs"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	flipterrors "go.flipt.io/flipt/errors"
	"go.uber.org/zap"
)

type S3ClientAPI interface {
	ListObjectsV2(context.Context, *s3.ListObjectsV2Input, ...func(*s3.Options)) (*s3.ListObjectsV2Output, error)
	GetObject(context.Context, *s3.GetObjectInput, ...func(*s3.Options)) (*s3.GetObjectOutput, error)
}

// FS is only for accessing files in a single bucket. The directory
// entries are cached. It is specifically intended for use by a source
// that calls fs.WalkDir and does not fully implement all fs operations
type FS struct {
	logger   *zap.Logger
	s3Client S3ClientAPI

	bucket   string
	dirEntry *Dir
	entries  []fs.DirEntry
}

// ensure FS implements fs.FS aka Open
var _ fs.FS = &FS{}

// ensure FS implements fs.StatFS aka Stat
var _ fs.StatFS = &FS{}

// ensure FS implements fs.ReadDirFS aka ReadDir
var _ fs.ReadDirFS = &FS{}

// New creates a FS for the single bucket
func New(logger *zap.Logger, s3Client S3ClientAPI, bucket string) (*FS, error) {
	return &FS{
		logger:   logger,
		s3Client: s3Client,
		bucket:   bucket,
	}, nil
}

// Open implements fs.FS. For the S3 filesystem, it fetches the object
// contents from s3.
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

	output, err := f.s3Client.GetObject(context.Background(),
		&s3.GetObjectInput{
			Bucket: &f.bucket,
			Key:    &name,
		})
	if err != nil {
		// try to return fs compatible error if possible
		if flipterrors.AsMatch[*types.NoSuchBucket](err) ||
			flipterrors.AsMatch[*types.NoSuchKey](err) ||
			flipterrors.AsMatch[*types.NotFound](err) {
			return nil, pathError
		}
		pathError.Err = err

		return nil, pathError
	}

	return &File{
		bucket:       f.bucket,
		key:          name,
		length:       output.ContentLength,
		body:         output.Body,
		lastModified: *output.LastModified,
	}, nil
}

// Stat implements fs.StatFS. For the s3 filesystem, this gets the
// objects in the s3 bucket and stores them for later use. Stat can
// only be called on the currect directory as the s3 filesystem only
// supports walking a single bucket configured at creation time.
func (f *FS) Stat(name string) (fs.FileInfo, error) {
	// Stat can only be called on the current directory
	if name != "." {
		return nil, &fs.PathError{
			Op:   "Stat",
			Path: name,
			Err:  fs.ErrInvalid,
		}
	}
	dirInfo := &FileInfo{
		name: name,
		size: 0,
	}
	f.dirEntry = &Dir{
		FileInfo: dirInfo,
	}
	output, err := f.s3Client.ListObjectsV2(context.Background(),
		&s3.ListObjectsV2Input{
			Bucket: &f.bucket,
		})
	if err != nil {
		return nil, err
	}

	f.entries = make([]fs.DirEntry, len(output.Contents))
	for i := range output.Contents {
		c := output.Contents[i]
		fi := &FileInfo{
			name:    *c.Key,
			size:    c.Size,
			modTime: *c.LastModified,
		}
		f.entries[i] = fi
		if dirInfo.modTime.IsZero() ||
			dirInfo.modTime.Compare(fi.modTime) < 0 {
			dirInfo.modTime = fi.modTime
		}
	}

	return f.dirEntry, nil
}

// ReadDir implements fs.ReadDirFS. For the s3 filesystem, this
// returns the previously fetched objects in the bucket. This can only
// be called on the current directory as the s3 filesystem does not
// support any kind of recursive directory structure
func (f *FS) ReadDir(name string) ([]fs.DirEntry, error) {
	// ReadDir can only be called after Stat and only on the
	// current directory, aka "." or the bucket
	if (name != "." && name != f.bucket) || f.entries == nil {
		return nil, &fs.PathError{
			Op:   "ReadDir",
			Path: name,
			Err:  fs.ErrInvalid,
		}
	}

	return f.entries, nil
}

type Dir struct {
	*FileInfo
}

// ensure Dir implements fs.FileInfo
var _ fs.FileInfo = &Dir{}

func (d *Dir) Stat() (fs.FileInfo, error) {
	return d.FileInfo, nil
}

func (d *Dir) Read(p []byte) (n int, err error) {
	return 0, io.EOF
}

func (d *Dir) Close() error {
	return nil
}

func (d *Dir) IsDir() bool {
	return true
}

func (d *Dir) Mode() fs.FileMode {
	return fs.ModeDir
}

// ensure FileInfo implements fs.FileInfo
var _ fs.FileInfo = &FileInfo{}

// ensure FileInfo implements fs.DirEntry
var _ fs.DirEntry = &FileInfo{}

type FileInfo struct {
	name    string
	size    int64
	modTime time.Time
}

func (fi *FileInfo) Name() string {
	return fi.name
}

func (fi *FileInfo) Size() int64 {
	return fi.size
}

func (fi *FileInfo) Type() fs.FileMode {
	return 0
}

func (fi *FileInfo) Mode() fs.FileMode {
	return fs.ModePerm
}

func (fi *FileInfo) ModTime() time.Time {
	return fi.modTime
}

func (fi *FileInfo) IsDir() bool {
	return false
}

func (fi *FileInfo) Sys() any {
	return nil
}
func (fi *FileInfo) Info() (fs.FileInfo, error) {
	return fi, nil
}

type File struct {
	bucket       string
	key          string
	length       int64
	body         io.ReadCloser
	lastModified time.Time
}

// ensure File implements the fs.File interface
var _ fs.File = &File{}

func (f *File) Stat() (fs.FileInfo, error) {
	return &FileInfo{
		name: f.key,
		size: f.length,
	}, nil
}

func (f *File) Read(p []byte) (int, error) {
	return f.body.Read(p)
}

func (f *File) Close() error {
	return f.body.Close()
}
