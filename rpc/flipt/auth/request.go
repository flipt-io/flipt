package auth

import (
	"go.flipt.io/flipt/rpc/flipt"
)

func (req *ListAuthenticationsRequest) Request() []flipt.Request {
	return []flipt.Request{{Scope: flipt.ScopeAuthentication, Action: flipt.ActionRead}}
}

func (req *GetAuthenticationRequest) Request() []flipt.Request {
	return []flipt.Request{{Scope: flipt.ScopeAuthentication, Action: flipt.ActionRead}}
}

func (req *DeleteAuthenticationRequest) Request() []flipt.Request {
	return []flipt.Request{{Scope: flipt.ScopeAuthentication, Action: flipt.ActionDelete}}
}
