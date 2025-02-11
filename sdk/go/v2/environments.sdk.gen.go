// Code generated by protoc-gen-go-flipt-sdk. DO NOT EDIT.

package sdk

import (
	context "context"
	environments "go.flipt.io/flipt/rpc/v2/environments"
)

type Environments struct {
	transport              environments.EnvironmentsServiceClient
	authenticationProvider ClientAuthenticationProvider
}

func (x *Environments) ListEnvironments(ctx context.Context, v *environments.ListEnvironmentsRequest) (*environments.ListEnvironmentsResponse, error) {
	ctx, err := authenticate(ctx, x.authenticationProvider)
	if err != nil {
		return nil, err
	}
	return x.transport.ListEnvironments(ctx, v)
}

func (x *Environments) GetNamespace(ctx context.Context, v *environments.GetNamespaceRequest) (*environments.NamespaceResponse, error) {
	ctx, err := authenticate(ctx, x.authenticationProvider)
	if err != nil {
		return nil, err
	}
	return x.transport.GetNamespace(ctx, v)
}

func (x *Environments) ListNamespaces(ctx context.Context, v *environments.ListNamespacesRequest) (*environments.ListNamespacesResponse, error) {
	ctx, err := authenticate(ctx, x.authenticationProvider)
	if err != nil {
		return nil, err
	}
	return x.transport.ListNamespaces(ctx, v)
}

func (x *Environments) CreateNamespace(ctx context.Context, v *environments.UpdateNamespaceRequest) (*environments.NamespaceResponse, error) {
	ctx, err := authenticate(ctx, x.authenticationProvider)
	if err != nil {
		return nil, err
	}
	return x.transport.CreateNamespace(ctx, v)
}

func (x *Environments) UpdateNamespace(ctx context.Context, v *environments.UpdateNamespaceRequest) (*environments.NamespaceResponse, error) {
	ctx, err := authenticate(ctx, x.authenticationProvider)
	if err != nil {
		return nil, err
	}
	return x.transport.UpdateNamespace(ctx, v)
}

func (x *Environments) DeleteNamespace(ctx context.Context, v *environments.DeleteNamespaceRequest) (*environments.DeleteNamespaceResponse, error) {
	ctx, err := authenticate(ctx, x.authenticationProvider)
	if err != nil {
		return nil, err
	}
	return x.transport.DeleteNamespace(ctx, v)
}

func (x *Environments) GetResource(ctx context.Context, v *environments.GetResourceRequest) (*environments.ResourceResponse, error) {
	ctx, err := authenticate(ctx, x.authenticationProvider)
	if err != nil {
		return nil, err
	}
	return x.transport.GetResource(ctx, v)
}

func (x *Environments) ListResources(ctx context.Context, v *environments.ListResourcesRequest) (*environments.ListResourcesResponse, error) {
	ctx, err := authenticate(ctx, x.authenticationProvider)
	if err != nil {
		return nil, err
	}
	return x.transport.ListResources(ctx, v)
}

func (x *Environments) CreateResource(ctx context.Context, v *environments.UpdateResourceRequest) (*environments.ResourceResponse, error) {
	ctx, err := authenticate(ctx, x.authenticationProvider)
	if err != nil {
		return nil, err
	}
	return x.transport.CreateResource(ctx, v)
}

func (x *Environments) UpdateResource(ctx context.Context, v *environments.UpdateResourceRequest) (*environments.ResourceResponse, error) {
	ctx, err := authenticate(ctx, x.authenticationProvider)
	if err != nil {
		return nil, err
	}
	return x.transport.UpdateResource(ctx, v)
}

func (x *Environments) DeleteResource(ctx context.Context, v *environments.DeleteResourceRequest) (*environments.DeleteResourceResponse, error) {
	ctx, err := authenticate(ctx, x.authenticationProvider)
	if err != nil {
		return nil, err
	}
	return x.transport.DeleteResource(ctx, v)
}
