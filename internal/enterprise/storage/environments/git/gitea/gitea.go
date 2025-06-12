// Flipt Enterprise-Only Feature
// This file contains functionality that is licensed under the Flipt Fair Core License (FCL).
// You may NOT use, modify, or distribute this file or its contents without a valid Enterprise license.
// For details: https://github.com/flipt-io/flipt/blob/v2/LICENSE

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
	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/internal/credentials"
	"go.flipt.io/flipt/internal/enterprise/storage/environments/git"
	serverenvs "go.flipt.io/flipt/internal/server/environments"
	"go.flipt.io/flipt/rpc/v2/environments"
	"go.uber.org/zap"
)

var _ git.SCM = (*SCM)(nil)

type Client interface {
	CreatePullRequest(owner, repo string, opt gitea.CreatePullRequestOption) (*gitea.PullRequest, *gitea.Response, error)
	CompareCommits(user, repo, prev, current string) (*gitea.Compare, *gitea.Response, error)
	ListRepoPullRequests(owner, repo string, opt gitea.ListPullRequestsOptions) ([]*gitea.PullRequest, *gitea.Response, error)
}

type SCM struct {
	logger    *zap.Logger
	client    Client
	repoOwner string
	repoName  string
}

type giteaOptions struct {
	httpClient *http.Client
	apiAuth    *credentials.APIAuth
}

type ClientOption func(*giteaOptions)

func WithApiAuth(apiAuth *credentials.APIAuth) ClientOption {
	return func(c *giteaOptions) {
		c.apiAuth = apiAuth
	}
}

func NewSCM(logger *zap.Logger, url, repoOwner, repoName string, opts ...ClientOption) (*SCM, error) {
	giteaOpts := &giteaOptions{
		httpClient: http.DefaultClient,
	}

	for _, opt := range opts {
		opt(giteaOpts)
	}

	clientOpts := []gitea.ClientOption{}

	if giteaOpts.httpClient != nil {
		clientOpts = append(clientOpts, gitea.SetHTTPClient(giteaOpts.httpClient))
	}

	if giteaOpts.apiAuth != nil {
		// Configure API client authentication
		apiAuth := giteaOpts.apiAuth
		switch apiAuth.Type() {
		case config.CredentialTypeAccessToken:
			// Use token for API operations
			clientOpts = append(clientOpts, gitea.SetToken(apiAuth.Token))
		case config.CredentialTypeBasic:
			// Use basic auth for API operations
			clientOpts = append(clientOpts, gitea.SetBasicAuth(apiAuth.Username, apiAuth.Password))
		default:
			return nil, fmt.Errorf("unsupported credential type: %T", apiAuth.Type())
		}
	}

	client, err := gitea.NewClient(url, clientOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create gitea client: %w", err)
	}

	return &SCM{
		logger:    logger.With(zap.String("repository", fmt.Sprintf("%s/%s", repoOwner, repoName)), zap.String("scm", "gitea")),
		repoOwner: repoOwner,
		repoName:  repoName,
		client:    client,
	}, nil
}

func (s *SCM) Propose(ctx context.Context, req git.ProposalRequest) (*environments.EnvironmentProposalDetails, error) {
	s.logger.Info("proposing pull request", zap.String("base", req.Base), zap.String("head", req.Head), zap.String("title", req.Title), zap.Bool("draft", req.Draft))
	if req.Draft {
		// gitea's way to say it's a draft PR. It actually could be customized by administrator and
		// [WIP] just is a default value.
		req.Title = fmt.Sprintf("[WIP] %s", req.Title)
	}

	pr, _, err := s.client.CreatePullRequest(s.repoOwner, s.repoName, gitea.CreatePullRequestOption{
		Base:  req.Base,
		Head:  req.Head,
		Title: req.Title,
		Body:  req.Body,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create pull request: %w", err)
	}

	s.logger.Info("pull request created", zap.String("pr", pr.HTMLURL), zap.String("state", string(pr.State)))

	return &environments.EnvironmentProposalDetails{
		Url:   pr.HTMLURL,
		State: environments.ProposalState_PROPOSAL_STATE_OPEN,
	}, nil
}

func (s *SCM) ListChanges(ctx context.Context, req git.ListChangesRequest) (*environments.ListBranchedEnvironmentChangesResponse, error) {
	s.logger.Info("listing changes", zap.String("base", req.Base), zap.String("head", req.Head))
	comparition, _, err := s.client.CompareCommits(s.repoOwner, s.repoName, req.Base, req.Head)
	if err != nil {
		return nil, fmt.Errorf("failed to compare branches: %w", err)
	}

	s.logger.Info("changes compared", zap.Int("commits", len(comparition.Commits)))

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
	var (
		details = map[string]*environments.EnvironmentProposalDetails{}
		prs     = s.listPRs(ctx, env.Key())
	)

	for pr := range prs.All() {
		branch := pr.Head.Ref
		state := environments.ProposalState_PROPOSAL_STATE_OPEN
		if pr.State == gitea.StateClosed {
			state = environments.ProposalState_PROPOSAL_STATE_CLOSED
			if pr.Merged != nil || pr.MergedCommitID != nil {
				state = environments.ProposalState_PROPOSAL_STATE_MERGED
			}
		}

		details[branch] = &environments.EnvironmentProposalDetails{
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
