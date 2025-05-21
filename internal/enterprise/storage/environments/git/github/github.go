// Flipt Enterprise-Only Feature
// This file contains functionality that is licensed under the Flipt Fair Core License (FCL).
// You may NOT use, modify, or distribute this file or its contents without a valid Enterprise license.
// For details: https://github.com/flipt-io/flipt/blob/v2/LICENSE

package github

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/go-github/v64/github"
	"go.flipt.io/flipt/internal/enterprise/storage/environments/git"
	"go.flipt.io/flipt/rpc/v2/environments"
)

var _ git.SCM = (*SCM)(nil)

type SCM struct {
	repoOwner string
	repoName  string
	client    *github.PullRequestsService
}

func NewSCM(ctx context.Context, repoOwner, repoName string, httpClient *http.Client) *SCM {
	return &SCM{
		repoOwner: repoOwner,
		repoName:  repoName,
		client:    github.NewClient(httpClient).PullRequests,
	}
}

func (s *SCM) Propose(ctx context.Context, req git.ProposalRequest) (*environments.ProposeEnvironmentResponse, error) {
	pr, _, err := s.client.Create(ctx, s.repoOwner, s.repoName, &github.NewPullRequest{
		Base:  github.String(req.Base),
		Head:  github.String(req.Head),
		Title: github.String(req.Title),
		Body:  github.String(req.Body),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create pull request: %w", err)
	}

	return &environments.ProposeEnvironmentResponse{
		Scm: environments.SCM_GITHUB_SCM,
		Url: pr.GetHTMLURL(),
	}, nil
}
