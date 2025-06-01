package gitea

import (
	"cmp"
	"context"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"

	"code.gitea.io/sdk/gitea"
	"go.flipt.io/flipt/internal/enterprise/storage/environments/git"
	serverenvs "go.flipt.io/flipt/internal/server/environments"
	"go.flipt.io/flipt/rpc/v2/environments"
)

var _ git.SCM = (*SCM)(nil)

type SCM struct {
	client    *gitea.Client
	repoOwner string
	repoName  string
}

type (
	ClientOption = gitea.ClientOption
)

func WithHttpClient(httpClient *http.Client) ClientOption {
	return gitea.SetHTTPClient(httpClient)
}

func NewSCM(url, repoOwner, repoName string, opts ...ClientOption) (*SCM, error) {
	client, err := gitea.NewClient(url, opts...)
	if err != nil {
		return nil, err
	}
	return &SCM{
		repoOwner: repoOwner,
		repoName:  repoName,
		client:    client,
	}, nil
}

func (s *SCM) Propose(ctx context.Context, req git.ProposalRequest) (*environments.EnvironmentProposalDetails, error) {
	title := req.Title
	if req.Draft {
		// gitea's way to say it's a draft PR. It actually could be customized by administrator and
		// [WIP] just is a default value.
		title = "[WIP] " + title
	}
	pr, _, err := s.client.CreatePullRequest(s.repoOwner, s.repoName, gitea.CreatePullRequestOption{
		Base:  req.Base,
		Head:  req.Head,
		Title: title,
		Body:  req.Body,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create pull request: %w", err)
	}
	return &environments.EnvironmentProposalDetails{
		Scm:   environments.SCM_GITHUB_SCM,
		Url:   pr.HTMLURL,
		State: environments.ProposalState_PROPOSAL_STATE_OPEN,
	}, nil
}

func (s *SCM) ListChanges(ctx context.Context, req git.ListChangesRequest) (*environments.ListBranchedEnvironmentChangesResponse, error) {
	comparition, _, err := s.client.CompareCommits(s.repoOwner, s.repoName, req.Base, req.Head)
	if err != nil {
		return nil, fmt.Errorf("failed to compare branches: %w", err)
	}
	var (
		changes []*environments.Change
		limit   = req.Limit
	)

	for _, commit := range comparition.Commits {
		if limit > 0 && int32(len(changes)) >= limit {
			break
		}

		change := &environments.Change{
			Revision:  commit.SHA,
			Message:   commit.RepoCommit.Message,
			ScmUrl:    &commit.HTMLURL,
			Timestamp: commit.Created.Format(time.RFC3339),
		}

		if commit.Author != nil {
			change.AuthorName = &commit.Author.FullName
			change.AuthorEmail = &commit.Author.Email
		}

		changes = append(changes, change)
	}

	slices.SortFunc(changes, func(i, j *environments.Change) int {
		return cmp.Compare(i.Timestamp, i.Timestamp)
	})

	return &environments.ListBranchedEnvironmentChangesResponse{
		Changes: changes,
	}, nil
}

func (s *SCM) ListProposals(ctx context.Context, env serverenvs.Environment) (map[string]*environments.EnvironmentProposalDetails, error) {
	details := map[string]*environments.EnvironmentProposalDetails{}
	prs, _, err := s.client.ListRepoPullRequests(s.repoOwner, s.repoName, gitea.ListPullRequestsOptions{
		State: gitea.StateAll,
	})
	if err != nil {
		return nil, err
	}

	for _, pr := range prs {
		branch := pr.Head.Ref
		if !strings.HasPrefix(branch, fmt.Sprintf("flipt/%s/", env.Key())) {
			continue
		}
		state := environments.ProposalState_PROPOSAL_STATE_OPEN
		if pr.State == gitea.StateClosed {
			state = environments.ProposalState_PROPOSAL_STATE_CLOSED
			if pr.Merged != nil || pr.MergedCommitID != nil {
				state = environments.ProposalState_PROPOSAL_STATE_MERGED
			}
		}

		details[branch] = &environments.EnvironmentProposalDetails{
			Scm:   environments.SCM_GITHUB_SCM,
			Url:   pr.HTMLURL,
			State: state,
		}
	}

	return details, nil
}
