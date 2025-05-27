package git

import (
	"context"

	"go.flipt.io/flipt/errors"
	serverenvs "go.flipt.io/flipt/internal/server/environments"
	"go.flipt.io/flipt/rpc/v2/environments"
)

type SCMNotImplemented struct{}

func (s *SCMNotImplemented) Propose(ctx context.Context, req ProposalRequest) (*environments.EnvironmentProposalDetails, error) {
	return nil, errors.ErrNotImplemented("SCM not implemented")
}

func (s *SCMNotImplemented) ListChanges(ctx context.Context, req ListChangesRequest) (*environments.ListBranchedEnvironmentChangesResponse, error) {
	return nil, errors.ErrNotImplemented("SCM not implemented")
}

func (s *SCMNotImplemented) ListProposals(ctx context.Context, env serverenvs.Environment) (map[string]*environments.EnvironmentProposalDetails, error) {
	return nil, errors.ErrNotImplemented("SCM not implemented")
}
