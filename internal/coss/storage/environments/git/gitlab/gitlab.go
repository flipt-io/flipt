// Flipt Commercial Open Source Feature
// This file contains functionality that is licensed under the Flipt Fair Core License (FCL).
// You may NOT use, modify, or distribute this file or its contents without a valid paid license.
// For details: https://github.com/flipt-io/flipt/blob/v2/LICENSE

package gitlab

import (
	"context"
	"fmt"
	"iter"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	gitlab "gitlab.com/gitlab-org/api/client-go"
	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/internal/coss/storage/environments/git"
	"go.flipt.io/flipt/internal/credentials"
	serverenvs "go.flipt.io/flipt/internal/server/environments"
	"go.flipt.io/flipt/rpc/v2/environments"
	"go.uber.org/zap"
)

var _ git.SCM = (*SCM)(nil)

// MergeRequestsService defines the interface for GitLab merge request operations used by SCM.
type MergeRequestsService interface {
	CreateMergeRequest(pid any, opt *gitlab.CreateMergeRequestOptions, options ...gitlab.RequestOptionFunc) (*gitlab.MergeRequest, *gitlab.Response, error)
	ListProjectMergeRequests(pid any, opt *gitlab.ListProjectMergeRequestsOptions, options ...gitlab.RequestOptionFunc) ([]*gitlab.BasicMergeRequest, *gitlab.Response, error)
}

// RepositoriesService defines the interface for GitLab repository operations used by SCM.
type RepositoriesService interface {
	Compare(pid any, opt *gitlab.CompareOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Compare, *gitlab.Response, error)
}

// SCM implements the git.SCM interface for GitLab.
type SCM struct {
	logger    *zap.Logger
	projectID string
	mrs       MergeRequestsService
	repos     RepositoriesService
}

type gitLabOptions struct {
	apiURL     *url.URL
	httpClient *http.Client
	apiAuth    *credentials.APIAuth
}

type ClientOption func(*gitLabOptions)

func WithApiURL(apiURL *url.URL) ClientOption {
	return func(c *gitLabOptions) {
		c.apiURL = apiURL
	}
}

func WithApiAuth(apiAuth *credentials.APIAuth) ClientOption {
	return func(c *gitLabOptions) {
		c.apiAuth = apiAuth
	}
}

// NewSCM creates a new GitLab SCM instance.
func NewSCM(ctx context.Context, logger *zap.Logger, repoOwner, repoName string, opts ...ClientOption) (*SCM, error) {
	gitlabOpts := &gitLabOptions{
		httpClient: http.DefaultClient,
	}

	for _, opt := range opts {
		opt(gitlabOpts)
	}

	var (
		clientOpts []gitlab.ClientOptionFunc
		client     *gitlab.Client
		err        error
	)

	if gitlabOpts.apiURL != nil {
		clientOpts = append(clientOpts, gitlab.WithBaseURL(gitlabOpts.apiURL.String()))
	}

	if gitlabOpts.httpClient != nil {
		clientOpts = append(clientOpts, gitlab.WithHTTPClient(gitlabOpts.httpClient))
	}

	if gitlabOpts.apiAuth != nil {
		// Configure API client authentication
		apiAuth := gitlabOpts.apiAuth
		switch apiAuth.Type() {
		case config.CredentialTypeAccessToken:
			// Use token for API operations
			client, err = gitlab.NewClient(apiAuth.Token, clientOpts...)
		case config.CredentialTypeBasic:
			// Use basic auth for API operations
			client, err = gitlab.NewBasicAuthClient(apiAuth.Username, apiAuth.Password, clientOpts...)
		default:
			return nil, fmt.Errorf("unsupported credential type: %T", apiAuth.Type())
		}
		if err != nil {
			return nil, fmt.Errorf("failed to create gitlab client: %w", err)
		}
	}

	// gitlab project ID can be a numeric ID or the repoOwner/repoName
	projectID := fmt.Sprintf("%s/%s", repoOwner, repoName)

	return &SCM{
		logger:    logger.With(zap.String("repository", projectID), zap.String("scm", "gitlab")),
		projectID: projectID,
		mrs:       client.MergeRequests,
		repos:     client.Repositories,
	}, nil
}

