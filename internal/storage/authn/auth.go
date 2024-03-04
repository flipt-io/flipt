package authn

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"time"

	"go.flipt.io/flipt/errors"
	"go.flipt.io/flipt/internal/containers"
	"go.flipt.io/flipt/internal/storage"
	"go.flipt.io/flipt/rpc/flipt/auth"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const decodedTokenLen = 32

// Store persists Authentication instances.
type Store interface {
	// CreateAuthentication creates a new instance of an Authentication and returns a unique clientToken
	// string which can be used to retrieve the Authentication again via GetAuthenticationByClientToken.
	CreateAuthentication(context.Context, *CreateAuthenticationRequest) (string, *auth.Authentication, error)
	// GetAuthenticationByClientToken retrieves an instance of Authentication from the backing
	// store using the provided clientToken string as the key.
	GetAuthenticationByClientToken(ctx context.Context, clientToken string) (*auth.Authentication, error)
	// GetAuthenticationByID retrieves an instance of Authentication from the backing
	// store using the provided id string.
	GetAuthenticationByID(ctx context.Context, id string) (*auth.Authentication, error)
	// ListAuthenticationsRequest retrieves a set of Authentication instances based on the provided
	// predicates with the supplied ListAuthenticationsRequest.
	ListAuthentications(context.Context, *storage.ListRequest[ListAuthenticationsPredicate]) (storage.ResultSet[*auth.Authentication], error)
	// DeleteAuthentications attempts to delete one or more Authentication instances from the backing store.
	// Use DeleteByID to construct a request to delete a single Authentication by ID string.
	// Use DeleteByMethod to construct a request to delete 0 or more Authentications by Method and optional expired before constraint.
	DeleteAuthentications(context.Context, *DeleteAuthenticationsRequest) error
	// ExpireAuthenticationByID attempts to expire an Authentication by ID string and the provided expiry time.
	ExpireAuthenticationByID(context.Context, string, *timestamppb.Timestamp) error
}

// CreateAuthenticationRequest is the argument passed when creating instances
// of an Authentication on a target AuthenticationStore.
type CreateAuthenticationRequest struct {
	Method    auth.Method
	ExpiresAt *timestamppb.Timestamp
	Metadata  map[string]string
	// ClientToken is an (optional) explicit client token to be associated with the authentication.
	// When it is not supplied a random token will be generated and returned instead.
	ClientToken string
}

// ListMethod can be passed to storage.NewListRequest.
// The request can then be used to predicate ListAuthentications by auth method.
func ListMethod(method auth.Method) ListAuthenticationsPredicate {
	return ListAuthenticationsPredicate{Method: &method}
}

// ListAuthenticationsPredicate contains the fields necessary to predicate a list operation
// on a authentications storage backend.
type ListAuthenticationsPredicate struct {
	Method *auth.Method
}

// DeleteAuthenticationsRequest is a request to delete one or more Authentication instances
// in a backing auth.Store.
type DeleteAuthenticationsRequest struct {
	ID            *string
	Method        *auth.Method
	ExpiredBefore *timestamppb.Timestamp
}

func (d *DeleteAuthenticationsRequest) Valid() error {
	if d.ID == nil && d.Method == nil && d.ExpiredBefore == nil {
		return errors.ErrInvalidf("id, method or expired-before timestamp is required")
	}

	return nil
}

// Delete constructs a new *DeleteAuthenticationsRequest using the provided options.
func Delete(opts ...containers.Option[DeleteAuthenticationsRequest]) *DeleteAuthenticationsRequest {
	req := &DeleteAuthenticationsRequest{}

	for _, opt := range opts {
		opt(req)
	}

	return req
}

// WithID is an option which predicates a delete with a specific authentication ID.
func WithID(id string) containers.Option[DeleteAuthenticationsRequest] {
	return func(r *DeleteAuthenticationsRequest) {
		r.ID = &id
	}
}

// WithMethod is an option which ensures a delete applies to Authentications of the provided method.
func WithMethod(method auth.Method) containers.Option[DeleteAuthenticationsRequest] {
	return func(r *DeleteAuthenticationsRequest) {
		r.Method = &method
	}
}

// WithExpiredBefore is an option which ensures a delete only applies to Auhentications
// with an expires_at timestamp occurring before the supplied timestamp.
func WithExpiredBefore(t time.Time) containers.Option[DeleteAuthenticationsRequest] {
	return func(r *DeleteAuthenticationsRequest) {
		r.ExpiredBefore = timestamppb.New(t)
	}
}

// GenerateRandomToken produces a URL safe base64 encoded string of random characters
// the data is sourced from a pseudo-random input stream
func GenerateRandomToken() string {
	var token [decodedTokenLen]byte
	if _, err := rand.Read(token[:]); err != nil {
		panic(err)
	}

	return base64.URLEncoding.EncodeToString(token[:])
}

// HashClientToken performs a SHA256 sum on the input string
// it returns the result as a URL safe base64 encoded string
func HashClientToken(token string) (string, error) {
	// produce SHA256 hash of token
	hash := sha256.New()
	if _, err := hash.Write([]byte(token)); err != nil {
		return "", fmt.Errorf("hashing client token: %w", err)
	}

	// base64(sha256sum)
	var (
		data = make([]byte, 0, base64.URLEncoding.EncodedLen(hash.Size()))
		buf  = bytes.NewBuffer(data)
		enc  = base64.NewEncoder(base64.URLEncoding, buf)
	)

	if _, err := enc.Write(hash.Sum(nil)); err != nil {
		return "", fmt.Errorf("hashing client token: %w", err)
	}

	if err := enc.Close(); err != nil {
		return "", fmt.Errorf("hashing client token: %w", err)
	}

	return buf.String(), nil
}
