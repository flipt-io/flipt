package gitlab

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gitlab "gitlab.com/gitlab-org/api/client-go"
	"go.flipt.io/flipt/internal/enterprise/storage/environments/git"
	serverenvsmock "go.flipt.io/flipt/internal/server/environments"
	rpcenv "go.flipt.io/flipt/rpc/v2/environments"
	"go.uber.org/zap"
)

func TestSCM_Propose(t *testing.T) {
	mockMR := NewMockMergeRequestsService(t)
	scm := &SCM{
		logger:    zap.NewNop(),
		projectID: "owner/repo",
		mrs:       mockMR,
		repos:     nil,
	}

	ctx := context.Background()
	req := git.ProposalRequest{
		Base:  "main",
		Head:  "feature",
		Title: "Test MR",
		Body:  "This is a test",
		Draft: false,
	}

	mr := &gitlab.MergeRequest{BasicMergeRequest: gitlab.BasicMergeRequest{WebURL: "http://example.com/mr", State: "opened"}}
	resp := &gitlab.Response{Response: &http.Response{Body: io.NopCloser(strings.NewReader("ok")), StatusCode: 201}}
	mockMR.EXPECT().CreateMergeRequest("owner/repo", &gitlab.CreateMergeRequestOptions{
		Title:        &req.Title,
		Description:  &req.Body,
		SourceBranch: &req.Head,
		TargetBranch: &req.Base,
	}).Return(mr, resp, nil)

	result, err := scm.Propose(ctx, req)
	require.NoError(t, err)
	assert.Equal(t, "http://example.com/mr", result.Url)
	assert.Equal(t, rpcenv.ProposalState_PROPOSAL_STATE_OPEN, result.State)
}

func TestSCM_Propose_Draft(t *testing.T) {
	mockMR := NewMockMergeRequestsService(t)
	scm := &SCM{
		logger:    zap.NewNop(),
		projectID: "owner/repo",
		mrs:       mockMR,
		repos:     nil,
	}

	ctx := context.Background()
	req := git.ProposalRequest{
		Base:  "main",
		Head:  "feature",
		Title: "Test MR",
		Body:  "This is a test",
		Draft: true,
	}
	draftTitle := "Draft: Test MR"
	mr := &gitlab.MergeRequest{BasicMergeRequest: gitlab.BasicMergeRequest{WebURL: "http://example.com/mr", State: "opened"}}
	resp := &gitlab.Response{Response: &http.Response{Body: io.NopCloser(strings.NewReader("ok")), StatusCode: 201}}
	mockMR.EXPECT().CreateMergeRequest("owner/repo", &gitlab.CreateMergeRequestOptions{
		Title:        &draftTitle,
		Description:  &req.Body,
		SourceBranch: &req.Head,
		TargetBranch: &req.Base,
	}).Return(mr, resp, nil)

	result, err := scm.Propose(ctx, req)
	require.NoError(t, err)
	assert.Equal(t, "http://example.com/mr", result.Url)
	assert.Equal(t, rpcenv.ProposalState_PROPOSAL_STATE_OPEN, result.State)
}

func TestSCM_Propose_Error(t *testing.T) {
	mockMR := NewMockMergeRequestsService(t)
	scm := &SCM{
		logger:    zap.NewNop(),
		projectID: "owner/repo",
		mrs:       mockMR,
		repos:     nil,
	}

	ctx := context.Background()
	req := git.ProposalRequest{}
	mockMR.EXPECT().CreateMergeRequest("owner/repo", &gitlab.CreateMergeRequestOptions{
		Title:        &req.Title,
		Description:  &req.Body,
		SourceBranch: &req.Head,
		TargetBranch: &req.Base,
	}).Return(nil, nil, errors.New("create error"))

	result, err := scm.Propose(ctx, req)
	require.Error(t, err)
	assert.Nil(t, result)
}

