package ecr

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/aws/aws-sdk-go-v2/service/ecrpublic"
	"oras.land/oras-go/v2/registry/remote/auth"
)

var ErrNoAWSECRAuthorizationData = errors.New("no ecr authorization data provided")

// PrivateClient interface defines methods for interacting with a private Amazon ECR registry
type PrivateClient interface {
	// GetAuthorizationToken retrieves an authorization token for accessing a private ECR registry
	GetAuthorizationToken(ctx context.Context, params *ecr.GetAuthorizationTokenInput, optFns ...func(*ecr.Options)) (*ecr.GetAuthorizationTokenOutput, error)
}

// PublicClient interface defines methods for interacting with a public Amazon ECR registry
type PublicClient interface {
	// GetAuthorizationToken retrieves an authorization token for accessing a public ECR registry
	GetAuthorizationToken(ctx context.Context, params *ecrpublic.GetAuthorizationTokenInput, optFns ...func(*ecrpublic.Options)) (*ecrpublic.GetAuthorizationTokenOutput, error)
}

// Client interface defines a generic method for getting an authorization token.
// This interface will be implemented by PrivateClient and PublicClient
type Client interface {
	// GetAuthorizationToken retrieves an authorization token for accessing an ECR registry (private or public)
	GetAuthorizationToken(ctx context.Context) (string, time.Time, error)
}

type publicClient struct {
	client PublicClient
}

func (r *publicClient) GetAuthorizationToken(ctx context.Context) (string, time.Time, error) {
	client := r.client
	if client == nil {
		cfg, err := config.LoadDefaultConfig(context.Background())
		if err != nil {
			return "", time.Time{}, err
		}
		client = ecrpublic.NewFromConfig(cfg)
	}
	response, err := client.GetAuthorizationToken(ctx, &ecrpublic.GetAuthorizationTokenInput{})
	if err != nil {
		return "", time.Time{}, err
	}
	authData := response.AuthorizationData
	if authData == nil {
		return "", time.Time{}, ErrNoAWSECRAuthorizationData
	}
	if authData.AuthorizationToken == nil {
		return "", time.Time{}, auth.ErrBasicCredentialNotFound
	}
	return *authData.AuthorizationToken, *authData.ExpiresAt, nil
}

type privateClient struct {
	client PrivateClient
}

func (r *privateClient) GetAuthorizationToken(ctx context.Context) (string, time.Time, error) {
	client := r.client
	if client == nil {
		cfg, err := config.LoadDefaultConfig(ctx)
		if err != nil {
			return "", time.Time{}, err
		}
		client = ecr.NewFromConfig(cfg)
	}
	response, err := client.GetAuthorizationToken(ctx, &ecr.GetAuthorizationTokenInput{})
	if err != nil {
		return "", time.Time{}, err
	}
	if len(response.AuthorizationData) == 0 {
		return "", time.Time{}, ErrNoAWSECRAuthorizationData
	}
	authData := response.AuthorizationData[0]

	if authData.AuthorizationToken == nil {
		return "", time.Time{}, auth.ErrBasicCredentialNotFound
	}
	return *authData.AuthorizationToken, *authData.ExpiresAt, nil
}
