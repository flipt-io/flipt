package client

import "go.flipt.io/flipt/rpc/flipt"

func (r *EvaluationNamespaceSnapshotRequest) Request() []flipt.Request {
	return []flipt.Request{
		flipt.NewRequest(flipt.ScopeNamespace, flipt.ActionRead, flipt.WithEnvironment(r.EnvironmentKey), flipt.WithNamespace(r.Key)),
	}
}

func (r *EvaluationNamespaceSnapshotRequest) GetNamespaceKey() string {
	return r.Key
}