func TestSCM_ListChanges(t *testing.T) {
	mockRepos := NewMockRepositoriesService(t)
	scm := &SCM{
		logger:    zap.NewNop(),
		projectID: "owner/repo",
		mrs:       nil,
		repos:     mockRepos,
	}

	ctx := context.Background()
	req := git.ListChangesRequest{Base: "main", Head: "feature", Limit: 1}

	commitTime := time.Now()
	commit := &gitlab.Commit{
		ID:          "sha",
		Message:     "commit message",
		WebURL:      "http://example.com/commit",
		AuthorName:  "author",
		AuthorEmail: "author@example.com",
		CreatedAt:   &commitTime,
	}
	comparison := &gitlab.Compare{Commits: []*gitlab.Commit{commit}}
	resp := &gitlab.Response{Response: &http.Response{Body: io.NopCloser(strings.NewReader("ok")), StatusCode: 200}}
	mockRepos.EXPECT().Compare("owner/repo", &gitlab.CompareOptions{
		From: &req.Base,
		To:   &req.Head,
	}).Return(comparison, resp, nil)

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
		projectID: "owner/repo",
		mrs:       nil,
		repos:     mockRepos,
	}

	ctx := context.Background()
	req := git.ListChangesRequest{Base: "main", Head: "feature"}
	mockRepos.EXPECT().Compare("owner/repo", &gitlab.CompareOptions{
		From: &req.Base,
		To:   &req.Head,
	}).Return(nil, nil, errors.New("compare error"))

	result, err := scm.ListChanges(ctx, req)
	require.Error(t, err)
	assert.Nil(t, result)
}

func TestSCM_ListProposals(t *testing.T) {
	mockMR := NewMockMergeRequestsService(t)
	scm := &SCM{
		logger:    zap.NewNop(),
		projectID: "owner/repo",
		mrs:       mockMR,
		repos:     nil,
	}

	ctx := context.Background()
	mockEnv := serverenvsmock.NewMockEnvironment(t)
	mockEnv.EXPECT().Configuration().Return(&rpcenv.EnvironmentConfiguration{Ref: "main"})
	mockEnv.EXPECT().Key().Return("testenv")

	branch := "flipt/testenv/feature"
	mr := &gitlab.BasicMergeRequest{
		SourceBranch: branch,
		WebURL:       "http://example.com/mr",
		State:        "opened",
	}
	mrsList := []*gitlab.BasicMergeRequest{mr}
	resp := &gitlab.Response{NextPage: 0}
	mockMR.EXPECT().ListProjectMergeRequests("owner/repo", &gitlab.ListProjectMergeRequestsOptions{
		TargetBranch: gitlab.Ptr("main"),
		State:        gitlab.Ptr("all"),
		ListOptions:  gitlab.ListOptions{PerPage: 100},
	}).Return(mrsList, resp, nil)

	result, err := scm.ListProposals(ctx, mockEnv)
	require.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Contains(t, result, branch)
	assert.Equal(t, "http://example.com/mr", result[branch].Url)
	assert.Equal(t, rpcenv.ProposalState_PROPOSAL_STATE_OPEN, result[branch].State)
}

func TestSCM_ListProposals_PrefixFilter(t *testing.T) {
	mockMR := NewMockMergeRequestsService(t)
	scm := &SCM{
		logger:    zap.NewNop(),
		projectID: "owner/repo",
		mrs:       mockMR,
		repos:     nil,
	}

	ctx := context.Background()
	mockEnv := serverenvsmock.NewMockEnvironment(t)
	mockEnv.EXPECT().Configuration().Return(&rpcenv.EnvironmentConfiguration{Ref: "main"})

	mr := &gitlab.BasicMergeRequest{
		SourceBranch: "otherprefix/testenv/feature",
		WebURL:       "http://example.com/mr",
		State:        "opened",
	}
	mrsList := []*gitlab.BasicMergeRequest{mr}
	resp := &gitlab.Response{NextPage: 0}
	mockMR.EXPECT().ListProjectMergeRequests("owner/repo", &gitlab.ListProjectMergeRequestsOptions{
		TargetBranch: gitlab.Ptr("main"),
		State:        gitlab.Ptr("all"),
		ListOptions:  gitlab.ListOptions{PerPage: 100},
	}).Return(mrsList, resp, nil)

	result, err := scm.ListProposals(ctx, mockEnv)
	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestSCM_ListProposals_ClosedVsOpen(t *testing.T) {
	mockMR := NewMockMergeRequestsService(t)
	scm := &SCM{
		logger:    zap.NewNop(),
		projectID: "owner/repo",
		mrs:       mockMR,
		repos:     nil,
	}

	ctx := context.Background()
	mockEnv := serverenvsmock.NewMockEnvironment(t)
	mockEnv.EXPECT().Configuration().Return(&rpcenv.EnvironmentConfiguration{Ref: "main"})
	mockEnv.EXPECT().Key().Return("testenv")

	branch := "flipt/testenv/feature"
	mrOpen := &gitlab.BasicMergeRequest{
		SourceBranch: branch,
		WebURL:       "http://example.com/mr-open",
		State:        "opened",
	}
	mrClosed := &gitlab.BasicMergeRequest{
		SourceBranch: branch,
		WebURL:       "http://example.com/mr-closed",
		State:        "closed",
	}
	mrsList := []*gitlab.BasicMergeRequest{mrClosed, mrOpen}
	resp := &gitlab.Response{NextPage: 0}
	mockMR.EXPECT().ListProjectMergeRequests("owner/repo", &gitlab.ListProjectMergeRequestsOptions{
		TargetBranch: gitlab.Ptr("main"),
		State:        gitlab.Ptr("all"),
		ListOptions:  gitlab.ListOptions{PerPage: 100},
	}).Return(mrsList, resp, nil)

	result, err := scm.ListProposals(ctx, mockEnv)
	require.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "http://example.com/mr-open", result[branch].Url)
	assert.Equal(t, rpcenv.ProposalState_PROPOSAL_STATE_OPEN, result[branch].State)
}

