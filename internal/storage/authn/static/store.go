package static

import (
	"context"
	"fmt"

	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/internal/storage/authn"
	"go.flipt.io/flipt/rpc/flipt"
	rpcauth "go.flipt.io/flipt/rpc/flipt/auth"
	"go.uber.org/zap"
)

var _ authn.Store = (*Store)(nil)

// Store is an implementation of storage.AuthenticationStore
//
// The Store primarily delegates all calls to another embedded store
// However, it does have a fixed set of static preconfigured credentials
// which are interogated on called to GetAuthenticationByClientToken to
// support static token credential lookup.
// Failure to find a static token here causes it to delegate to the embedded store.
type Store struct {
	authn.Store

	logger  *zap.Logger
	byToken map[string]*rpcauth.Authentication
}

// NewStore instantiates a new in-memory implementation of storage.AuthenticationStore
func NewStore(store authn.Store, logger *zap.Logger, storage config.AuthenticationMethodTokenStorage) (*Store, error) {
	s := &Store{
		Store:   store,
		logger:  logger,
		byToken: map[string]*rpcauth.Authentication{},
	}

	for name, token := range storage.Tokens {
		hashedToken, err := authn.HashClientToken(token.Credential)
		if err != nil {
			return nil, err
		}

		s.byToken[hashedToken] = &rpcauth.Authentication{
			Method:    rpcauth.Method_METHOD_TOKEN,
			Id:        name,
			Metadata:  token.Metadata,
			CreatedAt: flipt.Now(),
			UpdatedAt: flipt.Now(),
		}
	}

	return s, nil
}

// GetAuthenticationByClientToken retrieves an instance of Authentication from the backing
// store using the provided clientToken string as the key.
func (s *Store) GetAuthenticationByClientToken(ctx context.Context, clientToken string) (*rpcauth.Authentication, error) {
	hashedToken, err := authn.HashClientToken(clientToken)
	if err != nil {
		return nil, fmt.Errorf("getting authentication by token: %w", err)
	}

	authentication, ok := s.byToken[hashedToken]
	if ok {
		return authentication, nil
	}

	return s.Store.GetAuthenticationByClientToken(ctx, clientToken)
}
