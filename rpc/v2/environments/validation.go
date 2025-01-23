package configuration

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
	return requireType(r)
}

func (r *UpdateResourceRequest) Validate() error {
	return requireResource(r)
}

func (r *DeleteResourceRequest) Validate() error {
	return requireResource(r)
}

func requireKey(k keyed) error {
	if k.GetKey() == "" {
		return errors.ErrInvalid("key must not be empty")
	}

	return nil
}

func requireType(r typed) error {
	if r.GetNamespace() == "" {
		return errors.ErrInvalid("namespace must not be empty")
	}

	if r.GetType() == "" {
		return errors.ErrInvalid("type must not be empty")
	}

	return nil
}

func requireResource(r interface {
	keyed
	typed
}) error {
	if err := requireType(r); err != nil {
		return err
	}

	return requireKey(r)
}

type keyed interface {
	GetKey() string
}

type typed interface {
	GetType() string
	GetNamespace() string
}
