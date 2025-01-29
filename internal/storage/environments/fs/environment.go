package fs

import (
	"context"
	"os"
	"sync"

	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/internal/containers"
	"go.flipt.io/flipt/internal/server/environments"
	"go.flipt.io/flipt/internal/storage"
	storagefs "go.flipt.io/flipt/internal/storage/fs"
	"go.flipt.io/flipt/internal/storage/fs/local"
	rpcconfig "go.flipt.io/flipt/rpc/v2/environments"
	"go.uber.org/zap"
)

var _ environments.Environment = (*Environment)(nil)

type Environment struct {
	*local.Poller

	logger *zap.Logger
	cfg    *config.EnvironmentConfig

	mu      sync.RWMutex
	src     Filesystem
	storage Storage
	snap    *storagefs.Snapshot

	pollOpts []containers.Option[local.Poller]
}

func NewEnvironment(ctx context.Context, logger *zap.Logger, cfg *config.EnvironmentConfig, src Filesystem, storage Storage, opts ...containers.Option[Environment]) *Environment {
	env := &Environment{
		logger:  logger,
		cfg:     cfg,
		src:     src,
		storage: storage,
	}

	containers.ApplyAll(env, opts...)

	env.Poller = local.NewPoller(logger, ctx, func(ctx context.Context) error { env.update(ctx); return nil }, env.pollOpts...)
	go env.Poll()

	return env
}

// update fetches a new snapshot from the local filesystem
// and updates the current served reference via a write lock
func (e *Environment) update(_ context.Context) {
	if err := func() error {
		src := os.DirFS(e.cfg.Directory)
		conf, err := storagefs.GetConfig(src)
		if err != nil {
			return err
		}

		snap, err := storagefs.SnapshotFromFS(e.logger, conf, src)
		if err != nil {
			return err
		}

		e.mu.Lock()
		defer e.mu.Unlock()

		e.snap = snap

		return nil
	}(); err != nil {
		e.logger.Error("error performing update", zap.Error(err), zap.String("environment", e.cfg.Name))
	}
}

func (e *Environment) Name() string {
	return e.cfg.Name
}

func (s *Environment) GetNamespace(ctx context.Context, key string) (*rpcconfig.NamespaceResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ns, err := s.storage.GetNamespace(ctx, s.src, key)
	if err != nil {
		return nil, err
	}

	return &rpcconfig.NamespaceResponse{
		Namespace: ns,
	}, nil
}

func (s *Environment) ListNamespaces(ctx context.Context) (*rpcconfig.ListNamespacesResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	nl := &rpcconfig.ListNamespacesResponse{}

	items, err := s.storage.ListNamespaces(ctx, s.src)
	if err != nil {
		return nil, err
	}

	nl.Items = append(nl.Items, items...)

	return nl, nil
}

func (s *Environment) CreateNamespace(ctx context.Context, rev string, ns *rpcconfig.Namespace) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	return "", s.storage.PutNamespace(ctx, s.src, ns)
}

func (s *Environment) UpdateNamespace(ctx context.Context, rev string, ns *rpcconfig.Namespace) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	return "", s.storage.PutNamespace(ctx, s.src, ns)
}

func (s *Environment) DeleteNamespace(ctx context.Context, rev string, key string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	return "", s.storage.DeleteNamespace(ctx, s.src, key)
}

func (s *Environment) View(ctx context.Context, typ environments.ResourceType, fn environments.ViewFunc) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rstore, err := s.storage.Resource(typ)
	if err != nil {
		return err
	}

	return fn(ctx, &store{src: s.src, rstore: rstore})
}

func (s *Environment) Update(ctx context.Context, rev string, typ environments.ResourceType, fn environments.UpdateFunc) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	rstore, err := s.storage.Resource(typ)
	if err != nil {
		return "", err
	}

	return "", fn(ctx, &store{src: s.src, rstore: rstore})
}

type store struct {
	src    Filesystem
	rstore ResourceStorage
}

func (s *store) GetResource(ctx context.Context, namespace string, key string) (*rpcconfig.ResourceResponse, error) {
	resource, err := s.rstore.GetResource(ctx, s.src, namespace, key)
	if err != nil {
		return nil, err
	}

	return &rpcconfig.ResourceResponse{
		Resource: resource,
	}, nil
}

func (s *store) ListResources(ctx context.Context, namespace string) (*rpcconfig.ListResourcesResponse, error) {
	rs, err := s.rstore.ListResources(ctx, s.src, namespace)
	if err != nil {
		return nil, err
	}

	return &rpcconfig.ListResourcesResponse{
		Resources: rs,
	}, nil
}

func (s *store) CreateResource(ctx context.Context, r *rpcconfig.Resource) error {
	return s.rstore.PutResource(ctx, s.src, r)
}

func (s *store) UpdateResource(ctx context.Context, r *rpcconfig.Resource) error {
	return s.rstore.PutResource(ctx, s.src, r)
}

func (s *store) DeleteResource(ctx context.Context, namespace string, key string) error {
	return s.rstore.DeleteResource(ctx, s.src, namespace, key)
}

func (e *Environment) EvaluationStore() (storage.ReadOnlyStore, error) {
	return e.snap, nil
}
