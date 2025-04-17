package k8s

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/build/testing/integration"
	"go.flipt.io/build/testing/integration/authn/helpers"
)

func TestK8sAuth(t *testing.T) {
	integration.Harness(t, func(t *testing.T, opts integration.TestOpts) {
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
						helpers.CannotReadAnyIn(t, ctx, client, namespace.Key)
						helpers.CannotReadEnvironmentsAnyIn(t, ctx, opts.NoAuthClientV2(t))
					})

					t.Run("K8sClient", func(t *testing.T) {
						helpers.CanReadAllIn(t, ctx, opts.K8sClient(t), namespace.Key)
						helpers.CanReadEnvironmentsAllIn(t, ctx, opts.K8sClientV2(t))
					})
				})
			}
		})
	})
}
