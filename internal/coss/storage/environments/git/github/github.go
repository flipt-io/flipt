// Flipt Commercial Open Source Feature
// This file contains functionality that is licensed under the Flipt Fair Core License (FCL).
// You may NOT use, modify, or distribute this file or its contents without a valid paid license.
// For details: https://github.com/flipt-io/flipt/blob/v2/LICENSE

package github

import (
	"context"
	"encoding/base64"
	"fmt"
	"iter"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/google/go-github/v75/github"
	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/internal/coss/storage/environments/git"
	"go.flipt.io/flipt/internal/credentials"
	serverenvs "go.flipt.io/flipt/internal/server/environments"
	"go.flipt.io/flipt/rpc/v2/environments"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

var _ git.SCM = (*SCM)(nil)

// PullRequestsService defines the interface for GitHub pull request operations used by SCM.
type PullRequestsService interface {
	Create(ctx context.Context, owner, repo string, pr *github.NewPullRequest) (*github.PullRequest, *github.Response, error)
	List(ctx context.Context, owner, repo string, opts *github.PullRequestListOptions) ([]*github.PullRequest, *github.Response, error)
}

// RepositoriesService defines the interface for GitHub repository operations used by SCM.
type RepositoriesService interface {
	CompareCommits(ctx context.Context, owner, repo, base, head string, opts *github.ListOptions) (*github.CommitsComparison, *github.Response, error)
}

// SCM implements the git.SCM interface for GitHub.
type SCM struct {
	logger     *zap.Logger
	owner      string
	repository string
	prs        PullRequestsService
	repos      RepositoriesService
}

type gitHubOptions struct {
	apiURL     *url.URL
	httpClient *http.Client
	apiAuth    *credentials.APIAuth
}

type ClientOption func(*gitHubOptions)

func WithApiURL(apiURL *url.URL) ClientOption {
	return func(c *gitHubOptions) {
		// copied from go-github/github.go:WithEnterpriseURLs
		if !strings.HasSuffix(apiURL.Path, "/") {
			apiURL.Path += "/"
		}
		if !strings.HasSuffix(apiURL.Path, "/api/v3/") &&
			!strings.HasPrefix(apiURL.Host, "api.") &&
			!strings.Contains(apiURL.Host, ".api.") {
			apiURL.Path += "api/v3/"
		}

		c.apiURL = apiURL
	}
}

func WithApiAuth(apiAuth *credentials.APIAuth) ClientOption {
	return func(c *gitHubOptions) {
		c.apiAuth = apiAuth
	}
}

// NewSCM creates a new GitHub SCM instance.
func NewSCM(ctx context.Context, logger *zap.Logger, owner, repository string, opts ...ClientOption) (*SCM, error) {
	githubOpts := &gitHubOptions{
		httpClient: http.DefaultClient,
	}

	for _, opt := range opts {
		opt(githubOpts)
	}

	client := github.NewClient(githubOpts.httpClient)

	if githubOpts.apiURL != nil {
		client.BaseURL = githubOpts.apiURL
	}

	if githubOpts.apiAuth != nil {
		// Configure API client authentication
		apiAuth := githubOpts.apiAuth
		switch apiAuth.Type() {
		case config.CredentialTypeAccessToken:
			// Use token for API operations
			client = client.WithAuthToken(apiAuth.Token)
		case config.CredentialTypeBasic:
			// Use basic auth for API operations - convert to OAuth2 token format
			client = github.NewClient(oauth2.NewClient(ctx, oauth2.StaticTokenSource(
				&oauth2.Token{
					TokenType:   "Basic",
					AccessToken: base64.StdEncoding.EncodeToString(fmt.Appendf(nil, "%s:%s", apiAuth.Username, apiAuth.Password)),
				}),
			))
		default:
			return nil, fmt.Errorf("unsupported credential type: %T", apiAuth.Type())
		}
	}

	return &SCM{
		logger:     logger.With(zap.String("repository", fmt.Sprintf("%s/%s", owner, repository)), zap.String("scm", "github")),
		owner:      owner,
		repository: repository,
		prs:        client.PullRequests,
		repos:      client.Repositories,
	}, nil
}

// Propose creates a new pull request with the given request.
func (s *SCM) Propose(ctx context.Context, req git.ProposalRequest) (*environments.EnvironmentProposalDetails, error) {
	s.logger.Info("proposing pull request", zap.String("base", req.Base), zap.String("head", req.Head), zap.String("title", req.Title), zap.Bool("draft", req.Draft))

	pr, _, err := s.prs.Create(ctx, s.owner, s.repository, &github.NewPullRequest{
		Base:  github.Ptr(req.Base),
		Head:  github.Ptr(req.Head),
		Title: github.Ptr(req.Title),
		Body:  github.Ptr(req.Body),
		Draft: github.Ptr(req.Draft),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create pull request: %w", err)
	}

	s.logger.Info("pull request created", zap.String("pr", pr.GetHTMLURL()), zap.String("state", pr.GetState()))

	return &environments.EnvironmentProposalDetails{
		Url:   pr.GetHTMLURL(),
		State: environments.ProposalState_PROPOSAL_STATE_OPEN,
	}, nil
}

// ListChanges compares the base and head branches and returns the changes between them.
func (s *SCM) ListChanges(ctx context.Context, req git.ListChangesRequest) (*environments.ListBranchedEnvironmentChangesResponse, error) {
	s.logger.Info("listing changes", zap.String("base", req.Base), zap.String("head", req.Head))
	comparison, _, err := s.repos.CompareCommits(ctx, s.owner, s.repository, req.Base, req.Head, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to compare branches: %w", err)
	}

	s.logger.Debug("changes compared", zap.Int("commits", len(comparison.Commits)))

	var (
		changes []*environments.Change
		limit   = req.Limit
	)

	for _, commit := range comparison.Commits {
		if commit == nil || commit.GetCommit() == nil {
			continue
		}

		if limit > 0 && len(changes) >= int(limit) {
			break
		}

		change := &environments.Change{
			Revision: commit.GetSHA(),
			Message:  commit.GetCommit().GetMessage(),
			ScmUrl:   github.Ptr(commit.GetHTMLURL()),
		}

		if commit.GetCommit().GetAuthor() != nil {
			change.AuthorName = github.Ptr(commit.GetCommit().GetAuthor().GetName())
			change.AuthorEmail = github.Ptr(commit.GetCommit().GetAuthor().GetEmail())
			change.Timestamp = commit.GetCommit().GetAuthor().GetDate().Format(time.RFC3339)
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

// ListProposals lists all proposals for the given environment.
func (s *SCM) ListProposals(ctx context.Context, env serverenvs.Environment) (map[string]*environments.EnvironmentProposalDetails, error) {
	var (
		baseCfg = env.Configuration()
		prs     = s.listPRs(ctx, baseCfg.Ref)
		details = map[string]*environments.EnvironmentProposalDetails{}
	)

	s.logger.Debug("listing proposals for environment",
		zap.String("environment", env.Key()),
		zap.String("base", baseCfg.Ref))

	for pr := range prs.All() {
		branch := pr.Head.GetRef()

		s.logger.Debug("checking PR for flipt branch",
			zap.Int64("prID", pr.GetID()),
			zap.String("branch", branch),
			zap.String("expectedPrefix", fmt.Sprintf("flipt/%s/", env.Key())))

		if !strings.HasPrefix(branch, fmt.Sprintf("flipt/%s/", env.Key())) {
			continue
		}

		s.logger.Debug("found flipt PR",
			zap.Int64("prID", pr.GetID()),
			zap.String("branch", branch))

		if _, ok := details[branch]; ok {
			// we let existing PRs get replaced by other PRs for the same branch
			// if the existing PR is not in an open state
			if pr.GetState() != "open" {
				continue
			}
		}

		state := environments.ProposalState_PROPOSAL_STATE_OPEN
		if pr.GetState() == "closed" {
			state = environments.ProposalState_PROPOSAL_STATE_CLOSED
			if pr.GetMerged() || pr.MergeCommitSHA != nil {
				state = environments.ProposalState_PROPOSAL_STATE_MERGED
			}
		}

		details[branch] = &environments.EnvironmentProposalDetails{
			Url:   pr.GetHTMLURL(),
			State: state,
		}
	}

	s.logger.Debug("found proposals for environment",
		zap.String("environment", env.Key()),
		zap.Int("count", len(details)))

	return details, nil
}

type prs struct {
	logger     *zap.Logger
	ctx        context.Context
	client     PullRequestsService
	owner      string
	repository string
	base       string

	err error
}

func (s *SCM) listPRs(ctx context.Context, base string) *prs {
	return &prs{
		logger:     s.logger,
		ctx:        ctx,
		client:     s.prs,
		owner:      s.owner,
		repository: s.repository,
		base:       base,
		err:        nil,
	}
}

func (p *prs) Err() error {
	return p.err
}

func (p *prs) All() iter.Seq[*github.PullRequest] {
	return iter.Seq[*github.PullRequest](func(yield func(*github.PullRequest) bool) {
		p.logger.Debug("fetching pull requests with pagination",
			zap.String("owner", p.owner),
			zap.String("repository", p.repository),
			zap.String("base", p.base))

		opts := &github.PullRequestListOptions{
			Base: p.base,
			ListOptions: github.ListOptions{
				PerPage: 100,
			},
			State: "all",
		}

		var totalPRs int
		for {
			prs, resp, err := p.client.List(p.ctx, p.owner, p.repository, opts)
			if err != nil {
				p.err = err
				return
			}

			totalPRs += len(prs)

			for _, pr := range prs {
				if !strings.HasPrefix(pr.Head.GetRef(), "flipt/") {
					continue
				}

				if !yield(pr) {
					return
				}
			}

			if resp.NextPage == 0 {
				p.logger.Debug("retrieved pull requests from GitHub",
					zap.Int("totalPRs", totalPRs),
					zap.String("base", p.base))
				return
			}

			opts.Page = resp.NextPage
		}
	})
}
