package git

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"sync"
	"text/template"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/internal/gitfs"
	serverenvs "go.flipt.io/flipt/internal/server/environments"
	"go.flipt.io/flipt/internal/storage"
	environmentsfs "go.flipt.io/flipt/internal/storage/environments/fs"
	storagefs "go.flipt.io/flipt/internal/storage/fs"
	storagegit "go.flipt.io/flipt/internal/storage/git"
	rpcenvironments "go.flipt.io/flipt/rpc/v2/environments"
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

	mu   sync.RWMutex
	refs map[string]string

	head plumbing.Hash
	snap *storagefs.Snapshot
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
) (_ *Environment, err error) {
	return &Environment{
		logger:  logger,
		cfg:     cfg,
		repo:    repo,
		storage: storage,
		refs:    map[string]string{},
		snap:    storagefs.EmptySnapshot(),
	}, nil
}

func (e *Environment) Name() string {
	return e.cfg.Name
}

func (e *Environment) GetNamespace(ctx context.Context, key string) (resp *rpcenvironments.NamespaceResponse, err error) {
	err = e.repo.View(ctx, func(hash plumbing.Hash, fs environmentsfs.Filesystem) error {
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
	err = e.repo.View(ctx, func(hash plumbing.Hash, fs environmentsfs.Filesystem) error {
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
	hash, err := e.repo.UpdateAndPush(ctx, func(src environmentsfs.Filesystem) (string, error) {
		// chroot our filesystem to the configured directory
		src = environmentsfs.SubFilesystem(src, e.cfg.Directory)

		conf, err := storagefs.GetConfig(environmentsfs.ToFS(src))
		if err != nil {
			return "", err
		}

		change, err := fn(src)
		if err != nil {
			return "", err
		}

		return e.messageForChanges(conf.Templates.CommitMessageTemplate, change)
	}, storagegit.IfHeadMatches(&hash))
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

	if err := e.repo.View(ctx, func(hash plumbing.Hash, fs environmentsfs.Filesystem) error {
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
	hash, err = e.repo.UpdateAndPush(ctx, func(src environmentsfs.Filesystem) (string, error) {
		rstore, err := e.storage.Resource(typ)
		if err != nil {
			return "", fmt.Errorf("git storage update: %w", err)
		}

		// chroot our filesystem to the configured directory
		src = environmentsfs.SubFilesystem(src, e.cfg.Directory)

		conf, err := storagefs.GetConfig(environmentsfs.ToFS(src))
		if err != nil {
			return "", err
		}

		store := &store{typ: typ, rstore: rstore, base: hash, fs: src}
		if err := fn(ctx, store); err != nil {
			return "", err
		}

		return e.messageForChanges(conf.Templates.CommitMessageTemplate, store.changes...)
	}, storagegit.IfHeadMatches(&hash))
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
			Namespace: r.Namespace,
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
			Namespace: r.Namespace,
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
	return e.snap, nil
}

// Notify is called whenever the tracked branch is fetched and advances
func (e *Environment) Notify(ctx context.Context, head plumbing.Hash) error {
	// check if head has advanced
	if e.head == head {
		// head has not advanced so we skip building
		return nil
	}

	snap, err := e.buildSnapshot(head)
	if err != nil {
		e.logger.Error("updating snapshot",
			zap.Error(err),
			zap.String("environment", e.cfg.Name))
		return err
	}

	e.head = head
	e.snap = snap

	return nil

}

func (e *Environment) buildSnapshot(hash plumbing.Hash) (*storagefs.Snapshot, error) {
	var gfs fs.FS
	gfs, err := gitfs.NewFromRepoHash(e.logger, e.repo.Repository, hash)
	if err != nil {
		return nil, err
	}

	if e.cfg.Directory != "" {
		gfs, err = fs.Sub(gfs, e.cfg.Directory)
		if err != nil {
			return nil, err
		}
	}

	conf, err := storagefs.GetConfig(gfs)
	if err != nil {
		return nil, err
	}

	return storagefs.SnapshotFromFS(e.logger, conf, gfs)
}
