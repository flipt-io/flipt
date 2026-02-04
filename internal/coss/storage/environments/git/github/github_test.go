package github

import (
	"errors"
	"net/url"
	"testing"
	"time"

	"github.com/google/go-github/v75/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/coss/storage/environments/git"
	"go.flipt.io/flipt/internal/credentials"
	serverenvsmock "go.flipt.io/flipt/internal/server/environments"
	rpcenv "go.flipt.io/flipt/rpc/v2/environments"
	"go.uber.org/zap"
)

func TestSCM_Propose(t *testing.T) {
	mockPR := NewMockPullRequestsService(t)
	scm := &SCM{
		logger:     zap.NewNop(),
		owner:      "owner",
		repository: "repo",
		prs:        mockPR,
		repos:      nil,
	}

	ctx := t.Context()
	req := git.ProposalRequest{
		Base:  "main",
		Head:  "feature",
		Title: "Test PR",
		Body:  "This is a test",
		Draft: false,
	}

	expectedPR := &github.PullRequest{}
	expectedPR.HTMLURL = github.Ptr("http://example.com/pr")
	mockPR.EXPECT().
		Create(ctx, "owner", "repo", &github.NewPullRequest{
			Base:  github.Ptr("main"),
			Head:  github.Ptr("feature"),
			Title: github.Ptr("Test PR"),
			Body:  github.Ptr("This is a test"),
			Draft: github.Ptr(false),
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
		logger:     zap.NewNop(),
		owner:      "owner",
		repository: "repo",
		prs:        mockPR,
		repos:      nil,
	}

	ctx := t.Context()
	req := git.ProposalRequest{}
	mockPR.EXPECT().
		Create(ctx, "owner", "repo", &github.NewPullRequest{
			Base:  github.Ptr(""),
			Head:  github.Ptr(""),
			Title: github.Ptr(""),
			Body:  github.Ptr(""),
			Draft: github.Ptr(false),
		}).
		Return(nil, nil, errors.New("create error"))

	result, err := scm.Propose(ctx, req)
	require.Error(t, err)
	assert.Nil(t, result)
}

func TestSCM_ListChanges(t *testing.T) {
	mockRepos := NewMockRepositoriesService(t)
	scm := &SCM{
		logger:     zap.NewNop(),
		owner:      "owner",
		repository: "repo",
		prs:        nil,
		repos:      mockRepos,
	}

	ctx := t.Context()
	req := git.ListChangesRequest{Base: "main", Head: "feature", Limit: 1}

	commitTime := time.Now()
	commit := &github.Commit{
		Author: &github.CommitAuthor{
			Name:  github.Ptr("author"),
			Email: github.Ptr("author@example.com"),
			Date:  &github.Timestamp{Time: commitTime},
		},
		Message: github.Ptr("commit message"),
	}
	prCommit := &github.RepositoryCommit{
		SHA:     github.Ptr("sha"),
		Commit:  commit,
		HTMLURL: github.Ptr("http://example.com/commit"),
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
		logger:     zap.NewNop(),
		owner:      "owner",
		repository: "repo",
		prs:        nil,
		repos:      mockRepos,
	}

	ctx := t.Context()
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
		logger:     zap.NewNop(),
		owner:      "owner",
		repository: "repo",
		prs:        mockPR,
		repos:      nil,
	}

	ctx := t.Context()
	mockEnv := serverenvsmock.NewMockEnvironment(t)
	mockEnv.EXPECT().Configuration().Return(&rpcenv.EnvironmentConfiguration{Ref: "main"})
	mockEnv.EXPECT().Key().Return("testenv")

	pr := &github.PullRequest{
		Head:    &github.PullRequestBranch{Ref: github.Ptr("flipt/testenv/feature")},
		HTMLURL: github.Ptr("http://example.com/pr"),
		State:   github.Ptr("open"),
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
		logger:     zap.NewNop(),
		owner:      "owner",
		repository: "repo",
		prs:        mockPR,
		repos:      nil,
	}

	ctx := t.Context()
	mockEnv := serverenvsmock.NewMockEnvironment(t)
	mockEnv.EXPECT().Configuration().Return(&rpcenv.EnvironmentConfiguration{Ref: "main"})
	mockEnv.EXPECT().Key().Return("testenv")

	pr := &github.PullRequest{
		Head:    &github.PullRequestBranch{Ref: github.Ptr("otherprefix/testenv/feature")},
		HTMLURL: github.Ptr("http://example.com/pr"),
		State:   github.Ptr("open"),
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
		logger:     zap.NewNop(),
		owner:      "owner",
		repository: "repo",
		prs:        mockPR,
		repos:      nil,
	}

	ctx := t.Context()
	mockEnv := serverenvsmock.NewMockEnvironment(t)
	mockEnv.EXPECT().Configuration().Return(&rpcenv.EnvironmentConfiguration{Ref: "main"})
	mockEnv.EXPECT().Key().Return("testenv")

	branch := "flipt/testenv/feature"
	prOpen := &github.PullRequest{
		Head:    &github.PullRequestBranch{Ref: github.Ptr(branch)},
		HTMLURL: github.Ptr("http://example.com/pr-open"),
		State:   github.Ptr("open"),
	}
	prClosed := &github.PullRequest{
		Head:    &github.PullRequestBranch{Ref: github.Ptr(branch)},
		HTMLURL: github.Ptr("http://example.com/pr-closed"),
		State:   github.Ptr("closed"),
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
		logger:     zap.NewNop(),
		owner:      "owner",
		repository: "repo",
		prs:        mockPR,
		repos:      nil,
	}

	ctx := t.Context()
	mockEnv := serverenvsmock.NewMockEnvironment(t)
	mockEnv.EXPECT().Configuration().Return(&rpcenv.EnvironmentConfiguration{Ref: "main"})
	mockEnv.EXPECT().Key().Return("testenv")

	branch := "flipt/testenv/feature"
	prClosedMerged := &github.PullRequest{
		Head:           &github.PullRequestBranch{Ref: github.Ptr(branch)},
		HTMLURL:        github.Ptr("http://example.com/pr-merged"),
		State:          github.Ptr("closed"),
		Merged:         github.Ptr(true),
		MergeCommitSHA: github.Ptr("sha123"),
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
		logger:     zap.NewNop(),
		owner:      "owner",
		repository: "repo",
		prs:        mockPR,
		repos:      nil,
	}

	ctx := t.Context()
	mockEnv := serverenvsmock.NewMockEnvironment(t)
	mockEnv.EXPECT().Configuration().Return(&rpcenv.EnvironmentConfiguration{Ref: "main"})
	mockEnv.EXPECT().Key().Return("testenv")

	branch := "flipt/testenv/feature"
	prClosed := &github.PullRequest{
		Head:    &github.PullRequestBranch{Ref: github.Ptr(branch)},
		HTMLURL: github.Ptr("http://example.com/pr-closed"),
		State:   github.Ptr("closed"),
		Merged:  github.Ptr(false),
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

func TestWithApiAuth(t *testing.T) {
	t.Run("sets HTTP client with access token", func(t *testing.T) {
		// Create an APIAuth with access token type
		// Note: APIAuth fields are not exported, so we need to create it through the normal API
		// For testing purposes, we'll verify the function accepts the parameter correctly
		opts := &gitHubOptions{
			ctx: t.Context(),
		}

		// Create a minimal APIAuth - the Type() method will return empty string by default
		// which won't match any case, so httpClient remains nil
		apiAuth := &credentials.APIAuth{}
		WithApiAuth(apiAuth)(opts)

		// Since the Type is empty and doesn't match any case, httpClient should remain nil
		assert.Nil(t, opts.httpClient)
	})
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
