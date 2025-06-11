package github

import (
	"context"
	"errors"
	"net/url"
	"testing"
	"time"

	"github.com/google/go-github/v66/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	git "go.flipt.io/flipt/internal/enterprise/storage/environments/git"
	serverenvsmock "go.flipt.io/flipt/internal/server/environments"
	rpcenv "go.flipt.io/flipt/rpc/v2/environments"
	"go.uber.org/zap"
)

func TestSCM_Propose(t *testing.T) {
	mockPR := NewMockPullRequestsService(t)
	scm := &SCM{
		logger:    zap.NewNop(),
		repoOwner: "owner",
		repoName:  "repo",
		prs:       mockPR,
		repos:     nil,
	}

	ctx := context.Background()
	req := git.ProposalRequest{
		Base:  "main",
		Head:  "feature",
		Title: "Test PR",
		Body:  "This is a test",
		Draft: false,
	}

	expectedPR := &github.PullRequest{}
	expectedPR.HTMLURL = github.String("http://example.com/pr")
	mockPR.EXPECT().
		Create(ctx, "owner", "repo", &github.NewPullRequest{
			Base:  github.String("main"),
			Head:  github.String("feature"),
			Title: github.String("Test PR"),
			Body:  github.String("This is a test"),
			Draft: github.Bool(false),
		}).
		Return(expectedPR, nil, nil)

	result, err := scm.Propose(ctx, req)
	require.NoError(t, err)
	assert.Equal(t, "http://example.com/pr", result.Url)
	assert.Equal(t, rpcenv.ProposalState_PROPOSAL_STATE_OPEN, result.State)
}

func TestSCM_Propose_Error(t *testing.T) {
	mockPR := NewMockPullRequestsService(t)
	scm := &SCM{
		logger:    zap.NewNop(),
		repoOwner: "owner",
		repoName:  "repo",
		prs:       mockPR,
		repos:     nil,
	}

	ctx := context.Background()
	req := git.ProposalRequest{}
	mockPR.EXPECT().
		Create(ctx, "owner", "repo", &github.NewPullRequest{
			Base:  github.String(""),
			Head:  github.String(""),
			Title: github.String(""),
			Body:  github.String(""),
			Draft: github.Bool(false),
		}).
		Return(nil, nil, errors.New("create error"))

	result, err := scm.Propose(ctx, req)
	require.Error(t, err)
	assert.Nil(t, result)
}

func TestSCM_ListChanges(t *testing.T) {
	mockRepos := NewMockRepositoriesService(t)
	scm := &SCM{
		logger:    zap.NewNop(),
		repoOwner: "owner",
		repoName:  "repo",
		prs:       nil,
		repos:     mockRepos,
	}

	ctx := context.Background()
	req := git.ListChangesRequest{Base: "main", Head: "feature", Limit: 1}

	commitTime := time.Now()
	commit := &github.Commit{
		Author: &github.CommitAuthor{
			Name:  github.String("author"),
			Email: github.String("author@example.com"),
			Date:  &github.Timestamp{Time: commitTime},
		},
		Message: github.String("commit message"),
	}
	prCommit := &github.RepositoryCommit{
		SHA:     github.String("sha"),
		Commit:  commit,
		HTMLURL: github.String("http://example.com/commit"),
	}
	comparison := &github.CommitsComparison{
		Commits: []*github.RepositoryCommit{prCommit},
	}

	mockRepos.EXPECT().
		CompareCommits(ctx, "owner", "repo", "main", "feature", (*github.ListOptions)(nil)).
		Return(comparison, nil, nil)

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
	mockRepos := NewMockRepositoriesService(t)
	scm := &SCM{
		logger:    zap.NewNop(),
		repoOwner: "owner",
		repoName:  "repo",
		prs:       nil,
		repos:     mockRepos,
	}

	ctx := context.Background()
	req := git.ListChangesRequest{Base: "main", Head: "feature"}

	mockRepos.EXPECT().
		CompareCommits(ctx, "owner", "repo", "main", "feature", (*github.ListOptions)(nil)).
		Return(nil, nil, errors.New("compare error"))

	result, err := scm.ListChanges(ctx, req)
	require.Error(t, err)
	assert.Nil(t, result)
}

