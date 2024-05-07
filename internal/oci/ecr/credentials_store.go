package ecr

import (
	"context"
	"encoding/base64"
	"errors"
	"strings"
	"sync"
	"time"

	"oras.land/oras-go/v2/registry/remote/auth"
	"oras.land/oras-go/v2/registry/remote/credentials"
)

func defaultClientFunc(serverAddress string) Client {
	switch {
	case strings.HasPrefix(serverAddress, "public.ecr.aws"):
		return &publicClient{}
	default:
		return &privateClient{}
	}
}

func NewCredentialsStore() credentials.Store {
	return &CredentialsStore{
		cache:      map[string]cacheItem{},
		clientFunc: defaultClientFunc,
	}
}

type cacheItem struct {
	credential auth.Credential
	expiresAt  time.Time
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

// Put saves credentials into the store for the given server address.
func (s *CredentialsStore) Put(ctx context.Context, serverAddress string, cred auth.Credential) error {
	return errors.ErrUnsupported
}

// Delete removes credentials from the store for the given server address.
func (s *CredentialsStore) Delete(ctx context.Context, serverAddress string) error {
	return errors.ErrUnsupported
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
