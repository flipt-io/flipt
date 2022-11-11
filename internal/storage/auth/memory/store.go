package memory

import (
	"context"
	"fmt"
	"sort"
	"strconv"
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
func (s *Store) CreateAuthentication(_ context.Context, r *auth.CreateAuthenticationRequest) (string, *rpcauth.Authentication, error) {
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

func (s *Store) ListAuthentications(ctx context.Context, req *storage.ListRequest[auth.ListAuthenticationsPredicate]) (storage.ResultSet[*rpcauth.Authentication], error) {
	var set storage.ResultSet[*rpcauth.Authentication]

	// adjust the query parameters within normal bounds
	req.QueryParams.Normalize()

	// copy all auths into slice
	s.mu.Lock()
	set.Results = make([]*rpcauth.Authentication, 0, len(s.auths))
	for _, res := range s.auths {
		set.Results = append(set.Results, res)
	}
	s.mu.Unlock()

	// sort by created_at and specified order
	sort.Slice(set.Results, func(i, j int) bool {
		if req.QueryParams.Order != storage.OrderAsc {
			i, j = j, i
		}

		return set.Results[i].CreatedAt.AsTime().
			Before(set.Results[j].CreatedAt.AsTime())
	})

	// parse page token as an offset integer
	var offset int
	if v, err := strconv.ParseInt(req.QueryParams.PageToken, 10, 64); err == nil {
		offset = int(v)
	}

	// ensure end of page does not exceed entire set
	end := offset + int(req.QueryParams.Limit)
	if end > len(set.Results) {
		end = len(set.Results)
	} else if end < len(set.Results) {
		// set next page token given there are more entries
		set.NextPageToken = fmt.Sprintf("%d", end)
	}

	// reduce results set to requested page
	set.Results = set.Results[offset:end]

	return set, nil
}

func (s *Store) deleteMatching(fn func(*rpcauth.Authentication) bool) (count int) {
	for hashedToken, a := range s.auths {
		if fn(a) {
			delete(s.auths, hashedToken)
			count++
		}
	}

	return
}

func (s *Store) DeleteAuthentications(_ context.Context, req *auth.DeleteAuthenticationsRequest) error {
	if err := req.Valid(); err != nil {
		return fmt.Errorf("deleting authentications: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if req.ID != nil {
		if s.deleteMatching(func(a *rpcauth.Authentication) bool {
			return a.Id == *req.ID
		}) > 0 {
			return nil
		}

		return errors.ErrNotFoundf("authentication %q", *req.ID)
	}

	if req.Method != nil {
		_ = s.deleteMatching(func(a *rpcauth.Authentication) bool {
			return a.Method == req.Method.Method &&
				(req.Method.ExpiredBefore == nil || a.ExpiresAt.AsTime().Before(*req.Method.ExpiredBefore))
		})

		return nil
	}

	// should not be reachable

	return nil
}
