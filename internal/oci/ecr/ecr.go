package ecr

import (
	"context"
	"encoding/base64"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/aws/aws-sdk-go-v2/service/ecrpublic"
	"oras.land/oras-go/v2/registry/remote/auth"
)

var ErrNoAWSECRAuthorizationData = errors.New("no ecr authorization data provided")

type PrivateClient interface {
	GetAuthorizationToken(ctx context.Context, params *ecr.GetAuthorizationTokenInput, optFns ...func(*ecr.Options)) (*ecr.GetAuthorizationTokenOutput, error)
}
type PublicClient interface {
	GetAuthorizationToken(ctx context.Context, params *ecrpublic.GetAuthorizationTokenInput, optFns ...func(*ecrpublic.Options)) (*ecrpublic.GetAuthorizationTokenOutput, error)
}

type Client interface {
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

func CredentialFunc(registry string) auth.CredentialFunc {
	return new(registry).Credential
}

func new(registry string) *svc {
	r := &svc{}
	switch {
	case strings.HasPrefix(registry, "public.ecr.aws"):
		r.client = &publicClient{}
	default:
		r.client = &privateClient{}
	}
	return r
}

type svc struct {
	mu               sync.Mutex
	expiresAt        time.Time
	cachedCredential auth.Credential
	client           Client
}

func (r *svc) Credential(ctx context.Context, hostport string) (auth.Credential, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if time.Now().UTC().Before(r.expiresAt) {
		return r.cachedCredential, nil
	}
	token, expiresAt, err := r.client.GetAuthorizationToken(ctx)
	if err != nil {
		return auth.EmptyCredential, err
	}
	credential, err := r.extractCredential(token)
	if err != nil {
		return auth.EmptyCredential, err
	}
	r.cachedCredential = credential
	r.expiresAt = expiresAt
	return credential, nil
}

func (r *svc) extractCredential(token string) (auth.Credential, error) {
	output, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return auth.EmptyCredential, err
	}

	userpass := strings.SplitN(string(output), ":", 2)
	if len(userpass) != 2 {
		return auth.EmptyCredential, auth.ErrBasicCredentialNotFound
	}

	return auth.Credential{
		Username: userpass[0],
		Password: userpass[1],
	}, nil
}
