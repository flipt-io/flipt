package environments

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

// GetTypeUrl delegates to the underlying payloads GetTypeUrl method.
func (r *UpdateResourceRequest) GetTypeUrl() string {
	return r.Payload.GetTypeUrl()
}

func (r *UpdateResourceRequest) Request() []flipt.Request {
	return []flipt.Request{flipt.NewRequest(resourceFrom(r), flipt.ActionUpdate, flipt.WithNamespace(r.Namespace))}
}

func (r *DeleteResourceRequest) Request() []flipt.Request {
	return []flipt.Request{flipt.NewRequest(resourceFrom(r), flipt.ActionDelete, flipt.WithNamespace(r.Namespace))}
}

func resourceFrom(t typed) flipt.Resource {
	switch t.GetTypeUrl() {
	case "flipt.core.Flag":
		return flipt.ResourceFlag
	case "flipt.core.Segment":
		return flipt.ResourceSegment
	default:
		return flipt.ResourceUnknown
	}
}
