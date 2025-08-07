// Flipt Commercial Open Source Feature
// This file contains functionality that is licensed under the Flipt Fair Core License (FCL).
// You may NOT use, modify, or distribute this file or its contents without a valid paid license.
// For details: https://github.com/flipt-io/flipt/blob/v2/LICENSE

package bitbucket

import (
	"context"
	"fmt"
	"iter"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/ktrysmt/go-bitbucket"
	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/internal/coss/storage/environments/git"
	"go.flipt.io/flipt/internal/credentials"
	serverenvs "go.flipt.io/flipt/internal/server/environments"
	"go.flipt.io/flipt/rpc/v2/environments"
	"go.uber.org/zap"
)

var _ git.SCM = (*SCM)(nil)

// PullRequestsService defines the interface for BitBucket pull request operations used by SCM.
type PullRequestsService interface {
	Create(po *bitbucket.PullRequestsOptions) (any, error)
	Gets(po *bitbucket.PullRequestsOptions) (any, error)
}

// CommitsService defines the interface for BitBucket commits operations used by SCM.
type CommitsService interface {
	GetCommits(cmo *bitbucket.CommitsOptions) (any, error)
}

// SCM implements the git.SCM interface for BitBucket.
type SCM struct {
	logger   *zap.Logger
	owner    string
	repoSlug string
	prs      PullRequestsService
	commits  CommitsService
}

type bitBucketOptions struct {
	apiURL  string
	apiAuth *credentials.APIAuth
}

type ClientOption func(*bitBucketOptions)

func WithApiURL(apiURL string) ClientOption {
	return func(c *bitBucketOptions) {
		c.apiURL = apiURL
	}
}

func WithApiAuth(apiAuth *credentials.APIAuth) ClientOption {
	return func(c *bitBucketOptions) {
		c.apiAuth = apiAuth
	}
}

// NewSCM creates a new BitBucket SCM instance.
func NewSCM(ctx context.Context, logger *zap.Logger, owner, repoSlug string, opts ...ClientOption) (*SCM, error) {
	bitbucketOpts := &bitBucketOptions{
		apiURL: bitbucket.DEFAULT_BITBUCKET_API_BASE_URL,
	}

	for _, opt := range opts {
		opt(bitbucketOpts)
	}

	var client *bitbucket.Client

	if bitbucketOpts.apiAuth != nil {
		// Configure API client authentication
		apiAuth := bitbucketOpts.apiAuth
		switch apiAuth.Type() {
		case config.CredentialTypeAccessToken:
			// Use bearer token for API operations
			client = bitbucket.NewOAuthbearerToken(apiAuth.Token)
		case config.CredentialTypeBasic:
			// Use basic auth for API operations
			client = bitbucket.NewBasicAuth(apiAuth.Username, apiAuth.Password)
		default:
			return nil, fmt.Errorf("unsupported credential type: %T", apiAuth.Type())
		}
	} else {
		// Create an unauthenticated client (for public repos)
		client = bitbucket.NewBasicAuth("", "")
	}

	// Configure pagination settings - use optimal defaults
	// Similar to GitHub provider which uses PerPage: 100
	client.Pagelen = 50              // Optimal page size for BitBucket API
	client.LimitPages = 0            // No limit - fetch all pages
	client.DisableAutoPaging = false // Always enable auto-pagination for complete results

	// Set custom API URL if provided (for BitBucket Server instances)
	if bitbucketOpts.apiURL != bitbucket.DEFAULT_BITBUCKET_API_BASE_URL {
		if apiURL, err := url.Parse(bitbucketOpts.apiURL); err == nil {
			client.SetApiBaseURL(*apiURL)
		}
	}

	return &SCM{
		logger:   logger.With(zap.String("repository", fmt.Sprintf("%s/%s", owner, repoSlug)), zap.String("scm", "bitbucket")),
		owner:    owner,
		repoSlug: repoSlug,
		prs:      client.Repositories.PullRequests,
		commits:  client.Repositories.Commits,
	}, nil
}

// Propose creates a new pull request with the given request.
func (s *SCM) Propose(ctx context.Context, req git.ProposalRequest) (*environments.EnvironmentProposalDetails, error) {
	s.logger.Info("proposing pull request", zap.String("base", req.Base), zap.String("head", req.Head), zap.String("title", req.Title), zap.Bool("draft", req.Draft))

	title := req.Title
	// BitBucket API v2.0 doesn't support native draft PRs, use title prefix convention
	if req.Draft {
		title = "[DRAFT] " + title
	}

	pr, err := s.prs.Create(&bitbucket.PullRequestsOptions{
		Owner:             s.owner,
		RepoSlug:          s.repoSlug,
		Title:             title,
		Description:       req.Body,
		SourceBranch:      req.Head,
		DestinationBranch: req.Base,
		CloseSourceBranch: false, // Keep source branch by default
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create pull request: %w", err)
	}

	// Cast response to map to extract URL and state
	prMap, ok := pr.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("unexpected response format from BitBucket API")
	}

	var prURL string
	if links, ok := prMap["links"].(map[string]any); ok {
		if html, ok := links["html"].(map[string]any); ok {
			if href, ok := html["href"].(string); ok {
				prURL = href
			}
		}
	}

	s.logger.Info("pull request created", zap.String("pr", prURL), zap.String("state", "OPEN"))

	return &environments.EnvironmentProposalDetails{
		Url:   prURL,
		State: environments.ProposalState_PROPOSAL_STATE_OPEN,
	}, nil
}

