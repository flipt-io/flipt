package authz

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/build/testing/integration"
	"go.flipt.io/flipt/rpc/flipt"
	"go.flipt.io/flipt/rpc/flipt/auth"
	sdk "go.flipt.io/flipt/sdk/go"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func Common(t *testing.T, opts integration.TestOpts) {
	client := opts.TokenClient(t)

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
				t.Run("StaticToken", func(t *testing.T) {
					t.Run("NoRole", func(t *testing.T) {
						// ensure we can do resource specific operations across namespaces
						// with the admin token
						client := opts.TokenClient(t)
						cannotReadAnyIn(t, ctx, client, namespace.Key)
					})

					t.Run("Admin", func(t *testing.T) {
						// ensure we can do resource specific operations across namespaces
						// with the admin token
						client := opts.TokenClient(t, integration.WithRole("admin"))
						canReadAllIn(t, ctx, client, namespace.Key)
					})

					t.Run("Editor", func(t *testing.T) {
						// ensure we can do resource specific operations across namespaces
						// with the admin token
						client := opts.TokenClient(t, integration.WithRole("editor"))
						canReadAllIn(t, ctx, client, namespace.Key)
					})

					t.Run("Viewer", func(t *testing.T) {
						// ensure we can do resource specific operations across namespaces
						// with the admin token
						client := opts.TokenClient(t, integration.WithRole("viewer"))
						canReadAllIn(t, ctx, client, namespace.Key)
					})

					t.Run("NamespacedViewer", func(t *testing.T) {
						// ensure we can do resource specific operations across namespaces
						// with the admin token
						client := opts.TokenClient(t, integration.WithRole(fmt.Sprintf("%s_viewer", namespace.Expected)))
						cannotReadAnyIn(t, ctx, client, integration.Namespaces.OtherNamespaceFrom(namespace.Expected))
					})
				})
			})
		}

		t.Run("Expire Self", func(t *testing.T) {
			err := client.Auth().AuthenticationService().ExpireAuthenticationSelf(ctx, &auth.ExpireAuthenticationSelfRequest{
				ExpiresAt: flipt.Now(),
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

func listFlagIsAllowed(t *testing.T, ctx context.Context, client sdk.SDK, namespace string) {
	t.Helper()

	t.Run(fmt.Sprintf("ListFlags(namespace: %q)", namespace), func(t *testing.T) {
		// construct a new client using the previously obtained client token
		resp, err := client.Flipt().ListFlags(ctx, &flipt.ListFlagRequest{
			NamespaceKey: namespace,
		})
		require.NoError(t, err)
		require.NotEmpty(t, resp.Flags)
	})
}

func canReadAllIn(t *testing.T, ctx context.Context, client sdk.SDK, namespace string) {
	t.Run("CanReadAll", func(t *testing.T) {
		clientCallSet{
			can(GetNamespace(&flipt.GetNamespaceRequest{Key: namespace})),
			can(GetFlag(&flipt.GetFlagRequest{NamespaceKey: namespace, Key: "flag"})),
			can(ListFlags(&flipt.ListFlagRequest{NamespaceKey: namespace})),
			can(GetRule(&flipt.GetRuleRequest{NamespaceKey: namespace, FlagKey: "flag", Id: "id"})),
			can(ListRules(&flipt.ListRuleRequest{NamespaceKey: namespace, FlagKey: "flag"})),
			can(GetRollout(&flipt.GetRolloutRequest{NamespaceKey: namespace, FlagKey: "flag"})),
			can(ListRollouts(&flipt.ListRolloutRequest{NamespaceKey: namespace, FlagKey: "flag"})),
			can(GetSegment(&flipt.GetSegmentRequest{NamespaceKey: namespace, Key: "segment"})),
			can(ListSegments(&flipt.ListSegmentRequest{NamespaceKey: namespace})),
		}.assert(t, ctx, client)
	})
}

func cannotReadAnyIn(t *testing.T, ctx context.Context, client sdk.SDK, namespace string) {
	t.Run("CannotReadAny", func(t *testing.T) {
		clientCallSet{
			cannot(GetNamespace(&flipt.GetNamespaceRequest{Key: namespace})),
			cannot(GetFlag(&flipt.GetFlagRequest{NamespaceKey: namespace, Key: "flag"})),
			cannot(ListFlags(&flipt.ListFlagRequest{NamespaceKey: namespace})),
			cannot(GetRule(&flipt.GetRuleRequest{NamespaceKey: namespace, FlagKey: "flag", Id: "id"})),
			cannot(ListRules(&flipt.ListRuleRequest{NamespaceKey: namespace, FlagKey: "flag"})),
			cannot(GetRollout(&flipt.GetRolloutRequest{NamespaceKey: namespace, FlagKey: "flag"})),
			cannot(ListRollouts(&flipt.ListRolloutRequest{NamespaceKey: namespace, FlagKey: "flag"})),
			cannot(GetSegment(&flipt.GetSegmentRequest{NamespaceKey: namespace, Key: "segment"})),
			cannot(ListSegments(&flipt.ListSegmentRequest{NamespaceKey: namespace})),
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

func (s clientCallSet) assert(t *testing.T, ctx context.Context, client sdk.SDK) {
	for _, c := range s {
		assertIsAuthorized(t, c.call(t, ctx, client), c.authorized)
	}
}

type clientCall func(*testing.T, context.Context, sdk.SDK) error

func GetNamespace(in *flipt.GetNamespaceRequest) clientCall {
	return func(t *testing.T, ctx context.Context, s sdk.SDK) error {
		_, err := s.Flipt().GetNamespace(ctx, in)
		return err
	}
}

func ListNamespaces(in *flipt.ListNamespaceRequest) clientCall {
	return func(t *testing.T, ctx context.Context, s sdk.SDK) error {
		_, err := s.Flipt().ListNamespaces(ctx, in)
		return err
	}
}

func GetFlag(in *flipt.GetFlagRequest) clientCall {
	return func(t *testing.T, ctx context.Context, s sdk.SDK) error {
		_, err := s.Flipt().GetFlag(ctx, in)
		return err
	}
}

func ListFlags(in *flipt.ListFlagRequest) clientCall {
	return func(t *testing.T, ctx context.Context, s sdk.SDK) error {
		_, err := s.Flipt().ListFlags(ctx, in)
		return err
	}
}

func GetRule(in *flipt.GetRuleRequest) clientCall {
	return func(t *testing.T, ctx context.Context, s sdk.SDK) error {
		_, err := s.Flipt().GetRule(ctx, in)
		return err
	}
}

func ListRules(in *flipt.ListRuleRequest) clientCall {
	return func(t *testing.T, ctx context.Context, s sdk.SDK) error {
		_, err := s.Flipt().ListRules(ctx, in)
		return err
	}
}

func GetRollout(in *flipt.GetRolloutRequest) clientCall {
	return func(t *testing.T, ctx context.Context, s sdk.SDK) error {
		_, err := s.Flipt().GetRollout(ctx, in)
		return err
	}
}

func ListRollouts(in *flipt.ListRolloutRequest) clientCall {
	return func(t *testing.T, ctx context.Context, s sdk.SDK) error {
		_, err := s.Flipt().ListRollouts(ctx, in)
		return err
	}
}

func GetSegment(in *flipt.GetSegmentRequest) clientCall {
	return func(t *testing.T, ctx context.Context, s sdk.SDK) error {
		_, err := s.Flipt().GetSegment(ctx, in)
		return err
	}
}

func ListSegments(in *flipt.ListSegmentRequest) clientCall {
	return func(t *testing.T, ctx context.Context, s sdk.SDK) error {
		_, err := s.Flipt().ListSegments(ctx, in)
		return err
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
