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

// JWTValidator is a function that validates a JWT token.
// It is the common interface between the JWT and Kubernetes methods.
type JWTValidator interface {
	Validate(ctx context.Context, token string) (map[string]any, error)
}
