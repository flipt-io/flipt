package git

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"text/template"

	"github.com/docker/docker/pkg/namesgenerator"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"go.flipt.io/flipt/internal/config"
	serverenvs "go.flipt.io/flipt/internal/server/environments"
	"go.flipt.io/flipt/internal/storage"
	evaluation "go.flipt.io/flipt/internal/storage/environments/evaluation"
	environmentsfs "go.flipt.io/flipt/internal/storage/environments/fs"
	storagefs "go.flipt.io/flipt/internal/storage/fs"
	storagegit "go.flipt.io/flipt/internal/storage/git"
	rpcenvironments "go.flipt.io/flipt/rpc/v2/environments"
	rpcevaluation "go.flipt.io/flipt/rpc/v2/evaluation"
	"go.uber.org/zap"
)

var _ serverenvs.Environment = (*Environment)(nil)

// Environment is an implementation of the configuration servers Environment interface
// which is backed by a Git repository.
// The repository could be in-memory or on-disk and optionally may push
// to some upstream remote.
type Environment struct {
	logger *zap.Logger

	cfg     *config.EnvironmentConfig
	repo    *storagegit.Repository
	storage environmentsfs.Storage

	mu            sync.RWMutex
	branches      map[string]*Environment
	refs          map[string]string
	currentBranch string
	head          plumbing.Hash
	snap          *storagefs.Snapshot
	publisher     evaluation.SnapshotPublisher
}

// NewEnvironmentFromRepo takes a git repository and a set of typed resource storage implementations and exposes
// the controls necessary to get, list, put and delete both namespaces and their resources
// It optionally roots all changes to a target directory within the source repository
func NewEnvironmentFromRepo(
	ctx context.Context,
	logger *zap.Logger,
	cfg *config.EnvironmentConfig,
	repo *storagegit.Repository,
	storage environmentsfs.Storage,
	publisher evaluation.SnapshotPublisher,
) (_ *Environment, err error) {
	return &Environment{
		logger:        logger,
		cfg:           cfg,
		repo:          repo,
		storage:       storage,
		refs:          map[string]string{},
		snap:          storagefs.EmptySnapshot(),
		publisher:     publisher,
		currentBranch: repo.GetDefaultBranch(),
		branches:      map[string]*Environment{},
	}, nil
}

func (e *Environment) Key() string {
	return e.cfg.Name
}

func (e *Environment) Default() bool {
	return e.cfg.Default
}

func (e *Environment) Repository() *storagegit.Repository {
	return e.repo
}

func (e *Environment) Configuration() *rpcenvironments.EnvironmentConfiguration {
	return &rpcenvironments.EnvironmentConfiguration{
		Remote:    e.repo.GetRemote(),
		Branch:    e.currentBranch,
		Directory: e.cfg.Directory,
	}
}

// Branch creates a new branch from the current environment and returns a new Environment
// that is backed by the new branch.
// The new Environment is added to the branches map and the current branch is updated.
func (e *Environment) Branch(ctx context.Context) (serverenvs.Environment, error) {
	// generate an ID for the branched environment
	name := strings.ReplaceAll(namesgenerator.GetRandomName(0), "_", "")
	branchName := fmt.Sprintf("flipt/%s/%s", e.cfg.Name, name)

	cfg := *e.cfg
	cfg.Name = name

	if err := e.repo.CreateBranchIfNotExists(branchName, storagegit.WithBase(branchName)); err != nil {
		return nil, err
	}

	env, err := NewEnvironmentFromRepo(
		ctx,
		e.logger,
		&cfg,
		e.repo,
		e.storage,
		evaluation.NewNoopPublisher(),
	)

	if err != nil {
		return nil, err
	}

	env.currentBranch = branchName

	e.mu.Lock()
	defer e.mu.Unlock()

	e.branches[cfg.Name] = env

	return env, nil
}

func (e *Environment) Propose(ctx context.Context, branch serverenvs.Environment) (*rpcenvironments.ProposeEnvironmentResponse, error) {
	return nil, errors.New("not implemented")
}

func (e *Environment) GetNamespace(ctx context.Context, key string) (resp *rpcenvironments.NamespaceResponse, err error) {
	err = e.repo.View(ctx, e.currentBranch, func(hash plumbing.Hash, fs environmentsfs.Filesystem) error {
		ns, err := e.storage.GetNamespace(ctx, environmentsfs.SubFilesystem(fs, e.cfg.Directory), key)
		if err != nil {
			return err
		}

		resp = &rpcenvironments.NamespaceResponse{
			Namespace: ns,
			Revision:  hash.String(),
		}

		return nil
	})
	return
}

