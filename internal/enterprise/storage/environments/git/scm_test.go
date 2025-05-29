package git

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/errors"
	"go.flipt.io/flipt/internal/server/environments"
)

func TestSCMNotImplemented_Propose(t *testing.T) {
	scm := &SCMNotImplemented{}
	res, err := scm.Propose(context.Background(), ProposalRequest{})
	assert.Nil(t, res)
	require.Error(t, err)
	assert.ErrorIs(t, err, errors.ErrNotImplemented("SCM not implemented"))
}

func TestSCMNotImplemented_ListChanges(t *testing.T) {
	scm := &SCMNotImplemented{}
	res, err := scm.ListChanges(context.Background(), ListChangesRequest{})
	assert.Nil(t, res)
	require.Error(t, err)
	assert.ErrorIs(t, err, errors.ErrNotImplemented("SCM not implemented"))
}

func TestSCMNotImplemented_ListProposals(t *testing.T) {
	scm := &SCMNotImplemented{}
	mockEnv := &environments.MockEnvironment{}
	res, err := scm.ListProposals(context.Background(), mockEnv)
	assert.Nil(t, res)
	require.Error(t, err)
	assert.ErrorIs(t, err, errors.ErrNotImplemented("SCM not implemented"))
}
