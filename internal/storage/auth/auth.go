package auth

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"

	"go.flipt.io/flipt/internal/storage"
	"go.flipt.io/flipt/rpc/flipt/auth"
	rpcauth "go.flipt.io/flipt/rpc/flipt/auth"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const decodedTokenLen = 32

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
