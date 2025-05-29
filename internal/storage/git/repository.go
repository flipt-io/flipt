package git

import (
	"context"
	"fmt"
	"maps"
	"os"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/cache"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/storage"
	gitfilesystem "github.com/go-git/go-git/v5/storage/filesystem"
	"github.com/go-git/go-git/v5/storage/memory"
	"go.flipt.io/flipt/errors"
	"go.flipt.io/flipt/internal/containers"
	envsfs "go.flipt.io/flipt/internal/storage/environments/fs"
	"go.uber.org/zap"
)

type Repository struct {
	*git.Repository

	logger *zap.Logger

	mu                 sync.RWMutex
	remote             *config.RemoteConfig
	defaultBranch      string
	auth               transport.AuthMethod
	insecureSkipTLS    bool
	caBundle           []byte
	localPath          string
	readme             []byte
	sigName            string
	sigEmail           string
	maxOpenDescriptors int

	subs []Subscriber

	pollInterval time.Duration
	cancel       func()
	done         chan struct{}

	metrics repoMetrics
}

type Subscriber interface {
	Branches() []string
	Notify(ctx context.Context, heads map[string]string) error
}

func NewRepository(ctx context.Context, logger *zap.Logger, opts ...containers.Option[Repository]) (*Repository, error) {
	repo, empty, err := newRepository(ctx, logger, opts...)
	if err != nil {
		return nil, err
	}

	if empty {
		logger.Debug("repository empty, attempting to add and push a README")
		// add initial readme if repo is empty
		if _, err := repo.UpdateAndPush(ctx, repo.defaultBranch, func(fs envsfs.Filesystem) (string, error) {
			fi, err := fs.OpenFile("README.md", os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
			if err != nil {
				return "", err
			}

			if _, err := fi.Write(repo.readme); err != nil {
				return "", err
			}

			if err := fi.Close(); err != nil {
				return "", err
			}

			return "add initial README", nil
		}, UpdateWithInitialCommit); err != nil {
			return nil, err
		}
	}

	repo.startPolling(ctx)

	return repo, nil
}

// newRepository is a wrapper around the core *git.Repository
// It handles configuring a repository source appropriately based on our configuration
// It also exposes some common operations and ensures safe concurrent access while fetching and pushing
func newRepository(ctx context.Context, logger *zap.Logger, opts ...containers.Option[Repository]) (_ *Repository, empty bool, err error) {
	r := &Repository{
		logger:        logger,
		defaultBranch: "main",
		readme:        []byte(defaultReadmeContents),
		// we initialize with a noop function incase
		// we dont start the polling loop
		cancel: func() {},
		done:   make(chan struct{}),
	}

	containers.ApplyAll(r, opts...)

	var metricsOpts []containers.Option[repoMetrics]

	// we initially assume the repo is empty because we start
	// with an in-memory blank slate
	empty = true
	storage := (storage.Storer)(memory.NewStorage())
	r.Repository, err = git.InitWithOptions(storage, nil, git.InitOptions{
		DefaultBranch: plumbing.NewBranchReferenceName(r.defaultBranch),
	})
	if err != nil {
		return nil, empty, err
	}

	if r.localPath != "" {
		storage = gitfilesystem.NewStorageWithOptions(osfs.New(r.localPath), cache.NewObjectLRUDefault(), gitfilesystem.Options{
			MaxOpenDescriptors: r.maxOpenDescriptors,
		})

		entries, err := os.ReadDir(r.localPath)
		if empty = err != nil || len(entries) == 0; empty {
			// either its empty or there was an error opening the file
			if err != nil && !errors.Is(err, os.ErrNotExist) {
				return nil, empty, err
			}

			r.Repository, err = git.InitWithOptions(storage, nil, git.InitOptions{
				DefaultBranch: plumbing.NewBranchReferenceName(r.defaultBranch),
			})
			if err != nil {
				return nil, empty, err
			}
		} else {
			// opened successfully and there is contents so we assume not empty
			r.Repository, err = git.Open(storage, nil)
			if err != nil {
				return nil, empty, err
			}
		}
	}

	if r.remote != nil {
		if len(r.remote.URLs) == 0 {
			return nil, empty, errors.New("must supply at-least one remote URL")
		}

		metricsOpts = append(metricsOpts, withRemote(r.remote.URLs[0]))

		if _, err = r.CreateRemote(r.remote); err != nil {
			if !errors.Is(err, git.ErrRemoteExists) {
				return nil, empty, err
			}
		}

		// given an upstream has been configured we're going to start
		// by changing our assumption to the repository having contents
		empty = false

		// do an initial fetch to setup remote tracking branches
		if err := r.Fetch(ctx); err != nil {
			if !errors.Is(err, transport.ErrEmptyRemoteRepository) &&
				!errors.Is(err, git.NoMatchingRefSpecError{}) {
				return nil, empty, fmt.Errorf("performing initial fetch: %w", err)
			}

			// the remote was reachable but either its contents was completely empty
			// or our default branch doesn't exist and so we decide to seed it
			empty = true

			logger.Debug("initial fetch empty", zap.String("reference", r.defaultBranch), zap.Error(err))
		}
	}

	r.metrics = newRepoMetrics(metricsOpts...)

	if plumbing.IsHash(r.defaultBranch) {
		// if we still need to add an initial commit to the repository then we assume they couldn't
		// have predicted the initial hash and return reference not found
		if empty {
			return nil, empty, fmt.Errorf("target repository is empty: %w", plumbing.ErrReferenceNotFound)
		}

		return r, empty, r.Storer.SetReference(plumbing.NewHashReference(plumbing.HEAD, plumbing.NewHash(r.defaultBranch)))
	}

	return r, empty, nil
}

func (r *Repository) startPolling(ctx context.Context) {
	if r.pollInterval == 0 {
		close(r.done)
		return
	}

	go func() {
		defer close(r.done)

		ticker := time.NewTicker(r.pollInterval)
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := r.Fetch(ctx); err != nil {
					r.metrics.recordPollError(ctx)
					r.logger.Error("error performing fetch", zap.Error(err))
					continue
				}

				r.logger.Debug("fetch successful")
			}
		}
	}()
}

