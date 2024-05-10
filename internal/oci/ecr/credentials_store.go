package ecr

import (
	"context"
	"encoding/base64"
	"strings"
	"sync"
	"time"

	"oras.land/oras-go/v2/registry/remote/auth"
)

func defaultClientFunc(endpoint string) func(serverAddress string) Client {
	return func(serverAddress string) Client {
		switch {
		case strings.HasPrefix(serverAddress, "public.ecr.aws"):
			return NewPublicClient(endpoint)
		default:
			return NewPrivateClient(endpoint)
		}
	}
}

type cacheItem struct {
	credential auth.Credential
	expiresAt  time.Time
}

func NewCredentialsStore(endpoint string) *CredentialsStore {
	return &CredentialsStore{
		cache:      map[string]cacheItem{},
		clientFunc: defaultClientFunc(endpoint),
	}
}

type CredentialsStore struct {
	mu         sync.Mutex
	cache      map[string]cacheItem
	clientFunc func(serverAddress string) Client
}

// Get retrieves credentials from the store for the given server address.
func (s *CredentialsStore) Get(ctx context.Context, serverAddress string) (auth.Credential, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if item, ok := s.cache[serverAddress]; ok && time.Now().UTC().Before(item.expiresAt) {
		return item.credential, nil
	}

	token, expiresAt, err := s.clientFunc(serverAddress).GetAuthorizationToken(ctx)
	if err != nil {
		return auth.EmptyCredential, err
	}
	credential, err := s.extractCredential(token)
	if err != nil {
		return auth.EmptyCredential, err
	}
	s.cache[serverAddress] = cacheItem{credential, expiresAt}
	return credential, nil
}

func (s *CredentialsStore) extractCredential(token string) (auth.Credential, error) {
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
