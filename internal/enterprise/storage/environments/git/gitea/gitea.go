package gitea

import (
	"cmp"
	"context"
	"fmt"
	"iter"
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

type (
	ClientOption = gitea.ClientOption
	Client       interface {
		CreatePullRequest(owner, repo string, opt gitea.CreatePullRequestOption) (*gitea.PullRequest, *gitea.Response, error)
		CompareCommits(user, repo, prev, current string) (*gitea.Compare, *gitea.Response, error)
		ListRepoPullRequests(owner, repo string, opt gitea.ListPullRequestsOptions) ([]*gitea.PullRequest, *gitea.Response, error)
	}
)

type SCM struct {
	client    Client
	repoOwner string
	repoName  string
}

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
		Scm:   environments.SCM_GITEA_SCM,
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
		return cmp.Compare(i.Timestamp, j.Timestamp)
	})

	return &environments.ListBranchedEnvironmentChangesResponse{
		Changes: changes,
	}, nil
}

func (s *SCM) ListProposals(ctx context.Context, env serverenvs.Environment) (map[string]*environments.EnvironmentProposalDetails, error) {
	details := map[string]*environments.EnvironmentProposalDetails{}
	prs := s.listPRs(ctx, env.Key())

	for pr := range prs.All() {
		fmt.Printf("Found PR: %s\n", pr.HTMLURL)
		branch := pr.Head.Ref
		state := environments.ProposalState_PROPOSAL_STATE_OPEN
		if pr.State == gitea.StateClosed {
			state = environments.ProposalState_PROPOSAL_STATE_CLOSED
			if pr.Merged != nil || pr.MergedCommitID != nil {
				state = environments.ProposalState_PROPOSAL_STATE_MERGED
			}
		}

		details[branch] = &environments.EnvironmentProposalDetails{
			Scm:   environments.SCM_GITEA_SCM,
			Url:   pr.HTMLURL,
			State: state,
		}
	}

	return details, prs.Err()
}

func (s *SCM) listPRs(ctx context.Context, base string) *prs {
	return &prs{ctx, s.client, s.repoOwner, s.repoName, base, nil}
}

type prs struct {
	ctx       context.Context
	client    Client
	repoOwner string
	repoName  string
	base      string

	err error
}

func (p *prs) Err() error {
	return p.err
}

func (p *prs) All() iter.Seq[*gitea.PullRequest] {
	return iter.Seq[*gitea.PullRequest](func(yield func(*gitea.PullRequest) bool) {
		opts := gitea.ListPullRequestsOptions{
			State: gitea.StateAll,
		}
		opts.PageSize = 100
		prefix := fmt.Sprintf("flipt/%s/", p.base)
		for {
			prs, resp, err := p.client.ListRepoPullRequests(p.repoOwner, p.repoName, opts)
			if err != nil {
				p.err = err
				return
			}

			for _, pr := range prs {
				if !strings.HasPrefix(pr.Head.Ref, prefix) || pr.Base.Name != p.base {
					continue
				}

				if !yield(pr) {
					return
				}
			}

			if resp.NextPage == 0 {
				return
			}

			opts.Page = resp.NextPage
		}
	})
}
