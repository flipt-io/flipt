package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCredentialConfig_Validate_GitHubApp(t *testing.T) {
	t.Run("missing fields", func(t *testing.T) {
		config := CredentialConfig{
			Type: CredentialTypeGithubApp,
		}
		err := config.validate()
		require.Error(t, err)
		require.Contains(t, err.Error(), "credentials: github_app configuration non-empty value is required")
	})
	t.Run("missing both private key fields", func(t *testing.T) {
		config := CredentialConfig{
			Type: CredentialTypeGithubApp,
			GitHubApp: &GitHubAppConfig{
				ClientID:        "test-client-id",
				InstallationID:  12345,
				PrivateKeyPath:  "",
				PrivateKeyBytes: "",
			},
		}
		err := config.validate()
		require.Error(t, err)
		require.Contains(t, err.Error(), "github_app private_key_path or private_key_bytes")
	})

	t.Run("both private key fields provided", func(t *testing.T) {
		config := CredentialConfig{
			Type: CredentialTypeGithubApp,
			GitHubApp: &GitHubAppConfig{
				ClientID:        "test-client-id",
				InstallationID:  12345,
				PrivateKeyPath:  "/path/to/key.pem",
				PrivateKeyBytes: "-----BEGIN RSA PRIVATE KEY-----",
			},
		}
		err := config.validate()
		require.Error(t, err)
		require.Contains(t, err.Error(), "please provide exclusively one of private_key_path or private_key_bytes")
	})

	t.Run("only private key path provided", func(t *testing.T) {
		config := CredentialConfig{
			Type: CredentialTypeGithubApp,
			GitHubApp: &GitHubAppConfig{
				ClientID:        "test-client-id",
				InstallationID:  12345,
				PrivateKeyPath:  "/path/to/key.pem",
				PrivateKeyBytes: "",
			},
		}
		err := config.validate()
		require.NoError(t, err)
	})

	t.Run("only private key bytes provided", func(t *testing.T) {
		config := CredentialConfig{
			Type: CredentialTypeGithubApp,
			GitHubApp: &GitHubAppConfig{
				ClientID:        "test-client-id",
				InstallationID:  12345,
				PrivateKeyPath:  "",
				PrivateKeyBytes: "-----BEGIN RSA PRIVATE KEY-----",
			},
		}
		err := config.validate()
		require.NoError(t, err)
	})

	t.Run("missing client id", func(t *testing.T) {
		config := CredentialConfig{
			Type: CredentialTypeGithubApp,
			GitHubApp: &GitHubAppConfig{
				ClientID:        "",
				InstallationID:  12345,
				PrivateKeyBytes: "-----BEGIN RSA PRIVATE KEY-----",
			},
		}
		err := config.validate()
		require.Error(t, err)
		require.Contains(t, err.Error(), "github_app client_id")
	})

	t.Run("missing installation id", func(t *testing.T) {
		config := CredentialConfig{
			Type: CredentialTypeGithubApp,
			GitHubApp: &GitHubAppConfig{
				ClientID:        "test-client-id",
				InstallationID:  0,
				PrivateKeyBytes: "-----BEGIN RSA PRIVATE KEY-----",
			},
		}
		err := config.validate()
		require.Error(t, err)
		require.Contains(t, err.Error(), "github_app installation_id")
	})
}
