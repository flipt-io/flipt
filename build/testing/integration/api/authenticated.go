package api

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/build/testing/integration"
	"go.flipt.io/flipt/rpc/flipt/auth"
	sdk "go.flipt.io/flipt/sdk/go"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func Authenticated(t *testing.T, client sdk.SDK, opts integration.TestOpts) {
	t.Run("Authentication Methods", func(t *testing.T) {
		ctx := context.Background()

		t.Run("List methods", func(t *testing.T) {
			t.Log(`List methods (ensure at-least 1).`)

			methods, err := client.Auth().PublicAuthenticationService().ListAuthenticationMethods(ctx)

			if opts.AuthConfig.Required() && opts.AuthConfig == integration.AuthNamespaced {
				require.EqualError(t, err, "rpc error: code = Unauthenticated desc = request was not authenticated")
				return
			}

			require.NoError(t, err)

			assert.NotEmpty(t, methods)
		})

		t.Run("Get Self", func(t *testing.T) {
			authn, err := client.Auth().AuthenticationService().GetAuthenticationSelf(ctx)

			if opts.AuthConfig.Required() && opts.AuthConfig == integration.AuthNamespaced {
				require.EqualError(t, err, "rpc error: code = Unauthenticated desc = request was not authenticated")
				return
			}

			require.NoError(t, err)

			assert.NotEmpty(t, authn.Id)
		})

		t.Run("Static Token", func(t *testing.T) {
			t.Log(`Create token.`)

			t.Run("With name and description", func(t *testing.T) {
				resp, err := client.Auth().AuthenticationMethodTokenService().CreateToken(ctx, &auth.CreateTokenRequest{
					Name:        "Access Token",
					Description: "Some kind of access token.",
				})

				// a namespaced token should not be able to create any other tokens
				if opts.AuthConfig.Required() && opts.AuthConfig == integration.AuthNamespaced {
					require.EqualError(t, err, "rpc error: code = Unauthenticated desc = request was not authenticated")
					return
				}

				require.NoError(t, err)

				assert.NotEmpty(t, resp.ClientToken)
				assert.Equal(t, "Access Token", resp.Authentication.Metadata["io.flipt.auth.token.name"])
				assert.Equal(t, "Some kind of access token.", resp.Authentication.Metadata["io.flipt.auth.token.description"])
			})

			t.Run("With name and namespaceKey", func(t *testing.T) {
				// namespaced scoped tokens can only create tokens
				// in the same namespace
				// this ensures that the scope is appropriate for that condition
				namespace := opts.Namespace
				if namespace == "" {
					namespace = "some-namespace"
				}

				resp, err := client.Auth().AuthenticationMethodTokenService().CreateToken(ctx, &auth.CreateTokenRequest{
					Name:         "Scoped Access Token",
					NamespaceKey: namespace,
				})

				// a namespaced token should not be able to create any other tokens
				if opts.AuthConfig.Required() && opts.AuthConfig == integration.AuthNamespaced {
					require.EqualError(t, err, "rpc error: code = Unauthenticated desc = request was not authenticated")
					return
				}

				require.NoError(t, err)

				assert.NotEmpty(t, resp.ClientToken)
				assert.Equal(t, "Scoped Access Token", resp.Authentication.Metadata["io.flipt.auth.token.name"])
				assert.Empty(t, resp.Authentication.Metadata["io.flipt.auth.token.description"])
				assert.Equal(t, namespace, resp.Authentication.Metadata["io.flipt.auth.token.namespace"])
			})
		})

		t.Run("Expire Self", func(t *testing.T) {
			err := client.Auth().AuthenticationService().ExpireAuthenticationSelf(ctx, &auth.ExpireAuthenticationSelfRequest{
				ExpiresAt: timestamppb.Now(),
			})

			if opts.AuthConfig.Required() && opts.AuthConfig == integration.AuthNamespaced {
				require.EqualError(t, err, "rpc error: code = Unauthenticated desc = request was not authenticated")
				return
			}

			require.NoError(t, err)

			t.Log(`Ensure token is no longer valid.`)

			_, err = client.Auth().AuthenticationService().GetAuthenticationSelf(ctx)

			status, ok := status.FromError(err)
			require.True(t, ok)
			assert.Equal(t, codes.Unauthenticated, status.Code())
		})
	})
}
