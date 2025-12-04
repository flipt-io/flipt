package credentials

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizeSSHRemoteURL(t *testing.T) {
	tests := []struct {
		name      string
		remoteURL string
		sshUser   string
		want      string
		wantErr   string
	}{
		// Basic HTTPS to SSH conversion
		{
			name:      "https URL with git user",
			remoteURL: "https://github.com/org-name/some-test-repo.git",
			sshUser:   "git",
			want:      "git@github.com:org-name/some-test-repo.git",
		},
		{
			name:      "https URL without .git suffix",
			remoteURL: "https://github.com/org-name/some-test-repo",
			sshUser:   "git",
			want:      "git@github.com:org-name/some-test-repo",
		},
		{
			name:      "https URL with custom user",
			remoteURL: "https://github.com/org-name/some-test-repo.git",
			sshUser:   "myuser",
			want:      "myuser@github.com:org-name/some-test-repo.git",
		},

		// Plain host/path format
		{
			name:      "host/path format with git user",
			remoteURL: "github.com/org-name/some-test-repo.git",
			sshUser:   "git",
			want:      "git@github.com:org-name/some-test-repo.git",
		},
		{
			name:      "host/path format with empty user defaults to git",
			remoteURL: "github.com/org-name/some-test-repo.git",
			sshUser:   "",
			want:      "git@github.com:org-name/some-test-repo.git",
		},

		// Already valid SSH URLs
		{
			name:      "already valid SSH URL with matching user",
			remoteURL: "git@github.com:org-name/some-test-repo.git",
			sshUser:   "git",
			want:      "git@github.com:org-name/some-test-repo.git",
		},
		{
			name:      "already valid SSH URL with custom matching user",
			remoteURL: "deploy@gitlab.com:org/repo.git",
			sshUser:   "deploy",
			want:      "deploy@gitlab.com:org/repo.git",
		},

		// SSH protocol URLs
		{
			name:      "ssh:// protocol URL",
			remoteURL: "ssh://git@github.com/org-name/some-test-repo.git",
			sshUser:   "git",
			want:      "git@github.com:org-name/some-test-repo.git",
		},
		{
			name:      "ssh:// protocol URL without user",
			remoteURL: "ssh://github.com/org-name/some-test-repo.git",
			sshUser:   "git",
			want:      "git@github.com:org-name/some-test-repo.git",
		},

		// HTTP (non-HTTPS) URL
		{
			name:      "http URL converts to SSH",
			remoteURL: "http://github.com/org-name/some-test-repo.git",
			sshUser:   "git",
			want:      "git@github.com:org-name/some-test-repo.git",
		},

		// Whitespace handling
		{
			name:      "URL with leading/trailing whitespace",
			remoteURL: "  https://github.com/org/repo.git  ",
			sshUser:   "git",
			want:      "git@github.com:org/repo.git",
		},

		// Different Git providers
		{
			name:      "GitLab URL",
			remoteURL: "https://gitlab.com/group/project.git",
			sshUser:   "git",
			want:      "git@gitlab.com:group/project.git",
		},
		{
			name:      "Bitbucket URL",
			remoteURL: "https://bitbucket.org/team/repo.git",
			sshUser:   "git",
			want:      "git@bitbucket.org:team/repo.git",
		},
		{
			name:      "Self-hosted GitLab",
			remoteURL: "https://git.example.com/team/project.git",
			sshUser:   "git",
			want:      "git@git.example.com:team/project.git",
		},

		// URLs with explicit ports
		{
			name:      "https URL with port 443",
			remoteURL: "https://github.com:443/org/repo.git",
			sshUser:   "git",
			want:      "git@github.com:org/repo.git",
		},
		{
			name:      "ssh:// URL with port 22",
			remoteURL: "ssh://git@github.com:22/org/repo.git",
			sshUser:   "git",
			want:      "git@github.com:org/repo.git",
		},
		{
			name:      "ssh:// URL without user but with port",
			remoteURL: "ssh://github.com:22/org/repo.git",
			sshUser:   "git",
			want:      "git@github.com:org/repo.git",
		},
		{
			name:      "host:port/path format",
			remoteURL: "github.com:443/org/repo.git",
			sshUser:   "git",
			want:      "git@github.com:org/repo.git",
		},

		// Nested paths (GitLab subgroups, etc.)
		{
			name:      "GitLab nested subgroups",
			remoteURL: "https://gitlab.com/group/subgroup/project.git",
			sshUser:   "git",
			want:      "git@gitlab.com:group/subgroup/project.git",
		},
		{
			name:      "deeply nested path",
			remoteURL: "https://gitlab.com/org/team/product/service.git",
			sshUser:   "git",
			want:      "git@gitlab.com:org/team/product/service.git",
		},
		{
			name:      "ssh:// with nested path",
			remoteURL: "ssh://git@gitlab.com/group/subgroup/project.git",
			sshUser:   "git",
			want:      "git@gitlab.com:group/subgroup/project.git",
		},

		// Error cases
		{
			name:      "conflicting users in SSH URL",
			remoteURL: "git@github.com:org-name/some-test-repo.git",
			sshUser:   "other",
			wantErr:   `conflicting SSH users: URL contains "git" but credentials specify "other"`,
		},
		{
			name:      "conflicting users in ssh:// URL",
			remoteURL: "ssh://git@github.com/org/repo.git",
			sshUser:   "deploy",
			wantErr:   `conflicting SSH users: URL contains "git" but credentials specify "deploy"`,
		},
		{
			name:      "invalid URL without path",
			remoteURL: "github.com",
			sshUser:   "git",
			wantErr:   `invalid remote URL format: "github.com" (expected host/path)`,
		},
		{
			name:      "invalid ssh:// URL without path",
			remoteURL: "ssh://github.com",
			sshUser:   "git",
			wantErr:   `invalid ssh:// URL format`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NormalizeSSHRemoteURL(tt.remoteURL, tt.sshUser)

			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
