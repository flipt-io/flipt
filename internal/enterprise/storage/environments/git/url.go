package git

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	giturl "github.com/kubescape/go-git-url"
)

type URL struct {
	Owner string
	Repo  string
}

func ParseGitURL(rawURL string) (*URL, error) {
	gitURL, err := giturl.NewGitURL(rawURL)
	if err != nil {
		// fall back to ssh parsing
		owner, repo, err := parseSSHRepo(rawURL)
		if err != nil {
			return nil, fmt.Errorf("failed to parse git url: %w", err)
		}
		return &URL{Owner: owner, Repo: repo}, nil
	}

	return &URL{Owner: gitURL.GetOwnerName(), Repo: gitURL.GetRepoName()}, nil
}

func parseSSHRepo(rawURL string) (owner, repo string, err error) {
	// Handle SSH-style URL: git@gitea.example.com:owner/repo.git
	if strings.Contains(rawURL, ":") && strings.Contains(rawURL, "@") && !strings.HasPrefix(rawURL, "http") {
		// Example: git@gitea.example.com:owner/repo.git
		parts := strings.SplitN(rawURL, ":", 2)
		if len(parts) != 2 {
			return "", "", errors.New("invalid SSH URL format")
		}
		path := strings.TrimSuffix(parts[1], ".git")
		subParts := strings.SplitN(path, "/", 2)
		if len(subParts) != 2 {
			return "", "", errors.New("invalid SSH repo path")
		}
		return subParts[0], subParts[1], nil
	}

	// Try parsing as a regular URL
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", "", fmt.Errorf("invalid URL: %w", err)
	}

	path := strings.TrimSuffix(u.Path, ".git")
	path = strings.TrimPrefix(path, "/")
	parts := strings.SplitN(path, "/", 2)
	if len(parts) != 2 {
		return "", "", errors.New("invalid path structure")
	}

	return parts[0], parts[1], nil
}