func TestSCM_ListProposals_ClosedMerged(t *testing.T) {
	mockMR := NewMockMergeRequestsService(t)
	scm := &SCM{
		logger:    zap.NewNop(),
		projectID: "owner/repo",
		mrs:       mockMR,
		repos:     nil,
	}

	ctx := context.Background()
	mockEnv := serverenvsmock.NewMockEnvironment(t)
	mockEnv.EXPECT().Configuration().Return(&rpcenv.EnvironmentConfiguration{Ref: "main"})
	mockEnv.EXPECT().Key().Return("testenv")

	branch := "flipt/testenv/feature"
	mrClosedMerged := &gitlab.BasicMergeRequest{
		SourceBranch: branch,
		WebURL:       "http://example.com/mr-merged",
		State:        "merged",
	}
	mrsList := []*gitlab.BasicMergeRequest{mrClosedMerged}
	resp := &gitlab.Response{NextPage: 0}
	mockMR.EXPECT().ListProjectMergeRequests("owner/repo", &gitlab.ListProjectMergeRequestsOptions{
		TargetBranch: gitlab.Ptr("main"),
		State:        gitlab.Ptr("all"),
		ListOptions:  gitlab.ListOptions{PerPage: 100},
	}).Return(mrsList, resp, nil)

	result, err := scm.ListProposals(ctx, mockEnv)
	require.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "http://example.com/mr-merged", result[branch].Url)
	assert.Equal(t, rpcenv.ProposalState_PROPOSAL_STATE_MERGED, result[branch].State)
}

func TestSCM_ListProposals_ClosedNotMerged(t *testing.T) {
	mockMR := NewMockMergeRequestsService(t)
	scm := &SCM{
		logger:    zap.NewNop(),
		projectID: "owner/repo",
		mrs:       mockMR,
		repos:     nil,
	}

	ctx := context.Background()
	mockEnv := serverenvsmock.NewMockEnvironment(t)
	mockEnv.EXPECT().Configuration().Return(&rpcenv.EnvironmentConfiguration{Ref: "main"})
	mockEnv.EXPECT().Key().Return("testenv")

	branch := "flipt/testenv/feature"
	mrClosed := &gitlab.BasicMergeRequest{
		SourceBranch: branch,
		WebURL:       "http://example.com/mr-closed",
		State:        "closed",
	}
	mrsList := []*gitlab.BasicMergeRequest{mrClosed}
	resp := &gitlab.Response{NextPage: 0}
	mockMR.EXPECT().ListProjectMergeRequests("owner/repo", &gitlab.ListProjectMergeRequestsOptions{
		TargetBranch: gitlab.Ptr("main"),
		State:        gitlab.Ptr("all"),
		ListOptions:  gitlab.ListOptions{PerPage: 100},
	}).Return(mrsList, resp, nil)

	result, err := scm.ListProposals(ctx, mockEnv)
	require.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "http://example.com/mr-closed", result[branch].Url)
	assert.Equal(t, rpcenv.ProposalState_PROPOSAL_STATE_CLOSED, result[branch].State)
}
