package secrets

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseReference(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		want    Reference
		wantOK  bool
		wantErr string
	}{
		{
			name:   "provider and key",
			value:  "${secret:file:github-token}",
			want:   Reference{Provider: "file", Path: "github-token", Key: "github-token"},
			wantOK: true,
		},
		{
			name:   "nested path",
			value:  "${secret:vault:flipt/git/token}",
			want:   Reference{Provider: "vault", Path: "flipt/git/token", Key: "flipt/git/token"},
			wantOK: true,
		},
		{
			name:   "key without provider",
			value:  "${secret:github-token}",
			want:   Reference{Path: "github-token", Key: "github-token"},
			wantOK: true,
		},
		{
			name:    "too many segments",
			value:   "${secret:file:a:b}",
			wantOK:  true,
			wantErr: `invalid secret reference format "file:a:b"`,
		},
		{
			name:  "plain value",
			value: "ghp_abc123",
		},
		{
			name:  "empty value",
			value: "",
		},
		{
			name:  "environment reference",
			value: "${env:GITHUB_TOKEN}",
		},
		{
			name:  "reference embedded in a larger value",
			value: "prefix-${secret:file:github-token}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ref, ok, err := ParseReference(tt.value)
			assert.Equal(t, tt.wantOK, ok)

			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, ref)
		})
	}
}
