package evaluation

import "go.flipt.io/flipt/rpc/flipt"

func (r *EvaluationNamespaceSnapshotRequest) Request() []flipt.Request {
	return []flipt.Request{
		flipt.NewRequest(flipt.ScopeNamespace, flipt.ActionEvaluate, flipt.WithEnvironment(r.EnvironmentKey), flipt.WithNamespace(r.Key)),
	}
}

func (r *EvaluationNamespaceSnapshotRequest) GetNamespaceKey() string {
	return r.Key
}

func (r *EvaluationNamespaceSnapshotStreamRequest) Request() []flipt.Request {
	return []flipt.Request{
		flipt.NewRequest(flipt.ScopeNamespace, flipt.ActionEvaluate, flipt.WithEnvironment(r.EnvironmentKey), flipt.WithNamespace(r.Key)),
	}
}

func (r *EvaluationNamespaceSnapshotStreamRequest) GetNamespaceKey() string {
	return r.Key
}
