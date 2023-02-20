package auth

import (
	"context"
	"fmt"

	"go.flipt.io/flipt/internal/storage"
	rpcauth "go.flipt.io/flipt/rpc/flipt/auth"
)

type bootstrapOpt struct {
	clientToken string
}

type BootstrapOption func(*bootstrapOpt)

func WithClientToken(clientToken string) BootstrapOption {
	return func(o *bootstrapOpt) {
		o.clientToken = clientToken
	}
}

// Bootstrap creates an initial static authentication of type token
// if one does not already exist.
func Bootstrap(ctx context.Context, store Store, opts ...BootstrapOption) (string, error) {
	var o bootstrapOpt
	for _, opt := range opts {
		opt(&o)
	}

	set, err := store.ListAuthentications(ctx, storage.NewListRequest(ListWithMethod(rpcauth.Method_METHOD_TOKEN)))
	if err != nil {
		return "", fmt.Errorf("bootstrapping authentication store: %w", err)
	}

	// ensures we only create a token if no authentications of type token currently exist
	if len(set.Results) > 0 {
		return "", nil
	}

	req := &CreateAuthenticationRequest{
		Method: rpcauth.Method_METHOD_TOKEN,
		Metadata: map[string]string{
			"io.flipt.auth.token.name":        "initial_bootstrap_token",
			"io.flipt.auth.token.description": "Initial token created when bootstrapping authentication",
		},
	}

	// if a client token is provided, use it
	if o.clientToken != "" {
		req.ClientToken = o.clientToken
	}

	clientToken, _, err := store.CreateAuthentication(ctx, req)

	if err != nil {
		return "", fmt.Errorf("boostrapping authentication store: %w", err)
	}

	return clientToken, nil
}
