package evaluation

import "go.flipt.io/flipt/rpc/flipt"

func (x *EvaluationRequest) Request() []flipt.Request {
	return []flipt.Request{
		flipt.NewRequest(flipt.ScopeNamespace, flipt.ActionEvaluate,
			flipt.WithEnvironment(x.GetEnvironmentKey()),
			flipt.WithNamespace(x.GetNamespaceKey())),
	}
}

func (x *BatchEvaluationRequest) Request() []flipt.Request {
	requests := make([]flipt.Request, 0, len(x.GetRequests()))
	for _, r := range x.GetRequests() {
		requests = append(requests, flipt.NewRequest(flipt.ScopeNamespace, flipt.ActionEvaluate,
			flipt.WithEnvironment(r.GetEnvironmentKey()),
			flipt.WithNamespace(r.GetNamespaceKey())))
	}
	return requests
}
