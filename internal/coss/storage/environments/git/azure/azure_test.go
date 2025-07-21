package azure

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	azuregit "github.com/microsoft/azure-devops-go-api/azuredevops/v7/git"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/coss/storage/environments/git"
	serverenvsmock "go.flipt.io/flipt/internal/server/environments"
	rpcenv "go.flipt.io/flipt/rpc/v2/environments"
	"go.uber.org/zap"
)

func toPtr[T any](t testing.TB, p T) *T {
	t.Helper()
	return &p
}

func TestSCM_Propose(t *testing.T) {
	mockClient := NewMockClient(t)
	scm := &SCM{
		logger:    zap.NewNop(),
		repoOwner: "owner",
		repoName:  "repo",
		client:    mockClient,
	}

	ctx := context.Background()
	req := git.ProposalRequest{
		Base:  "main",
		Head:  "feature",
		Title: "Test PR",
		Body:  "This is a test",
		Draft: false,
	}

	mockPR := &azuregit.GitPullRequest{Url: toPtr(t, "http://example.com/pr"), Status: toPtr(t, azuregit.PullRequestStatusValues.Active)}
	mockClient.EXPECT().CreatePullRequest(mock.Anything, mock.Anything).Return(mockPR, nil)

	result, err := scm.Propose(ctx, req)
	require.NoError(t, err)
	assert.Equal(t, "http://example.com/pr", result.Url)
	assert.Equal(t, rpcenv.ProposalState_PROPOSAL_STATE_OPEN, result.State)
}

func TestSCM_Propose_Error(t *testing.T) {
	mockClient := NewMockClient(t)
	scm := &SCM{
		logger:    zap.NewNop(),
		repoOwner: "owner",
		repoName:  "repo",
		client:    mockClient,
	}

	ctx := context.Background()
	req := git.ProposalRequest{}

	mockClient.EXPECT().CreatePullRequest(mock.Anything, mock.Anything).Return(nil, errors.New("create error"))

	result, err := scm.Propose(ctx, req)
	require.Error(t, err)
	assert.Nil(t, result)
}

func TestSCM_ListChanges(t *testing.T) {
	mockClient := NewMockClient(t)
	scm := &SCM{
		logger:      zap.NewNop(),
		repoOwner:   "owner",
		repoName:    "repo",
		repoProject: "project",
		client:      mockClient,
	}

	ctx := context.Background()
	req := git.ListChangesRequest{Base: "main", Head: "feature", Limit: 1}

	commitTime := time.Now()
	commit := azuregit.GitCommitRef{
		CommitId: toPtr(t, "sha"),
		Comment:  toPtr(t, "commit message"),
		Author: &azuregit.GitUserDate{
			Email: toPtr(t, "author@example.com"),
			Date:  &azuredevops.Time{Time: commitTime},
			Name:  toPtr(t, "author"),
		},
		RemoteUrl: toPtr(t, "http://example.com/commit"),
	}
	comparison := []azuregit.GitCommitRef{commit}

	mockClient.EXPECT().GetCommits(mock.Anything, azuregit.GetCommitsArgs{
		RepositoryId: toPtr(t, scm.repoName),
		Project:      toPtr(t, scm.repoProject),
		SearchCriteria: &azuregit.GitQueryCommitsCriteria{
			IncludeLinks: toPtr(t, true),
			ItemVersion: &azuregit.GitVersionDescriptor{
				Version:     &req.Base,
				VersionType: &azuregit.GitVersionTypeValues.Branch,
			},
			CompareVersion: &azuregit.GitVersionDescriptor{
				Version:     &req.Head,
				VersionType: &azuregit.GitVersionTypeValues.Branch,
			},
		},
	}).Return(&comparison, nil)

	result, err := scm.ListChanges(ctx, req)
	require.NoError(t, err)
	assert.Len(t, result.Changes, 1)
	assert.Equal(t, "sha", result.Changes[0].Revision)
	assert.Equal(t, "commit message", result.Changes[0].Message)
	assert.Equal(t, "author", *result.Changes[0].AuthorName)
	assert.Equal(t, "author@example.com", *result.Changes[0].AuthorEmail)
	assert.Equal(t, "http://example.com/commit", *result.Changes[0].ScmUrl)
}

