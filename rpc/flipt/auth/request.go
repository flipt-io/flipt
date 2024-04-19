package auth

import (
	"go.flipt.io/flipt/rpc/flipt"
)

func (req *CreateTokenRequest) Request() flipt.Request {
	return flipt.NewRequest(flipt.SubjectToken, flipt.ActionCreate)
}
