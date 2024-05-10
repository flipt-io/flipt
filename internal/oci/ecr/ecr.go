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

// Credential returns a Credential() function that can be used by auth.Client.
func Credential(store *CredentialsStore) auth.CredentialFunc {
	return func(ctx context.Context, hostport string) (auth.Credential, error) {
		return store.Get(ctx, hostport)
	}
}

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

func NewPublicClient(endpoint string) Client {
	return &publicClient{endpoint: endpoint}
}

type publicClient struct {
	client   PublicClient
	endpoint string
}

func (r *publicClient) GetAuthorizationToken(ctx context.Context) (string, time.Time, error) {
	client := r.client
	if client == nil {
		cfg, err := config.LoadDefaultConfig(context.Background())
		if err != nil {
			return "", time.Time{}, err
		}
		client = ecrpublic.NewFromConfig(cfg, func(o *ecrpublic.Options) {
			if r.endpoint != "" {
				o.BaseEndpoint = &r.endpoint
			}
		})
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

func NewPrivateClient(endpoint string) Client {
	return &privateClient{endpoint: endpoint}
}

type privateClient struct {
	client   PrivateClient
	endpoint string
}

func (r *privateClient) GetAuthorizationToken(ctx context.Context) (string, time.Time, error) {
	client := r.client
	if client == nil {
		cfg, err := config.LoadDefaultConfig(ctx)
		if err != nil {
			return "", time.Time{}, err
		}
		client = ecr.NewFromConfig(cfg, func(o *ecr.Options) {
			if r.endpoint != "" {
				o.BaseEndpoint = &r.endpoint
			}
		})
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
