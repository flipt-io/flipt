package auth

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
	rpcauth "go.flipt.io/flipt/rpc/flipt/auth"
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
	// ListAuthenticationsRequest retrieves a set of Authentication instances based on the provided
	// predicates with the supplied ListAuthenticationsRequest.
	ListAuthentications(context.Context, *storage.ListRequest[ListAuthenticationsPredicate]) (storage.ResultSet[*auth.Authentication], error)
	// DeleteAuthentications attempts to delete one or more Authentication instances from the backing store.
	// Use DeleteByID to construct a request to delete a single Authentication by ID string.
	// Use DeleteByMethod to construct a request to delete 0 or more Authentications by Method and optional expired before constraint.
	DeleteAuthentications(context.Context, *DeleteAuthenticationsRequest) error
}

// CreateAuthenticationRequest is the argument passed when creating instances
// of an Authentication on a target AuthenticationStore.
type CreateAuthenticationRequest struct {
	Method    auth.Method
	ExpiresAt *timestamppb.Timestamp
	Metadata  map[string]string
}

// ListWithMethod can be passed to storage.NewListRequest.
// The request can then be used to predicate ListAuthentications by auth method.
func ListWithMethod(method rpcauth.Method) storage.ListOption[ListAuthenticationsPredicate] {
	return func(r *storage.ListRequest[ListAuthenticationsPredicate]) {
		r.Predicate.Method = &method
	}
}

// ListAuthenticationsPredicate contains the fields necessary to predicate a list operation
// on a authentications storage backend.
type ListAuthenticationsPredicate struct {
	Method *auth.Method
}

// DeleteAuthenticationsRequest is a request to delete one or more Authentication instances
// in a backing auth.Store.
type DeleteAuthenticationsRequest struct {
	ID     *string
	Method *ByMethod
}

func (d *DeleteAuthenticationsRequest) Valid() error {
	if d.ID == nil && d.Method == nil {
		return errors.ErrInvalidf("delete is not predicated")
	}

	return nil
}

// DeleteByID returns a *DeleteAuthenticationsRequest which identifies a single instance to be
// deleted identified by the configured id string.
func DeleteByID(id string) *DeleteAuthenticationsRequest {
	return &DeleteAuthenticationsRequest{
		ID: &id,
	}
}

// ByMethod is a structure which contain predices used to identify a set of Authentications
// within a backing store. Primarily it identifies a particular auth method as a predicate.
// It also contain an optional ExpiredBefore timestamp to filter the target set.
type ByMethod struct {
	auth.Method
	ExpiredBefore *time.Time
}

// ExpiredBefore is an option which sets the ExpiredBefore timestamp on a ByMethod predicate.
func ExpiredBefore(t time.Time) containers.Option[ByMethod] {
	return func(m *ByMethod) {
		m.ExpiredBefore = &t
	}
}

// DeleteByMethod constructs a DeleteAuthenticationsRequest which identifies a set of Authentication
// instaces using the ByMethod predicate.
func DeleteByMethod(method auth.Method, opts ...containers.Option[ByMethod]) *DeleteAuthenticationsRequest {
	req := &DeleteAuthenticationsRequest{
		Method: &ByMethod{
			Method: method,
		},
	}

	containers.ApplyAll(req.Method, opts...)

	return req
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
