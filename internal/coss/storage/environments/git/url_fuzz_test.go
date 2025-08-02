//go:build go1.18
// +build go1.18

package git

import (
	"testing"
)

func FuzzParseGitURL(f *testing.F) {
	// Add seed corpus with valid URLs
	seeds := []string{
		"https://github.com/owner/repo.git",
		"https://gitlab.com/owner/repo",
		"git@github.com:owner/repo.git",
		"git@gitlab.example.com:owner/repo.git",
		"ssh://git@github.com/owner/repo.git",
		"https://bitbucket.org/owner/repo",
		"git@bitbucket.org:owner/repo.git",
		"https://dev.azure.com/org/project/_git/repo",
		"git@ssh.dev.azure.com:v3/org/project/repo",
		"https://gitea.example.com/owner/repo.git",
		"git@gitea.example.com:owner/repo.git",
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	// Add some edge cases
	f.Add("")
	f.Add("not-a-url")
	f.Add("https://")
	f.Add("git@")
	f.Add(":")
	f.Add("@:")
	f.Add("git@host:")
	f.Add("https://host/")
	f.Add("file:///local/path")

	f.Fuzz(func(t *testing.T, rawURL string) {
		// Recover from panics
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("ParseGitURL panicked with input '%s': %v", rawURL, r)
			}
		}()

		// Test that ParseGitURL doesn't panic
		_, _ = ParseGitURL(rawURL)
	})
}

func FuzzParseSSHRepo(f *testing.F) {
	// Add seed corpus with SSH-style URLs
	seeds := []string{
		"git@github.com:owner/repo.git",
		"git@gitlab.com:owner/repo",
		"user@host.com:owner/repo.git",
		"git@ssh.dev.azure.com:v3/org/project/repo",
		"git@bitbucket.org:owner/repo.git",
		"deploy@gitea.example.com:owner/repo.git",
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	// Add edge cases
	f.Add("")
	f.Add("@")
	f.Add(":")
	f.Add("@:")
	f.Add("user@")
	f.Add("@host:")
	f.Add("user@host")
	f.Add("user@host:")
	f.Add("user@host:/")
	f.Add("user@host:repo")
	f.Add("user@host:owner/")
	f.Add("user@host:/owner/repo")
	f.Add("https://github.com/owner/repo.git") // Should fall back to regular URL parsing

	f.Fuzz(func(t *testing.T, rawURL string) {
		// Recover from panics
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("parseSSHRepo panicked with input '%s': %v", rawURL, r)
			}
		}()

		// Test that parseSSHRepo doesn't panic
		_, _ = parseSSHRepo(rawURL)
	})
}