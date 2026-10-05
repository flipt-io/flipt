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
			name: "secret_ref",
			webhook: &IncomingWebhookConfig{
				SecretRef: &SecretReference{Provider: "file", Path: "webhooks", Key: "gitlab"},
			},
		},
		{
			name:    "neither secret nor secret_ref",
			webhook: &IncomingWebhookConfig{},
			wantErr: "environments: scm webhook: exactly one of secret or secret_ref is required",
		},
		{
			name: "both secret and secret_ref",
			webhook: &IncomingWebhookConfig{
				Secret:    "s3cr3t",
				SecretRef: &SecretReference{Provider: "file", Path: "webhooks", Key: "gitlab"},
			},
			wantErr: "environments: scm webhook: exactly one of secret or secret_ref is required",
		},
		{
			name: "secret_ref missing provider",
			webhook: &IncomingWebhookConfig{
				SecretRef: &SecretReference{Path: "webhooks", Key: "gitlab"},
			},
			wantErr: "environments: scm webhook: secret_ref: secret_reference: provider non-empty value is required",
		},
		{
			name: "secret_ref missing path",
			webhook: &IncomingWebhookConfig{
				SecretRef: &SecretReference{Provider: "file", Key: "gitlab"},
			},
			wantErr: "environments: scm webhook: secret_ref: secret_reference: path non-empty value is required",
		},
		{
			name: "secret_ref missing key",
			webhook: &IncomingWebhookConfig{
				SecretRef: &SecretReference{Provider: "file", Path: "webhooks"},
			},
			wantErr: "environments: scm webhook: secret_ref: secret_reference: key non-empty value is required",
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
