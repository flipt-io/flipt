package gitea

import (
	"context"
	"errors"
	"testing"
	"time"

	"code.gitea.io/sdk/gitea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/enterprise/storage/environments/git"
	serverenvsmock "go.flipt.io/flipt/internal/server/environments"
	rpcenv "go.flipt.io/flipt/rpc/v2/environments"
	"go.uber.org/zap"
)

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

	mockPR := &gitea.PullRequest{HTMLURL: "http://example.com/pr"}
	mockClient.EXPECT().CreatePullRequest("owner", "repo", mock.Anything).Return(mockPR, nil, nil)

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

	mockClient.EXPECT().CreatePullRequest("owner", "repo", mock.Anything).Return(nil, nil, errors.New("create error"))

	result, err := scm.Propose(ctx, req)
	require.Error(t, err)
	assert.Nil(t, result)
}

func TestSCM_ListChanges(t *testing.T) {
	mockClient := NewMockClient(t)
	scm := &SCM{
		logger:    zap.NewNop(),
		repoOwner: "owner",
		repoName:  "repo",
		client:    mockClient,
	}

	ctx := context.Background()
	req := git.ListChangesRequest{Base: "main", Head: "feature", Limit: 1}

	commitTime := time.Now()
	commit := &gitea.Commit{
		CommitMeta: &gitea.CommitMeta{
			SHA:     "sha",
			Created: commitTime,
		},
		RepoCommit: &gitea.RepoCommit{
			Message: "commit message",
		},
		HTMLURL: "http://example.com/commit",

		Author: &gitea.User{
			FullName: "author",
			Email:    "author@example.com",
		},
	}
	comparison := &gitea.Compare{Commits: []*gitea.Commit{commit}}

	mockClient.EXPECT().CompareCommits("owner", "repo", "main", "feature").Return(comparison, nil, nil)

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
		logger:    zap.NewNop(),
		repoOwner: "owner",
		repoName:  "repo",
		client:    mockClient,
	}

	ctx := context.Background()
	req := git.ListChangesRequest{Base: "main", Head: "feature"}

	mockClient.EXPECT().CompareCommits("owner", "repo", "main", "feature").Return(nil, nil, errors.New("compare error"))

	result, err := scm.ListChanges(ctx, req)
	require.Error(t, err)
	assert.Nil(t, result)
}

func TestSCM_ListProposals(t *testing.T) {
	mockClient := NewMockClient(t)
	scm := &SCM{
		repoOwner: "owner",
		repoName:  "repo",
		client:    mockClient,
	}

	ctx := context.Background()
	mockEnv := serverenvsmock.NewMockEnvironment(t)
	mockEnv.EXPECT().Key().Return("testenv")

	branch := "flipt/testenv/feature"
	pr := &gitea.PullRequest{
		Head:    &gitea.PRBranchInfo{Ref: branch},
		Base:    &gitea.PRBranchInfo{Name: "testenv"},
		HTMLURL: "http://example.com/pr",
		State:   "open",
	}
	prs := []*gitea.PullRequest{pr}

	mockClient.EXPECT().ListRepoPullRequests("owner", "repo", mock.Anything).Return(prs, &gitea.Response{}, nil)

	result, err := scm.ListProposals(ctx, mockEnv)
	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.Contains(t, result, branch)
	assert.Equal(t, "http://example.com/pr", result[branch].Url)
	assert.Equal(t, rpcenv.ProposalState_PROPOSAL_STATE_OPEN, result[branch].State)
}

