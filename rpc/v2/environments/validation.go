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
	return requireKey(r)
}

func (r *UpdateResourceRequest) Validate() error {
	return requireKey(r)
}

func (r *DeleteResourceRequest) Validate() error {
	return requireKey(r)
}

func requireKey(k keyed) error {
	if k.GetKey() == "" {
		return errors.ErrInvalid("key must not be empty")
	}

	return nil
}

type keyed interface {
	GetKey() string
}
