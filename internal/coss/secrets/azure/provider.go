// Flipt Commercial Open Source Feature
// This file contains functionality that is licensed under the Flipt Fair Core License (FCL).
// You may NOT use, modify, or distribute this file or its contents without a valid paid license.
// For details: https://github.com/flipt-io/flipt/blob/v2/LICENSE

package azure

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"
	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/internal/secrets"
	"go.uber.org/zap"
)

func init() {
	// Register Azure Key Vault provider factory
	secrets.RegisterProviderFactory("azure", func(cfg *config.Config, logger *zap.Logger) (secrets.Provider, error) {
		if cfg.Secrets.Providers.Azure == nil {
			return nil, fmt.Errorf("azure provider configuration not found")
		}

		return NewProvider(cfg.Secrets.Providers.Azure.VaultURL, logger)
	})
}

// Provider implements secrets.Provider for Azure Key Vault.
type Provider struct {
	client *azsecrets.Client
	logger *zap.Logger
}

// NewProvider creates a new Azure Key Vault provider.
// Authentication uses DefaultAzureCredential which supports managed identity,
// environment variables (AZURE_CLIENT_ID, AZURE_TENANT_ID, AZURE_CLIENT_SECRET),
// Azure CLI, and other credential sources.
func NewProvider(vaultURL string, logger *zap.Logger) (*Provider, error) {
	var (
		cred       azcore.TokenCredential
		clientOpts *azsecrets.ClientOptions
	)

	if os.Getenv("AZURE_KEYVAULT_EMULATOR") != "" {
		logger.Info("connecting to Azure Key Vault emulator",
			zap.String("vault_url", vaultURL))

		// Use fake credential for emulator testing
		cred = &fakeCredential{}
		clientOpts = &azsecrets.ClientOptions{
			ClientOptions: azcore.ClientOptions{
				Transport: &http.Client{
					Transport: &http.Transport{
						TLSClientConfig: &tls.Config{
							InsecureSkipVerify: true, //nolint:gosec // emulator only
						},
					},
				},
			},
			DisableChallengeResourceVerification: true,
		}
	} else {
		var err error
		cred, err = azidentity.NewDefaultAzureCredential(nil)
		if err != nil {
			return nil, fmt.Errorf("creating azure credential: %w", err)
		}
	}

	client, err := azsecrets.NewClient(vaultURL, cred, clientOpts)
	if err != nil {
		return nil, fmt.Errorf("creating azure key vault client: %w", err)
	}

	return &Provider{
		client: client,
		logger: logger,
	}, nil
}

// fakeCredential implements azcore.TokenCredential for emulator testing.
type fakeCredential struct{}

func (f *fakeCredential) GetToken(_ context.Context, _ policy.TokenRequestOptions) (azcore.AccessToken, error) {
	return azcore.AccessToken{Token: "fake-token"}, nil
}

// GetSecret retrieves a secret from Azure Key Vault.
func (p *Provider) GetSecret(ctx context.Context, path string) (*secrets.Secret, error) {
	p.logger.Debug("reading secret from azure key vault",
		zap.String("path", path))

	// Empty version string retrieves the latest version
	resp, err := p.client.GetSecret(ctx, path, "", nil)
	if err != nil {
		return nil, fmt.Errorf("getting secret: %w", err)
	}

	data := make(map[string][]byte)
	if resp.Value != nil {
		data["value"] = []byte(*resp.Value)
	}

	metadata := map[string]string{}
	if resp.ID != nil {
		metadata["id"] = string(*resp.ID)
	}

	return &secrets.Secret{
		Path:     path,
		Data:     data,
		Metadata: metadata,
	}, nil
}

// ListSecrets returns all secret names in Azure Key Vault matching the given prefix.
func (p *Provider) ListSecrets(ctx context.Context, pathPrefix string) ([]string, error) {
	p.logger.Debug("listing secrets from azure key vault",
		zap.String("prefix", pathPrefix))

	pager := p.client.NewListSecretPropertiesPager(nil)

	var paths []string
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("listing secrets: %w", err)
		}

		for _, secret := range page.Value {
			if secret.ID == nil {
				continue
			}

			// Use the Name() method on the ID type to extract the secret name
			name := secret.ID.Name()
			if pathPrefix == "" || strings.HasPrefix(name, pathPrefix) {
				paths = append(paths, name)
			}
		}
	}

	return paths, nil
}
