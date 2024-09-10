package evaluation

import "go.flipt.io/flipt/rpc/flipt"

func (r *EvaluationNamespaceSnapshotRequest) Request() []flipt.Request {
	return []flipt.Request{
		flipt.NewRequest(flipt.ResourceFlag, flipt.ActionRead, flipt.WithNamespace(r.Key)),
		flipt.NewRequest(flipt.ResourceSegment, flipt.ActionRead, flipt.WithNamespace(r.Key)),
	}
}

func (r *EvaluationNamespaceSnapshotRequest) GetNamespaceKey() string {
	return r.Key
}
