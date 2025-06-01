package fs

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/server/environments"
	api "go.flipt.io/flipt/rpc/v2/environments"
)

func Test_DefaultFliptConfig(t *testing.T) {
	conf := DefaultFliptConfig()
	t.Run("CommitMessageTemplate", func(t *testing.T) {
		for _, test := range []struct {
			name     string
			changes  []Change
			expected string
		}{
			{
				name: "update a namespace",
				changes: []Change{
					{Verb: VerbUpdate, Resource: Resource{
						Type: environments.NewResourceType("flipt.config", "Namespace"),
						Key:  "default",
					}},
				},
				expected: "update namespace default",
			},
			{
				name: "create a flag",
				changes: []Change{
					{Verb: VerbCreate, Resource: Resource{
						Type:      environments.NewResourceType("flipt.core", "Flag"),
						Namespace: "default",
						Key:       "some-flag",
					}},
				},
				expected: "create flag default/some-flag",
			},
			{
				name: "multiple",
				changes: []Change{
					{Verb: VerbDelete, Resource: Resource{
						Type:      environments.NewResourceType("flipt.core", "Segment"),
						Namespace: "default",
						Key:       "some-segment",
					}},
					{Verb: VerbCreate, Resource: Resource{
						Type:      environments.NewResourceType("flipt.core", "Segment"),
						Namespace: "default",
						Key:       "some-other-segment",
					}},
				},
				expected: `updated multiple resources

delete segment default/some-segment
create segment default/some-other-segment`,
			},
		} {
			t.Run(test.name, func(t *testing.T) {
				buf := &bytes.Buffer{}
				err := conf.Templates.CommitMessageTemplate.Execute(buf, struct {
					Changes []Change
				}{
					Changes: test.changes,
				})
				require.NoError(t, err)

				assert.Equal(t, test.expected, buf.String())
			})
		}
	})

	t.Run("ProposalTitleTemplate", func(t *testing.T) {
		dir := "qwerty"
		tests := []struct {
			name          string
			input         templateContext
			expectedTitle string
			expectedBody  string
		}{
			{
				"no directory",
				templateContext{
					Base:   &api.EnvironmentConfiguration{Ref: "main"},
					Branch: &api.EnvironmentConfiguration{Ref: "some-branch"},
				},
				"Flipt: Update features on main",
				`This pull request updates Flipt resources on branch main.

🟢 **Source:**
- Branch: some-branch

🎯 **Target:**
- Branch: main

👀 Please review the changes and merge if everything looks good.`,
			},
			{
				"defined directory",
				templateContext{
					Base:   &api.EnvironmentConfiguration{Ref: "main", Directory: &dir},
					Branch: &api.EnvironmentConfiguration{Ref: "some-branch", Directory: &dir},
				},
				"Flipt: Update features in qwerty on main",
				`This pull request updates Flipt resources in qwerty on branch main.

🟢 **Source:**
- Directory: qwerty
- Branch: some-branch

🎯 **Target:**
- Directory: qwerty
- Branch: main

👀 Please review the changes and merge if everything looks good.`,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				buf := &bytes.Buffer{}
				err := conf.Templates.ProposalTitleTemplate.Execute(buf, tt.input)
				require.NoError(t, err)
				assert.Equal(t, tt.expectedTitle, buf.String())
				buf.Reset()
				err = conf.Templates.ProposalBodyTemplate.Execute(buf, tt.input)
				require.NoError(t, err)
				assert.Equal(t, tt.expectedBody, buf.String())
			})
		}
	})
}

type templateContext struct {
	Base   *api.EnvironmentConfiguration
	Branch *api.EnvironmentConfiguration
}
