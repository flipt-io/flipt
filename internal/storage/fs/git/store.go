package git

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"sync"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/storage/memory"
	"go.flipt.io/flipt/internal/containers"
	"go.flipt.io/flipt/internal/gitfs"
	"go.flipt.io/flipt/internal/storage"
	storagefs "go.flipt.io/flipt/internal/storage/fs"
	"go.uber.org/zap"
)

// ensure that the git *Store implements storage.Store
var _ storagefs.SnapshotStore = (*SnapshotStore)(nil)

// SnapshotStore is an implementation of storage.SnapshotStore
// This implementation is backed by a Git repository and it tracks an upstream reference.
// When subscribing to this source, the upstream reference is tracked
// by polling the upstream on a configurable interval.
type SnapshotStore struct {
	io.Closer

	logger *zap.Logger
	repo   *git.Repository

	mu   sync.RWMutex
	snap storage.ReadOnlyStore

	url             string
	ref             string
	hash            plumbing.Hash
	auth            transport.AuthMethod
	caBundle        []byte
	insecureSkipTLS bool
	pollOpts        []containers.Option[storagefs.Poller]
}

// WithRef configures the target reference to be used when fetching
// and building fs.FS implementations.
// If it is a valid hash, then the fixed SHA value is used.
// Otherwise, it is treated as a reference in the origin upstream.
func WithRef(ref string) containers.Option[SnapshotStore] {
	return func(s *SnapshotStore) {
		if plumbing.IsHash(ref) {
			s.hash = plumbing.NewHash(ref)
			return
		}

		s.ref = ref
	}
}

// WithPollOptions configures the poller used to trigger update procedures
func WithPollOptions(opts ...containers.Option[storagefs.Poller]) containers.Option[SnapshotStore] {
	return func(s *SnapshotStore) {
		s.pollOpts = append(s.pollOpts, opts...)
	}
}

// WithAuth returns an option which configures the auth method used
// by the provided source.
func WithAuth(auth transport.AuthMethod) containers.Option[SnapshotStore] {
	return func(s *SnapshotStore) {
		s.auth = auth
	}
}

// WithInsecureTLS returns an option which configures the insecure TLS
// setting for the provided source.
func WithInsecureTLS(insecureSkipTLS bool) containers.Option[SnapshotStore] {
	return func(s *SnapshotStore) {
		s.insecureSkipTLS = insecureSkipTLS
	}
}

// WithCABundle returns an option which configures the CA Bundle used for
// validating the TLS connection to the provided source.
func WithCABundle(caCertBytes []byte) containers.Option[SnapshotStore] {
	return func(s *SnapshotStore) {
		if caCertBytes != nil {
			s.caBundle = caCertBytes
		}
	}
}

// NewSnapshotStore constructs and configures a Store.
// The store uses the connection and credential details provided to build
// fs.FS implementations around a target git repository.
func NewSnapshotStore(ctx context.Context, logger *zap.Logger, url string, opts ...containers.Option[SnapshotStore]) (_ *SnapshotStore, err error) {
	store := &SnapshotStore{
		Closer: io.NopCloser(nil),
		logger: logger.With(zap.String("repository", url)),
		url:    url,
		ref:    "main",
	}
	containers.ApplyAll(store, opts...)

	field := zap.Stringer("ref", plumbing.NewBranchReferenceName(store.ref))
	if store.hash != plumbing.ZeroHash {
		field = zap.Stringer("SHA", store.hash)
	}
	store.logger = store.logger.With(field)

	store.repo, err = git.Clone(memory.NewStorage(), nil, &git.CloneOptions{
		Auth:            store.auth,
		URL:             store.url,
		CABundle:        store.caBundle,
		InsecureSkipTLS: store.insecureSkipTLS,
	})
	if err != nil {
		return nil, err
	}

	// fetch snapshot at-least once before returning store
	// to ensure we have some state to serve
	if err := store.get(ctx); err != nil {
		return nil, err
	}

	// if the reference is a static hash then it is immutable
	// if we have already fetched it once, there is not point updating again
	if store.hash == plumbing.ZeroHash {
		poller := storagefs.NewPoller(store.logger, ctx, store.update, store.pollOpts...)
		store.Closer = poller
		go poller.Poll()
	}

	return store, nil
}

// String returns an identifier string for the store type.
func (*SnapshotStore) String() string {
	return "git"
}

// View accepts a function which takes a *StoreSnapshot.
// The SnapshotStore will supply a snapshot which is valid
// for the lifetime of the provided function call.
func (s *SnapshotStore) View(fn func(storage.ReadOnlyStore) error) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return fn(s.snap)
}

// update fetches from the remote and given that a the target reference
// HEAD updates to a new revision, it builds a snapshot and updates it
// on the store.
func (s *SnapshotStore) update(ctx context.Context) (bool, error) {
	if err := s.repo.Fetch(&git.FetchOptions{
		Auth: s.auth,
		RefSpecs: []config.RefSpec{
			config.RefSpec(fmt.Sprintf(
				"+%s:%s",
				plumbing.NewBranchReferenceName(s.ref),
				plumbing.NewRemoteReferenceName("origin", s.ref),
			)),
		},
	}); err != nil {
		if !errors.Is(err, git.NoErrAlreadyUpToDate) {
			return false, err
		}

		return false, nil
	}

	if err := s.get(ctx); err != nil {
		return false, err
	}

	return true, nil
}

// get builds a new store snapshot based on the configure Git remote and reference.
func (s *SnapshotStore) get(context.Context) (err error) {
	var fs fs.FS
	if s.hash != plumbing.ZeroHash {
		fs, err = gitfs.NewFromRepoHash(s.logger, s.repo, s.hash)
	} else {
		fs, err = gitfs.NewFromRepo(s.logger, s.repo, gitfs.WithReference(plumbing.NewRemoteReferenceName("origin", s.ref)))
	}
	if err != nil {
		return err
	}

	snap, err := storagefs.SnapshotFromFS(s.logger, fs)
	if err != nil {
		return err
	}

	s.mu.Lock()
	s.snap = snap
	s.mu.Unlock()

	return nil
}
