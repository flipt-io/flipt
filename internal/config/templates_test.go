package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTemplatesConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  TemplatesConfig
		wantErr bool
		errMsg  string
	}{
		{
			name:    "empty config is valid",
			config:  TemplatesConfig{},
			wantErr: false,
		},
		{
			name: "valid commit_message template",
			config: TemplatesConfig{
				CommitMessage: "{{ (index .Changes 0) }} [skip ci]",
			},
			wantErr: false,
		},
		{
			name: "valid proposal_title template",
			config: TemplatesConfig{
				ProposalTitle: "Flipt: Update features on {{.Base.Ref}}",
			},
			wantErr: false,
		},
		{
			name: "valid proposal_body template",
			config: TemplatesConfig{
				ProposalBody: "Source: {{.Branch.Ref}}\nTarget: {{.Base.Ref}}",
			},
			wantErr: false,
		},
		{
			name: "all templates valid",
			config: TemplatesConfig{
				CommitMessage: "{{ (index .Changes 0) }}",
				ProposalTitle: "PR: {{.Base.Ref}}",
				ProposalBody:  "Body text",
			},
			wantErr: false,
		},
		{
			name: "invalid commit_message template - unclosed brace",
			config: TemplatesConfig{
				CommitMessage: "{{ .Changes",
			},
			wantErr: true,
			errMsg:  "invalid commit_message template",
		},
		{
			name: "invalid proposal_title template - bad syntax",
			config: TemplatesConfig{
				ProposalTitle: "{{if}}",
			},
			wantErr: true,
			errMsg:  "invalid proposal_title template",
		},
		{
			name: "invalid proposal_body template - unclosed action",
			config: TemplatesConfig{
				ProposalBody: "{{ range .Items }}",
			},
			wantErr: true,
			errMsg:  "invalid proposal_body template",
		},
		{
			name: "first invalid template reports error",
			config: TemplatesConfig{
				CommitMessage: "{{ .Valid }}",
				ProposalTitle: "{{if}}", // invalid
				ProposalBody:  "{{if}}", // also invalid
			},
			wantErr: true,
			errMsg:  "invalid proposal_title template",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