// ListChanges compares the base and head branches and returns the changes between them.
func (s *SCM) ListChanges(ctx context.Context, req git.ListChangesRequest) (*environments.ListBranchedEnvironmentChangesResponse, error) {
	s.logger.Info("listing changes", zap.String("base", req.Base), zap.String("head", req.Head))

	// Use commits endpoint to get commits between base and head
	// BitBucket API uses include/exclude pattern for commit range queries
	commitsResp, err := s.commits.GetCommits(&bitbucket.CommitsOptions{
		Owner:    s.owner,
		RepoSlug: s.repoSlug,
		Include:  req.Head,
		Exclude:  req.Base,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get commits: %w", err)
	}

	// Cast response to expected format
	commitsMap, ok := commitsResp.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("unexpected response format from BitBucket commits API")
	}

	var changes []*environments.Change
	limit := req.Limit

	if values, ok := commitsMap["values"].([]any); ok {
		s.logger.Debug("changes compared", zap.Int("commits", len(values)))

		for _, commitVal := range values {
			if limit > 0 && len(changes) >= int(limit) {
				break
			}

			commit, ok := commitVal.(map[string]any)
			if !ok {
				continue
			}

			change := &environments.Change{}

			// Extract commit hash
			if hash, ok := commit["hash"].(string); ok {
				change.Revision = hash
			}

			// Extract commit message
			if message, ok := commit["message"].(string); ok {
				change.Message = message
			}

			// Extract commit URL
			if links, ok := commit["links"].(map[string]any); ok {
				if html, ok := links["html"].(map[string]any); ok {
					if href, ok := html["href"].(string); ok {
						change.ScmUrl = &href
					}
				}
			}

			// Extract author information
			if author, ok := commit["author"].(map[string]any); ok {
				if user, ok := author["user"].(map[string]any); ok {
					if displayName, ok := user["display_name"].(string); ok {
						change.AuthorName = &displayName
					}
				}
				// BitBucket commits don't always include email in the response
			}

			// Extract timestamp
			if date, ok := commit["date"].(string); ok {
				if parsedTime, err := time.Parse(time.RFC3339, date); err == nil {
					change.Timestamp = parsedTime.Format(time.RFC3339)
				}
			}

			changes = append(changes, change)
		}
	}

	// Sort changes by timestamp descending if not empty
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
		// Extract branch name from PR
		branch := ""
		if source, ok := pr["source"].(map[string]any); ok {
			if branchInfo, ok := source["branch"].(map[string]any); ok {
				if name, ok := branchInfo["name"].(string); ok {
					branch = name
				}
			}
		}

		if !strings.HasPrefix(branch, fmt.Sprintf("flipt/%s/", env.Key())) {
			continue
		}

		if _, ok := details[branch]; ok {
			// We let existing PRs get replaced by other PRs for the same branch
			// if the existing PR is not in an open state
			if state, ok := pr["state"].(string); ok && state != "OPEN" {
				continue
			}
		}

		state := environments.ProposalState_PROPOSAL_STATE_OPEN
		if prState, ok := pr["state"].(string); ok {
			switch prState {
			case "MERGED":
				state = environments.ProposalState_PROPOSAL_STATE_MERGED
			case "DECLINED", "SUPERSEDED":
				state = environments.ProposalState_PROPOSAL_STATE_CLOSED
			default:
				state = environments.ProposalState_PROPOSAL_STATE_OPEN
			}
		}

		var prURL string
		if links, ok := pr["links"].(map[string]any); ok {
			if html, ok := links["html"].(map[string]any); ok {
				if href, ok := html["href"].(string); ok {
					prURL = href
				}
			}
		}

		details[branch] = &environments.EnvironmentProposalDetails{
			Url:   prURL,
			State: state,
		}
	}

	s.logger.Debug("found proposals for environment",
		zap.String("environment", env.Key()),
		zap.Int("count", len(details)))

	return details, nil
}

type prs struct {
	logger   *zap.Logger
	ctx      context.Context
	client   PullRequestsService
	owner    string
	repoSlug string
	base     string

	err error
}

func (s *SCM) listPRs(ctx context.Context, base string) *prs {
	return &prs{s.logger, ctx, s.prs, s.owner, s.repoSlug, base, nil}
}

func (p *prs) Err() error {
	return p.err
}

func (p *prs) All() iter.Seq[map[string]any] {
	return iter.Seq[map[string]any](func(yield func(map[string]any) bool) {
		p.logger.Debug("fetching pull requests with automatic pagination",
			zap.String("owner", p.owner),
			zap.String("repoSlug", p.repoSlug),
			zap.String("base", p.base))

		// Get pull requests for the repository
		// Use Bitbucket's query API to filter PRs server-side for better performance
		// Query for PRs where source.branch.name starts with "flipt/" and destination matches our base
		// The go-bitbucket library automatically handles pagination
		prsResp, err := p.client.Gets(&bitbucket.PullRequestsOptions{
			Owner:    p.owner,
			RepoSlug: p.repoSlug,
			States:   []string{"OPEN", "MERGED", "DECLINED", "SUPERSEDED"},
			Query:    fmt.Sprintf(`source.branch.name ~ "flipt/" AND destination.branch.name = "%s"`, p.base),
		})
		if err != nil {
			p.err = fmt.Errorf("failed to fetch pull requests: %w", err)
			return
		}

		// Cast response to expected format
		prsMap, ok := prsResp.(map[string]any)
		if !ok {
			p.err = fmt.Errorf("unexpected response format from BitBucket PRs API")
			return
		}

		values, ok := prsMap["values"].([]any)
		if !ok {
			p.err = fmt.Errorf("unexpected response format: missing 'values' field")
			return
		}

		// All PRs returned should already match our criteria due to server-side filtering
		// We still iterate through them to yield each one
		for _, prVal := range values {
			pr, ok := prVal.(map[string]any)
			if !ok {
				continue
			}

			if !yield(pr) {
				return
			}
		}
	})
}
