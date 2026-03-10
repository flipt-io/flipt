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

func (r *CopyNamespaceRequest) Validate() error {
	if r.GetNamespaceKey() == "" {
		return errors.ErrInvalid("namespace_key must not be empty")
	}
	return nil
}

func (r *BulkApplyResourcesRequest) Validate() error {
	if len(r.GetNamespaceKeys()) == 0 {
		return errors.ErrInvalid("namespace_keys must not be empty")
	}
	return nil
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