func TestSCM_ListProposals(t *testing.T) {
	mockPR := NewMockPullRequestsService(t)
	scm := &SCM{
		logger:    zap.NewNop(),
		repoOwner: "owner",
		repoName:  "repo",
		prs:       mockPR,
		repos:     nil,
	}

	ctx := context.Background()
	mockEnv := serverenvsmock.NewMockEnvironment(t)
	mockEnv.EXPECT().Configuration().Return(&rpcenv.EnvironmentConfiguration{Ref: "main"})
	mockEnv.EXPECT().Key().Return("testenv")

	pr := &github.PullRequest{
		Head:    &github.PullRequestBranch{Ref: github.String("flipt/testenv/feature")},
		HTMLURL: github.String("http://example.com/pr"),
		State:   github.String("open"),
	}
	prsList := []*github.PullRequest{pr}
	mockPR.EXPECT().
		List(ctx, "owner", "repo", mock.Anything).
		Return(prsList, &github.Response{NextPage: 0}, nil)

	result, err := scm.ListProposals(ctx, mockEnv)
	require.NoError(t, err)
	assert.Len(t, result, 1)
	branch := "flipt/testenv/feature"
	assert.Contains(t, result, branch)
	assert.Equal(t, "http://example.com/pr", result[branch].Url)
	assert.Equal(t, rpcenv.ProposalState_PROPOSAL_STATE_OPEN, result[branch].State)
}

func TestSCM_ListProposals_PrefixFilter(t *testing.T) {
	mockPR := NewMockPullRequestsService(t)
	scm := &SCM{
		logger:    zap.NewNop(),
		repoOwner: "owner",
		repoName:  "repo",
		prs:       mockPR,
		repos:     nil,
	}

	ctx := context.Background()
	mockEnv := serverenvsmock.NewMockEnvironment(t)
	mockEnv.EXPECT().Configuration().Return(&rpcenv.EnvironmentConfiguration{Ref: "main"})

	pr := &github.PullRequest{
		Head:    &github.PullRequestBranch{Ref: github.String("otherprefix/testenv/feature")},
		HTMLURL: github.String("http://example.com/pr"),
		State:   github.String("open"),
	}
	prsList := []*github.PullRequest{pr}
	mockPR.EXPECT().
		List(ctx, "owner", "repo", mock.Anything).
		Return(prsList, &github.Response{NextPage: 0}, nil)

	result, err := scm.ListProposals(ctx, mockEnv)
	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestSCM_ListProposals_ClosedVsOpen(t *testing.T) {
	mockPR := NewMockPullRequestsService(t)
	scm := &SCM{
		logger:    zap.NewNop(),
		repoOwner: "owner",
		repoName:  "repo",
		prs:       mockPR,
		repos:     nil,
	}

	ctx := context.Background()
	mockEnv := serverenvsmock.NewMockEnvironment(t)
	mockEnv.EXPECT().Configuration().Return(&rpcenv.EnvironmentConfiguration{Ref: "main"})
	mockEnv.EXPECT().Key().Return("testenv")

	branch := "flipt/testenv/feature"
	prOpen := &github.PullRequest{
		Head:    &github.PullRequestBranch{Ref: github.String(branch)},
		HTMLURL: github.String("http://example.com/pr-open"),
		State:   github.String("open"),
	}
	prClosed := &github.PullRequest{
		Head:    &github.PullRequestBranch{Ref: github.String(branch)},
		HTMLURL: github.String("http://example.com/pr-closed"),
		State:   github.String("closed"),
	}
	prsList := []*github.PullRequest{prClosed, prOpen}
	mockPR.EXPECT().
		List(ctx, "owner", "repo", mock.Anything).
		Return(prsList, &github.Response{NextPage: 0}, nil)

	result, err := scm.ListProposals(ctx, mockEnv)
	require.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "http://example.com/pr-open", result[branch].Url)
	assert.Equal(t, rpcenv.ProposalState_PROPOSAL_STATE_OPEN, result[branch].State)
}

