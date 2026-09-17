package names

import (
	"testing"

	"github.com/go-git/go-git/v6/plumbing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRefSlug(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		// names that are already valid must not change
		{name: "simple", in: "production", want: "production"},
		{name: "hyphens and dots", in: "my--env.v1", want: "my--env.v1"},
		{name: "leading hyphen", in: "-env", want: "-env"},
		{name: "unicode letters", in: "größe", want: "größe"},
		{name: "at sign", in: "team@flipt", want: "team@flipt"},

		// names with characters git does not permit
		{name: "space", in: "Dev Playground", want: "Dev-Playground"},
		{name: "tab and newline", in: "a\tb\nc", want: "a-b-c"},
		{name: "slash", in: "my/feature", want: "my-feature"},
		{name: "control character", in: "a\x01b", want: "a-b"},
		{name: "del character", in: "a\x7fb", want: "a-b"},
		{name: "special characters", in: "a~b^c:d?e*f[g\\h", want: "a-b-c-d-e-f-g-h"},
		{name: "double dot", in: "release..1", want: "release.-1"},
		{name: "triple dot", in: "a...b", want: "a.-.b"},
		{name: "leading dot", in: ".hidden", want: "-hidden"},
		{name: "trailing dot", in: "env.", want: "env-"},
		{name: "at brace", in: "a@{b", want: "a@-b"},
		{name: "only at sign", in: "@", want: "-"},
		{name: "lock suffix", in: "env.lock", want: "env-lock"},
		{name: "only lock suffix", in: ".lock", want: "-lock"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RefSlug(tt.in)
			assert.Equal(t, tt.want, got)

			// the slug must always form a valid git branch name in the
			// position Flipt uses it: flipt/<environment>/<branch>
			ref := plumbing.NewBranchReferenceName("flipt/" + got + "/branch")
			require.NoError(t, ref.Validate(), "slug %q must be a valid reference component", got)

			// the slug of a slug is the slug itself
			assert.Equal(t, got, RefSlug(got))
		})
	}
}
