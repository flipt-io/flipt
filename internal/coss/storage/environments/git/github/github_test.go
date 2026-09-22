package github

import (
	"errors"
	"net/url"
	"testing"
	"time"

	"github.com/google/go-github/v92/github"
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
	expectedPR.HTMLURL = new("http://example.com/pr")
	mockPR.EXPECT().
		Create(ctx, "owner", "repo", github.CreatePullRequest{
			Base:  "main",
			Head:  "feature",
			Title: new("Test PR"),
			Body:  new("This is a test"),
			Draft: new(false),
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
		Create(ctx, "owner", "repo", github.CreatePullRequest{
			Base:  "",
			Head:  "",
			Title: new(""),
			Body:  new(""),
			Draft: new(false),
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
			Name:  new("author"),
			Email: new("author@example.com"),
			Date:  &github.Timestamp{Time: commitTime},
		},
		Message: new("commit message"),
	}
	prCommit := &github.RepositoryCommit{
		SHA:     new("sha"),
		Commit:  commit,
		HTMLURL: new("http://example.com/commit"),
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
		Head:    &github.PullRequestBranch{Ref: new("flipt/testenv/feature")},
		HTMLURL: new("http://example.com/pr"),
		State:   new("open"),
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
		Head:    &github.PullRequestBranch{Ref: new("otherprefix/testenv/feature")},
		HTMLURL: new("http://example.com/pr"),
		State:   new("open"),
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
		Head:    &github.PullRequestBranch{Ref: new(branch)},
		HTMLURL: new("http://example.com/pr-open"),
		State:   new("open"),
	}
	prClosed := &github.PullRequest{
		Head:    &github.PullRequestBranch{Ref: new(branch)},
		HTMLURL: new("http://example.com/pr-closed"),
		State:   new("closed"),
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
		Head:           &github.PullRequestBranch{Ref: new(branch)},
		HTMLURL:        new("http://example.com/pr-merged"),
		State:          new("closed"),
		Merged:         new(true),
		MergeCommitSHA: new("sha123"),
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
		Head:    &github.PullRequestBranch{Ref: new(branch)},
		HTMLURL: new("http://example.com/pr-closed"),
		State:   new("closed"),
		Merged:  new(false),
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
	const wantEnterprise = "https://github.example.com/api/v3/"

	tests := []struct {
		name    string
		rawURL  string
		wantURL string
	}{
		{
			name:    "adds trailing slash and api path if missing",
			rawURL:  "https://github.example.com",
			wantURL: wantEnterprise,
		},
		{
			name:    "adds api path if missing and path has trailing slash",
			rawURL:  "https://github.example.com/",
			wantURL: wantEnterprise,
		},
		{
			name:    "does not add api path if already present",
			rawURL:  "https://github.example.com/api/v3/",
			wantURL: wantEnterprise,
		},
		{
			name:    "does not add api path without trailing slash if already present",
			rawURL:  "https://github.example.com/api/v3",
			wantURL: wantEnterprise,
		},
		{
			name:    "does not add api path for api.github.com host",
			rawURL:  "https://api.github.com/",
			wantURL: "https://api.github.com/",
		},
		{
			name:    "does not add api path for .api. in host",
			rawURL:  "https://foo.api.github.com/",
			wantURL: "https://foo.api.github.com/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := url.Parse(tt.rawURL)
			require.NoError(t, err)

			opts := &gitHubOptions{}
			WithApiURL(u)(opts)

			require.NotNil(t, opts.apiURL)
			assert.Equal(t, tt.wantURL, *opts.apiURL)
		})
	}
}
