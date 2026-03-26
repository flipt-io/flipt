package environments

import "go.flipt.io/flipt/errors"

const maxBulkApplyTargetPairs = 100

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

func (r *CompareEnvironmentsRequest) Validate() error {
	if r.GetNamespaceKey() == "" {
		return errors.ErrInvalid("namespace_key must not be empty")
	}
	if r.GetTargetEnvironmentKey() == "" {
		return errors.ErrInvalid("target_environment_key must not be empty")
	}
	if r.GetTargetNamespaceKey() == "" {
		return errors.ErrInvalid("target_namespace_key must not be empty")
	}
	return nil
}

func (r *BulkApplyResourcesRequest) Validate() error {
	if r.GetEnvironmentKey() == "" && len(r.GetEnvironmentKeys()) == 0 {
		return errors.ErrInvalid("environment_key or environment_keys must not be empty")
	}

	if len(r.GetNamespaceKeys()) == 0 {
		return errors.ErrInvalid("namespace_keys must not be empty")
	}

	environmentCount := len(r.GetEnvironmentKeys())
	if environmentCount == 0 {
		environmentCount = 1
	}

	if len(r.GetNamespaceKeys())*environmentCount > maxBulkApplyTargetPairs {
		return errors.ErrInvalidf(
			"bulk apply exceeds max target pairs (%d)",
			maxBulkApplyTargetPairs,
		)
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
