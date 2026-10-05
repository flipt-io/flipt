package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSCMConfig_ValidateWebhook(t *testing.T) {
	tests := []struct {
		name    string
		webhook *IncomingWebhookConfig
		wantErr string
	}{
		{
			name: "no webhook",
		},
		{
			name:    "secret",
			webhook: &IncomingWebhookConfig{Secret: "s3cr3t"},
		},
		{
			name:    "secret reference",
			webhook: &IncomingWebhookConfig{Secret: "${secret:file:gitlab-webhook}"},
		},
		{
			name:    "missing secret",
			webhook: &IncomingWebhookConfig{},
			wantErr: "environments: scm webhook: secret is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scm := SCMConfig{Type: GitLabSCMType, Webhook: tt.webhook}

			err := scm.validate()
			if tt.wantErr == "" {
				require.NoError(t, err)
				return
			}

			require.Error(t, err)
			assert.Equal(t, tt.wantErr, err.Error())
		})
	}
}
