package method

import (
	"context"

	"go.flipt.io/flipt/errors"
	"google.golang.org/grpc/metadata"
)

// CallbackValidateState validates the state for the callback request on both OIDC and GitHub as
// an OAuth provider.
func CallbackValidateState(ctx context.Context, state string) error {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return errors.ErrUnauthenticatedf("missing state parameter")
	}

	fstate, ok := md["flipt_client_state"]
	if !ok || len(state) < 1 {
		return errors.ErrUnauthenticatedf("missing state parameter")
	}

	if state != fstate[0] {
		return errors.ErrUnauthenticatedf("unexpected state parameter")
	}

	return nil
}

// FlowVerifier returns the secret bound to the current OIDC login by the
// VerifierCookieKey cookie.
func FlowVerifier(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", errors.ErrUnauthenticatedf("missing verifier")
	}

	verifier := md.Get(VerifierCookieKey)
	if len(verifier) < 1 || verifier[0] == "" {
		return "", errors.ErrUnauthenticatedf("missing verifier")
	}

	return verifier[0], nil
}
