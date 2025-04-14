package authn

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/build/testing/integration"
	"go.flipt.io/flipt/rpc/flipt"
	"go.flipt.io/flipt/rpc/v2/environments"
	sdk "go.flipt.io/flipt/sdk/go"
	sdkv2 "go.flipt.io/flipt/sdk/go/v2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func Common(t *testing.T, opts integration.TestOpts) {
	client := opts.BootstrapClient(t)

	t.Run("Authentication Methods", func(t *testing.T) {
		ctx := context.Background()

		t.Run("List methods", func(t *testing.T) {
			t.Log(`List methods (ensure at-least 1).`)

			methods, err := client.Auth().PublicAuthenticationService().ListAuthenticationMethods(ctx)
			require.NoError(t, err)
			assert.NotEmpty(t, methods)
		})

		t.Run("Get Self", func(t *testing.T) {
			authn, err := client.Auth().AuthenticationService().GetAuthenticationSelf(ctx)
			require.NoError(t, err)
			assert.NotEmpty(t, authn.Id)
		})

		for _, namespace := range integration.Namespaces {
			t.Run(fmt.Sprintf("InNamespace(%q)", namespace.Key), func(t *testing.T) {
				t.Run("NoAuth", func(t *testing.T) {
					client := opts.NoAuthClient(t)
					cannotReadAnyIn(t, ctx, client, namespace.Key)
					cannotReadEnvironmentsAnyIn(t, ctx, opts.ClientV2(t), namespace.Key)
				})

				t.Run("StaticToken", func(t *testing.T) {
					// ensure we can do resource specific operations across namespaces
					canReadAllIn(t, ctx, client, namespace.Key)
					canReadEnvironmentsAllIn(t, ctx, opts.TokenClientV2(t), namespace.Key)
				})

				t.Run("K8sClient", func(t *testing.T) {
					canReadAllIn(t, ctx, opts.K8sClient(t), namespace.Key)
					// k8s auth is not supported for sdkv2
				})

				t.Run("JWTClient", func(t *testing.T) {
					canReadAllIn(t, ctx, opts.JWTClient(t), namespace.Key)
					canReadEnvironmentsAllIn(t, ctx, opts.JWTClientV2(t), namespace.Key)
				})
			})
		}
	})
}

func canReadAllIn(t *testing.T, ctx context.Context, client sdk.SDK, namespace string) {
	t.Run("CanReadAll", func(t *testing.T) {
		_, err := client.Flipt().ListFlags(ctx, &flipt.ListFlagRequest{
			NamespaceKey: namespace,
		})
		assertErrIsUnauthenticated(t, err, false)
	})
}

func cannotReadAnyIn(t *testing.T, ctx context.Context, client sdk.SDK, namespace string) {
	t.Run("CannotReadAny", func(t *testing.T) {
		_, err := client.Flipt().ListFlags(ctx, &flipt.ListFlagRequest{
			NamespaceKey: namespace,
		})
		assertErrIsUnauthenticated(t, err, true)
	})
}

func canReadEnvironmentsAllIn(t *testing.T, ctx context.Context, client sdkv2.SDK, namespace string) {
	t.Run("EnvironmentsCanReadAll", func(t *testing.T) {
		// ensure we can do resource specific operations across namespaces
		client := client.Environments()
		_, err := client.ListEnvironments(ctx, &environments.ListEnvironmentsRequest{})
		assertErrIsUnauthenticated(t, err, false)
	})
}

func cannotReadEnvironmentsAnyIn(t *testing.T, ctx context.Context, client sdkv2.SDK, _namespace string) {
	t.Run("EnvironmentsCannotReadAny", func(t *testing.T) {
		// ensure we cannot do resource specific operations across namespaces
		client := client.Environments()
		_, err := client.ListEnvironments(ctx, &environments.ListEnvironmentsRequest{})
		assertErrIsUnauthenticated(t, err, true)
	})
}

func assertErrIsUnauthenticated(t *testing.T, err error, unauthenticated bool) {
	if !assert.Equal(
		t,
		codes.Unauthenticated == status.Code(err),
		unauthenticated,
		"expected unauthenticated error",
	) {
		t.Logf("error: %v", err)
	}
}
