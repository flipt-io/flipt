package oci

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"time"

	"github.com/opencontainers/go-digest"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"go.flipt.io/flipt/internal/containers"
	ocifs "go.flipt.io/flipt/internal/oci/fs"
	"go.uber.org/zap"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content"
	"oras.land/oras-go/v2/content/memory"
	"oras.land/oras-go/v2/registry"
	"oras.land/oras-go/v2/registry/remote"
)

var ErrDigestSeen = errors.New("digest already seen")

type Source struct {
	logger   *zap.Logger
	interval time.Duration

	current digest.Digest
	ref     registry.Reference
	remote  *remote.Repository
	local   *memory.Store
}

// NewSource constructs and configures a Source.
// The source uses the connection and credential details provided to build
// fs.FS implementations around a target git repository.
func NewSource(logger *zap.Logger, reference string, opts ...containers.Option[Source]) (_ *Source, err error) {
	ref, err := registry.ParseReference(reference)
	if err != nil {
		return nil, err
	}

	repo, err := remote.NewRepository(fmt.Sprintf("%s/%s", ref.Registry, ref.Repository))
	if err != nil {
		return nil, err
	}

	repo.PlainHTTP = true

	src := &Source{
		logger:   logger.With(zap.String("repository", ref.String())),
		interval: 30 * time.Second,
		ref:      ref,
		remote:   repo,
		local:    memory.New(),
	}
	containers.ApplyAll(src, opts...)

	return src, nil
}

// WithPollInterval configures the interval in which origin is polled to
// discover any updates to the target reference.
func WithPollInterval(tick time.Duration) containers.Option[Source] {
	return func(s *Source) {
		s.interval = tick
	}
}

func (s *Source) String() string {
	return "oci"
}

// Get builds a single instance of an fs.FS
func (s *Source) Get() (fs.FS, error) {
	manifest, err := s.copyFromRemote(context.Background())
	if err != nil {
		return nil, err
	}

	if s.current != "" {
		manifest := manifest
		manifest.Annotations = map[string]string{}
		bytes, err := json.Marshal(&manifest)
		if err != nil {
			return nil, err
		}

		newDigest := digest.FromBytes(bytes)
		if newDigest == s.current {
			return nil, ErrDigestSeen
		}

		s.current = newDigest
	}

	return ocifs.New(s.local, manifest), nil
}

// Subscribe feeds implementations of fs.FS onto the provided channel.
// It should block until the provided context is cancelled (it will be called in a goroutine).
// It should close the provided channel before it returns.
func (s *Source) Subscribe(ctx context.Context, ch chan<- fs.FS) {
	defer close(ch)

	ticker := time.NewTicker(s.interval)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			fs, err := s.Get()
			if err != nil {
				if errors.Is(err, ErrDigestSeen) {
					s.logger.Debug("store already up to date")
					continue
				}

				s.logger.Error("failed resolving upstream", zap.Error(err))
				continue
			}

			ch <- fs

			s.logger.Debug("fetched new reference from remote")
		}
	}
}

func (s *Source) copyFromRemote(ctx context.Context) (m v1.Manifest, _ error) {
	desc, err := oras.Copy(ctx, s.remote, s.ref.Reference, s.local, s.ref.Reference, oras.DefaultCopyOptions)
	if err != nil {
		return m, err
	}

	bytes, err := content.FetchAll(ctx, s.local, desc)
	if err != nil {
		return m, err
	}

	return m, json.Unmarshal(bytes, &m)
}
