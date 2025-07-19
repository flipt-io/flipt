// Flipt Commercial Open Source Feature
// This file contains functionality that is licensed under the Flipt Fair Core License (FCL).
// You may NOT use, modify, or distribute this file or its contents without a valid paid license.
// For details: https://github.com/flipt-io/flipt/blob/v2/LICENSE
package azure

import (
	"context"
	"fmt"
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
}

type ClientOption func(*azureOptions)

func WithApiURL(apiURL *url.URL) ClientOption {
	return func(c *azureOptions) {
		c.apiURL = apiURL
	}
}

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

func NewSCM(logger *zap.Logger, project, repoName string, opts ...ClientOption) (*SCM, error) {
	options := &azureOptions{}
	for _, opt := range opts {
		opt(options)
	}

	connection := azuredevops.NewPatConnection(options.apiURL.String(), options.personalAccessToken)
	client, err := azuregit.NewClient(context.TODO(), connection)
	if err != nil {
		return nil, fmt.Errorf("failed to create azure git client: %w", err)
	}

	return &SCM{
		logger:   logger.With(zap.String("repository", fmt.Sprintf("%s/%s", project, repoName)), zap.String("scm", "azure git")),
		client:   client,
		project:  project,
		repoName: repoName,
		baseURL:  options.apiURL.String(),
	}, nil
}

type SCM struct {
	logger   *zap.Logger
	client   Client
	project  string
	repoName string
	baseURL  string
}

func (s *SCM) ListChanges(ctx context.Context, req git.ListChangesRequest) (*environments.ListBranchedEnvironmentChangesResponse, error) {
	changes := []*environments.Change{}
	includeLinks := true
	commits, err := s.client.GetCommits(ctx, azuregit.GetCommitsArgs{
		RepositoryId: &s.repoName,
		Project:      &s.project,
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
		changes = append(changes, &environments.Change{
			Revision:  *commit.CommitId,
			Message:   *commit.Comment,
			ScmUrl:    commit.RemoteUrl,
			Timestamp: commit.Author.Date.Time.Format(time.RFC3339),
		})
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
		Project:      &s.project,
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
	details := map[string]*environments.EnvironmentProposalDetails{}
	targetRefName := fmt.Sprintf("refs/heads/%s", env.Configuration().Ref)
	includeLinks := true
	prs, err := s.client.GetPullRequests(ctx, azuregit.GetPullRequestsArgs{
		RepositoryId: &s.repoName,
		Project:      &s.project,
		SearchCriteria: &azuregit.GitPullRequestSearchCriteria{
			TargetRefName: &targetRefName,
			IncludeLinks:  &includeLinks,
		},
	})
	if err != nil {
		return nil, err
	}
	for _, pr := range *prs {
		branch := strings.TrimPrefix(*pr.SourceRefName, "refs/heads/")
		state := environments.ProposalState_PROPOSAL_STATE_OPEN
		if pr.Status == &azuregit.PullRequestStatusValues.Abandoned {
			state = environments.ProposalState_PROPOSAL_STATE_CLOSED
		} else if pr.Status == &azuregit.PullRequestStatusValues.Completed {
			state = environments.ProposalState_PROPOSAL_STATE_MERGED
		}

		webURL := fmt.Sprintf("%s/%s/_git/%s/pullrequest/%d",
			s.baseURL,
			s.project,
			s.repoName,
			*pr.PullRequestId,
		)
		details[branch] = &environments.EnvironmentProposalDetails{
			Url:   webURL,
			State: state,
		}
	}
	return details, nil
}
