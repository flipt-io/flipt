package fs

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/server/environments"
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
}
