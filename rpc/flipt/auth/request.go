package auth

import (
	"go.flipt.io/flipt/rpc/flipt"
)

func (req *CreateTokenRequest) Request() []flipt.Request {
	return []flipt.Request{flipt.NewRequest(flipt.ResourceAuthentication, flipt.ActionCreate, flipt.WithSubject(flipt.SubjectToken))}
}

func (req *ListAuthenticationsRequest) Request() []flipt.Request {
	return []flipt.Request{flipt.NewRequest(flipt.ResourceAuthentication, flipt.ActionRead)}
}

func (req *GetAuthenticationRequest) Request() []flipt.Request {
	return []flipt.Request{flipt.NewRequest(flipt.ResourceAuthentication, flipt.ActionRead)}
}

func (req *DeleteAuthenticationRequest) Request() []flipt.Request {
	return []flipt.Request{flipt.NewRequest(flipt.ResourceAuthentication, flipt.ActionDelete)}
}
