package memory

import (
	"context"
	"fmt"
	"sync"

	"github.com/gofrs/uuid"
	"go.flipt.io/flipt/errors"
	"go.flipt.io/flipt/internal/storage"
	"go.flipt.io/flipt/internal/storage/auth"
	rpcauth "go.flipt.io/flipt/rpc/flipt/auth"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Store is an in-memory implementation of storage.AuthenticationStore
//
// Authentications are stored in a map by hashedClientToken.
// Access to the map is protected by a mutex, meaning this is implementation
// is safe to use concurrently.
type Store struct {
	mu    sync.Mutex
	auths map[string]*rpcauth.Authentication

	now           func() *timestamppb.Timestamp
	generateID    func() string
	generateToken func() string
}

// Option is a type which configures a *Store
type Option func(*Store)

// NewStore instantiates a new in-memory implementation of storage.AuthenticationStore
func NewStore(opts ...Option) *Store {
	store := &Store{
		auths: map[string]*rpcauth.Authentication{},
		now:   timestamppb.Now,
		generateID: func() string {
			return uuid.Must(uuid.NewV4()).String()
		},
		generateToken: auth.GenerateRandomToken,
	}

	for _, opt := range opts {
		opt(store)
	}

	return store
}

// WithNowFunc overrides the stores now() function used to obtain
// a protobuf timestamp representative of the current time of evaluation.
func WithNowFunc(fn func() *timestamppb.Timestamp) Option {
	return func(s *Store) {
		s.now = fn
	}
}

// WithTokenGeneratorFunc overrides the stores token generator function
// used to generate new random token strings as client tokens, when
// creating new instances of Authentication.
// The default is a pseudo-random string of bytes base64 encoded.
func WithTokenGeneratorFunc(fn func() string) Option {
	return func(s *Store) {
		s.generateToken = fn
	}
}

// WithIDGeneratorFunc overrides the stores ID generator function
// used to generate new random ID strings, when creating new instances
// of Authentications.
// The default is a string containing a valid UUID (V4).
func WithIDGeneratorFunc(fn func() string) Option {
	return func(s *Store) {
		s.generateID = fn
	}
}

// CreateAuthentication creates a new instance of an Authentication and returns a unique clientToken
// string which can be used to retrieve the Authentication again via GetAuthenticationByClientToken.
func (s *Store) CreateAuthentication(_ context.Context, r *storage.CreateAuthenticationRequest) (string, *rpcauth.Authentication, error) {
	if r.ExpiresAt != nil && !r.ExpiresAt.IsValid() {
		return "", nil, errors.ErrInvalidf("invalid expiry time: %v", r.ExpiresAt)
	}

	var (
		now            = s.now()
		clientToken    = s.generateToken()
		authentication = &rpcauth.Authentication{
			Id:        s.generateID(),
			Method:    r.Method,
			Metadata:  r.Metadata,
			ExpiresAt: r.ExpiresAt,
			CreatedAt: now,
			UpdatedAt: now,
		}
	)

	hashedToken, err := auth.HashClientToken(clientToken)
	if err != nil {
		return "", nil, fmt.Errorf("creating authentication: %w", err)
	}

	s.mu.Lock()
	s.auths[hashedToken] = authentication
	s.mu.Unlock()

	return clientToken, authentication, nil
}

// GetAuthenticationByClientToken retrieves an instance of Authentication from the backing
// store using the provided clientToken string as the key.
func (s *Store) GetAuthenticationByClientToken(ctx context.Context, clientToken string) (*rpcauth.Authentication, error) {
	hashedToken, err := auth.HashClientToken(clientToken)
	if err != nil {
		return nil, fmt.Errorf("getting authentication by token: %w", err)
	}

	s.mu.Lock()
	authentication, ok := s.auths[hashedToken]
	s.mu.Unlock()
	if !ok {
		return nil, errors.ErrNotFoundf("getting authentication by token")
	}

	return authentication, nil
}
