package auth

import (
	"context"
	"fmt"

	"go.flipt.io/flipt/internal/storage"
	rpcauth "go.flipt.io/flipt/rpc/flipt/auth"
)

// Bootstrap creates an initial static authentication of type token
// if one does not already exist.
func Bootstrap(ctx context.Context, store storage.AuthenticationStore) (string, error) {
	req := storage.NewListRequest(ListWithMethod(rpcauth.Method_METHOD_TOKEN))
	set, err := store.ListAuthentications(ctx, req)
	if err != nil {
		return "", fmt.Errorf("bootstrapping authentication store: %w", err)
	}

	// if at-least auth token exists we do not create one
	if len(set.Results) > 0 {
		return "", nil
	}

	clientToken, _, err := store.CreateAuthentication(ctx, &storage.CreateAuthenticationRequest{
		Method: rpcauth.Method_METHOD_TOKEN,
		Metadata: map[string]string{
			"io.flipt.auth.token.name":        "initial_bootstrap_token",
			"io.flipt.auth.token.description": "Initial token created when bootstrapping authentication",
		},
	})

	if err != nil {
		return "", fmt.Errorf("boostrapping authentication store: %w", err)
	}

	return clientToken, nil
}
