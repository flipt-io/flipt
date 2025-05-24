// Flipt Enterprise-Only Feature
// This file contains functionality that is licensed under the Flipt Fair Core License (FCL).
// You may NOT use, modify, or distribute this file or its contents without a valid Enterprise license.
// For details: https://github.com/flipt-io/flipt/blob/v2/LICENSE

package github

import (
	"context"
	"fmt"
	"net/http"
	"sort"

	"github.com/google/go-github/v64/github"
	"go.flipt.io/flipt/internal/enterprise/storage/environments/git"
	"go.flipt.io/flipt/rpc/v2/environments"
)

var _ git.SCM = (*SCM)(nil)

type SCM struct {
	repoOwner string
	repoName  string
	client    *github.PullRequestsService
	repos     *github.RepositoriesService
}

func NewSCM(ctx context.Context, repoOwner, repoName string, httpClient *http.Client) *SCM {
	ghClient := github.NewClient(httpClient)
	return &SCM{
		repoOwner: repoOwner,
		repoName:  repoName,
		client:    ghClient.PullRequests,
		repos:     ghClient.Repositories,
	}
}

func (s *SCM) Propose(ctx context.Context, req git.ProposalRequest) (*environments.ProposeEnvironmentResponse, error) {
	pr, _, err := s.client.Create(ctx, s.repoOwner, s.repoName, &github.NewPullRequest{
		Base:  github.String(req.Base),
		Head:  github.String(req.Head),
		Title: github.String(req.Title),
		Body:  github.String(req.Body),
		Draft: github.Bool(req.Draft),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create pull request: %w", err)
	}

	return &environments.ProposeEnvironmentResponse{
		Scm: environments.SCM_GITHUB_SCM,
		Url: pr.GetHTMLURL(),
	}, nil
}

func (s *SCM) ListChanges(ctx context.Context, req git.ListChangesRequest) (*environments.ListBranchedEnvironmentChangesResponse, error) {
	comparison, _, err := s.repos.CompareCommits(ctx, s.repoOwner, s.repoName, req.Base, req.Head, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to compare branches: %w", err)
	}

	var (
		changes []*environments.Change
		limit   = req.Limit
	)

	for _, commit := range comparison.Commits {
		if commit == nil || commit.GetCommit() == nil {
			continue
		}

		if limit > 0 && int32(len(changes)) >= limit {
			break
		}

		change := &environments.Change{
			Revision: commit.GetSHA(),
			Message:  commit.GetCommit().GetMessage(),
			ScmUrl:   github.String(commit.GetHTMLURL()),
		}

		if commit.GetCommit().GetAuthor() != nil {
			change.AuthorName = github.String(commit.GetCommit().GetAuthor().GetName())
			change.AuthorEmail = github.String(commit.GetCommit().GetAuthor().GetEmail())
			change.Timestamp = commit.GetCommit().GetAuthor().GetDate().Format("2006-01-02T15:04:05Z07:00")
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