func (e *Environment) ListNamespaces(ctx context.Context) (resp *rpcenvironments.ListNamespacesResponse, err error) {
	err = e.repo.View(ctx, e.currentBranch, func(hash plumbing.Hash, fs environmentsfs.Filesystem) error {
		items, err := e.storage.ListNamespaces(ctx, environmentsfs.SubFilesystem(fs, e.cfg.Directory))
		if err != nil {
			return err
		}

		resp = &rpcenvironments.ListNamespacesResponse{
			Items:    items,
			Revision: hash.String(),
		}

		return nil
	})
	return
}

func (e *Environment) CreateNamespace(ctx context.Context, rev string, ns *rpcenvironments.Namespace) (string, error) {
	return e.updateNamespace(ctx, rev, func(fs environmentsfs.Filesystem) (storagefs.Change, error) {
		if err := e.storage.PutNamespace(ctx, fs, ns); err != nil {
			return storagefs.Change{}, err
		}

		return storagefs.Change{
			Verb:     storagefs.VerbCreate,
			Resource: storagefs.Resource{Type: serverenvs.NewResourceType("flipt.config", "Namespace"), Key: ns.Key},
		}, nil
	})
}

func (e *Environment) UpdateNamespace(ctx context.Context, rev string, ns *rpcenvironments.Namespace) (string, error) {
	return e.updateNamespace(ctx, rev, func(fs environmentsfs.Filesystem) (storagefs.Change, error) {
		if err := e.storage.PutNamespace(ctx, fs, ns); err != nil {
			return storagefs.Change{}, err
		}

		return storagefs.Change{
			Verb:     storagefs.VerbUpdate,
			Resource: storagefs.Resource{Type: serverenvs.NewResourceType("flipt.config", "Namespace"), Key: ns.Key},
		}, nil
	})
}

func (e *Environment) DeleteNamespace(ctx context.Context, rev, key string) (string, error) {
	return e.updateNamespace(ctx, rev, func(fs environmentsfs.Filesystem) (storagefs.Change, error) {
		if err := e.storage.DeleteNamespace(ctx, fs, key); err != nil {
			return storagefs.Change{}, err
		}

		return storagefs.Change{
			Verb:     storagefs.VerbDelete,
			Resource: storagefs.Resource{Type: serverenvs.NewResourceType("flipt.config", "Namespace"), Key: key},
		}, nil
	})
}

func (e *Environment) updateNamespace(ctx context.Context, rev string, fn func(environmentsfs.Filesystem) (storagefs.Change, error)) (string, error) {
	hash := plumbing.NewHash(rev)
	hash, err := e.repo.UpdateAndPush(ctx, e.currentBranch, func(src environmentsfs.Filesystem) (string, error) {
		// chroot our filesystem to the configured directory
		src = environmentsfs.SubFilesystem(src, e.cfg.Directory)

		conf, err := storagefs.GetConfig(e.logger, environmentsfs.ToFS(src))
		if err != nil {
			return "", err
		}

		change, err := fn(src)
		if err != nil {
			return "", err
		}

		return e.messageForChanges(conf.Templates.CommitMessageTemplate, change)
	}, storagegit.UpdateIfHeadMatches(&hash))
	if err != nil {
		return "", err
	}

	return hash.String(), nil
}

func (e *Environment) View(ctx context.Context, typ serverenvs.ResourceType, fn serverenvs.ViewFunc) error {
	rstore, err := e.storage.Resource(typ)
	if err != nil {
		return fmt.Errorf("git storage view: %w", err)
	}

	if err := e.repo.View(ctx, e.currentBranch, func(hash plumbing.Hash, fs environmentsfs.Filesystem) error {
		return fn(ctx, &store{typ: typ, rstore: rstore, base: hash, fs: environmentsfs.SubFilesystem(fs, e.cfg.Directory)})
	}); err != nil {
		return err
	}

	return nil
}

func (e *Environment) Update(ctx context.Context, rev string, typ serverenvs.ResourceType, fn serverenvs.UpdateFunc) (_ string, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("resource type %q: %w", typ, err)
		}
	}()

	hash := plumbing.NewHash(rev)
	hash, err = e.repo.UpdateAndPush(ctx, e.currentBranch, func(src environmentsfs.Filesystem) (string, error) {
		rstore, err := e.storage.Resource(typ)
		if err != nil {
			return "", fmt.Errorf("git storage update: %w", err)
		}

		// chroot our filesystem to the configured directory
		src = environmentsfs.SubFilesystem(src, e.cfg.Directory)

		conf, err := storagefs.GetConfig(e.logger, environmentsfs.ToFS(src))
		if err != nil {
			return "", err
		}

		store := &store{typ: typ, rstore: rstore, base: hash, fs: src}
		if err := fn(ctx, store); err != nil {
			return "", err
		}

		return e.messageForChanges(conf.Templates.CommitMessageTemplate, store.changes...)
	}, storagegit.UpdateIfHeadMatches(&hash))
	if err != nil {
		return "", err
	}

	return hash.String(), nil
}

