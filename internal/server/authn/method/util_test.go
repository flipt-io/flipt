package method

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/errors"
	"google.golang.org/grpc/metadata"
)

func TestCallbackValidateState(t *testing.T) {
	t.Run("missing state parameter", func(t *testing.T) {
		err := CallbackValidateState(context.Background(), "")
		require.ErrorIs(t, err, errors.ErrUnauthenticated)
		require.Contains(t, err.Error(), "missing state parameter")
	})

	t.Run("missing metadata parameter", func(t *testing.T) {
		err := CallbackValidateState(context.Background(), "state123")
		require.ErrorIs(t, err, errors.ErrUnauthenticated)
		require.Contains(t, err.Error(), "missing metadata parameter")
	})

	t.Run("missing client state in metadata", func(t *testing.T) {
		ctx := metadata.NewIncomingContext(context.Background(), metadata.MD{})
		err := CallbackValidateState(ctx, "state123")
		require.ErrorIs(t, err, errors.ErrUnauthenticated)
		require.Contains(t, err.Error(), "missing client state in metadata")
	})

	t.Run("state mismatch", func(t *testing.T) {
		ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("flipt_client_state", "otherstate"))
		err := CallbackValidateState(ctx, "state123")
		require.ErrorIs(t, err, errors.ErrUnauthenticated)
		require.Contains(t, err.Error(), "unexpected state parameter")
	})

	t.Run("state matches", func(t *testing.T) {
		ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("flipt_client_state", "state123"))
		err := CallbackValidateState(ctx, "state123")
		require.NoError(t, err)
	})
}
