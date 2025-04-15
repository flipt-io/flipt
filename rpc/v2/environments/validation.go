package environments

import "go.flipt.io/flipt/errors"

func (r *GetNamespaceRequest) Validate() error {
	if err := requireEnvironment(r); err != nil {
		return err
	}

	return requireKey(r)
}

func (r *UpdateNamespaceRequest) Validate() error {
	if err := requireEnvironment(r); err != nil {
		return err
	}

	return requireKey(r)
}

func (r *DeleteNamespaceRequest) Validate() error {
	if err := requireEnvironment(r); err != nil {
		return err
	}

	return requireKey(r)
}

func (r *GetResourceRequest) Validate() error {
	return requireResource(r)
}

func (r *ListResourcesRequest) Validate() error {
	return requireNamespace(r)
}

func (r *UpdateResourceRequest) Validate() error {
	return requireResource(r)
}

func (r *DeleteResourceRequest) Validate() error {
	return requireResource(r)
}

func requireEnvironment(r environmented) error {
	if r.GetEnvironmentKey() == "" {
		return errors.ErrInvalid("environment key must not be empty")
	}

	return nil
}

func requireNamespace(r namespaced) error {
	if r.GetNamespaceKey() == "" {
		return errors.ErrInvalid("namespace key must not be empty")
	}

	return nil
}

func requireKey(k keyed) error {
	if k.GetKey() == "" {
		return errors.ErrInvalid("key must not be empty")
	}

	return nil
}

func requireResource(r resourced) error {
	if err := requireEnvironment(r); err != nil {
		return err
	}

	if err := requireNamespace(r); err != nil {
		return err
	}

	return requireKey(r)
}

type environmented interface {
	GetEnvironmentKey() string
}

type namespaced interface {
	GetNamespaceKey() string
}

type keyed interface {
	GetKey() string
}

type resourced interface {
	environmented
	namespaced
	keyed
}
