package git

import (
	"context"

	"go.flipt.io/flipt/errors"
	"go.flipt.io/flipt/rpc/v2/environments"
)

type SCMNotImplemented struct{}

func (s *SCMNotImplemented) Propose(ctx context.Context, req ProposalRequest) (*environments.ProposeEnvironmentResponse, error) {
	return nil, errors.ErrNotImplemented("SCM not implemented")
}

func (s *SCMNotImplemented) ListChanges(ctx context.Context, req ListChangesRequest) (*environments.ListBranchedEnvironmentChangesResponse, error) {
	return nil, errors.ErrNotImplemented("SCM not implemented")
}
