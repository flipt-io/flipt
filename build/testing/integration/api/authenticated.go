package api

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/rpc/flipt/auth"
	sdk "go.flipt.io/flipt/sdk/go"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func Authenticated(t *testing.T, client sdk.SDK) {
	t.Run("Authentication Methods", func(t *testing.T) {
		ctx := context.Background()

		t.Log(`List methods (ensure at-least 1).`)

		methods, err := client.Auth().PublicAuthenticationService().ListAuthenticationMethods(ctx)
		require.NoError(t, err)

		assert.NotEmpty(t, methods)

		t.Run("Get Self", func(t *testing.T) {
			authn, err := client.Auth().AuthenticationService().GetAuthenticationSelf(ctx)
			require.NoError(t, err)

			assert.NotEmpty(t, authn.Id)
		})

		t.Run("Static Token", func(t *testing.T) {
			t.Log(`Create token.`)

			resp, err := client.Auth().AuthenticationMethodTokenService().CreateToken(ctx, &auth.CreateTokenRequest{
				Name:        "Access Token",
				Description: "Some kind of access token.",
			})
			require.NoError(t, err)

			assert.NotEmpty(t, resp.ClientToken)
			assert.Equal(t, "Access Token", resp.Authentication.Metadata["io.flipt.auth.token.name"])
			assert.Equal(t, "Some kind of access token.", resp.Authentication.Metadata["io.flipt.auth.token.description"])
		})

		t.Run("Expire Self", func(t *testing.T) {
			err := client.Auth().AuthenticationService().ExpireAuthenticationSelf(ctx, &auth.ExpireAuthenticationSelfRequest{
				ExpiresAt: timestamppb.Now(),
			})
			require.NoError(t, err)

			t.Log(`Ensure token is no longer valid.`)

			_, err = client.Auth().AuthenticationService().GetAuthenticationSelf(ctx)

			status, ok := status.FromError(err)
			require.True(t, ok)
			assert.Equal(t, codes.Unauthenticated, status.Code())
		})
	})
}
