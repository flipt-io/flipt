package method

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/errors"
	"google.golang.org/grpc/metadata"
)

func TestCallbackValidateState(t *testing.T) {
	t.Run("missing state parameter", func(t *testing.T) {
		err := CallbackValidateState(t.Context(), "")
		var unauthErr errors.ErrUnauthenticated
		require.ErrorAs(t, err, &unauthErr)
		require.Contains(t, err.Error(), "missing state parameter")
	})

	t.Run("missing metadata parameter", func(t *testing.T) {
		err := CallbackValidateState(t.Context(), "state123")
		var unauthErr errors.ErrUnauthenticated
		require.ErrorAs(t, err, &unauthErr)
		require.Contains(t, err.Error(), "missing metadata parameter")
	})

	t.Run("missing client state in metadata", func(t *testing.T) {
		ctx := metadata.NewIncomingContext(t.Context(), metadata.MD{})
		err := CallbackValidateState(ctx, "state123")
		var unauthErr errors.ErrUnauthenticated
		require.ErrorAs(t, err, &unauthErr)
		require.Contains(t, err.Error(), "missing client state in metadata")
	})

	t.Run("state mismatch", func(t *testing.T) {
		ctx := metadata.NewIncomingContext(t.Context(), metadata.Pairs("flipt_client_state", "otherstate"))
		err := CallbackValidateState(ctx, "state123")
		var unauthErr errors.ErrUnauthenticated
		require.ErrorAs(t, err, &unauthErr)
		require.Contains(t, err.Error(), "unexpected state parameter")
	})

	t.Run("state matches", func(t *testing.T) {
		ctx := metadata.NewIncomingContext(t.Context(), metadata.Pairs("flipt_client_state", "state123"))
		err := CallbackValidateState(ctx, "state123")
		require.NoError(t, err)
	})
}
