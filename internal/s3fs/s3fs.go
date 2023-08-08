package s3fs

import (
	"context"
	"io"
	"io/fs"
	"strings"
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

	// configuration
	bucket string
	prefix string

	// cached entries
	dirEntry *Dir
}

// ensure FS implements fs.FS aka Open
var _ fs.FS = &FS{}

// ensure FS implements fs.StatFS aka Stat
var _ fs.StatFS = &FS{}

// ensure FS implements fs.ReadDirFS aka ReadDir
var _ fs.ReadDirFS = &FS{}

// New creates a FS for the single bucket
func New(logger *zap.Logger, s3Client S3ClientAPI, bucket string, prefix string) (*FS, error) {
	return &FS{
		logger:   logger,
		s3Client: s3Client,
		bucket:   bucket,
		prefix:   prefix,
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

	// If a prefix is not provided, prepend the prefix. This
	// allows s3fs to support `.flipt.yml` under the prefix
	if f.prefix != "" && !strings.HasPrefix(name, f.prefix) {
		name = f.prefix + name
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

	// AWS S3 does not store the last modified time for the bucket
	// anywhere. We'd have to iterate through all the objects to
	// calculate it, which doesn't seem worth it.

	return f.dirEntry, nil
}

// ReadDir implements fs.ReadDirFS. This can only be called on the
// current directory as the s3 filesystem does not support any kind of
// recursive directory structure
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

	// If a prefix is provided, only list objects with that prefix
	// This lets the user configure a portion of a bucket for
	// feature flags, simulating a subdirectory.
	//
	// See https://docs.aws.amazon.com/AmazonS3/latest/userguide/using-prefixes.html
	var prefix *string
	if f.prefix != "" {
		prefix = &f.prefix
	}

	// instead of caching the entries in Open, fetch them here so
	// if the list is large, they are not stored on the FS object.
	entries := []fs.DirEntry{}

	// loop until all results are retrieved, but don't loop more
	// than 100 times (creating 100,000 entries) as a safety
	// measure to ensure we don't run out of memory and/or loop
	// forever
	var continuationToken *string
	for i := 0; i < 100; i++ {
		output, err := f.s3Client.ListObjectsV2(context.Background(),
			&s3.ListObjectsV2Input{
				Bucket:            &f.bucket,
				Prefix:            prefix,
				ContinuationToken: continuationToken,
			})
		if err != nil {
			return nil, err
		}

		for i := range output.Contents {
			c := output.Contents[i]
			fi := &FileInfo{
				name:    *c.Key,
				size:    c.Size,
				modTime: *c.LastModified,
			}
			entries = append(entries, fi)
		}
		if !output.IsTruncated {
			return entries, nil
		}
		continuationToken = output.NextContinuationToken
	}

	// We looped more than 100 times. Instead of silently
	// truncating, return an error. Should we return a custom
	// error?
	return nil, &fs.PathError{
		Op:   "ReadDir",
		Path: name,
		Err:  fs.ErrClosed,
	}
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
