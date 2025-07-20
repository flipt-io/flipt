// Flipt Commercial Open Source Feature
// This file contains functionality that is licensed under the Flipt Fair Core License (FCL).
// You may NOT use, modify, or distribute this file or its contents without a valid paid license.
// For details: https://github.com/flipt-io/flipt/blob/v2/LICENSE
package azure

import (
	"context"
	"fmt"
	"iter"
	"net/url"
	"strings"
	"time"

	"github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	azuregit "github.com/microsoft/azure-devops-go-api/azuredevops/v7/git"
	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/internal/coss/storage/environments/git"
	"go.flipt.io/flipt/internal/credentials"
	serverenvs "go.flipt.io/flipt/internal/server/environments"
	"go.flipt.io/flipt/rpc/v2/environments"
	"go.uber.org/zap"
)

var _ git.SCM = (*SCM)(nil)

type Client interface {
	CreatePullRequest(context.Context, azuregit.CreatePullRequestArgs) (*azuregit.GitPullRequest, error)
	GetPullRequests(context.Context, azuregit.GetPullRequestsArgs) (*[]azuregit.GitPullRequest, error)
	GetCommits(context.Context, azuregit.GetCommitsArgs) (*[]azuregit.GitCommitRef, error)
}

type azureOptions struct {
	apiURL              *url.URL
	personalAccessToken string
	logger              *zap.Logger
}

type ClientOption func(*azureOptions)

// WithApiURL sets the API URL for the Azure DevOps client.
func WithApiURL(apiURL *url.URL) ClientOption {
	return func(c *azureOptions) {
		c.apiURL = apiURL
	}
}

// WithApiAuth sets the API authentication credentials for the Azure DevOps client.
func WithApiAuth(apiAuth *credentials.APIAuth) ClientOption {
	return func(c *azureOptions) {
		if apiAuth == nil {
			return
		}
		switch apiAuth.Type() {
		case config.CredentialTypeBasic:
			c.personalAccessToken = apiAuth.Password
		case config.CredentialTypeAccessToken:
			c.personalAccessToken = apiAuth.Token
		}
	}
}

// WithLogger sets the logger for the SCM client.
func WithLogger(logger *zap.Logger) ClientOption {
	return func(c *azureOptions) {
		if logger != nil {
			c.logger = logger
		}
	}
}

// NewSCM creates a new SCM instance for Azure DevOps Git.
func NewSCM(ctx context.Context, repoOwner, repoProject, repoName string, opts ...ClientOption) (*SCM, error) {
	options := &azureOptions{
		logger: zap.NewNop(),
	}
	for _, opt := range opts {
		opt(options)
	}

	connection := azuredevops.NewPatConnection(fmt.Sprintf("%s/%s", options.apiURL.String(), repoOwner), options.personalAccessToken)
	client, err := azuregit.NewClient(ctx, connection)
	if err != nil {
		return nil, fmt.Errorf("failed to create azure git client: %w", err)
	}

	return &SCM{
		logger:      options.logger.With(zap.String("repository", fmt.Sprintf("%s/%s", repoProject, repoName)), zap.String("scm", "azure git")),
		client:      client,
		repoOwner:   repoOwner,
		repoProject: repoProject,
		repoName:    repoName,
		baseURL:     options.apiURL.String(),
	}, nil
}

type SCM struct {
	logger      *zap.Logger
	client      Client
	repoOwner   string
	repoProject string
	repoName    string
	baseURL     string
}

