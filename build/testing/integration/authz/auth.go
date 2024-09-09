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
	"go.flipt.io/flipt/rpc/flipt/auth"
	"go.flipt.io/flipt/rpc/flipt/evaluation"
	sdk "go.flipt.io/flipt/sdk/go"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func Common(t *testing.T, opts integration.TestOpts) {
	client := opts.TokenClient(t)

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
				for _, test := range []struct {
					name   string
					client func(*testing.T, ...integration.ClientOpt) sdk.SDK
				}{
					{"StaticToken", opts.TokenClient},
					{"JWT", opts.JWTClient},
					{"K8s", opts.K8sClient},
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
							cannotReadAnyIn(t, ctx, client, integration.Namespaces.OtherNamespaceFrom(namespace.Expected))
							// cannot write namespaces
							cannotWriteNamespaces(t, ctx, client)
							// cannot write in namespaces either
							cannotWriteNamespacedIn(t, ctx, client, namespace.Key)
						})
					})
				}
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

func canWriteNamespaces(t *testing.T, ctx context.Context, client sdk.SDK) {
	t.Run("CanWriteNamespaces", func(t *testing.T) {
		namespace := fmt.Sprintf("%x", rand.Int63())
		clientCallSet{
			can(CreateNamespace(&flipt.CreateNamespaceRequest{Key: namespace, Name: namespace})),
			can(UpdateNamespace(&flipt.UpdateNamespaceRequest{Key: namespace, Name: namespace})),
			can(DeleteNamespace(&flipt.DeleteNamespaceRequest{Key: namespace})),
		}.assert(t, ctx, client)
	})
}

func cannotWriteNamespaces(t *testing.T, ctx context.Context, client sdk.SDK) {
	t.Run("CannotWriteNamespaces", func(t *testing.T) {
		namespace := fmt.Sprintf("%x", rand.Int63())
		clientCallSet{
			cannot(CreateNamespace(&flipt.CreateNamespaceRequest{Key: namespace, Name: namespace})),
			cannot(UpdateNamespace(&flipt.UpdateNamespaceRequest{Key: namespace, Name: namespace})),
			cannot(DeleteNamespace(&flipt.DeleteNamespaceRequest{Key: namespace})),
		}.assert(t, ctx, client)
	})
}

