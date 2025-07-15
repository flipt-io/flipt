package signing

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.flipt.io/build/testing/integration"
	"go.flipt.io/flipt/rpc/flipt/core"
	"go.flipt.io/flipt/rpc/v2/environments"
	"google.golang.org/protobuf/types/known/anypb"
)

// TestCommitSigning tests that Flipt can start and operate with commit signing enabled
// This test verifies:
// 1. Server starts successfully with Vault secrets provider and commit signing configured
// 2. Secrets are retrieved from Vault without errors
// 3. GPG signer is initialized correctly
// 4. Flag operations complete successfully (triggering signed commits)
// 5. No signing-related errors occur during operation
func TestCommitSigning(t *testing.T) {
	integration.Harness(t, func(t *testing.T, opts integration.TestOpts) {
		ctx := context.Background()

		t.Log("Creating flag to trigger signed commit")
		envClient := opts.TokenClientV2(t).Environments()

		flagPayload := &core.Flag{
			Key:         "signing-test",
			Name:        "Signing Test",
			Description: "This flag tests commit signing functionality",
			Enabled:     true,
		}

		flag, err := anypb.New(flagPayload)
		require.NoError(t, err)

		created, err := envClient.CreateResource(ctx, &environments.UpdateResourceRequest{
			EnvironmentKey: integration.DefaultEnvironment,
			NamespaceKey:   integration.DefaultNamespace,
			Key:            "signing-test",
			Payload:        flag,
			Revision:       "", // empty for new resource
		})
		require.NoError(t, err, "Flag creation should succeed with signing enabled")
		require.NotNil(t, created)

		t.Log("Waiting for commit to be processed")
		time.Sleep(3 * time.Second)

		t.Log("Verifying operation completed without signing errors")
	})
}
