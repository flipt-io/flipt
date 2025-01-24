package environments

import "go.flipt.io/flipt/errors"

func (r *GetNamespaceRequest) Validate() error {
	return requireKey(r)
}

func (r *UpdateNamespaceRequest) Validate() error {
	return requireKey(r)
}

func (r *DeleteNamespaceRequest) Validate() error {
	return requireKey(r)
}

func (r *GetResourceRequest) Validate() error {
	return requireResource(r)
}

func (r *ListResourcesRequest) Validate() error {
	if err := requireType(r); err != nil {
		return err
	}

	return requireNamespace(r)
}

func (r *UpdateResourceRequest) Validate() error {
	return requireResource(r)
}

func (r *DeleteResourceRequest) Validate() error {
	return requireResource(r)
}

func requireKey(k keyed) error {
	if k.GetEnvironment() == "" {
		return errors.ErrInvalid("environment must not be empty")
	}

	if k.GetKey() == "" {
		return errors.ErrInvalid("key must not be empty")
	}

	return nil
}

func requireType(r typed) error {
	if r.GetTypeUrl() == "" {
		return errors.ErrInvalid("type must not be empty")
	}

	return nil
}

func requireNamespace(r namespaced) error {
	if r.GetNamespace() == "" {
		return errors.ErrInvalid("namespace must not be empty")
	}

	return nil
}

func requireResource(r interface {
	keyed
	namespaced
	typed
}) error {
	if err := requireType(r); err != nil {
		return err
	}

	if err := requireNamespace(r); err != nil {
		return err
	}

	return requireKey(r)
}

type keyed interface {
	GetEnvironment() string
	GetKey() string
}

type typed interface {
	GetTypeUrl() string
}

type namespaced interface {
	GetNamespace() string
}