func TestSCM_ListChanges_Error(t *testing.T) {
	mockClient := NewMockClient(t)
	scm := &SCM{
		logger:      zap.NewNop(),
		repoOwner:   "owner",
		repoProject: "project",
		repoName:    "repo",
		client:      mockClient,
	}

	ctx := context.Background()
	req := git.ListChangesRequest{Base: "main", Head: "feature"}

	mockClient.EXPECT().GetCommits(mock.Anything, mock.Anything).Return(nil, errors.New("compare error"))
	result, err := scm.ListChanges(ctx, req)
	require.Error(t, err)
	assert.Nil(t, result)
}

func TestSCM_ListProposals(t *testing.T) {
	mockClient := NewMockClient(t)
	scm := &SCM{
		baseURL:     "http://example.com",
		repoOwner:   "owner",
		repoProject: "project",
		repoName:    "repo",
		client:      mockClient,
		logger:      zap.NewNop(),
	}

	ctx := context.Background()
	mockEnv := serverenvsmock.NewMockEnvironment(t)
	mockEnv.EXPECT().Configuration().Return(&rpcenv.EnvironmentConfiguration{Ref: "main"})
	mockEnv.EXPECT().Key().Return("testenv")

	branchOpen := "flipt/testenv/feature-open"
	branchClosed := "flipt/testenv/feature-closed"
	prOpen := azuregit.GitPullRequest{
		SourceRefName: toPtr(t, fmt.Sprintf("refs/heads/%s", branchOpen)),
		PullRequestId: toPtr(t, 123),
		Status:        &azuregit.PullRequestStatusValues.Active,
	}
	prClosed := azuregit.GitPullRequest{
		SourceRefName: toPtr(t, fmt.Sprintf("refs/heads/%s", branchClosed)),
		PullRequestId: toPtr(t, 124),
		Status:        &azuregit.PullRequestStatusValues.Completed,
	}

	prOther := azuregit.GitPullRequest{
		SourceRefName: toPtr(t, "refs/heads/other-feature"),
		PullRequestId: toPtr(t, 125),
		Status:        &azuregit.PullRequestStatusValues.Completed,
	}
	prs := []azuregit.GitPullRequest{prOpen, prClosed, prOther}

	mockClient.EXPECT().GetPullRequests(mock.Anything, azuregit.GetPullRequestsArgs{
		RepositoryId: &scm.repoName,
		Project:      &scm.repoProject,
		Top:          toPtr(t, 100),
		Skip:         toPtr(t, 0),
		SearchCriteria: &azuregit.GitPullRequestSearchCriteria{
			TargetRefName: toPtr(t, "refs/heads/main"),
			IncludeLinks:  toPtr(t, true),
		},
	}).Return(&prs, nil)

	prsEmpty := []azuregit.GitPullRequest{}
	mockClient.EXPECT().GetPullRequests(mock.Anything, azuregit.GetPullRequestsArgs{
		RepositoryId: &scm.repoName,
		Project:      &scm.repoProject,
		Top:          toPtr(t, 100),
		Skip:         toPtr(t, 100),
		SearchCriteria: &azuregit.GitPullRequestSearchCriteria{
			TargetRefName: toPtr(t, "refs/heads/main"),
			IncludeLinks:  toPtr(t, true),
		},
	}).Return(&prsEmpty, nil)

	result, err := scm.ListProposals(ctx, mockEnv)
	require.NoError(t, err)
	require.Len(t, result, 2)
	assert.Contains(t, result, branchOpen)
	assert.Contains(t, result, branchClosed)
	assert.Equal(t, "http://example.com/owner/project/_git/repo/pullrequest/123", result[branchOpen].Url)
	assert.Equal(t, rpcenv.ProposalState_PROPOSAL_STATE_OPEN, result[branchOpen].State)
	assert.Equal(t, rpcenv.ProposalState_PROPOSAL_STATE_MERGED, result[branchClosed].State)
}
