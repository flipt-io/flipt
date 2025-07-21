package git

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	giturl "github.com/kubescape/go-git-url"
)

type URL interface {
	GetOwnerName() string
	GetRepoName() string
	GetHostName() string
}

type gitURL struct {
	giturl.IGitURL
}

var _ URL = &sshURL{}

type sshURL struct {
	Owner string
	Repo  string
	Host  string
}

func (u *sshURL) GetOwnerName() string {
	return u.Owner
}

func (u *sshURL) GetRepoName() string {
	return u.Repo
}

func (u *sshURL) GetHostName() string {
	return u.Host
}

func ParseGitURL(rawURL string) (URL, error) {
	gitURL, err := giturl.NewGitURL(rawURL)
	if err != nil {
		// fall back to ssh parsing
		sshURL, err := parseSSHRepo(rawURL)
		if err != nil {
			return nil, fmt.Errorf("failed to parse git url: %w", err)
		}
		return sshURL, nil
	}

	return gitURL, nil
}

func parseSSHRepo(rawURL string) (*sshURL, error) {
	// Handle SSH-style URL: git@gitea.example.com:owner/repo.git
	if strings.Contains(rawURL, ":") && strings.Contains(rawURL, "@") && !strings.HasPrefix(rawURL, "http") {
		// Example: git@gitea.example.com:owner/repo.git
		parts := strings.SplitN(rawURL, ":", 2)
		if len(parts) != 2 {
			return nil, errors.New("invalid SSH URL format")
		}
		path := strings.TrimSuffix(parts[1], ".git")
		subParts := strings.SplitN(path, "/", 2)
		if len(subParts) != 2 {
			return nil, errors.New("invalid SSH repo path")
		}
		return &sshURL{Owner: subParts[0], Repo: subParts[1]}, nil
	}

	// Try parsing as a regular URL
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	path := strings.TrimSuffix(u.Path, ".git")
	path = strings.TrimPrefix(path, "/")
	parts := strings.SplitN(path, "/", 2)
	if len(parts) != 2 {
		return nil, errors.New("invalid path structure")
	}

	return &sshURL{Owner: parts[0], Repo: parts[1], Host: u.Host}, nil
}
