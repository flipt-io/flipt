package git

import (
	"testing"
	"text/template"

	"github.com/go-git/go-git/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/internal/server/environments"
	storagefs "go.flipt.io/flipt/internal/storage/fs"
)

func Test_Environment_messageForChanges(t *testing.T) {
	for _, test := range []struct {
		name            string
		changes         []storagefs.Change
		tmpl            *template.Template
		expectedMessage string
		expectedErr     error
	}{
		{
			name:        "no changes",
			expectedErr: git.ErrEmptyCommit,
		},
		{
			name: "single change",
			changes: []storagefs.Change{
				{
					Verb: storagefs.VerbCreate,
					Resource: storagefs.Resource{
						Type: environments.NewResourceType("flipt.config", "Namespace"),
						Key:  "default",
					},
				},
			},
			expectedMessage: `create namespace default`,
		},
		{
			name: "multiple changes",
			changes: []storagefs.Change{
				{
					Verb: storagefs.VerbCreate,
					Resource: storagefs.Resource{
						Type:      environments.NewResourceType("flipt.core", "Flag"),
						Namespace: "default",
						Key:       "someFeature",
					},
				},
				{
					Verb: storagefs.VerbDelete,
					Resource: storagefs.Resource{
						Type:      environments.NewResourceType("flipt.core", "Segment"),
						Namespace: "default",
						Key:       "someSegment",
					},
				},
			},
			expectedMessage: `updated multiple resources

create flag default/someFeature
delete segment default/someSegment`,
		},
		{
			name: "single change",
			changes: []storagefs.Change{
				{
					Verb: storagefs.VerbUpdate,
					Resource: storagefs.Resource{
						Type:      environments.NewResourceType("flipt.core", "Flag"),
						Namespace: "default",
						Key:       "some-flag",
					},
				},
			},
			tmpl: template.Must(template.New("commitMessage").Parse(`{{- if eq (len .Changes) 1 }}
        {{- printf "feat(flipt/%s): %s" .Environment.Name (index .Changes 0) }}
        {{- else -}}
        updated multiple resources
        {{ range $change := .Changes }}
        {{- printf "feat(flipt/%s): %s" .Environment.Name $change }}
        {{- end }}
        {{- end }}`)),
			expectedMessage: `feat(flipt/production): update flag default/some-flag`,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			env := &Environment{
				cfg: &config.EnvironmentConfig{
					Name: "production",
				},
			}

			template := test.tmpl
			if template == nil {
				template = storagefs.DefaultFliptConfig().Templates.CommitMessageTemplate
			}

			msg, err := env.messageForChanges(template, test.changes...)
			if test.expectedErr != nil {
				require.ErrorIs(t, err, test.expectedErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, test.expectedMessage, msg)
		})
	}
}
