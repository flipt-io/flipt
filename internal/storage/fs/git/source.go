package git

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"strings"
	"time"

	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/storage"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/gofrs/uuid"
	"go.flipt.io/flipt/internal/containers"
	"go.flipt.io/flipt/internal/ext"
	"go.flipt.io/flipt/internal/gitfs"
	storagefs "go.flipt.io/flipt/internal/storage/fs"
	"go.flipt.io/flipt/rpc/flipt"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

var _ storagefs.Source = (*Source)(nil)

// Source is an implementation of storage/fs.FSSource
// This implementation is backed by a Git repository and it tracks an upstream reference.
// When subscribing to this source, the upstream reference is tracked
// by polling the upstream on a configurable interval.
type Source struct {
	logger *zap.Logger
	repo   *git.Repository
	store  storage.Storer

	url      string
	ref      string
	hash     plumbing.Hash
	interval time.Duration
	auth     transport.AuthMethod
}

// WithRef configures the target reference to be used when fetching
// and building fs.FS implementations.
// If it is a valid hash, then the fixed SHA value is used.
// Otherwise, it is treated as a reference in the origin upstream.
func WithRef(ref string) containers.Option[Source] {
	return func(s *Source) {
		if plumbing.IsHash(ref) {
			s.hash = plumbing.NewHash(ref)
			return
		}

		s.ref = ref
	}
}

// WithPollInterval configures the interval in which origin is polled to
// discover any updates to the target reference.
func WithPollInterval(tick time.Duration) containers.Option[Source] {
	return func(s *Source) {
		s.interval = tick
	}
}

// WithAuth returns an option which configures the auth method used
// by the provided source.
func WithAuth(auth transport.AuthMethod) containers.Option[Source] {
	return func(s *Source) {
		s.auth = auth
	}
}

// NewSource constructs and configures a Source.
// The source uses the connection and credential details provided to build
// fs.FS implementations around a target git repository.
func NewSource(logger *zap.Logger, url string, opts ...containers.Option[Source]) (_ *Source, err error) {
	source := &Source{
		logger:   logger.With(zap.String("repository", url)),
		url:      url,
		ref:      "main",
		interval: 30 * time.Second,
	}
	containers.ApplyAll(source, opts...)

	field := zap.Stringer("ref", plumbing.NewBranchReferenceName(source.ref))
	if source.hash != plumbing.ZeroHash {
		field = zap.Stringer("SHA", source.hash)
	}
	source.logger = source.logger.With(field)

	source.store = memory.NewStorage()
	source.repo, err = git.Clone(source.store, nil, &git.CloneOptions{
		Auth: source.auth,
		URL:  source.url,
	})
	if err != nil {
		return nil, err
	}

	return source, nil
}

// String returns an identifier string for the store type.
func (*Source) String() string {
	return "git"
}

// Get builds a new *storagefs.Snapshot based on the configure Git remote and reference.
func (s *Source) Get() (_ *storagefs.Snapshot, err error) {
	var fs fs.FS
	if s.hash != plumbing.ZeroHash {
		fs, err = gitfs.NewFromRepoHash(s.logger, s.repo, s.hash)
		if err != nil {
			return nil, err
		}
	} else {
		fs, err = gitfs.NewFromRepo(s.logger, s.repo, gitfs.WithReference(plumbing.NewRemoteReferenceName("origin", s.ref)))
		if err != nil {
			return nil, err
		}
	}

	return storagefs.SnapshotFromFS(s.logger, fs)
}

// Subscribe feeds gitfs implementations of fs.FS onto the provided channel.
// It blocks until the provided context is cancelled (it will be called in a goroutine).
// It closes the provided channel before it returns.
func (s *Source) Subscribe(ctx context.Context, ch chan<- *storagefs.Snapshot) {
	defer close(ch)

	// NOTE: theres is no point subscribing to updates for a git Hash
	// as it is atomic and will never change.
	if s.hash != plumbing.ZeroHash {
		s.logger.Info("skipping subscribe as static SHA has been configured")
		return
	}

	ticker := time.NewTicker(s.interval)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.logger.Debug("fetching from remote")
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
				if errors.Is(err, git.NoErrAlreadyUpToDate) {
					s.logger.Debug("store already up to date")
					continue
				}

				s.logger.Error("failed fetching remote", zap.Error(err))
				continue
			}

			snap, err := s.Get()
			if err != nil {
				s.logger.Error("failed creating gitfs", zap.Error(err))
				continue
			}

			ch <- snap

			s.logger.Debug("finished fetching from remote")
		}
	}
}