func (s *SCM) ListChanges(ctx context.Context, req git.ListChangesRequest) (*environments.ListBranchedEnvironmentChangesResponse, error) {
	changes := []*environments.Change{}
	includeLinks := true
	commits, err := s.client.GetCommits(ctx, azuregit.GetCommitsArgs{
		RepositoryId: &s.repoName,
		Project:      &s.repoProject,
		SearchCriteria: &azuregit.GitQueryCommitsCriteria{
			IncludeLinks: &includeLinks,
			ItemVersion: &azuregit.GitVersionDescriptor{
				Version:     &req.Base,
				VersionType: &azuregit.GitVersionTypeValues.Branch,
			},
			CompareVersion: &azuregit.GitVersionDescriptor{
				Version:     &req.Head,
				VersionType: &azuregit.GitVersionTypeValues.Branch,
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to compare branches: %w", err)
	}

	for _, commit := range *commits {
		change := &environments.Change{
			Revision: *commit.CommitId,
			Message:  *commit.Comment,
			ScmUrl:   commit.RemoteUrl,
		}
		if commit.Author != nil {
			if commit.Author.Date != nil {
				change.Timestamp = commit.Author.Date.Time.Format(time.RFC3339)
			}
			change.AuthorName = commit.Author.Name
			change.AuthorEmail = commit.Author.Email
		}
		changes = append(changes, change)
	}
	return &environments.ListBranchedEnvironmentChangesResponse{
		Changes: changes,
	}, nil
}

func (s *SCM) Propose(ctx context.Context, req git.ProposalRequest) (*environments.EnvironmentProposalDetails, error) {
	s.logger.Info("proposing pull request", zap.String("base", req.Base), zap.String("head", req.Head), zap.String("title", req.Title), zap.Bool("draft", req.Draft))
	sourceRefName := fmt.Sprintf("refs/heads/%s", req.Head)
	targetRefName := fmt.Sprintf("refs/heads/%s", req.Base)
	pr, err := s.client.CreatePullRequest(ctx, azuregit.CreatePullRequestArgs{
		GitPullRequestToCreate: &azuregit.GitPullRequest{
			Title:         &req.Title,
			Description:   &req.Body,
			IsDraft:       &req.Draft,
			SourceRefName: &sourceRefName,
			TargetRefName: &targetRefName,
		},
		RepositoryId: &s.repoName,
		Project:      &s.repoProject,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create pull request: %w", err)
	}
	s.logger.Info("pull request created", zap.String("pr", *pr.Url), zap.String("state", string(*pr.Status)))
	return &environments.EnvironmentProposalDetails{
		Url:   *pr.Url,
		State: environments.ProposalState_PROPOSAL_STATE_OPEN,
	}, nil
}

func (s *SCM) ListProposals(ctx context.Context, env serverenvs.Environment) (map[string]*environments.EnvironmentProposalDetails, error) {
	var (
		details = map[string]*environments.EnvironmentProposalDetails{}
		prs     = &prs{ctx: ctx, client: s.client, repoProject: s.repoProject, repoName: s.repoName, base: env.Configuration().Ref}
	)

	for pr := range prs.All() {
		branch := strings.TrimPrefix(*pr.SourceRefName, "refs/heads/")
		state := environments.ProposalState_PROPOSAL_STATE_OPEN
		if pr.Status == &azuregit.PullRequestStatusValues.Abandoned {
			state = environments.ProposalState_PROPOSAL_STATE_CLOSED
		} else if pr.Status == &azuregit.PullRequestStatusValues.Completed {
			state = environments.ProposalState_PROPOSAL_STATE_MERGED
		}

		webURL := fmt.Sprintf("%s/%s/%s/_git/%s/pullrequest/%d",
			s.baseURL,
			s.repoOwner,
			s.repoProject,
			s.repoName,
			*pr.PullRequestId,
		)
		details[branch] = &environments.EnvironmentProposalDetails{
			Url:   webURL,
			State: state,
		}
	}

	return details, prs.Err()
}

type prs struct {
	ctx         context.Context
	client      Client
	repoProject string
	repoName    string
	base        string

	err error
}

func (p *prs) Err() error {
	return p.err
}

func (p *prs) All() iter.Seq[*azuregit.GitPullRequest] {
	return iter.Seq[*azuregit.GitPullRequest](func(yield func(*azuregit.GitPullRequest) bool) {
		targetRefName := fmt.Sprintf("refs/heads/%s", p.base)
		includeLinks := true
		top := 100
		skip := 0
		criteria := &azuregit.GitPullRequestSearchCriteria{
			TargetRefName: &targetRefName,
			IncludeLinks:  &includeLinks,
		}

		for {
			prs, err := p.client.GetPullRequests(p.ctx, azuregit.GetPullRequestsArgs{
				RepositoryId:   &p.repoName,
				Project:        &p.repoProject,
				SearchCriteria: criteria,
				Top:            &top,
				Skip:           &skip,
			})
			if err != nil {
				p.err = err
				return
			}
			if prs == nil || len(*prs) == 0 {
				return
			}
			for _, pr := range *prs {
				if !yield(&pr) {
					return
				}
			}
			skip += top
		}
	})
}
