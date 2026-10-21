package analytics

import "go.flipt.io/flipt/rpc/flipt"

// Request authorizes reading flag analytics as reading the flag.
func (r *GetFlagEvaluationsCountRequest) Request() []flipt.Request {
	return []flipt.Request{
		flipt.NewRequest(flipt.ResourceFlag, flipt.ActionRead, flipt.WithNamespace(r.NamespaceKey), flipt.WithSubject(flipt.SubjectFlag)),
	}
}
