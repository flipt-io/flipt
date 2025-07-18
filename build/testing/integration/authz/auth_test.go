package authz

import (
	"context"
	"fmt"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/build/testing/integration"
	"go.flipt.io/flipt/rpc/flipt"
	"go.flipt.io/flipt/rpc/flipt/core"
	"go.flipt.io/flipt/rpc/flipt/evaluation"
	"go.flipt.io/flipt/rpc/v2/environments"
	sdkv2 "go.flipt.io/flipt/sdk/go/v2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/anypb"
)

func TestAuthz(t *testing.T) {
	integration.Harness(t, func(t *testing.T, opts integration.TestOpts) {
		client := opts.BootstrapClient(t)

		t.Run("Evaluation", func(t *testing.T) {
			ctx := context.Background()

			t.Run("Boolean", func(t *testing.T) {
				_, err := client.Evaluation().Boolean(ctx, &evaluation.EvaluationRequest{
					FlagKey: "flag_boolean",
					Context: map[string]string{
						"in_segment": "segment_001",
					},
				})
				require.NoError(t, err)
			})

			t.Run("Variant", func(t *testing.T) {
				_, err := client.Evaluation().Variant(ctx, &evaluation.EvaluationRequest{
					FlagKey: "flag_001",
					Context: map[string]string{
						"in_segment": "segment_001",
					},
				})
				require.NoError(t, err)
			})
		})

		t.Run("Authentication Methods", func(t *testing.T) {
			ctx := context.Background()

			t.Run("List methods", func(t *testing.T) {
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
					for _, test := range []struct {
						name   string
						client func(*testing.T, ...integration.ClientOpt) sdkv2.SDK
					}{
						{"StaticToken", opts.TokenClientV2},
						{"JWT", opts.JWTClientV2},
					} {
						t.Run(test.name, func(t *testing.T) {
							t.Run("NoRole", func(t *testing.T) {
								// ensure we cannot do any read specific operations across namespaces
								// with a token with no role
								client := test.client(t)
								cannotReadAnyIn(t, ctx, client, namespace.Key)
								cannotWriteNamespaces(t, ctx, client)
								cannotWriteNamespacedIn(t, ctx, client, namespace.Key)
							})

							t.Run("Admin", func(t *testing.T) {
								// ensure admin can do all the things
								client := test.client(t, integration.WithRole("admin"))
								canReadAllIn(t, ctx, client, namespace.Key)
								canWriteNamespaces(t, ctx, client)
								canWriteNamespacedIn(t, ctx, client, namespace.Key)
							})

							t.Run("Editor", func(t *testing.T) {
								client := test.client(t, integration.WithRole("editor"))
								// can read everywhere (including namespaces)
								canReadAllIn(t, ctx, client, namespace.Key)
								// cannot write namespaces
								cannotWriteNamespaces(t, ctx, client)
								// but can write in namespaces
								canWriteNamespacedIn(t, ctx, client, namespace.Key)
							})

							t.Run("Viewer", func(t *testing.T) {
								client := test.client(t, integration.WithRole("viewer"))
								// can read everywhere (including namespaces)
								canReadAllIn(t, ctx, client, namespace.Key)
								// cannot write namespaces
								cannotWriteNamespaces(t, ctx, client)
								// cannot write in namespaces either
								cannotWriteNamespacedIn(t, ctx, client, namespace.Key)
							})

							t.Run("NamespacedViewer", func(t *testing.T) {
								// ensure we cannot do read specific operations across other namespaces
								// with the namespaced viewer role token
								client := test.client(t, integration.WithRole(fmt.Sprintf("%s_viewer", namespace.Expected)))
								// can read in designated namespace
								canReadAllIn(t, ctx, client, namespace.Key)
								// cannot read in other namespace
								cannotReadAnyIn(t, ctx, client, integration.Namespaces.OtherKeyFrom(namespace.Key))
								// cannot write namespaces
								cannotWriteNamespaces(t, ctx, client)
								// cannot write in namespaces either
								cannotWriteNamespacedIn(t, ctx, client, namespace.Key)
							})
						})
					}
				})
			}
		})
	})
}

func canReadAllIn(t *testing.T, ctx context.Context, client sdkv2.SDK, namespace string) {
	t.Run("CanReadAll", func(t *testing.T) {
		clientCallSet{
			can(ListFlags(&flipt.ListFlagRequest{
				EnvironmentKey: "default",
				NamespaceKey:   namespace,
			})),
		}.assert(t, ctx, client)
	})
}