// Propose creates a new merge request with the given request.
func (s *SCM) Propose(ctx context.Context, req git.ProposalRequest) (*environments.EnvironmentProposalDetails, error) {
	s.logger.Info("proposing pull request", zap.String("base", req.Base), zap.String("head", req.Head), zap.String("title", req.Title), zap.Bool("draft", req.Draft))

	createOpts := &gitlab.CreateMergeRequestOptions{
		Title:        &req.Title,
		Description:  &req.Body,
		SourceBranch: &req.Head,
		TargetBranch: &req.Base,
	}

	if req.Draft {
		// gitlab's way of marking a MR as a draft via the API
		// https://forum.gitlab.com/t/creating-draft-mr-with-the-api/55019/3
		createOpts.Title = gitlab.Ptr(fmt.Sprintf("Draft: %s", req.Title))
	}

	mr, _, err := s.mrs.CreateMergeRequest(s.projectID, createOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to create merge request: %w", err)
	}

	s.logger.Info("pull request created", zap.String("pr", mr.WebURL), zap.String("state", mr.State))

	return &environments.EnvironmentProposalDetails{
		Url:   mr.WebURL,
		State: environments.ProposalState_PROPOSAL_STATE_OPEN,
	}, nil
}

// ListChanges compares the base and head branches and returns the changes between them.
func (s *SCM) ListChanges(ctx context.Context, req git.ListChangesRequest) (*environments.ListBranchedEnvironmentChangesResponse, error) {
	s.logger.Info("listing changes", zap.String("base", req.Base), zap.String("head", req.Head))
	compareOpts := &gitlab.CompareOptions{
		From: &req.Base,
		To:   &req.Head,
	}

	comparison, _, err := s.repos.Compare(s.projectID, compareOpts)
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
			Revision: commit.ID,
			Message:  commit.Message,
			ScmUrl:   &commit.WebURL,
		}

		if commit.AuthorName != "" {
			change.AuthorName = &commit.AuthorName
		}
		if commit.AuthorEmail != "" {
			change.AuthorEmail = &commit.AuthorEmail
		}
		if commit.CreatedAt != nil {
			change.Timestamp = commit.CreatedAt.Format(time.RFC3339)
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
		mrs     = s.listMRs(ctx, baseCfg.Ref)
		details = map[string]*environments.EnvironmentProposalDetails{}
	)

	s.logger.Debug("listing proposals for environment",
		zap.String("environment", env.Key()),
		zap.String("base", baseCfg.Ref))

	for mr := range mrs.All() {
		branch := mr.SourceBranch

		s.logger.Debug("checking MR for flipt branch",
			zap.Int("mrID", mr.ID),
			zap.String("branch", branch),
			zap.String("expectedPrefix", fmt.Sprintf("flipt/%s/", env.Key())))

		if !strings.HasPrefix(branch, fmt.Sprintf("flipt/%s/", env.Key())) {
			continue
		}

		s.logger.Debug("found flipt MR",
			zap.Int("mrID", mr.ID),
			zap.String("branch", branch))

		if _, ok := details[branch]; ok {
			// we let existing MRs get replaced by other MRs for the same branch
			// if the existing MR is not in an open state
			if mr.State != "opened" {
				continue
			}
		}

		state := environments.ProposalState_PROPOSAL_STATE_OPEN
		switch mr.State {
		case "closed":
			state = environments.ProposalState_PROPOSAL_STATE_CLOSED
		case "merged":
			state = environments.ProposalState_PROPOSAL_STATE_MERGED
		}

		details[branch] = &environments.EnvironmentProposalDetails{
			Url:   mr.WebURL,
			State: state,
		}
	}

	s.logger.Debug("found proposals for environment",
		zap.String("environment", env.Key()),
		zap.Int("count", len(details)))

	return details, nil
}

type mrs struct {
	logger    *zap.Logger
	ctx       context.Context
	client    MergeRequestsService
	projectID string
	base      string

	err error
}

func (s *SCM) listMRs(ctx context.Context, base string) *mrs {
	return &mrs{s.logger, ctx, s.mrs, s.projectID, base, nil}
}

func (m *mrs) Err() error {
	return m.err
}

func (m *mrs) All() iter.Seq[*gitlab.BasicMergeRequest] {
	return iter.Seq[*gitlab.BasicMergeRequest](func(yield func(*gitlab.BasicMergeRequest) bool) {
		opts := &gitlab.ListProjectMergeRequestsOptions{
			TargetBranch: &m.base,
			State:        gitlab.Ptr("all"),
			ListOptions: gitlab.ListOptions{
				PerPage: 100,
			},
		}

		for {
			mrs, resp, err := m.client.ListProjectMergeRequests(m.projectID, opts)
			if err != nil {
				m.err = err
				return
			}

			for _, mr := range mrs {
				if !strings.HasPrefix(mr.SourceBranch, "flipt/") {
					continue
				}

				if !yield(mr) {
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