func TestSCM_ListProposals_PrefixFilter(t *testing.T) {
	mockClient := NewMockClient(t)
	scm := &SCM{
		logger:    zap.NewNop(),
		repoOwner: "owner",
		repoName:  "repo",
		client:    mockClient,
	}

	ctx := context.Background()
	mockEnv := serverenvsmock.NewMockEnvironment(t)
	mockEnv.EXPECT().Key().Return("testenv")

	pr := &gitea.PullRequest{
		Head:    &gitea.PRBranchInfo{Ref: "otherprefix/testenv/feature"},
		Base:    &gitea.PRBranchInfo{Name: "testenv"},
		HTMLURL: "http://example.com/pr",
		State:   gitea.StateOpen,
	}
	prs := []*gitea.PullRequest{pr}

	mockClient.EXPECT().ListRepoPullRequests("owner", "repo", mock.Anything).Return(prs, &gitea.Response{}, nil)

	result, err := scm.ListProposals(ctx, mockEnv)
	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestSCM_ListProposals_ClosedVsOpen(t *testing.T) {
	mockClient := NewMockClient(t)
	scm := &SCM{
		logger:    zap.NewNop(),
		repoOwner: "owner",
		repoName:  "repo",
		client:    mockClient,
	}

	ctx := context.Background()
	mockEnv := serverenvsmock.NewMockEnvironment(t)
	mockEnv.EXPECT().Key().Return("testenv")

	branch := "flipt/testenv/feature"
	prOpen := &gitea.PullRequest{
		Head:    &gitea.PRBranchInfo{Ref: branch},
		Base:    &gitea.PRBranchInfo{Name: "testenv"},
		HTMLURL: "http://example.com/pr-open",
		State:   gitea.StateOpen,
	}
	prClosed := &gitea.PullRequest{
		Head:    &gitea.PRBranchInfo{Ref: branch},
		Base:    &gitea.PRBranchInfo{Name: "testenv"},
		HTMLURL: "http://example.com/pr-closed",
		State:   gitea.StateClosed,
	}
	prs := []*gitea.PullRequest{prClosed, prOpen}

	mockClient.EXPECT().ListRepoPullRequests("owner", "repo", mock.Anything).Return(prs, &gitea.Response{}, nil)

	result, err := scm.ListProposals(ctx, mockEnv)
	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.Equal(t, "http://example.com/pr-open", result[branch].Url)
	assert.Equal(t, rpcenv.ProposalState_PROPOSAL_STATE_OPEN, result[branch].State)
}

func TestSCM_ListProposals_ClosedMerged(t *testing.T) {
	mockClient := NewMockClient(t)
	scm := &SCM{
		logger:    zap.NewNop(),
		repoOwner: "owner",
		repoName:  "repo",
		client:    mockClient,
	}

	ctx := context.Background()
	mockEnv := serverenvsmock.NewMockEnvironment(t)
	mockEnv.EXPECT().Key().Return("testenv")

	branch := "flipt/testenv/feature"
	mergeCommitSha := "sha123"
	prClosedMerged := &gitea.PullRequest{
		Head:           &gitea.PRBranchInfo{Ref: branch},
		Base:           &gitea.PRBranchInfo{Name: "testenv"},
		HTMLURL:        "http://example.com/pr-merged",
		State:          "closed",
		HasMerged:      true,
		MergedCommitID: &mergeCommitSha,
	}
	prs := []*gitea.PullRequest{prClosedMerged}

	mockClient.EXPECT().ListRepoPullRequests("owner", "repo", mock.Anything).Return(prs, &gitea.Response{}, nil)

	result, err := scm.ListProposals(ctx, mockEnv)
	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.Equal(t, "http://example.com/pr-merged", result[branch].Url)
	assert.Equal(t, rpcenv.ProposalState_PROPOSAL_STATE_MERGED, result[branch].State)
}

func TestSCM_ListProposals_ClosedNotMerged(t *testing.T) {
	mockClient := NewMockClient(t)
	scm := &SCM{
		logger:    zap.NewNop(),
		repoOwner: "owner",
		repoName:  "repo",
		client:    mockClient,
	}

	ctx := context.Background()
	mockEnv := serverenvsmock.NewMockEnvironment(t)
	mockEnv.EXPECT().Key().Return("testenv")

	branch := "flipt/testenv/feature"
	prClosed := &gitea.PullRequest{
		Head:      &gitea.PRBranchInfo{Ref: branch},
		Base:      &gitea.PRBranchInfo{Name: "testenv"},
		HTMLURL:   "http://example.com/pr-closed",
		State:     "closed",
		HasMerged: false,
	}
	prs := []*gitea.PullRequest{prClosed}

	mockClient.EXPECT().ListRepoPullRequests("owner", "repo", mock.Anything).Return(prs, &gitea.Response{}, nil)

	result, err := scm.ListProposals(ctx, mockEnv)
	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.Equal(t, "http://example.com/pr-closed", result[branch].Url)
	assert.Equal(t, rpcenv.ProposalState_PROPOSAL_STATE_CLOSED, result[branch].State)
}
