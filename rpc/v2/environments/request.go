package environments

import (
	"go.flipt.io/flipt/rpc/flipt"
)

func (r *ListEnvironmentsRequest) Request() []flipt.Request {
	return []flipt.Request{flipt.NewRequest(flipt.ScopeEnvironment, flipt.ActionRead, flipt.WithNoEnvironment())}
}

func (r *BranchEnvironmentRequest) Request() []flipt.Request {
	return []flipt.Request{flipt.NewRequest(flipt.ScopeEnvironment, flipt.ActionRead, flipt.WithEnvironment(r.EnvironmentKey))}
}

func (r *DeleteBranchEnvironmentRequest) Request() []flipt.Request {
	return []flipt.Request{flipt.NewRequest(flipt.ScopeEnvironment, flipt.ActionRead, flipt.WithEnvironment(r.EnvironmentKey))}
}

func (r *ListEnvironmentBranchesRequest) Request() []flipt.Request {
	return []flipt.Request{flipt.NewRequest(flipt.ScopeEnvironment, flipt.ActionRead, flipt.WithEnvironment(r.EnvironmentKey))}
}

func (r *ListBranchedEnvironmentChangesRequest) Request() []flipt.Request {
	return []flipt.Request{flipt.NewRequest(flipt.ScopeEnvironment, flipt.ActionRead, flipt.WithEnvironment(r.EnvironmentKey))}
}

func (r *ProposeEnvironmentRequest) Request() []flipt.Request {
	return []flipt.Request{flipt.NewRequest(flipt.ScopeEnvironment, flipt.ActionRead, flipt.WithEnvironment(r.EnvironmentKey))}
}

func (r *GetNamespaceRequest) Request() []flipt.Request {
	return []flipt.Request{flipt.NewRequest(flipt.ScopeEnvironment, flipt.ActionRead, flipt.WithEnvironment(r.EnvironmentKey), flipt.WithNamespace(r.Key))}
}

func (r *ListNamespacesRequest) Request() []flipt.Request {
	return []flipt.Request{flipt.NewRequest(flipt.ScopeEnvironment, flipt.ActionRead, flipt.WithEnvironment(r.EnvironmentKey), flipt.WithNoNamespace())}
}

func (r *UpdateNamespaceRequest) Request() []flipt.Request {
	return []flipt.Request{flipt.NewRequest(flipt.ScopeEnvironment, flipt.ActionUpdate, flipt.WithEnvironment(r.EnvironmentKey), flipt.WithNamespace(r.Key))}
}

func (r *DeleteNamespaceRequest) Request() []flipt.Request {
	return []flipt.Request{flipt.NewRequest(flipt.ScopeEnvironment, flipt.ActionDelete, flipt.WithEnvironment(r.EnvironmentKey), flipt.WithNamespace(r.Key))}
}

func (r *GetResourceRequest) Request() []flipt.Request {
	return []flipt.Request{flipt.NewRequest(flipt.ScopeNamespace, flipt.ActionRead, flipt.WithEnvironment(r.EnvironmentKey), flipt.WithNamespace(r.NamespaceKey))}
}

func (r *ListResourcesRequest) Request() []flipt.Request {
	return []flipt.Request{flipt.NewRequest(flipt.ScopeNamespace, flipt.ActionRead, flipt.WithEnvironment(r.EnvironmentKey), flipt.WithNamespace(r.NamespaceKey))}
}

func (r *UpdateResourceRequest) Request() []flipt.Request {
	return []flipt.Request{flipt.NewRequest(flipt.ScopeNamespace, flipt.ActionUpdate, flipt.WithEnvironment(r.EnvironmentKey), flipt.WithNamespace(r.NamespaceKey))}
}

func (r *DeleteResourceRequest) Request() []flipt.Request {
	return []flipt.Request{flipt.NewRequest(flipt.ScopeNamespace, flipt.ActionDelete, flipt.WithEnvironment(r.EnvironmentKey), flipt.WithNamespace(r.NamespaceKey))}
}
