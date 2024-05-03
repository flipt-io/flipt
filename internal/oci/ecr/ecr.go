package ecr

import (
	"context"
	"encoding/base64"
	"errors"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"oras.land/oras-go/v2/registry/remote/auth"
)

var ErrNoAWSECRAuthorizationData = errors.New("no ecr authorization data provided")

type Client interface {
	GetAuthorizationToken(ctx context.Context, params *ecr.GetAuthorizationTokenInput, optFns ...func(*ecr.Options)) (*ecr.GetAuthorizationTokenOutput, error)
}

type ECR struct {
	client Client
}

func (r *ECR) CredentialFunc(registry string) auth.CredentialFunc {
	return r.Credential
}

func (r *ECR) Credential(ctx context.Context, hostport string) (auth.Credential, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return auth.EmptyCredential, err
	}
	r.client = ecr.NewFromConfig(cfg)
	return r.fetchCredential(ctx)
}

func (r *ECR) fetchCredential(ctx context.Context) (auth.Credential, error) {
	response, err := r.client.GetAuthorizationToken(ctx, &ecr.GetAuthorizationTokenInput{})
	if err != nil {
		return auth.EmptyCredential, err
	}
	if len(response.AuthorizationData) == 0 {
		return auth.EmptyCredential, ErrNoAWSECRAuthorizationData
	}
	token := response.AuthorizationData[0].AuthorizationToken

	if token == nil {
		return auth.EmptyCredential, auth.ErrBasicCredentialNotFound
	}

	output, err := base64.StdEncoding.DecodeString(*token)
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
