package object

import (
	"context"
	"fmt"
	"net/url"

	s3v2 "github.com/aws/aws-sdk-go-v2/service/s3"
	gcaws "gocloud.dev/aws"
	gcblob "gocloud.dev/blob"
	_ "gocloud.dev/blob/azureblob"
	"gocloud.dev/blob/gcsblob"
	"gocloud.dev/blob/s3blob"
)

// s3Schema is a custom scheme for gocloud blob which works with
// how we interact with s3 (supports interfacing with minio)
const (
	s3Schema          = "s3i"
	googlecloudSchema = "googlecloud"
)

func init() {
	gcblob.DefaultURLMux().RegisterBucket(s3Schema, new(urlSessionOpener))
}

type urlSessionOpener struct{}

func (o *urlSessionOpener) OpenBucketURL(ctx context.Context, u *url.URL) (*gcblob.Bucket, error) {
	cfg, err := gcaws.V2ConfigFromURLParams(ctx, u.Query())
	if err != nil {
		return nil, fmt.Errorf("open bucket %v: %w", u, err)
	}
	clientV2 := s3v2.NewFromConfig(cfg, func(o *s3v2.Options) {
		o.UsePathStyle = true
	})
	return s3blob.OpenBucketV2(ctx, clientV2, u.Host, &s3blob.Options{})
}

// OpenBucket opens the bucket identified by the URL given.
//
// See the URLOpener documentation in driver subpackages for
// details on supported URL formats, and https://gocloud.dev/concepts/urls/
// for more information.
func OpenBucket(ctx context.Context, urlstr *url.URL) (*gcblob.Bucket, error) {
	urlCopy := *urlstr
	urlCopy.Scheme = remapScheme(urlstr.Scheme)
	return gcblob.DefaultURLMux().OpenBucketURL(ctx, &urlCopy)
}

func remapScheme(scheme string) string {
	switch scheme {
	case s3blob.Scheme:
		return s3Schema
	case googlecloudSchema:
		return gcsblob.Scheme
	default:
		return scheme
	}
}

func SupportedSchemes() []string {
	return append(gcblob.DefaultURLMux().BucketSchemes(), googlecloudSchema)
}
