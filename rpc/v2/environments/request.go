package configuration

import (
	"go.flipt.io/flipt/rpc/flipt"
)

func (r *GetNamespaceRequest) Request() []flipt.Request {
	return []flipt.Request{flipt.NewRequest(flipt.ResourceNamespace, flipt.ActionRead, flipt.WithNamespace(r.Key))}
}

func (r *ListNamespacesRequest) Request() []flipt.Request {
	return []flipt.Request{flipt.NewRequest(flipt.ResourceNamespace, flipt.ActionRead)}
}

func (r *UpdateNamespaceRequest) Request() []flipt.Request {
	return []flipt.Request{flipt.NewRequest(flipt.ResourceNamespace, flipt.ActionUpdate, flipt.WithNamespace(r.Key))}
}

func (r *DeleteNamespaceRequest) Request() []flipt.Request {
	return []flipt.Request{flipt.NewRequest(flipt.ResourceNamespace, flipt.ActionDelete, flipt.WithNamespace(r.Key))}
}

func (r *GetResourceRequest) Request() []flipt.Request {
	return []flipt.Request{flipt.NewRequest(resourceFrom(r), flipt.ActionRead, flipt.WithNamespace(r.Namespace))}
}

func (r *ListResourcesRequest) Request() []flipt.Request {
	return []flipt.Request{flipt.NewRequest(resourceFrom(r), flipt.ActionRead, flipt.WithNamespace(r.Namespace))}
}

func (r *UpdateResourceRequest) Request() []flipt.Request {
	return []flipt.Request{flipt.NewRequest(resourceFrom(r), flipt.ActionUpdate, flipt.WithNamespace(r.Namespace))}
}

func (r *DeleteResourceRequest) Request() []flipt.Request {
	return []flipt.Request{flipt.NewRequest(resourceFrom(r), flipt.ActionDelete, flipt.WithNamespace(r.Namespace))}
}

func (r *ListEnvironmentsRequest) Request() []flipt.Request {
	return []flipt.Request{flipt.NewRequest(flipt.ResourceEnvironment, flipt.ActionRead)}
}

func (r *GetCurrentEnvironmentRequest) Request() []flipt.Request {
	return []flipt.Request{flipt.NewRequest(flipt.ResourceEnvironment, flipt.ActionRead)}
}

func (r *BranchEnvironmentRequest) Request() []flipt.Request {
	return []flipt.Request{flipt.NewRequest(flipt.ResourceEnvironment, flipt.ActionRead)}
}

func (r *ListEnvironmentBranchesRequest) Request() []flipt.Request {
	return []flipt.Request{flipt.NewRequest(flipt.ResourceEnvironment, flipt.ActionRead)}
}

func (r *ProposeEvironmentRequest) Request() []flipt.Request {
	return []flipt.Request{flipt.NewRequest(flipt.ResourceEnvironment, flipt.ActionRead)}
}

func (r *ListEnvironmentChangesRequest) Request() []flipt.Request {
	return []flipt.Request{flipt.NewRequest(flipt.ResourceEnvironment, flipt.ActionRead)}
}

func (r *ListCurrentEnvironmentChangesRequest) Request() []flipt.Request {
	return []flipt.Request{flipt.NewRequest(flipt.ResourceEnvironment, flipt.ActionRead)}
}

func (r *NotifySourceRequest) Request() []flipt.Request {
	return []flipt.Request{}
}

func resourceFrom(t typed) flipt.Resource {
	switch t.GetType() {
	case "flipt.core.Flag":
		return flipt.ResourceFlag
	case "flipt.core.Segment":
		return flipt.ResourceSegment
	default:
		return flipt.ResourceUnknown
	}
}