func (e *Environment) messageForChanges(tmpl *template.Template, changes ...storagefs.Change) (string, error) {
	if len(changes) == 0 {
		return "", fmt.Errorf("committing and pushing: %w", git.ErrEmptyCommit)
	}

	buf := &bytes.Buffer{}
	if err := tmpl.Execute(buf, struct {
		Environment *config.EnvironmentConfig
		Changes     []storagefs.Change
	}{
		Environment: e.cfg,
		Changes:     changes,
	}); err != nil {
		return "", err
	}

	return buf.String(), nil
}

type store struct {
	typ    serverenvs.ResourceType
	rstore environmentsfs.ResourceStorage

	base    plumbing.Hash
	fs      environmentsfs.Filesystem
	changes []storagefs.Change
}

func (s *store) GetResource(ctx context.Context, namespace string, key string) (*rpcenvironments.ResourceResponse, error) {
	resource, err := s.rstore.GetResource(ctx, s.fs, namespace, key)
	if err != nil {
		return nil, err
	}

	return &rpcenvironments.ResourceResponse{
		Resource: resource,
		Revision: s.base.String(),
	}, nil
}

func (s *store) ListResources(ctx context.Context, namespace string) (*rpcenvironments.ListResourcesResponse, error) {
	rs, err := s.rstore.ListResources(ctx, s.fs, namespace)
	if err != nil {
		return nil, err
	}

	return &rpcenvironments.ListResourcesResponse{
		Resources: rs,
		Revision:  s.base.String(),
	}, nil
}

func (s *store) CreateResource(ctx context.Context, r *rpcenvironments.Resource) error {
	s.changes = append(s.changes, storagefs.Change{
		Verb: storagefs.VerbCreate,
		Resource: storagefs.Resource{
			Type:      s.typ,
			Namespace: r.NamespaceKey,
			Key:       r.Key,
		},
	})

	return s.rstore.PutResource(ctx, s.fs, r)
}

func (s *store) UpdateResource(ctx context.Context, r *rpcenvironments.Resource) error {
	s.changes = append(s.changes, storagefs.Change{
		Verb: storagefs.VerbUpdate,
		Resource: storagefs.Resource{
			Type:      s.typ,
			Namespace: r.NamespaceKey,
			Key:       r.Key,
		},
	})

	return s.rstore.PutResource(ctx, s.fs, r)
}

func (s *store) DeleteResource(ctx context.Context, namespace string, key string) error {
	s.changes = append(s.changes, storagefs.Change{
		Verb: storagefs.VerbDelete,
		Resource: storagefs.Resource{
			Type:      s.typ,
			Namespace: namespace,
			Key:       key,
		},
	})

	return s.rstore.DeleteResource(ctx, s.fs, namespace, key)
}

func (e *Environment) EvaluationStore() (storage.ReadOnlyStore, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	return e.snap, nil
}

func (e *Environment) EvaluationNamespaceSnapshot(ctx context.Context, ns string) (*rpcevaluation.EvaluationNamespaceSnapshot, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	return e.snap.EvaluationNamespaceSnapshot(ctx, ns)
}

func (e *Environment) EvaluationNamespaceSnapshotSubscribe(ctx context.Context, ns string, ch chan<- *rpcevaluation.EvaluationNamespaceSnapshot) (io.Closer, error) {
	return e.publisher.Subscribe(ctx, ns, ch)
}

// Notify is called whenever the tracked branch is fetched and advances
func (e *Environment) Notify(ctx context.Context, head plumbing.Hash) error {
	// check if head has advanced
	if e.head == head {
		// head has not advanced so we skip building
		return nil
	}

	snap, err := e.buildSnapshot(ctx, head)
	if err != nil {
		e.logger.Error("updating snapshot",
			zap.Error(err),
			zap.String("environment", e.cfg.Name))
		return err
	}

	if err := e.publisher.Publish(snap); err != nil {
		e.logger.Error("publishing snapshot",
			zap.Error(err),
			zap.String("environment", e.cfg.Name))
		return err
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	e.head = head
	e.snap = snap

	return nil

}

// buildSnapshot builds a snapshot from the current branch for all namespaces in the environment
func (e *Environment) buildSnapshot(ctx context.Context, hash plumbing.Hash) (snap *storagefs.Snapshot, err error) {
	return snap, e.repo.View(ctx, e.currentBranch, func(hash plumbing.Hash, fs environmentsfs.Filesystem) error {
		if e.cfg.Directory != "" {
			fs = environmentsfs.SubFilesystem(fs, e.cfg.Directory)
		}

		iofs := environmentsfs.ToFS(fs)
		conf, err := storagefs.GetConfig(e.logger, iofs)
		if err != nil {
			return err
		}

		snap, err = storagefs.SnapshotFromFS(e.logger, conf, iofs)
		return err
	}, storagegit.ViewWithHash(hash))
}