func (r *Repository) GetRemote() string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.remote == nil {
		return ""
	}

	return r.remote.URLs[0]
}

func (r *Repository) GetDefaultBranch() string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.defaultBranch
}

func (r *Repository) Close() error {
	r.cancel()

	<-r.done

	return nil
}

func (r *Repository) Subscribe(sub Subscriber) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.subs = append(r.subs, sub)
}

func (r *Repository) fetchHeads() []string {
	heads := map[string]struct{}{r.defaultBranch: {}}
	for _, sub := range r.subs {
		for _, head := range sub.Branches() {
			heads[head] = struct{}{}
		}
	}

	return slices.Collect(maps.Keys(heads))
}

// Fetch does a fetch for the requested head names on a configured remote.
// If the remote is not defined, then it is a silent noop.
// Iff specific is explicitly requested then only the heads in specific are fetched.
// Otherwise, it fetches all previously tracked head references.
func (r *Repository) Fetch(ctx context.Context, specific ...string) (err error) {
	if r.remote == nil {
		return nil
	}

	updatedRefs := map[string]plumbing.Hash{}
	r.mu.Lock()
	defer func() {
		r.mu.Unlock()
		r.updateSubs(ctx, updatedRefs)
	}()

	heads := specific
	if len(heads) == 0 {
		heads = r.fetchHeads()
	}

	var refSpecs = []config.RefSpec{}

	for _, head := range heads {
		refSpec := config.RefSpec(
			fmt.Sprintf("+%s:%s",
				plumbing.NewBranchReferenceName(head),
				plumbing.NewRemoteReferenceName(r.remote.Name, head),
			),
		)

		r.logger.Debug("preparing refspec for fetch", zap.Stringer("refspec", refSpec))

		refSpecs = append(refSpecs, refSpec)
	}

	if err := r.FetchContext(ctx, &git.FetchOptions{
		RemoteName:      r.remote.Name,
		Auth:            r.auth,
		CABundle:        r.caBundle,
		InsecureSkipTLS: r.insecureSkipTLS,
		RefSpecs:        refSpecs,
	}); err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
		return err
	}

	allRefs, err := r.References()
	if err != nil {
		return err
	}

	if err := allRefs.ForEach(func(ref *plumbing.Reference) error {
		// we're only interested in updates to remotes
		if !ref.Name().IsRemote() {
			return nil
		}

		for _, head := range heads {
			name := strings.TrimPrefix(ref.Name().String(), "refs/remotes/origin/")
			if refMatch(name, head) {
				updatedRefs[name] = ref.Hash()
			}
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}

// ViewOptions are options for the View method.
type ViewOptions struct {
	hash plumbing.Hash
}

// ViewWithHash configures a call to View with a specific hash.
func ViewWithHash(hash plumbing.Hash) containers.Option[ViewOptions] {
	return func(vo *ViewOptions) {
		vo.hash = hash
	}
}

// View reads the head of the given branch and passes the resulting git tree via
// the envsfs.Filesystem abstraction to the provided function.
func (r *Repository) View(
	ctx context.Context,
	branch string,
	fn func(hash plumbing.Hash, fs envsfs.Filesystem) error,
	opts ...containers.Option[ViewOptions],
) (err error) {
	var vopts ViewOptions
	containers.ApplyAll(&vopts, opts...)

	finished := r.metrics.recordView(ctx, branch)

	r.mu.RLock()
	defer func() {
		r.mu.RUnlock()

		finished(err)
	}()

	hash := vopts.hash
	if hash == plumbing.ZeroHash {
		hash, err = r.Resolve(branch)
		if err != nil {
			return err
		}
	}

	r.logger.Debug("view", zap.String("branch", branch), zap.Stringer("hash", hash))

	fs, err := r.newFilesystem(hash)
	if err != nil {
		return err
	}

	return fn(hash, fs)
}

type UpdateAndPushOptions struct {
	initialCommit bool
	ifHeadMatches *plumbing.Hash
}

// UpdateWithInitialCommit configures a call to UpdateAndPush to intentionally
// create an initial commit
func UpdateWithInitialCommit(uapo *UpdateAndPushOptions) {
	uapo.initialCommit = true
}

// UpdateIfHeadMatches predicates that an update should return an error early if the target branch
// does not match the supplied hash.
// This allows for updates to attempt a form of optimistic update and retry in the case of a conflict.
func UpdateIfHeadMatches(hash *plumbing.Hash) containers.Option[UpdateAndPushOptions] {
	return func(uapo *UpdateAndPushOptions) {
		uapo.ifHeadMatches = hash
	}
}

// UpdateAndPush calls the provided function with a Filesystem implementation which intercepts any write
// operations and builds the changes into a commit.
// Given an upstream remote is configured, the commit is also pushed to the given branch.
func (r *Repository) UpdateAndPush(
	ctx context.Context,
	branch string,
	fn func(fs envsfs.Filesystem) (string, error),
	opts ...containers.Option[UpdateAndPushOptions],
) (hash plumbing.Hash, err error) {
	var (
		finished = r.metrics.recordUpdate(ctx, branch)
		options  UpdateAndPushOptions
		commit   *object.Commit
	)

	containers.ApplyAll(&options, opts...)

	r.logger.Debug("starting update and push operation",
		zap.String("branch", branch),
		zap.Bool("initialCommit", options.initialCommit),
		zap.Any("remoteConfig", r.remote))

	r.mu.Lock()
	defer func() {
		r.mu.Unlock()
		if commit != nil {
			// update references
			r.updateSubs(ctx, map[string]plumbing.Hash{branch: commit.Hash})
		}
		finished(err)
	}()

	if !options.initialCommit {
		// for non initial commits we start by resolving the current head
		hash, err = r.Resolve(branch)
		if err != nil {
			return plumbing.ZeroHash, err
		}

		r.logger.Debug("resolved current head",
			zap.String("branch", branch),
			zap.Stringer("hash", hash))
	}

	if options.ifHeadMatches != nil && !options.ifHeadMatches.IsZero() && *options.ifHeadMatches != hash {
		r.logger.Warn("head revision mismatch",
			zap.Stringer("expected", *options.ifHeadMatches),
			zap.Stringer("actual", hash))
		return hash, errors.ErrConflictf("expected head revision %q has changed (now %q)", *options.ifHeadMatches, hash)
	}

	// if rev == nil then hash will be the zero hash
	fs, err := r.newFilesystem(hash)
	if err != nil {
		return hash, err
	}

	msg, err := fn(fs)
	if err != nil {
		return hash, err
	}

	r.logger.Debug("creating commit", zap.String("message", msg))
	commit, err = fs.commit(ctx, msg)
	if err != nil {
		return hash, err
	}

	r.logger.Debug("commit created successfully", zap.Stringer("hash", commit.Hash))

	if r.remote != nil {
		local := plumbing.NewBranchReferenceName(branch)
		r.logger.Debug("setting local reference",
			zap.Stringer("reference", local),
			zap.Stringer("hash", commit.Hash))

		if err := r.Storer.SetReference(plumbing.NewHashReference(local, commit.Hash)); err != nil {
			return hash, err
		}

		r.logger.Debug("pushing to remote",
			zap.String("remoteName", r.remote.Name),
			zap.Strings("remoteURLs", r.remote.URLs),
			zap.Stringer("refSpec", config.RefSpec(fmt.Sprintf("%[1]s:%[1]s", local))))

		if err := r.PushContext(ctx, &git.PushOptions{
			RemoteName:      r.remote.Name,
			Auth:            r.auth,
			CABundle:        r.caBundle,
			InsecureSkipTLS: r.insecureSkipTLS,
			RefSpecs: []config.RefSpec{
				config.RefSpec(fmt.Sprintf("%[1]s:%[1]s", local)),
			},
		}); err != nil {
			return hash, err
		}
	} else {
		r.logger.Debug("skipping push - no remote configured")
	}

	remoteName := "origin"
	if r.remote != nil {
		remoteName = r.remote.Name
	}

	// update remote tracking reference to match
	remoteRef := plumbing.NewHashReference(
		plumbing.NewRemoteReferenceName(remoteName, branch),
		commit.Hash)

	r.logger.Debug("setting remote tracking reference",
		zap.Stringer("reference", remoteRef.Name()),
		zap.Stringer("hash", commit.Hash))

	if err := r.Storer.SetReference(remoteRef); err != nil {
		return hash, err
	}

	r.logger.Debug("update and push completed successfully",
		zap.String("branch", branch),
		zap.Stringer("hash", commit.Hash))

	return commit.Hash, nil
}

func (r *Repository) updateSubs(ctx context.Context, refs map[string]plumbing.Hash) {
	// update subscribers for each matching ref
	for _, sub := range r.subs {
		matched := map[string]string{}
		for ref, hash := range refs {
			for _, branch := range sub.Branches() {
				if refMatch(ref, branch) {
					matched[ref] = hash.String()
				}
			}
		}

		if err := sub.Notify(ctx, matched); err != nil {
			r.metrics.recordUpdateSubsError(ctx)

			r.logger.Error("while updating subscriber", zap.Error(err))
		}
	}
}

func refMatch(ref, pattern string) bool {
	if !strings.Contains(pattern, "*") {
		return ref == pattern
	}

	return strings.HasPrefix(ref, pattern[:strings.Index(pattern, "*")]) //nolint:gocritic
}

func (r *Repository) ResolveHead() (plumbing.Hash, error) {
	return r.Resolve(r.defaultBranch)
}

func (r *Repository) Resolve(branch string) (plumbing.Hash, error) {
	reference, err := r.Repository.Reference(plumbing.NewRemoteReferenceName("origin", branch), true)
	if err != nil {
		return plumbing.ZeroHash, err
	}

	return reference.Hash(), nil
}

type CreateBranchOptions struct {
	base string
}

func WithBase(name string) containers.Option[CreateBranchOptions] {
	return func(cbo *CreateBranchOptions) {
		cbo.base = name
	}
}

func (r *Repository) CreateBranchIfNotExists(branch string, opts ...containers.Option[CreateBranchOptions]) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	remoteName := "origin"
	if r.remote != nil {
		remoteName = r.remote.Name
	}

	remoteRef := plumbing.NewRemoteReferenceName(remoteName, branch)
	if _, err := r.Reference(remoteRef, true); err == nil {
		// reference already exists
		return nil
	}

	opt := CreateBranchOptions{base: r.defaultBranch}

	containers.ApplyAll(&opt, opts...)

	reference, err := r.Repository.Reference(plumbing.NewRemoteReferenceName(remoteName, opt.base), true)
	if err != nil {
		return fmt.Errorf("base reference %q not found: %w", opt.base, err)
	}

	return r.Storer.SetReference(plumbing.NewHashReference(remoteRef,
		reference.Hash()))
}

func (r *Repository) DeleteBranch(ctx context.Context, branch string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Delete local branch
	localRef := plumbing.NewBranchReferenceName(branch)
	if err := r.Storer.RemoveReference(localRef); err != nil {
		return fmt.Errorf("failed to delete local branch: %w", err)
	}

	// Delete remote branch if remote is configured
	if r.remote != nil {
		refSpec := config.RefSpec(fmt.Sprintf(":%s", plumbing.NewBranchReferenceName(branch)))
		err := r.PushContext(ctx, &git.PushOptions{
			RemoteName:      r.remote.Name,
			Auth:            r.auth,
			CABundle:        r.caBundle,
			InsecureSkipTLS: r.insecureSkipTLS,
			RefSpecs:        []config.RefSpec{refSpec},
		})
		if err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
			return fmt.Errorf("failed to delete remote branch: %w", err)
		}
	}

	return nil
}

