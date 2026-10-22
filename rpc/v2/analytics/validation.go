package analytics

import "go.flipt.io/flipt/errors"

// Validate requires the environment and namespace, so the scope authorized
// (which defaults blank keys) is the scope queried.
func (r *GetFlagEvaluationsCountRequest) Validate() error {
	return requireScope(r)
}

// Validate requires the environment and namespace, so the scope authorized
// (which defaults blank keys) is the scope queried.
func (r *GetBatchFlagEvaluationsCountRequest) Validate() error {
	return requireScope(r)
}

func requireScope(s scoped) error {
	if s.GetEnvironmentKey() == "" {
		return errors.ErrInvalid("environment_key must not be empty")
	}

	if s.GetNamespaceKey() == "" {
		return errors.ErrInvalid("namespace_key must not be empty")
	}

	return nil
}

type scoped interface {
	GetEnvironmentKey() string
	GetNamespaceKey() string
}