func canWriteNamespaces(t *testing.T, ctx context.Context, client sdkv2.SDK) {
	t.Run("CanWriteNamespaces", func(t *testing.T) {
		namespace := fmt.Sprintf("%x", rand.Int63())
		clientCallSet{
			can(CreateNamespace(&environments.UpdateNamespaceRequest{
				EnvironmentKey: "default",
				Key:            namespace,
			})),
		}.assert(t, ctx, client)
	})
}

func cannotWriteNamespaces(t *testing.T, ctx context.Context, client sdkv2.SDK) {
	t.Run("CannotWriteNamespaces", func(t *testing.T) {
		namespace := fmt.Sprintf("%x", rand.Int63())
		clientCallSet{
			cannot(CreateNamespace(&environments.UpdateNamespaceRequest{
				EnvironmentKey: "default",
				Key:            namespace,
			})),
		}.assert(t, ctx, client)
	})
}

func canWriteNamespacedIn(t *testing.T, ctx context.Context, client sdkv2.SDK, namespace string) {
	t.Run("CanWriteNamespacedIn", func(t *testing.T) {
		flagKey := fmt.Sprintf("%x", rand.Int63())
		flagPayload := &core.Flag{
			Key:         flagKey,
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
		}

		flag, err := anypb.New(flagPayload)
		require.NoError(t, err)

		clientCallSet{
			can(CreateResource(&environments.UpdateResourceRequest{
				EnvironmentKey: "default",
				NamespaceKey:   namespace,
				Key:            flagKey,
				Payload:        flag,
			})),
		}.assert(t, ctx, client)
	})
}

func cannotReadAnyIn(t *testing.T, ctx context.Context, client sdkv2.SDK, namespace string) {
	t.Run("CannotReadAny", func(t *testing.T) {
		clientCallSet{
			cannot(ListFlags(&flipt.ListFlagRequest{
				EnvironmentKey: "default",
				NamespaceKey:   namespace,
			})),
		}.assert(t, ctx, client)
	})
}

func cannotWriteNamespacedIn(t *testing.T, ctx context.Context, client sdkv2.SDK, namespace string) {
	t.Run("CannotWriteNamespacedIn", func(t *testing.T) {
		flagKey := fmt.Sprintf("%x", rand.Int63())
		flagPayload := &core.Flag{
			Key:         flagKey,
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
		}

		flag, err := anypb.New(flagPayload)
		require.NoError(t, err)

		clientCallSet{
			cannot(CreateResource(&environments.UpdateResourceRequest{
				EnvironmentKey: "default",
				NamespaceKey:   namespace,
				Key:            flagKey,
				Payload:        flag,
			})),
		}.assert(t, ctx, client)
	})
}

type clientCallSet []isAuthorized

type isAuthorized struct {
	call       clientCall
	authorized bool
}

func can(c clientCall) isAuthorized    { return isAuthorized{c, true} }
func cannot(c clientCall) isAuthorized { return isAuthorized{c, false} }

func (s clientCallSet) assert(t *testing.T, ctx context.Context, client sdkv2.SDK) {
	for _, c := range s {
		assertIsAuthorized(t, c.call(t, ctx, client), c.authorized)
	}
}

type clientCall func(*testing.T, context.Context, sdkv2.SDK) error

func ListFlags(in *flipt.ListFlagRequest) clientCall {
	return func(t *testing.T, ctx context.Context, s sdkv2.SDK) error {
		_, err := s.Environments().ListResources(ctx, &environments.ListResourcesRequest{
			EnvironmentKey: in.EnvironmentKey,
			NamespaceKey:   in.NamespaceKey,
			TypeUrl:        "flipt.core.Flag",
		})
		if err != nil {
			return fmt.Errorf("ListFlags: %w", err)
		}

		return nil
	}
}

func CreateNamespace(in *environments.UpdateNamespaceRequest) clientCall {
	return func(t *testing.T, ctx context.Context, s sdkv2.SDK) error {
		_, err := s.Environments().CreateNamespace(ctx, in)
		if err != nil {
			return fmt.Errorf("CreateNamespace: %w", err)
		}

		return nil
	}
}

func CreateResource(in *environments.UpdateResourceRequest) clientCall {
	return func(t *testing.T, ctx context.Context, s sdkv2.SDK) error {
		_, err := s.Environments().CreateResource(ctx, in)
		if err != nil {
			return fmt.Errorf("CreateResource: %w", err)
		}

		return nil
	}
}

func assertIsAuthorized(t *testing.T, err error, authorized bool) {
	t.Helper()
	if !authorized {
		assert.Equal(t, codes.PermissionDenied, status.Code(err), err)
		return
	}

	if err != nil {
		code := status.Code(err)
		assert.NotEqual(t, codes.Unauthenticated, code, err)
		assert.NotEqual(t, codes.PermissionDenied, code, err)
	}
}
