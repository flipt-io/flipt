// Code generated by protoc-gen-go-flipt-sdk. DO NOT EDIT.

package grpc

import (
	flipt "go.flipt.io/flipt/rpc/flipt"
	auth "go.flipt.io/flipt/rpc/flipt/auth"
	evaluation "go.flipt.io/flipt/rpc/flipt/evaluation"
	meta "go.flipt.io/flipt/rpc/flipt/meta"
	_go "go.flipt.io/flipt/sdk/go"
	grpc "google.golang.org/grpc"
)

var _ _go.Transport = Transport{}

type Transport struct {
	cc grpc.ClientConnInterface
}

func NewTransport(cc grpc.ClientConnInterface) Transport {
	return Transport{cc: cc}
}

type authClient struct {
	cc grpc.ClientConnInterface
}

func (t authClient) PublicAuthenticationServiceClient() auth.PublicAuthenticationServiceClient {
	return auth.NewPublicAuthenticationServiceClient(t.cc)
}

func (t authClient) AuthenticationServiceClient() auth.AuthenticationServiceClient {
	return auth.NewAuthenticationServiceClient(t.cc)
}

func (t authClient) AuthenticationMethodTokenServiceClient() auth.AuthenticationMethodTokenServiceClient {
	return auth.NewAuthenticationMethodTokenServiceClient(t.cc)
}

func (t authClient) AuthenticationMethodOIDCServiceClient() auth.AuthenticationMethodOIDCServiceClient {
	return auth.NewAuthenticationMethodOIDCServiceClient(t.cc)
}

func (t authClient) AuthenticationMethodKubernetesServiceClient() auth.AuthenticationMethodKubernetesServiceClient {
	return auth.NewAuthenticationMethodKubernetesServiceClient(t.cc)
}

func (t authClient) AuthenticationMethodOAuthServiceClient() auth.AuthenticationMethodOAuthServiceClient {
	return auth.NewAuthenticationMethodOAuthServiceClient(t.cc)
}

func (t Transport) AuthClient() _go.AuthClient {
	return authClient{cc: t.cc}
}

func (t Transport) EvaluationClient() evaluation.EvaluationServiceClient {
	return evaluation.NewEvaluationServiceClient(t.cc)
}

func (t Transport) FliptClient() flipt.FliptClient {
	return flipt.NewFliptClient(t.cc)
}

func (t Transport) MetaClient() meta.MetadataServiceClient {
	return meta.NewMetadataServiceClient(t.cc)
}
