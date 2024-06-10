package authn

import (
	"context"
	"fmt"
	"time"

	"go.flipt.io/flipt/internal/storage"
	rpcauth "go.flipt.io/flipt/rpc/flipt/auth"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type bootstrapOpt struct {
	token      string
	expiration time.Duration
	metadata   map[string]string
}

// BootstrapOption is a type which configures the bootstrap or initial static token.
type BootstrapOption func(*bootstrapOpt)

// WithToken overrides the generated token with the provided token.
func WithToken(token string) BootstrapOption {
	return func(o *bootstrapOpt) {
		o.token = token
	}
}

// WithExpiration sets the expiration of the generated token.
func WithExpiration(expiration time.Duration) BootstrapOption {
	return func(o *bootstrapOpt) {
		o.expiration = expiration
	}
}

// WithMetadataAttribute can be used to add additional metadata k/v pairs
// to the resulting bootstrap token
func WithMetadataAttribute(key, value string) BootstrapOption {
	return func(bo *bootstrapOpt) {
		bo.metadata[key] = value
	}
}

// Bootstrap creates an initial static authentication of type token
// if one does not already exist.
func Bootstrap(ctx context.Context, store Store, opts ...BootstrapOption) (string, error) {
	o := bootstrapOpt{
		metadata: map[string]string{
			"io.flipt.auth.token.name":        "initial_bootstrap_token",
			"io.flipt.auth.token.description": "Initial token created when bootstrapping authentication",
		},
	}
	for _, opt := range opts {
		opt(&o)
	}

	set, err := store.ListAuthentications(ctx, storage.ListWithOptions(ListMethod(rpcauth.Method_METHOD_TOKEN)))
	if err != nil {
		return "", fmt.Errorf("bootstrapping authentication store: %w", err)
	}

	// ensures we only create a token if no authentications of type token currently exist
	if len(set.Results) > 0 {
		return "", nil
	}

	req := &CreateAuthenticationRequest{
		Method:   rpcauth.Method_METHOD_TOKEN,
		Metadata: o.metadata,
	}

	// if a client token is provided, use it
	if o.token != "" {
		req.ClientToken = o.token
	}

	// if an expiration is provided, use it
	if o.expiration != 0 {
		req.ExpiresAt = timestamppb.New(time.Now().Add(o.expiration))
	}

	clientToken, _, err := store.CreateAuthentication(ctx, req)

	if err != nil {
		return "", fmt.Errorf("boostrapping authentication store: %w", err)
	}

	return clientToken, nil
}
