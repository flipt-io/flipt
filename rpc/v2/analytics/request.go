package analytics

import (
	"go.flipt.io/flipt/rpc/flipt"
)

// Request authorizes reading flag analytics as a read on the namespace.
func (r *GetFlagEvaluationsCountRequest) Request() []flipt.Request {
	return []flipt.Request{flipt.NewRequest(flipt.ScopeNamespace, flipt.ActionRead, flipt.WithEnvironment(r.EnvironmentKey), flipt.WithNamespace(r.NamespaceKey))}
}

// Request authorizes reading flag analytics as a read on the namespace.
func (r *GetBatchFlagEvaluationsCountRequest) Request() []flipt.Request {
	return []flipt.Request{flipt.NewRequest(flipt.ScopeNamespace, flipt.ActionRead, flipt.WithEnvironment(r.EnvironmentKey), flipt.WithNamespace(r.NamespaceKey))}
}