func canWriteNamespacedIn(t *testing.T, ctx context.Context, client sdk.SDK, namespace string) {
	t.Run("CanWriteNamespacedIn", func(t *testing.T) {
		flag := fmt.Sprintf("%x", rand.Int63())
		segment := fmt.Sprintf("%x", rand.Int63())
		clientCallSet{
			can(CreateFlag(&flipt.CreateFlagRequest{NamespaceKey: namespace, Key: flag})),
			can(UpdateFlag(&flipt.UpdateFlagRequest{NamespaceKey: namespace, Key: flag})),
			can(CreateVariant(&flipt.CreateVariantRequest{NamespaceKey: namespace, FlagKey: flag, Key: flag})),
			can(UpdateVariant(&flipt.UpdateVariantRequest{NamespaceKey: namespace, Id: "abcdef"})),
			can(CreateSegment(&flipt.CreateSegmentRequest{NamespaceKey: namespace, Key: segment})),
			can(UpdateSegment(&flipt.UpdateSegmentRequest{NamespaceKey: namespace, Key: segment})),
			can(CreateConstraint(&flipt.CreateConstraintRequest{NamespaceKey: namespace, SegmentKey: segment})),
			can(UpdateConstraint(&flipt.UpdateConstraintRequest{NamespaceKey: namespace, Id: "abcdef"})),
			can(CreateRule(&flipt.CreateRuleRequest{NamespaceKey: namespace, FlagKey: flag, SegmentKey: segment})),
			can(UpdateRule(&flipt.UpdateRuleRequest{NamespaceKey: namespace, Id: "abcdef"})),
			can(OrderRules(&flipt.OrderRulesRequest{NamespaceKey: namespace, RuleIds: []string{"acdef"}})),
			can(CreateRollout(&flipt.CreateRolloutRequest{NamespaceKey: namespace, FlagKey: flag})),
			can(UpdateRollout(&flipt.UpdateRolloutRequest{NamespaceKey: namespace, Id: "abcdef"})),
			can(OrderRollouts(&flipt.OrderRolloutsRequest{NamespaceKey: namespace, RolloutIds: []string{"acdef"}})),
			// deletes
			can(DeleteFlag(&flipt.DeleteFlagRequest{NamespaceKey: namespace, Key: flag})),
			can(DeleteVariant(&flipt.DeleteVariantRequest{NamespaceKey: namespace, Id: "abcdef"})),
			can(DeleteSegment(&flipt.DeleteSegmentRequest{NamespaceKey: namespace, Key: segment})),
			can(DeleteConstraint(&flipt.DeleteConstraintRequest{NamespaceKey: namespace, Id: "abcdef"})),
			can(DeleteRule(&flipt.DeleteRuleRequest{NamespaceKey: namespace, Id: "abcdef"})),
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

func cannotWriteNamespacedIn(t *testing.T, ctx context.Context, client sdk.SDK, namespace string) {
	t.Run("CannotWriteNamespacedIn", func(t *testing.T) {
		flag := fmt.Sprintf("%x", rand.Int63())
		segment := fmt.Sprintf("%x", rand.Int63())
		clientCallSet{
			cannot(CreateFlag(&flipt.CreateFlagRequest{NamespaceKey: namespace, Key: flag})),
			cannot(UpdateFlag(&flipt.UpdateFlagRequest{NamespaceKey: namespace, Key: flag})),
			cannot(CreateVariant(&flipt.CreateVariantRequest{NamespaceKey: namespace, FlagKey: flag, Key: flag})),
			cannot(UpdateVariant(&flipt.UpdateVariantRequest{NamespaceKey: namespace, Id: "abcdef"})),
			cannot(CreateSegment(&flipt.CreateSegmentRequest{NamespaceKey: namespace, Key: segment})),
			cannot(UpdateSegment(&flipt.UpdateSegmentRequest{NamespaceKey: namespace, Key: segment})),
			cannot(CreateConstraint(&flipt.CreateConstraintRequest{NamespaceKey: namespace, SegmentKey: segment})),
			cannot(UpdateConstraint(&flipt.UpdateConstraintRequest{NamespaceKey: namespace, Id: "abcdef"})),
			cannot(CreateRule(&flipt.CreateRuleRequest{NamespaceKey: namespace, FlagKey: flag, SegmentKey: segment})),
			cannot(UpdateRule(&flipt.UpdateRuleRequest{NamespaceKey: namespace, Id: "abcdef"})),
			cannot(OrderRules(&flipt.OrderRulesRequest{NamespaceKey: namespace, RuleIds: []string{"acdef"}})),
			cannot(CreateRollout(&flipt.CreateRolloutRequest{NamespaceKey: namespace, FlagKey: flag})),
			cannot(UpdateRollout(&flipt.UpdateRolloutRequest{NamespaceKey: namespace, Id: "abcdef"})),
			cannot(OrderRollouts(&flipt.OrderRolloutsRequest{NamespaceKey: namespace, RolloutIds: []string{"acdef"}})),
			// deletes
			cannot(DeleteFlag(&flipt.DeleteFlagRequest{NamespaceKey: namespace, Key: flag})),
			cannot(DeleteVariant(&flipt.DeleteVariantRequest{NamespaceKey: namespace, Id: "abcdef"})),
			cannot(DeleteSegment(&flipt.DeleteSegmentRequest{NamespaceKey: namespace, Key: segment})),
			cannot(DeleteConstraint(&flipt.DeleteConstraintRequest{NamespaceKey: namespace, Id: "abcdef"})),
			cannot(DeleteRule(&flipt.DeleteRuleRequest{NamespaceKey: namespace, Id: "abcdef"})),
			cannot(DeleteRollout(&flipt.DeleteRolloutRequest{NamespaceKey: namespace, Id: "abcdef"})),
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
		return fmt.Errorf("GetNamespace: %w", err)
	}
}

func ListNamespaces(in *flipt.ListNamespaceRequest) clientCall {
	return func(t *testing.T, ctx context.Context, s sdk.SDK) error {
		_, err := s.Flipt().ListNamespaces(ctx, in)
		return fmt.Errorf("ListNamespaces: %w", err)
	}
}

func CreateNamespace(in *flipt.CreateNamespaceRequest) clientCall {
	return func(t *testing.T, ctx context.Context, s sdk.SDK) error {
		_, err := s.Flipt().CreateNamespace(ctx, in)
		return fmt.Errorf("CreateNamespace: %w", err)
	}
}

func UpdateNamespace(in *flipt.UpdateNamespaceRequest) clientCall {
	return func(t *testing.T, ctx context.Context, s sdk.SDK) error {
		_, err := s.Flipt().UpdateNamespace(ctx, in)
		return fmt.Errorf("UpdateNamespace: %w", err)
	}
}

func DeleteNamespace(in *flipt.DeleteNamespaceRequest) clientCall {
	return func(t *testing.T, ctx context.Context, s sdk.SDK) error {
		return fmt.Errorf("DeleteNamespace: %w", s.Flipt().DeleteNamespace(ctx, in))
	}
}

func GetFlag(in *flipt.GetFlagRequest) clientCall {
	return func(t *testing.T, ctx context.Context, s sdk.SDK) error {
		_, err := s.Flipt().GetFlag(ctx, in)
		return fmt.Errorf("GetFlag: %w", err)
	}
}

func ListFlags(in *flipt.ListFlagRequest) clientCall {
	return func(t *testing.T, ctx context.Context, s sdk.SDK) error {
		_, err := s.Flipt().ListFlags(ctx, in)
		return fmt.Errorf("ListFlags: %w", err)
	}
}

func CreateFlag(in *flipt.CreateFlagRequest) clientCall {
	return func(t *testing.T, ctx context.Context, s sdk.SDK) error {
		_, err := s.Flipt().CreateFlag(ctx, in)
		return fmt.Errorf("CreateFlag: %w", err)
	}
}

func UpdateFlag(in *flipt.UpdateFlagRequest) clientCall {
	return func(t *testing.T, ctx context.Context, s sdk.SDK) error {
		_, err := s.Flipt().UpdateFlag(ctx, in)
		return fmt.Errorf("UpdateFlag: %w", err)
	}
}

func DeleteFlag(in *flipt.DeleteFlagRequest) clientCall {
	return func(t *testing.T, ctx context.Context, s sdk.SDK) error {
		return fmt.Errorf("DeleteFlag: %w", s.Flipt().DeleteFlag(ctx, in))
	}
}

func CreateVariant(in *flipt.CreateVariantRequest) clientCall {
	return func(t *testing.T, ctx context.Context, s sdk.SDK) error {
		_, err := s.Flipt().CreateVariant(ctx, in)
		return fmt.Errorf("CreateVariant: %w", err)
	}
}

func UpdateVariant(in *flipt.UpdateVariantRequest) clientCall {
	return func(t *testing.T, ctx context.Context, s sdk.SDK) error {
		_, err := s.Flipt().UpdateVariant(ctx, in)
		return fmt.Errorf("UpdateVariant: %w", err)
	}
}

func DeleteVariant(in *flipt.DeleteVariantRequest) clientCall {
	return func(t *testing.T, ctx context.Context, s sdk.SDK) error {
		return fmt.Errorf("DeleteVariant: %w", s.Flipt().DeleteVariant(ctx, in))
	}
}

func GetRule(in *flipt.GetRuleRequest) clientCall {
	return func(t *testing.T, ctx context.Context, s sdk.SDK) error {
		_, err := s.Flipt().GetRule(ctx, in)
		return fmt.Errorf("GetRule: %w", err)
	}
}

func ListRules(in *flipt.ListRuleRequest) clientCall {
	return func(t *testing.T, ctx context.Context, s sdk.SDK) error {
		_, err := s.Flipt().ListRules(ctx, in)
		return fmt.Errorf("ListRules: %w", err)
	}
}

func CreateRule(in *flipt.CreateRuleRequest) clientCall {
	return func(t *testing.T, ctx context.Context, s sdk.SDK) error {
		_, err := s.Flipt().CreateRule(ctx, in)
		return fmt.Errorf("CreateRule: %w", err)
	}
}

func UpdateRule(in *flipt.UpdateRuleRequest) clientCall {
	return func(t *testing.T, ctx context.Context, s sdk.SDK) error {
		_, err := s.Flipt().UpdateRule(ctx, in)
		return fmt.Errorf("UpdateRule: %w", err)
	}
}

func DeleteRule(in *flipt.DeleteRuleRequest) clientCall {
	return func(t *testing.T, ctx context.Context, s sdk.SDK) error {
		return fmt.Errorf("DeleteRule: %w", s.Flipt().DeleteRule(ctx, in))
	}
}

func OrderRules(in *flipt.OrderRulesRequest) clientCall {
	return func(t *testing.T, ctx context.Context, s sdk.SDK) error {
		return fmt.Errorf("OrderRules: %w", s.Flipt().OrderRules(ctx, in))
	}
}

func CreateDistribution(in *flipt.CreateDistributionRequest) clientCall {
	return func(t *testing.T, ctx context.Context, s sdk.SDK) error {
		_, err := s.Flipt().CreateDistribution(ctx, in)
		return fmt.Errorf("CreateDistribution: %w", err)
	}
}

func UpdateDistribution(in *flipt.UpdateDistributionRequest) clientCall {
	return func(t *testing.T, ctx context.Context, s sdk.SDK) error {
		_, err := s.Flipt().UpdateDistribution(ctx, in)
		return fmt.Errorf("UpdateDistribution: %w", err)
	}
}

func DeleteDistribution(in *flipt.DeleteDistributionRequest) clientCall {
	return func(t *testing.T, ctx context.Context, s sdk.SDK) error {
		return fmt.Errorf("DeleteDistribution: %w", s.Flipt().DeleteDistribution(ctx, in))
	}
}

func CreateRollout(in *flipt.CreateRolloutRequest) clientCall {
	return func(t *testing.T, ctx context.Context, s sdk.SDK) error {
		_, err := s.Flipt().CreateRollout(ctx, in)
		return fmt.Errorf("CreateRollout: %w", err)
	}
}
func UpdateRollout(in *flipt.UpdateRolloutRequest) clientCall {
	return func(t *testing.T, ctx context.Context, s sdk.SDK) error {
		_, err := s.Flipt().UpdateRollout(ctx, in)
		return fmt.Errorf("UpdateRollout: %w", err)
	}
}

func DeleteRollout(in *flipt.DeleteRolloutRequest) clientCall {
	return func(t *testing.T, ctx context.Context, s sdk.SDK) error {
		return fmt.Errorf("DeleteRollout: %w", s.Flipt().DeleteRollout(ctx, in))
	}
}

func OrderRollouts(in *flipt.OrderRolloutsRequest) clientCall {
	return func(t *testing.T, ctx context.Context, s sdk.SDK) error {
		return fmt.Errorf("OrderRollouts: %w", s.Flipt().OrderRollouts(ctx, in))
	}
}

func GetRollout(in *flipt.GetRolloutRequest) clientCall {
	return func(t *testing.T, ctx context.Context, s sdk.SDK) error {
		_, err := s.Flipt().GetRollout(ctx, in)
		return fmt.Errorf("GetRollout: %w", err)
	}
}

func ListRollouts(in *flipt.ListRolloutRequest) clientCall {
	return func(t *testing.T, ctx context.Context, s sdk.SDK) error {
		_, err := s.Flipt().ListRollouts(ctx, in)
		return fmt.Errorf("ListRollouts: %w", err)
	}
}

func GetSegment(in *flipt.GetSegmentRequest) clientCall {
	return func(t *testing.T, ctx context.Context, s sdk.SDK) error {
		_, err := s.Flipt().GetSegment(ctx, in)
		return fmt.Errorf("GetSegment: %w", err)
	}
}

func ListSegments(in *flipt.ListSegmentRequest) clientCall {
	return func(t *testing.T, ctx context.Context, s sdk.SDK) error {
		_, err := s.Flipt().ListSegments(ctx, in)
		return fmt.Errorf("ListSegments: %w", err)
	}
}

func CreateSegment(in *flipt.CreateSegmentRequest) clientCall {
	return func(t *testing.T, ctx context.Context, s sdk.SDK) error {
		_, err := s.Flipt().CreateSegment(ctx, in)
		return fmt.Errorf("CreateSegment: %w", err)
	}
}

func UpdateSegment(in *flipt.UpdateSegmentRequest) clientCall {
	return func(t *testing.T, ctx context.Context, s sdk.SDK) error {
		_, err := s.Flipt().UpdateSegment(ctx, in)
		return fmt.Errorf("UpdateSegment: %w", err)
	}
}

func DeleteSegment(in *flipt.DeleteSegmentRequest) clientCall {
	return func(t *testing.T, ctx context.Context, s sdk.SDK) error {
		return fmt.Errorf("DeleteSegment: %w", s.Flipt().DeleteSegment(ctx, in))
	}
}

func CreateConstraint(in *flipt.CreateConstraintRequest) clientCall {
	return func(t *testing.T, ctx context.Context, s sdk.SDK) error {
		_, err := s.Flipt().CreateConstraint(ctx, in)
		return fmt.Errorf("CreateConstraint: %w", err)
	}
}

func UpdateConstraint(in *flipt.UpdateConstraintRequest) clientCall {
	return func(t *testing.T, ctx context.Context, s sdk.SDK) error {
		_, err := s.Flipt().UpdateConstraint(ctx, in)
		return fmt.Errorf("UpdateConstraint: %w", err)
	}
}

func DeleteConstraint(in *flipt.DeleteConstraintRequest) clientCall {
	return func(t *testing.T, ctx context.Context, s sdk.SDK) error {
		return fmt.Errorf("DeleteConstraint: %w", s.Flipt().DeleteConstraint(ctx, in))
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
