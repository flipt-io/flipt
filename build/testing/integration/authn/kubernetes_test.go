package authn

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/build/testing/integration"
	"go.flipt.io/flipt/rpc/flipt"
)

func TestKubernetesAuthenticationCaching(t *testing.T) {
	integration.Harness(t, func(t *testing.T, opts integration.TestOpts) {
		client := opts.K8sClient(t)

		// Make multiple requests to ensure caching works
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// First request should succeed and cache the token
		_, err := client.Flipt().ListFlags(ctx, &flipt.ListFlagRequest{})
		require.NoError(t, err)

		// Second request should use cached token
		_, err = client.Flipt().ListFlags(ctx, &flipt.ListFlagRequest{})
		require.NoError(t, err)

		// Third request should still use cached token
		_, err = client.Flipt().ListFlags(ctx, &flipt.ListFlagRequest{})
		require.NoError(t, err)
	})
}

func TestKubernetesAuthenticationMetadata(t *testing.T) {
	integration.Harness(t, func(t *testing.T, opts integration.TestOpts) {
		client := opts.K8sClient(t)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Get authentication info to verify metadata
		auth, err := client.Auth().AuthenticationService().GetAuthenticationSelf(ctx)
		require.NoError(t, err)

		// Verify the expected metadata fields are present
		assert.Contains(t, auth.Metadata, "io.flipt.auth.k8s.namespace")
		assert.Contains(t, auth.Metadata, "io.flipt.auth.k8s.pod.name")
		assert.Contains(t, auth.Metadata, "io.flipt.auth.k8s.pod.uid")
		assert.Contains(t, auth.Metadata, "io.flipt.auth.k8s.serviceaccount.name")
		assert.Contains(t, auth.Metadata, "io.flipt.auth.k8s.serviceaccount.uid")
	})
}