func (r *Repository) newFilesystem(hash plumbing.Hash) (_ *filesystem, err error) {
	return newFilesystem(
		r.logger,
		r.Storer,
		withSignature(r.sigName, r.sigEmail),
		withBaseCommit(hash),
	)
}

func WithRemote(name, url string) containers.Option[Repository] {
	return func(r *Repository) {
		r.remote = &config.RemoteConfig{
			Name: "origin",
			URLs: []string{url},
		}
	}
}

// WithDefaultBranch configures the default branch used to initially seed
// the repo, or base other branches on when they're not already present
// in the upstream.
func WithDefaultBranch(ref string) containers.Option[Repository] {
	return func(s *Repository) {
		s.defaultBranch = ref
	}
}

// WithAuth returns an option which configures the auth method used
// by the provided source.
func WithAuth(auth transport.AuthMethod) containers.Option[Repository] {
	return func(s *Repository) {
		s.auth = auth
	}
}

// WithInsecureTLS returns an option which configures the insecure TLS
// setting for the provided source.
func WithInsecureTLS(insecureSkipTLS bool) containers.Option[Repository] {
	return func(s *Repository) {
		s.insecureSkipTLS = insecureSkipTLS
	}
}

// WithCABundle returns an option which configures the CA Bundle used for
// validating the TLS connection to the provided source.
func WithCABundle(caCertBytes []byte) containers.Option[Repository] {
	return func(s *Repository) {
		if caCertBytes != nil {
			s.caBundle = caCertBytes
		}
	}
}

// WithFilesystemStorage configures the Git repository to clone into
// the local filesystem, instead of the default which is in-memory.
// The provided path is location for the dotgit folder.
func WithFilesystemStorage(path string) containers.Option[Repository] {
	return func(r *Repository) {
		r.localPath = path
	}
}

// WithSignature sets the default signature name and email when the signature
// cannot be derived from the request context.
func WithSignature(name, email string) containers.Option[Repository] {
	return func(r *Repository) {
		r.sigName = name
		r.sigEmail = email
	}
}

// WithInterval sets the period between automatic fetches from the upstream (if a remote is configured)
func WithInterval(interval time.Duration) containers.Option[Repository] {
	return func(r *Repository) {
		r.pollInterval = interval
	}
}

// WithMaxOpenDescriptors sets the maximum number of open file descriptors when using filesystem backed storage
func WithMaxOpenDescriptors(n int) containers.Option[Repository] {
	return func(r *Repository) {
		r.maxOpenDescriptors = n
	}
}

const defaultReadmeContents = `Flipt Configuration Repository
==============================

This repository contains Flipt feature flag configuration.
Each directory containing a file named features.yaml represents a namespace.`