func (s *Source) Propose(ctx context.Context, r *flipt.ProposeRequest) (*flipt.Proposal, error) {
	// validate revision
	if !plumbing.IsHash(r.Revision) {
		return nil, fmt.Errorf("ref is not valid hash: %q", r.Revision)
	}

	hash := plumbing.NewHash(r.Revision)

	fs, err := gitfs.NewFromRepoHash(s.logger, s.repo, hash)
	if err != nil {
		return nil, err
	}

	files, err := storagefs.ListStateFiles(s.logger, fs)
	if err != nil {
		return nil, err
	}

	s.logger.Debug("opening state files for writing", zap.Strings("paths", files))

	// retrieve referenced namespace ext.Document types (indexed by fi name)
	var docs map[string]document
	for _, file := range files {
		fi, err := fs.Open(file)
		if err != nil {
			return nil, err
		}

		defer fi.Close()

		doc := new(ext.Document)
		if err := yaml.NewDecoder(fi).Decode(doc); err != nil {
			return nil, err
		}

		// set namespace to default if empty in document
		if doc.Namespace == "" {
			doc.Namespace = "default"
		}

		docs[doc.Namespace] = document{doc, file}
	}

	// open repository on store with in-memory workspace
	repo, err := git.Open(s.store, memfs.New())
	if err != nil {
		return nil, err
	}

	work, err := repo.Worktree()
	if err != nil {
		return nil, err
	}

	proposal := &flipt.Proposal{
		Id: uuid.Must(uuid.NewV4()).String(),
	}

	// create proposal branch (flipt/proposal/$id)
	branch := fmt.Sprintf("flipt/proposal/%s", proposal.Id)
	if err := work.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(branch),
		Create: true,
		Hash:   hash,
	}); err != nil {
		return nil, err
	}

	// for each requested change
	// make it to the associated document
	// and then add and commit the difference
	for _, req := range r.Requests {
		document, message, err := update(docs, req)
		if err != nil {
			return nil, err
		}

		if document.Document == nil {
			if _, err := work.Remove(document.path); err != nil {
				return nil, err
			}
		} else {
			fi, err := work.Filesystem.OpenFile(document.path, os.O_CREATE|os.O_TRUNC|os.O_RDWR, os.ModePerm)
			if err != nil {
				return nil, err
			}
			defer fi.Close()

			if err = yaml.NewEncoder(fi).Encode(document.Document); err != nil {
				return nil, err
			}

			_, err = work.Add(document.path)
			if err != nil {
				return nil, err
			}
		}

		_, err = work.Commit(message, &git.CommitOptions{})
		if err != nil {
			return nil, err
		}
	}

	// push to proposed branch
	repo.PushContext(ctx, &git.PushOptions{
		Auth: s.auth,
		RefSpecs: []config.RefSpec{
			config.RefSpec(fmt.Sprintf("%s:%s", branch, branch)),
		},
	})

	// open PR

	return proposal, nil
}

type document struct {
	*ext.Document
	path string
}

