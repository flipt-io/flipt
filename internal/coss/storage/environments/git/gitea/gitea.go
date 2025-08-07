// Flipt Commercial Open Source Feature
// This file contains functionality that is licensed under the Flipt Fair Core License (FCL).
// You may NOT use, modify, or distribute this file or its contents without a valid paid license.
// For details: https://github.com/flipt-io/flipt/blob/v2/LICENSE

package gitea

import (
	"context"
	"fmt"
	"iter"
	"sort"
	"strings"
	"time"

	"code.gitea.io/sdk/gitea"
	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/internal/coss/storage/environments/git"
	"go.flipt.io/flipt/internal/credentials"
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
	logger     *zap.Logger
	client     Client
	owner      string
	repository string
}

type giteaOptions struct {
	apiAuth *credentials.APIAuth
}

type ClientOption func(*giteaOptions)

func WithApiAuth(apiAuth *credentials.APIAuth) ClientOption {
	return func(c *giteaOptions) {
		c.apiAuth = apiAuth
	}
}

func NewSCM(ctx context.Context, logger *zap.Logger, url, owner, repository string, opts ...ClientOption) (*SCM, error) {
	giteaOpts := &giteaOptions{}

	for _, opt := range opts {
		opt(giteaOpts)
	}

	clientOpts := []gitea.ClientOption{}

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
		logger:     logger.With(zap.String("repository", fmt.Sprintf("%s/%s", owner, repository)), zap.String("scm", "gitea")),
		owner:      owner,
		repository: repository,
		client:     client,
	}, nil
}

func (s *SCM) Propose(ctx context.Context, req git.ProposalRequest) (*environments.EnvironmentProposalDetails, error) {
	s.logger.Info("proposing pull request", zap.String("base", req.Base), zap.String("head", req.Head), zap.String("title", req.Title), zap.Bool("draft", req.Draft))
	if req.Draft {
		// gitea's way to say it's a draft PR. It actually could be customized by administrator and
		// [WIP] just is a default value.
		req.Title = fmt.Sprintf("[WIP] %s", req.Title)
	}

	pr, _, err := s.client.CreatePullRequest(s.owner, s.repository, gitea.CreatePullRequestOption{
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
	comparison, _, err := s.client.CompareCommits(s.owner, s.repository, req.Base, req.Head)
	if err != nil {
		return nil, fmt.Errorf("failed to compare branches: %w", err)
	}

	s.logger.Debug("changes compared", zap.Int("commits", len(comparison.Commits)))

	var (
		changes []*environments.Change
		limit   = req.Limit
	)

	for _, commit := range comparison.Commits {
		if limit > 0 && len(changes) >= int(limit) {
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

	// sort changes by timestamp descending if not empty
	if len(changes) > 0 {
		sort.Slice(changes, func(i, j int) bool {
			return changes[i].Timestamp > changes[j].Timestamp
		})
	}

	return &environments.ListBranchedEnvironmentChangesResponse{
		Changes: changes,
	}, nil
}

func (s *SCM) ListProposals(ctx context.Context, env serverenvs.Environment) (map[string]*environments.EnvironmentProposalDetails, error) {
	var (
		baseCfg = env.Configuration()
		details = map[string]*environments.EnvironmentProposalDetails{}
		prs     = s.listPRs(ctx, baseCfg.Ref)
	)

	s.logger.Debug("listing proposals for environment",
		zap.String("environment", env.Key()),
		zap.String("base", baseCfg.Ref))

	for pr := range prs.All() {
		branch := pr.Head.Ref

		if !strings.HasPrefix(branch, fmt.Sprintf("flipt/%s/", env.Key())) {
			continue
		}

		if _, ok := details[branch]; ok {
			// we let existing PRs get replaced by other PRs for the same branch
			// if the existing PR is not in an open state
			if pr.State != gitea.StateOpen {
				continue
			}
		}
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

	s.logger.Debug("found proposals for environment",
		zap.String("environment", env.Key()),
		zap.Int("count", len(details)))

	return details, prs.Err()
}

func (s *SCM) listPRs(ctx context.Context, base string) *prs {
	return &prs{
		logger:     s.logger,
		ctx:        ctx,
		client:     s.client,
		owner:      s.owner,
		repository: s.repository,
		base:       base,
		err:        nil,
	}
}

type prs struct {
	logger     *zap.Logger
	ctx        context.Context
	client     Client
	owner      string
	repository string
	base       string

	err error
}

func (p *prs) Err() error {
	return p.err
}

func (p *prs) All() iter.Seq[*gitea.PullRequest] {
	return iter.Seq[*gitea.PullRequest](func(yield func(*gitea.PullRequest) bool) {
		p.logger.Debug("fetching pull requests with pagination",
			zap.String("owner", p.owner),
			zap.String("repository", p.repository),
			zap.String("base", p.base))

		opts := gitea.ListPullRequestsOptions{
			State: gitea.StateAll,
		}

		// PageSize is embedded under opts.ListOptions
		opts.PageSize = 100

		var totalPRs int
		for {
			prs, resp, err := p.client.ListRepoPullRequests(p.owner, p.repository, opts)
			if err != nil {
				p.err = fmt.Errorf("failed to fetch pull requests: %w", err)
				return
			}

			totalPRs += len(prs)

			for _, pr := range prs {
				if !yield(pr) {
					return
				}
			}

			if resp.NextPage == 0 {
				p.logger.Debug("retrieved pull requests from Gitea",
					zap.Int("totalPRs", totalPRs),
					zap.String("base", p.base))
				return
			}

			opts.Page = resp.NextPage
		}
	})
}
