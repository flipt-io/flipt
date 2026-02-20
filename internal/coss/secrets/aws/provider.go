// Flipt Commercial Open Source Feature
// This file contains functionality that is licensed under the Flipt Fair Core License (FCL).
// You may NOT use, modify, or distribute this file or its contents without a valid paid license.
// For details: https://github.com/flipt-io/flipt/blob/v2/LICENSE

package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"
	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/internal/secrets"
	"go.uber.org/zap"
)

func init() {
	// Register AWS Secrets Manager provider factory
	secrets.RegisterProviderFactory("aws", func(cfg *config.Config, logger *zap.Logger) (secrets.Provider, error) {
		if cfg.Secrets.Providers.AWS == nil {
			return nil, fmt.Errorf("aws provider configuration not found")
		}

		return NewProvider(cfg.Secrets.Providers.AWS.EndpointURL, logger)
	})
}

// Provider implements secrets.Provider for AWS Secrets Manager.
type Provider struct {
	client *secretsmanager.Client
	logger *zap.Logger
}

// NewProvider creates a new AWS Secrets Manager provider.
// Region is determined by the AWS SDK's default configuration (e.g., AWS_DEFAULT_REGION environment variable).
func NewProvider(endpointURL string, logger *zap.Logger) (*Provider, error) {
	ctx := context.Background()

	cfg, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("loading aws config: %w", err)
	}

	var opts []func(*secretsmanager.Options)
	if endpointURL != "" {
		opts = append(opts, func(o *secretsmanager.Options) {
			o.BaseEndpoint = aws.String(endpointURL)
		})
	}

	client := secretsmanager.NewFromConfig(cfg, opts...)

	return &Provider{
		client: client,
		logger: logger,
	}, nil
}

// GetSecret retrieves a secret from AWS Secrets Manager.
func (p *Provider) GetSecret(ctx context.Context, path string) (*secrets.Secret, error) {
	p.logger.Debug("reading secret from aws secrets manager",
		zap.String("path", path))

	result, err := p.client.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(path),
	})
	if err != nil {
		return nil, fmt.Errorf("getting secret value: %w", err)
	}

	data := make(map[string][]byte)
	if result.SecretString != nil {
		data["value"] = []byte(*result.SecretString)
	} else if result.SecretBinary != nil {
		data["value"] = result.SecretBinary
	}

	return &secrets.Secret{
		Path: path,
		Data: data,
		Metadata: map[string]string{
			"arn": *result.ARN,
		},
	}, nil
}

// ListSecrets returns all secret names matching the prefix in AWS Secrets Manager.
func (p *Provider) ListSecrets(ctx context.Context, pathPrefix string) ([]string, error) {
	p.logger.Debug("listing secrets from aws secrets manager",
		zap.String("prefix", pathPrefix))

	input := &secretsmanager.ListSecretsInput{}
	if pathPrefix != "" {
		input.Filters = []types.Filter{
			{
				Key:    types.FilterNameStringTypeName,
				Values: []string{pathPrefix},
			},
		}
	}

	var paths []string
	paginator := secretsmanager.NewListSecretsPaginator(p.client, input)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("listing secrets: %w", err)
		}

		for _, secret := range page.SecretList {
			if secret.Name != nil {
				paths = append(paths, *secret.Name)
			}
		}
	}

	return paths, nil
}