func update(docs map[string]document, req *flipt.AnyRequest) (document, string, error) {
	switch r := req.Request.(type) {
	case *flipt.AnyRequest_CreateNamespace:
		_, ok := docs[r.CreateNamespace.Key]
		if ok {
			return document{}, r.CreateNamespace.Key, fmt.Errorf("namespace %q already exists", r.CreateNamespace.Key)
		}

		document := document{
			Document: &ext.Document{
				Namespace: r.CreateNamespace.Key,
			},
			path: fmt.Sprintf("%s.features.yml", strings.ToLower(r.CreateNamespace.Key)),
		}

		return document, fmt.Sprintf("feat(flipt): create namespace %q", r.CreateNamespace.Key), nil
	case *flipt.AnyRequest_UpdateNamespace:
		// TODO(georgemac): figure out what to do with update namespace
		doc, err := getNamespace(docs, r.UpdateNamespace.Key)
		if err != nil {
			return document{}, "", err
		}

		return doc, "", nil
	case *flipt.AnyRequest_DeleteNamespace:
		doc, err := getNamespace(docs, r.DeleteNamespace.Key)
		if err != nil {
			return document{}, "", err
		}

		// empty document will signify delete file
		return document{path: doc.path}, fmt.Sprintf("feat(flipt): delete namespace %q", r.DeleteNamespace.Key), nil
	case *flipt.AnyRequest_CreateFlag:
		doc, err := getNamespace(docs, r.CreateFlag.NamespaceKey)
		if err != nil {
			return document{}, "", err
		}

		fqn := path.Join(doc.Namespace, r.CreateFlag.Key)

		_, _, ok := find(func(f *ext.Flag) bool {
			return f.Key == r.CreateFlag.Key
		}, doc.Flags...)
		if ok {
			return document{}, "", fmt.Errorf("flag %q already exists", fqn)
		}

		doc.Flags = append(doc.Flags, &ext.Flag{
			Name:        r.CreateFlag.Name,
			Description: r.CreateFlag.Description,
		})

		return doc, fmt.Sprintf("feat(flipt): create flag %q", fqn), nil
	case *flipt.AnyRequest_UpdateFlag:
		doc, err := getNamespace(docs, r.UpdateFlag.NamespaceKey)
		if err != nil {
			return document{}, "", err
		}

		fqn := path.Join(doc.Namespace, r.UpdateFlag.Key)

		flag, _, ok := find(func(f *ext.Flag) bool {
			return f.Key == r.UpdateFlag.Key
		}, doc.Flags...)
		if !ok {
			return document{}, "", fmt.Errorf("flag %q does not exist", fqn)
		}

		flag.Name = r.UpdateFlag.Name
		flag.Description = r.UpdateFlag.Description

		return doc, fmt.Sprintf("feat(flipt): update flag %q", fqn), nil
	case *flipt.AnyRequest_DeleteFlag:
		doc, err := getNamespace(docs, r.DeleteFlag.NamespaceKey)
		if err != nil {
			return document{}, "", err
		}

		fqn := path.Join(doc.Namespace, r.DeleteFlag.Key)

		var ok bool
		doc.Flags, ok = remove(func(f *ext.Flag) bool {
			return f.Key == r.DeleteFlag.Key
		}, doc.Flags...)
		if !ok {
			return document{}, "", fmt.Errorf("flag %q does not exist", fqn)
		}

		return doc, fmt.Sprintf("feat(flipt): delete flag %q", fqn), nil
	case *flipt.AnyRequest_CreateVariant:
	case *flipt.AnyRequest_UpdateVariant:
	case *flipt.AnyRequest_DeleteVariant:
	case *flipt.AnyRequest_CreateRule:
	case *flipt.AnyRequest_UpdateRule:
	case *flipt.AnyRequest_OrderRules:
	case *flipt.AnyRequest_DeleteRule:
	case *flipt.AnyRequest_CreateDistribution:
	case *flipt.AnyRequest_UpdateDistribution:
	case *flipt.AnyRequest_DeleteDistribution:
	case *flipt.AnyRequest_CreateSegment:
	case *flipt.AnyRequest_UpdateSegment:
	case *flipt.AnyRequest_DeleteSegment:
	case *flipt.AnyRequest_CreateConstraint:
	case *flipt.AnyRequest_UpdateConstraint:
	case *flipt.AnyRequest_DeleteConstraint:
	}

	return document{}, "", fmt.Errorf("unexpected type %T", req.Request)
}

func getNamespace(docs map[string]document, key string) (document, error) {
	doc, ok := docs[key]
	if !ok {
		return document{}, fmt.Errorf("namespace %q does not exist", key)
	}

	return doc, nil
}

func remove[T any](match func(T) bool, ts ...T) ([]T, bool) {
	_, idx, ok := find(match, ts...)
	if !ok {
		return ts, false
	}

	return append(ts[0:idx], ts[idx+1:]...), true
}

func find[T any](match func(T) bool, ts ...T) (t T, _ int, _ bool) {
	for i, t := range ts {
		if match(t) {
			return t, i, true
		}
	}

	return t, -1, false
}
