package s3

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"go.flipt.io/flipt/internal/containers"
	"go.flipt.io/flipt/internal/s3fs"
	storagefs "go.flipt.io/flipt/internal/storage/fs"
	"go.uber.org/zap"
)

// Source represents an implementation of an fs.SnapshotSource
// This implementation is backed by an S3 bucket
type Source struct {
	logger *zap.Logger
	s3     *s3.Client

	endpoint string
	region   string
	bucket   string
	prefix   string
	interval time.Duration
}

// NewSource constructs a Source.
func NewSource(logger *zap.Logger, bucket string, opts ...containers.Option[Source]) (*Source, error) {
	s := &Source{
		logger:   logger,
		bucket:   bucket,
		interval: 60 * time.Second,
	}

	containers.ApplyAll(s, opts...)

	s3opts := make([](func(*config.LoadOptions) error), 0)
	if s.region != "" {
		s3opts = append(s3opts, config.WithRegion(s.region))
	}
	if s.endpoint != "" {
		customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
			if service == s3.ServiceID {
				return aws.Endpoint{
					PartitionID:       "aws",
					URL:               s.endpoint,
					HostnameImmutable: true,
					SigningRegion:     s.region,
				}, nil
			}
			return aws.Endpoint{}, fmt.Errorf("unknown endpoint requested")
		})
		s3opts = append(s3opts, config.WithEndpointResolverWithOptions(customResolver))
	}
	cfg, err := config.LoadDefaultConfig(context.Background(),
		s3opts...)
	if err != nil {
		return nil, err
	}
	s.s3 = s3.NewFromConfig(cfg)

	return s, nil
}

// WithPrefix configures the prefix for s3
func WithPrefix(prefix string) containers.Option[Source] {
	return func(s *Source) {
		s.prefix = prefix
	}
}

// WithRegion configures the region for s3
func WithRegion(region string) containers.Option[Source] {
	return func(s *Source) {
		s.region = region
	}
}

// WithEndpoint configures the region for s3
func WithEndpoint(endpoint string) containers.Option[Source] {
	return func(s *Source) {
		s.endpoint = endpoint
	}
}

// WithPollInterval configures the interval in which we will restore
// the s3 fs.
func WithPollInterval(tick time.Duration) containers.Option[Source] {
	return func(s *Source) {
		s.interval = tick
	}
}

// Get returns a *sourcefs.StoreSnapshot for the local filesystem.
func (s *Source) Get() (*storagefs.StoreSnapshot, error) {
	fs, err := s3fs.New(s.logger, s.s3, s.bucket, s.prefix)
	if err != nil {
		return nil, err
	}

	return storagefs.SnapshotFromFS(s.logger, fs)
}

// Subscribe feeds S3 populated *StoreSnapshot instances onto the provided channel.
// It blocks until the provided context is cancelled.
func (s *Source) Subscribe(ctx context.Context, ch chan<- *storagefs.StoreSnapshot) {
	defer close(ch)

	ticker := time.NewTicker(s.interval)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			snap, err := s.Get()
			if err != nil {
				s.logger.Error("error getting file system from directory", zap.Error(err))
				continue
			}

			s.logger.Debug("updating local store snapshot")

			ch <- snap
		}
	}
}

// String returns an identifier string for the store type.
func (s *Source) String() string {
	return "s3"
}
