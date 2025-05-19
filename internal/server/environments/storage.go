package environments

import (
	"context"
	"fmt"
	"io"
	"iter"
	"strings"
	"sync"

	"go.flipt.io/flipt/errors"
	"go.flipt.io/flipt/internal/common"
	"go.flipt.io/flipt/internal/storage"
	"go.flipt.io/flipt/rpc/flipt"
	"go.flipt.io/flipt/rpc/v2/environments"
	"go.flipt.io/flipt/rpc/v2/evaluation"
	"go.uber.org/zap"
)

type ResourceType struct {
	Package string
	Name    string
}

func NewResourceType(pkg, name string) ResourceType {
	return ResourceType{pkg, name}
}

func ParseResourceType(typ string) (t ResourceType, err error) {
	parts := strings.Split(strings.TrimPrefix(typ, "type.googleapis.com/"), ".")
	if len(parts) == 0 {
		return t, fmt.Errorf("unexpected package type %q", typ)
	}

	return ResourceType{
		Package: strings.Join(parts[:len(parts)-1], "."),
		Name:    parts[len(parts)-1],
	}, nil
}

func (r ResourceType) String() string {
	return fmt.Sprintf("%s.%s", r.Package, r.Name)
}

type Environment interface {
	Key() string
	Default() bool
	Configuration() *environments.EnvironmentConfiguration
	Branch(ctx context.Context) (Environment, error)
	Propose(ctx context.Context, branch Environment) (*environments.ProposeEnvironmentResponse, error)

	// Namespaces

	GetNamespace(_ context.Context, key string) (*environments.NamespaceResponse, error)
	ListNamespaces(context.Context) (*environments.ListNamespacesResponse, error)
	CreateNamespace(_ context.Context, rev string, _ *environments.Namespace) (string, error)
	UpdateNamespace(_ context.Context, rev string, _ *environments.Namespace) (string, error)
	DeleteNamespace(_ context.Context, rev, key string) (string, error)

	// Resources

	View(_ context.Context, typ ResourceType, fn ViewFunc) error
	Update(_ context.Context, rev string, typ ResourceType, fn UpdateFunc) (string, error)

	// Evaluation

	EvaluationStore() (storage.ReadOnlyStore, error)
	EvaluationNamespaceSnapshot(context.Context, string) (*evaluation.EvaluationNamespaceSnapshot, error)
	EvaluationNamespaceSnapshotSubscribe(context.Context, string, chan<- *evaluation.EvaluationNamespaceSnapshot) (io.Closer, error)
}

type ViewFunc func(context.Context, ResourceStoreView) error

type UpdateFunc func(context.Context, ResourceStore) error

type ResourceStoreView interface {
	GetResource(_ context.Context, namespace, key string) (*environments.ResourceResponse, error)
	ListResources(_ context.Context, namespace string) (*environments.ListResourcesResponse, error)
}

type ResourceStore interface {
	ResourceStoreView

	CreateResource(context.Context, *environments.Resource) error
	UpdateResource(context.Context, *environments.Resource) error
	DeleteResource(_ context.Context, namespace, key string) error
}

type EnvironmentStore struct {
	logger     *zap.Logger
	byKey      map[string]Environment
	defaultEnv Environment
	mu         sync.RWMutex
}

func NewEnvironmentStore(logger *zap.Logger, envs ...Environment) (*EnvironmentStore, error) {
	store := &EnvironmentStore{
		logger: logger,
		byKey:  map[string]Environment{},
	}

	for _, env := range envs {
		store.byKey[env.Key()] = env
		if env.Default() {
			store.defaultEnv = env
		}
	}

	if store.defaultEnv == nil {
		env, ok := store.byKey[flipt.DefaultEnvironment]
		switch {
		case ok:
			store.defaultEnv = env
		case len(envs) == 1:
			store.defaultEnv = envs[0]
		default:
			return nil, errors.New("explicit default environment required")
		}
	}

	return store, nil
}

func (e *EnvironmentStore) List(ctx context.Context) iter.Seq[Environment] {
	e.mu.RLock()
	defer e.mu.RUnlock()

	return iter.Seq[Environment](func(yield func(Environment) bool) {
		for _, env := range e.byKey {
			if !yield(env) {
				return
			}
		}
	})
}

func (e *EnvironmentStore) Add(env Environment) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.byKey[env.Key()] = env
	return nil
}

func (e *EnvironmentStore) Branch(ctx context.Context, base string) (Environment, error) {
	baseEnv, err := e.Get(ctx, base)
	if err != nil {
		return nil, err
	}

	branchEnv, err := baseEnv.Branch(ctx)
	if err != nil {
		return nil, err
	}

	e.Add(branchEnv)

	return branchEnv, nil
}

func (e *EnvironmentStore) Propose(ctx context.Context, base string, branch string) (*environments.ProposeEnvironmentResponse, error) {
	baseEnv, err := e.Get(ctx, base)
	if err != nil {
		return nil, err
	}

	branchEnv, err := e.Get(ctx, branch)
	if err != nil {
		return nil, err
	}

	return baseEnv.Propose(ctx, branchEnv)
}

// Get returns the environment identified by key.
func (e *EnvironmentStore) Get(ctx context.Context, key string) (Environment, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	env, ok := e.byKey[key]
	if !ok {
		return nil, errors.ErrNotFoundf("environment: %q", key)
	}

	return env, nil
}

// GetFromContext returns the environment identified by name from the context or the default environment if no name is provided.
func (e *EnvironmentStore) GetFromContext(ctx context.Context) Environment {
	env, ok := common.FliptEnvironmentFromContext(ctx)
	if ok {
		ee, err := e.Get(ctx, env)
		if err != nil {
			e.logger.Error("failed to get environment from context", zap.String("environment", env), zap.Error(err))
			return e.defaultEnv
		}
		return ee
	}

	return e.defaultEnv
}
