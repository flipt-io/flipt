package environments

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/build/testing/integration"
	"go.flipt.io/flipt/rpc/flipt/core"
	"go.flipt.io/flipt/rpc/flipt/evaluation"
	environments "go.flipt.io/flipt/rpc/v2/environments"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

func ptr[T any](t T) *T {
	return &t
}

func API(t *testing.T, ctx context.Context, opts integration.TestOpts) {
	var (
		envClient  = opts.TokenClientV2(t).Environments()
		evalClient = opts.BootstrapClient(t).Evaluation()
		revision   string
	)

	t.Run("Namespaces", func(t *testing.T) {
		nl, err := envClient.ListNamespaces(
			ctx,
			&environments.ListNamespacesRequest{Environment: integration.DefaultEnvironment},
		)
		require.NoError(t, err)

		revision = nl.Revision

		t.Run(`Ensure default cannot be created.`, func(t *testing.T) {
			_, err := envClient.CreateNamespace(ctx, &environments.UpdateNamespaceRequest{
				Environment: integration.DefaultEnvironment,
				Key:         integration.DefaultNamespace,
				Name:        "Default",
				Revision:    revision,
			})
			require.EqualError(t, err, "rpc error: code = AlreadyExists desc = create namespace \"default\"")

			t.Log(`Ensure default cannot be deleted.`)

			_, err = envClient.DeleteNamespace(ctx, &environments.DeleteNamespaceRequest{
				Environment: integration.DefaultEnvironment,
				Key:         integration.DefaultNamespace,
				Revision:    revision,
			})
			require.EqualError(t, err, "rpc error: code = InvalidArgument desc = namespace \"default\" is protected")
		})

		t.Log(`Create namespace.`)

		created, err := envClient.CreateNamespace(ctx, &environments.UpdateNamespaceRequest{
			Environment: integration.DefaultEnvironment,
			Key:         integration.AlternativeNamespace,
			Name:        "alternative",
			Revision:    revision,
		})
		require.NoError(t, err)

		assert.Equal(t, integration.AlternativeNamespace, created.Namespace.Key)
		assert.Equal(t, "alternative", created.Namespace.Name)

		revision = created.Revision

		t.Log(`Get namespace by key.`)

		retrieved, err := envClient.GetNamespace(ctx, &environments.GetNamespaceRequest{
			Environment: integration.DefaultEnvironment,
			Key:         integration.AlternativeNamespace,
		})
		require.NoError(t, err)

		assert.Equal(t, created.Namespace.Name, retrieved.Namespace.Name)

		t.Log(`List namespaces.`)

		namespaces, err := envClient.ListNamespaces(ctx, &environments.ListNamespacesRequest{
			Environment: integration.DefaultEnvironment,
		})
		require.NoError(t, err)

		assert.Len(t, namespaces.Items, 2)

		assert.Equal(t, integration.DefaultNamespace, namespaces.Items[0].Key)
		assert.Equal(t, integration.AlternativeNamespace, namespaces.Items[1].Key)

		t.Log(`Update namespace.`)

		updated, err := envClient.UpdateNamespace(ctx, &environments.UpdateNamespaceRequest{
			Environment: integration.DefaultEnvironment,
			Key:         integration.AlternativeNamespace,
			Name:        "production",
			Description: ptr("Some kind of description"),
			Revision:    revision,
		})
		require.NoError(t, err)

		assert.Equal(t, ptr("Some kind of description"), updated.Namespace.Description)

		revision = updated.Revision
	})

	for _, namespace := range (integration.NamespaceExpectations{
		{Key: integration.DefaultNamespace, Expected: integration.DefaultNamespace},
		{Key: integration.AlternativeNamespace, Expected: integration.AlternativeNamespace},
	}) {
		t.Run(fmt.Sprintf("namespace %q", namespace.Expected), func(t *testing.T) {
			t.Run("Flags and Variants", func(t *testing.T) {
				{
					t.Log("Create a new enabled flag with key \"test\".")

					flagPayload := &core.Flag{
						Key:         "test",
						Name:        "Test",
						Description: "This is a test flag",
						Enabled:     true,
						Variants: []*core.Variant{
							{
								Key:  "foo",
								Name: "Foo",
							},
							{
								Key:  "bar",
								Name: "Bar",
							},
						},
						Rules:          []*core.Rule{},
						Rollouts:       []*core.Rollout{},
						DefaultVariant: ptr("foo"),
					}

					flag, err := anypb.New(flagPayload)
					require.NoError(t, err)

					enabled, err := envClient.CreateResource(ctx, &environments.UpdateResourceRequest{
						Environment: integration.DefaultEnvironment,
						Namespace:   namespace.Key,
						Key:         "test",
						Payload:     flag,
						Revision:    revision,
					})
					require.NoError(t, err)

					assert.Equal(t, "test", enabled.Resource.Key)

					var enabledFlag core.Flag
					err = anypb.UnmarshalTo(enabled.Resource.Payload, &enabledFlag, proto.UnmarshalOptions{})
					require.NoError(t, err)

					enabled.Resource.Payload.TypeUrl = strings.TrimPrefix(enabled.Resource.Payload.TypeUrl, "type.googleapis.com/")

					assert.Equal(t, "Test", enabledFlag.Name)
					assert.Equal(t, "This is a test flag", enabledFlag.Description)
					assert.True(t, enabledFlag.Enabled, "Flag should be enabled")
					assert.Equal(t, []*core.Variant{
						{
							Key:  "foo",
							Name: "Foo",
						},
						{
							Key:  "bar",
							Name: "Bar",
						},
					}, enabledFlag.Variants)
					assert.Equal(t, ptr("foo"), enabledFlag.DefaultVariant)

					revision = enabled.Revision

					resp, err := evalClient.Variant(ctx, &evaluation.EvaluationRequest{
						NamespaceKey: namespace.Key,
						FlagKey:      "test",
					})
					require.NoError(t, err)

					// it shouldn't match as we have no segments
					assert.Equal(t, false, resp.Match)
				}

				{
					t.Log("Create a new flag in a disabled state.")

					disabledFlagAny, err := anypb.New(&core.Flag{
						Key:         "disabled",
						Name:        "Disabled",
						Description: "This is a disabled test flag",
						Enabled:     false,
					})
					require.NoError(t, err)

					disabled, err := envClient.CreateResource(ctx, &environments.UpdateResourceRequest{
						Environment: integration.DefaultEnvironment,
						Namespace:   namespace.Key,
						Key:         "disabled",
						Payload:     disabledFlagAny,
						Revision:    revision,
					})
					require.NoError(t, err)

					var disabledFlag core.Flag
					err = anypb.UnmarshalTo(disabled.Resource.Payload, &disabledFlag, proto.UnmarshalOptions{})
					require.NoError(t, err)

					assert.Equal(t, "disabled", disabledFlag.Key)
					assert.Equal(t, "Disabled", disabledFlag.Name)
					assert.Equal(t, "This is a disabled test flag", disabledFlag.Description)
					assert.False(t, disabledFlag.Enabled, "Flag should be disabled")

					revision = disabled.Revision
				}

				{
					t.Log("Create a new enabled boolean flag with key \"boolean_enabled\".")

					booleanFlagAny, err := anypb.New(&core.Flag{
						Type:        core.FlagType_BOOLEAN_FLAG_TYPE,
						Key:         "boolean_enabled",
						Name:        "Boolean Enabled",
						Description: "This is an enabled boolean test flag",
						Enabled:     true,
					})
					require.NoError(t, err)

					stripAnyTypePrefix(t, booleanFlagAny)

					booleanEnabled, err := envClient.CreateResource(ctx, &environments.UpdateResourceRequest{
						Environment: integration.DefaultEnvironment,
						Namespace:   namespace.Key,
						Key:         "boolean_enabled",
						Payload:     booleanFlagAny,
						Revision:    revision,
					})
					require.NoError(t, err)

					var booleanFlag core.Flag
					err = anypb.UnmarshalTo(booleanEnabled.Resource.Payload, &booleanFlag, proto.UnmarshalOptions{})
					require.NoError(t, err)

					assert.Equal(t, "boolean_enabled", booleanFlag.Key)
					assert.Equal(t, "Boolean Enabled", booleanFlag.Name)
					assert.Equal(t, "This is an enabled boolean test flag", booleanFlag.Description)
					assert.True(t, booleanFlag.Enabled, "Flag should be enabled")

					revision = booleanEnabled.Revision

					fetchedBooleanEnabled, err := envClient.GetResource(ctx, &environments.GetResourceRequest{
						Environment: integration.DefaultEnvironment,
						Namespace:   namespace.Key,
						Key:         "boolean_enabled",
						TypeUrl:     "flipt.core.Flag",
					})
					require.NoError(t, err)

					assert.Equal(t, booleanEnabled, fetchedBooleanEnabled)

					resp, err := evalClient.Boolean(ctx, &evaluation.EvaluationRequest{
						NamespaceKey: namespace.Key,
						FlagKey:      "boolean_enabled",
					})
					require.NoError(t, err)

					assert.Equal(t, true, resp.Enabled)
				}

				{
					t.Log("Create a new flag in a disabled state.")

					disabledBooleanFlagAny, err := anypb.New(&core.Flag{
						Type:        core.FlagType_BOOLEAN_FLAG_TYPE,
						Key:         "boolean_disabled",
						Name:        "Boolean Disabled",
						Description: "This is a disabled boolean test flag",
						Enabled:     false,
					})
					require.NoError(t, err)

					stripAnyTypePrefix(t, disabledBooleanFlagAny)

					booleanDisabled, err := envClient.CreateResource(ctx, &environments.UpdateResourceRequest{
						Environment: integration.DefaultEnvironment,
						Key:         "boolean_disabled",
						Namespace:   namespace.Key,
						Payload:     disabledBooleanFlagAny,
						Revision:    revision,
					})
					require.NoError(t, err)

					var booleanDisabledFlag core.Flag
					err = anypb.UnmarshalTo(booleanDisabled.Resource.Payload, &booleanDisabledFlag, proto.UnmarshalOptions{})
					require.NoError(t, err)

					assert.Equal(t, "boolean_disabled", booleanDisabledFlag.Key)
					assert.Equal(t, "Boolean Disabled", booleanDisabledFlag.Name)
					assert.Equal(t, "This is a disabled boolean test flag", booleanDisabledFlag.Description)
					assert.False(t, booleanDisabledFlag.Enabled, "Flag should be disabled")

					revision = booleanDisabled.Revision

					fetchedBooleanDisabled, err := envClient.GetResource(ctx, &environments.GetResourceRequest{
						Environment: integration.DefaultEnvironment,
						Namespace:   namespace.Key,
						Key:         "boolean_disabled",
						TypeUrl:     "flipt.core.Flag",
					})
					require.NoError(t, err)

					assert.Equal(t, booleanDisabled, fetchedBooleanDisabled)
				}

				{
					t.Log("Retrieve flag with key \"test\".")

					found, err := envClient.GetResource(ctx, &environments.GetResourceRequest{
						Environment: integration.DefaultEnvironment,
						Namespace:   namespace.Key,
						Key:         "test",
						TypeUrl:     "flipt.core.Flag",
					})
					require.NoError(t, err)

					var foundFlag core.Flag
					err = anypb.UnmarshalTo(found.Resource.Payload, &foundFlag, proto.UnmarshalOptions{})
					require.NoError(t, err)

					// expect to find our original enabled flag response
					// but with the latest revision being reported
					assert.Equal(t, revision, found.Revision)
				}
			})
		})
	}
}

func namespaceIsDefault(ns string) bool {
	return ns == "" || ns == "default"
}

func stripAnyTypePrefix(t *testing.T, a *anypb.Any) {
	t.Helper()
	a.TypeUrl = strings.TrimPrefix(a.TypeUrl, "type.googleapis.com/")
}
