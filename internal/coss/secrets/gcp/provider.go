// Flipt Commercial Open Source Feature
// This file contains functionality that is licensed under the Flipt Fair Core License (FCL).
// You may NOT use, modify, or distribute this file or its contents without a valid paid license.
// For details: https://github.com/flipt-io/flipt/blob/v2/LICENSE

package gcp

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/internal/secrets"
	"go.uber.org/zap"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func init() {
	// Register GCP Secret Manager provider factory
	secrets.RegisterProviderFactory("gcp", func(cfg *config.Config, logger *zap.Logger) (secrets.Provider, error) {
		if cfg.Secrets.Providers.GCP == nil {
			return nil, fmt.Errorf("gcp provider configuration not found")
		}

		return NewProvider(cfg.Secrets.Providers.GCP.Project, cfg.Secrets.Providers.GCP.Credentials, logger)
	})
}

// Provider implements secrets.Provider for Google Cloud Secret Manager.
type Provider struct {
	client  *secretmanager.Client
	project string
	logger  *zap.Logger
}

// NewProvider creates a new GCP Secret Manager provider.
// If the SECRET_MANAGER_EMULATOR_HOST environment variable is set, the client
// connects to the emulator using insecure credentials instead of real GCP.
func NewProvider(project, credentials string, logger *zap.Logger) (*Provider, error) {
	ctx := context.Background()

	var opts []option.ClientOption

	if emulatorHost := os.Getenv("SECRET_MANAGER_EMULATOR_HOST"); emulatorHost != "" {
		// Connect to emulator with insecure credentials
		logger.Info("connecting to GCP Secret Manager emulator",
			zap.String("host", emulatorHost))

		conn, err := grpc.NewClient(
			emulatorHost,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if err != nil {
			return nil, fmt.Errorf("creating grpc connection to emulator: %w", err)
		}

		opts = append(opts, option.WithGRPCConn(conn))
	} else if credentials != "" {
		opts = append(opts, option.WithCredentialsFile(credentials))
	}

	client, err := secretmanager.NewClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("creating gcp secret manager client: %w", err)
	}

	return &Provider{
		client:  client,
		project: project,
		logger:  logger,
	}, nil
}

// GetSecret retrieves a secret from GCP Secret Manager.
func (p *Provider) GetSecret(ctx context.Context, path string) (*secrets.Secret, error) {
	name := fmt.Sprintf("projects/%s/secrets/%s/versions/latest", p.project, path)

	p.logger.Debug("reading secret from gcp secret manager",
		zap.String("path", path),
		zap.String("project", p.project))

	result, err := p.client.AccessSecretVersion(ctx, &secretmanagerpb.AccessSecretVersionRequest{
		Name: name,
	})
	if err != nil {
		return nil, fmt.Errorf("accessing secret version: %w", err)
	}

	return &secrets.Secret{
		Path: path,
		Data: map[string][]byte{
			"value": result.Payload.Data,
		},
		Metadata: map[string]string{
			"name": result.Name,
		},
	}, nil
}

// ListSecrets returns all secret paths matching the prefix in GCP Secret Manager.
func (p *Provider) ListSecrets(ctx context.Context, pathPrefix string) ([]string, error) {
	parent := fmt.Sprintf("projects/%s", p.project)

	p.logger.Debug("listing secrets from gcp secret manager",
		zap.String("prefix", pathPrefix),
		zap.String("project", p.project))

	var filter string
	if pathPrefix != "" {
		filter = fmt.Sprintf("name:%s", pathPrefix)
	}

	it := p.client.ListSecrets(ctx, &secretmanagerpb.ListSecretsRequest{
		Parent: parent,
		Filter: filter,
	})

	secretPrefix := fmt.Sprintf("projects/%s/secrets/", p.project)

	var paths []string
	for {
		secret, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("listing secrets: %w", err)
		}

		// Strip the full resource name prefix to return just the secret name
		name := strings.TrimPrefix(secret.Name, secretPrefix)
		paths = append(paths, name)
	}

	return paths, nil
}

// Close closes the GCP Secret Manager client.
func (p *Provider) Close() error {
	return p.client.Close()
}
