package environments

import (
	"context"
	"fmt"
	"strings"

	"go.flipt.io/flipt/errors"
	"go.flipt.io/flipt/rpc/v2/environments"
)

type ResourceType struct {
	Package string
	Name    string
}

func NewResourceType(pkg, name string) ResourceType {
	return ResourceType{pkg, name}
}

func ParseResourceType(typ string) (t ResourceType, err error) {
	parts := strings.Split(typ, ".")
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
	Name() string

	// Namespaces

	GetNamespace(_ context.Context, key string) (*environments.NamespaceResponse, error)
	ListNamespaces(context.Context) (*environments.ListNamespacesResponse, error)
	CreateNamespace(_ context.Context, rev string, _ *environments.Namespace) (string, error)
	UpdateNamespace(_ context.Context, rev string, _ *environments.Namespace) (string, error)
	DeleteNamespace(_ context.Context, rev, key string) (string, error)

	// Resources

	View(_ context.Context, typ ResourceType, fn ViewFunc) error
	Update(_ context.Context, rev string, typ ResourceType, fn UpdateFunc) (string, error)
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
	byName map[string]Environment
}

func NewEnvironmentStore(envs ...Environment) *EnvironmentStore {
	store := &EnvironmentStore{
		byName: map[string]Environment{},
	}

	for _, env := range envs {
		store.byName[env.Name()] = env
	}

	return store
}

// Get returns the environment identified on the context via the matching host
func (e *EnvironmentStore) Get(ctx context.Context, name string) (Environment, error) {
	env, ok := e.byName[name]
	if !ok {
		return nil, errors.ErrNotFoundf("environment: %q", name)
	}

	return env, nil
}