func TestSCM_ListProposals_ClosedMerged(t *testing.T) {
	mockPR := NewMockPullRequestsService(t)
	scm := &SCM{
		logger:    zap.NewNop(),
		repoOwner: "owner",
		repoName:  "repo",
		prs:       mockPR,
		repos:     nil,
	}

	ctx := context.Background()
	mockEnv := serverenvsmock.NewMockEnvironment(t)
	mockEnv.EXPECT().Configuration().Return(&rpcenv.EnvironmentConfiguration{Ref: "main"})
	mockEnv.EXPECT().Key().Return("testenv")

	branch := "flipt/testenv/feature"
	prClosedMerged := &github.PullRequest{
		Head:           &github.PullRequestBranch{Ref: github.String(branch)},
		HTMLURL:        github.String("http://example.com/pr-merged"),
		State:          github.String("closed"),
		Merged:         github.Bool(true),
		MergeCommitSHA: github.String("sha123"),
	}
	prsList := []*github.PullRequest{prClosedMerged}
	mockPR.EXPECT().
		List(ctx, "owner", "repo", mock.Anything).
		Return(prsList, &github.Response{NextPage: 0}, nil)

	result, err := scm.ListProposals(ctx, mockEnv)
	require.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "http://example.com/pr-merged", result[branch].Url)
	assert.Equal(t, rpcenv.ProposalState_PROPOSAL_STATE_MERGED, result[branch].State)
}

func TestSCM_ListProposals_ClosedNotMerged(t *testing.T) {
	mockPR := NewMockPullRequestsService(t)
	scm := &SCM{
		logger:    zap.NewNop(),
		repoOwner: "owner",
		repoName:  "repo",
		prs:       mockPR,
		repos:     nil,
	}

	ctx := context.Background()
	mockEnv := serverenvsmock.NewMockEnvironment(t)
	mockEnv.EXPECT().Configuration().Return(&rpcenv.EnvironmentConfiguration{Ref: "main"})
	mockEnv.EXPECT().Key().Return("testenv")

	branch := "flipt/testenv/feature"
	prClosed := &github.PullRequest{
		Head:    &github.PullRequestBranch{Ref: github.String(branch)},
		HTMLURL: github.String("http://example.com/pr-closed"),
		State:   github.String("closed"),
		Merged:  github.Bool(false),
	}
	prsList := []*github.PullRequest{prClosed}
	mockPR.EXPECT().
		List(ctx, "owner", "repo", mock.Anything).
		Return(prsList, &github.Response{NextPage: 0}, nil)

	result, err := scm.ListProposals(ctx, mockEnv)
	require.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "http://example.com/pr-closed", result[branch].Url)
	assert.Equal(t, rpcenv.ProposalState_PROPOSAL_STATE_CLOSED, result[branch].State)
}

func TestWithApiURL(t *testing.T) {
	t.Run("adds trailing slash if missing", func(t *testing.T) {
		u, _ := url.Parse("https://github.example.com")
		opts := &gitHubOptions{}
		WithApiURL(u)(opts)
		assert.Equal(t, "https://github.example.com/api/v3/", opts.apiURL.String())
	})

	t.Run("does not add /api/v3/ if already present", func(t *testing.T) {
		u, _ := url.Parse("https://github.example.com/api/v3/")
		opts := &gitHubOptions{}
		WithApiURL(u)(opts)
		assert.Equal(t, "https://github.example.com/api/v3/", opts.apiURL.String())
	})

	t.Run("adds /api/v3/ if missing and path has trailing slash", func(t *testing.T) {
		u, _ := url.Parse("https://github.example.com/")
		opts := &gitHubOptions{}
		WithApiURL(u)(opts)
		assert.Equal(t, "https://github.example.com/api/v3/", opts.apiURL.String())
	})

	t.Run("does not add /api/v3/ for api.github.com host", func(t *testing.T) {
		u, _ := url.Parse("https://api.github.com/")
		opts := &gitHubOptions{}
		WithApiURL(u)(opts)
		assert.Equal(t, "https://api.github.com/", opts.apiURL.String())
	})

	t.Run("does not add /api/v3/ for .api. in host", func(t *testing.T) {
		u, _ := url.Parse("https://foo.api.github.com/")
		opts := &gitHubOptions{}
		WithApiURL(u)(opts)
		assert.Equal(t, "https://foo.api.github.com/", opts.apiURL.String())
	})
}
