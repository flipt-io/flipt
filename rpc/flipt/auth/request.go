package auth

import (
	"go.flipt.io/flipt/rpc/flipt"
)

func (req *CreateTokenRequest) Request() flipt.Request {
	return flipt.NewRequest(flipt.ResourceAuthentication, flipt.ActionCreate, flipt.WithSubject(flipt.SubjectToken))
}
