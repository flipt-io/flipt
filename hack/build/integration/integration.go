package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/rpc/flipt"
	"go.flipt.io/flipt/rpc/flipt/auth"
	sdk "go.flipt.io/flipt/sdk/go"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func Core(t *testing.T, fn func(t *testing.T) (sdk.SDK, string)) {
	t.Run("Core Suite", func(t *testing.T) {
		client, namespace := fn(t)

		ctx := context.Background()

		t.Run("Namespaces", func(t *testing.T) {
			if namespace != "" {
				t.Log(`Create namespace.`)

				created, err := client.Flipt().CreateNamespace(ctx, &flipt.CreateNamespaceRequest{
					Key:  namespace,
					Name: namespace,
				})
				require.NoError(t, err)

				assert.Equal(t, namespace, created.Key)
				assert.Equal(t, namespace, created.Name)

				t.Log(`Get namespace by key.`)

				retrieved, err := client.Flipt().GetNamespace(ctx, &flipt.GetNamespaceRequest{
					Key: namespace,
				})

				assert.Equal(t, created.Name, retrieved.Name)

				t.Log(`List namespaces.`)

				namespaces, err := client.Flipt().ListNamespaces(ctx, &flipt.ListNamespaceRequest{})
				require.NoError(t, err)

				assert.Len(t, namespaces.Namespaces, 2)

				assert.Equal(t, "default", namespaces.Namespaces[0].Key)
				assert.Equal(t, namespace, namespaces.Namespaces[1].Key)

				t.Log(`Update namespace.`)

				updated, err := client.Flipt().UpdateNamespace(ctx, &flipt.UpdateNamespaceRequest{
					Key:         namespace,
					Name:        namespace,
					Description: "Some kind of description",
				})
				require.NoError(t, err)

				assert.Equal(t, "Some kind of description", updated.Description)
			} else {
				t.Log(`Ensure default cannot be created.`)

				_, err := client.Flipt().CreateNamespace(ctx, &flipt.CreateNamespaceRequest{
					Key:  "default",
					Name: "Default",
				})
				require.EqualError(t, err, "rpc error: code = InvalidArgument desc = namespace \"default\" is not unique")

				t.Log(`Ensure default can be updated.`)

				updated, err := client.Flipt().UpdateNamespace(ctx, &flipt.UpdateNamespaceRequest{
					Key:         "default",
					Name:        "Default",
					Description: "Some namespace description.",
				})
				require.NoError(t, err)
				assert.Equal(t, "Default", updated.Name)

				t.Log(`Ensure default cannot be deleted.`)

				err = client.Flipt().DeleteNamespace(ctx, &flipt.DeleteNamespaceRequest{
					Key: "default",
				})
				require.EqualError(t, err, "rpc error: code = InvalidArgument desc = namespace \"default\" is protected")
			}
		})

		t.Run("Flags and Variants", func(t *testing.T) {
			t.Log("Create a new flag with key \"test\".")

			created, err := client.Flipt().CreateFlag(ctx, &flipt.CreateFlagRequest{
				NamespaceKey: namespace,
				Key:          "test",
				Name:         "Test",
				Description:  "This is a test flag",
				Enabled:      true,
			})
			require.NoError(t, err)

			assert.Equal(t, "test", created.Key)
			assert.Equal(t, "Test", created.Name)
			assert.Equal(t, "This is a test flag", created.Description)
			assert.True(t, created.Enabled, "Flag should be enabled")

			t.Log("Retrieve flag with key \"test\".")

			flag, err := client.Flipt().GetFlag(ctx, &flipt.GetFlagRequest{
				NamespaceKey: namespace,
				Key:          "test",
			})
			require.NoError(t, err)

			assert.Equal(t, created, flag)

			t.Log("Update flag with key \"test\".")

			updated, err := client.Flipt().UpdateFlag(ctx, &flipt.UpdateFlagRequest{
				NamespaceKey: namespace,
				Key:          created.Key,
				Name:         "Test 2",
				Description:  created.Description,
				Enabled:      true,
			})
			require.NoError(t, err)

			assert.Equal(t, "Test 2", updated.Name)

			t.Log("List all flags.")

			flags, err := client.Flipt().ListFlags(ctx, &flipt.ListFlagRequest{
				NamespaceKey: namespace,
			})
			require.NoError(t, err)

			assert.Len(t, flags.Flags, 1)
			assert.Equal(t, updated.Key, flags.Flags[0].Key)
			assert.Equal(t, updated.Name, flags.Flags[0].Name)
			assert.Equal(t, updated.Description, flags.Flags[0].Description)

			for _, key := range []string{"one", "two"} {
				t.Logf("Create variant with key %q.", key)

				createdVariant, err := client.Flipt().CreateVariant(ctx, &flipt.CreateVariantRequest{
					NamespaceKey: namespace,
					Key:          key,
					Name:         key,
					FlagKey:      "test",
				})
				require.NoError(t, err)

				assert.Equal(t, key, createdVariant.Key)
				assert.Equal(t, key, createdVariant.Name)
			}

			t.Log("Get flag \"test\" and check variants.")

			flag, err = client.Flipt().GetFlag(ctx, &flipt.GetFlagRequest{
				NamespaceKey: namespace,
				Key:          "test",
			})
			require.NoError(t, err)

			assert.Len(t, flag.Variants, 2)

			t.Log(`Update variant "one" (rename from "one" to "One").`)

			updatedVariant, err := client.Flipt().UpdateVariant(ctx, &flipt.UpdateVariantRequest{
				NamespaceKey: namespace,
				FlagKey:      "test",
				Id:           flag.Variants[0].Id,
				Key:          "one",
				Name:         "One",
			})
			require.NoError(t, err)

			assert.Equal(t, "one", updatedVariant.Key)
			assert.Equal(t, "One", updatedVariant.Name)
		})

		t.Run("Segments and Constraints", func(t *testing.T) {
			t.Log(`Create segment "everyone".`)

			createdSegment, err := client.Flipt().CreateSegment(ctx, &flipt.CreateSegmentRequest{
				NamespaceKey: namespace,
				Key:          "everyone",
				Name:         "Everyone",
				MatchType:    flipt.MatchType_ALL_MATCH_TYPE,
			})
			require.NoError(t, err)

			assert.Equal(t, "everyone", createdSegment.Key)
			assert.Equal(t, "Everyone", createdSegment.Name)
			assert.Equal(t, flipt.MatchType_ALL_MATCH_TYPE, createdSegment.MatchType)

			t.Log(`Get segment "everyone".`)

			retrievedSegment, err := client.Flipt().GetSegment(ctx, &flipt.GetSegmentRequest{
				NamespaceKey: namespace,
				Key:          "everyone",
			})
			require.NoError(t, err)

			assert.Equal(t, createdSegment.Key, retrievedSegment.Key)
			assert.Equal(t, createdSegment.Name, retrievedSegment.Name)
			assert.Equal(t, createdSegment.Description, retrievedSegment.Description)
			assert.Equal(t, createdSegment.MatchType, createdSegment.MatchType)

			t.Log(`Update segment "everyone" (rename "Everyone" to "All the peeps").`)

			updatedSegment, err := client.Flipt().UpdateSegment(ctx, &flipt.UpdateSegmentRequest{
				NamespaceKey: namespace,
				Key:          "everyone",
				Name:         "All the peeps",
			})
			require.NoError(t, err)

			assert.Equal(t, "everyone", updatedSegment.Key)
			assert.Equal(t, "All the peeps", updatedSegment.Name)
			assert.Equal(t, flipt.MatchType_ALL_MATCH_TYPE, updatedSegment.MatchType)

			for _, constraint := range []struct {
				property, operator, value string
			}{
				{"foo", "eq", "bar"},
				{"fizz", "neq", "buzz"},
			} {
				t.Logf(`Create constraint with key %q %s %q.`,
					constraint.property,
					constraint.operator,
					constraint.value,
				)

				createdConstraint, err := client.Flipt().CreateConstraint(ctx, &flipt.CreateConstraintRequest{
					NamespaceKey: namespace,
					SegmentKey:   "everyone",
					Type:         flipt.ComparisonType_STRING_COMPARISON_TYPE,
					Property:     constraint.property,
					Operator:     constraint.operator,
					Value:        constraint.value,
				})
				require.NoError(t, err)

				assert.Equal(t, constraint.property, createdConstraint.Property)
				assert.Equal(t, constraint.operator, createdConstraint.Operator)
				assert.Equal(t, constraint.value, createdConstraint.Value)
			}

			t.Log(`Get segment "everyone" with constraints.`)

			retrievedSegment, err = client.Flipt().GetSegment(ctx, &flipt.GetSegmentRequest{
				NamespaceKey: namespace,
				Key:          "everyone",
			})
			require.NoError(t, err)

			assert.Len(t, retrievedSegment.Constraints, 2)
			assert.Equal(t, "foo", retrievedSegment.Constraints[0].Property)
			assert.Equal(t, "fizz", retrievedSegment.Constraints[1].Property)

			t.Log(`Update constraint value (from "bar" to "baz").`)
			updatedConstraint, err := client.Flipt().UpdateConstraint(ctx, &flipt.UpdateConstraintRequest{
				NamespaceKey: namespace,
				SegmentKey:   "everyone",
				Type:         retrievedSegment.Constraints[0].Type,
				Id:           retrievedSegment.Constraints[0].Id,
				Property:     retrievedSegment.Constraints[0].Property,
				Operator:     retrievedSegment.Constraints[0].Operator,
				Value:        "baz",
			})
			require.NoError(t, err)

			assert.Equal(t, flipt.ComparisonType_STRING_COMPARISON_TYPE, updatedConstraint.Type)
			assert.Equal(t, "foo", updatedConstraint.Property)
			assert.Equal(t, "eq", updatedConstraint.Operator)
			assert.Equal(t, "baz", updatedConstraint.Value)
		})

		t.Run("Rules and Distributions", func(t *testing.T) {
			t.Log(`Create rule "rank 1".`)

			ruleOne, err := client.Flipt().CreateRule(ctx, &flipt.CreateRuleRequest{
				NamespaceKey: namespace,
				FlagKey:      "test",
				SegmentKey:   "everyone",
				Rank:         1,
			})

			require.NoError(t, err)

			assert.Equal(t, "test", ruleOne.FlagKey)
			assert.Equal(t, "everyone", ruleOne.SegmentKey)
			assert.Equal(t, int32(1), ruleOne.Rank)

			t.Log(`Get rules "rank 1".`)

			retrievedRule, err := client.Flipt().GetRule(ctx, &flipt.GetRuleRequest{
				NamespaceKey: namespace,
				FlagKey:      "test",
				Id:           ruleOne.Id,
			})
			require.NoError(t, err)

			assert.Equal(t, "test", retrievedRule.FlagKey)
			assert.Equal(t, "everyone", retrievedRule.SegmentKey)
			assert.Equal(t, int32(1), retrievedRule.Rank)

			t.Log(`Create rule "rank 2".`)

			ruleTwo, err := client.Flipt().CreateRule(ctx, &flipt.CreateRuleRequest{
				NamespaceKey: namespace,
				FlagKey:      "test",
				SegmentKey:   "everyone",
				Rank:         2,
			})

			require.NoError(t, err)

			assert.Equal(t, "test", ruleTwo.FlagKey)
			assert.Equal(t, "everyone", ruleTwo.SegmentKey)
			assert.Equal(t, int32(2), ruleTwo.Rank)

			// ensure you can not link flags and segments from different namespaces.
			if namespace != "" {
				t.Log(`Ensure that rules can only link entities in the same namespace.`)
				_, err = client.Flipt().CreateRule(ctx, &flipt.CreateRuleRequest{
					FlagKey:    "test",
					SegmentKey: "everyone",
					Rank:       2,
				})

				msg := "rpc error: code = NotFound desc = flag \"default/test\" or segment \"default/everyone\" not found"
				require.EqualError(t, err, msg)
			}

			t.Log(`List rules.`)

			allRules, err := client.Flipt().ListRules(ctx, &flipt.ListRuleRequest{
				NamespaceKey: namespace,
				FlagKey:      "test",
			})
			require.NoError(t, err)

			assert.Len(t, allRules.Rules, 2)

			assert.Equal(t, ruleOne.Id, allRules.Rules[0].Id)
			assert.Equal(t, ruleTwo.Id, allRules.Rules[1].Id)

			t.Log(`Re-order rules.`)

			err = client.Flipt().OrderRules(ctx, &flipt.OrderRulesRequest{
				NamespaceKey: namespace,
				FlagKey:      "test",
				RuleIds:      []string{ruleTwo.Id, ruleOne.Id},
			})
			require.NoError(t, err)

			t.Log(`List rules again.`)

			allRules, err = client.Flipt().ListRules(ctx, &flipt.ListRuleRequest{
				NamespaceKey: namespace,
				FlagKey:      "test",
			})
			require.NoError(t, err)

			assert.Len(t, allRules.Rules, 2)

			// ensure the order has switched
			assert.Equal(t, ruleTwo.Id, allRules.Rules[0].Id)
			assert.Equal(t, int32(1), allRules.Rules[0].Rank)
			assert.Equal(t, ruleOne.Id, allRules.Rules[1].Id)
			assert.Equal(t, int32(2), allRules.Rules[1].Rank)

			t.Log(`Create distribution "rollout 100".`)

			flag, err := client.Flipt().GetFlag(ctx, &flipt.GetFlagRequest{
				NamespaceKey: namespace,
				Key:          "test",
			})
			require.NoError(t, err)

			distribution, err := client.Flipt().CreateDistribution(ctx, &flipt.CreateDistributionRequest{
				NamespaceKey: namespace,
				FlagKey:      "test",
				RuleId:       ruleTwo.Id,
				VariantId:    flag.Variants[0].Id,
				Rollout:      100,
			})
			require.NoError(t, err)

			if namespace != "" {
				t.Log(`Ensure that distributions and all entities associated with them are part of same namespace.`)
				var (
					flagKey    = "defaultflag"
					segmentKey = "defaultsegment"
				)

				defaultFlag, err := client.Flipt().CreateFlag(ctx, &flipt.CreateFlagRequest{
					Key:         flagKey,
					Name:        "DefaultFlag",
					Description: "default flag description",
					Enabled:     true,
				})
				require.NoError(t, err)

				defaultVariant, err := client.Flipt().CreateVariant(ctx, &flipt.CreateVariantRequest{
					Key:         "defaultvariant",
					Name:        "DefaultVariant",
					FlagKey:     flagKey,
					Description: "default variant description",
				})
				require.NoError(t, err)

				defaultSegment, err := client.Flipt().CreateSegment(ctx, &flipt.CreateSegmentRequest{
					Key:         segmentKey,
					Name:        "DefaultSegment",
					Description: "default segment description",
					MatchType:   flipt.MatchType_ANY_MATCH_TYPE,
				})
				require.NoError(t, err)

				defaultRule, err := client.Flipt().CreateRule(ctx, &flipt.CreateRuleRequest{
					FlagKey:    flagKey,
					SegmentKey: segmentKey,
					Rank:       1,
				})
				require.NoError(t, err)

				defaultDistribution, err := client.Flipt().CreateDistribution(ctx, &flipt.CreateDistributionRequest{
					FlagKey:   flagKey,
					RuleId:    defaultRule.Id,
					VariantId: defaultVariant.Id,
					Rollout:   100,
				})
				require.NoError(t, err)

				_, err = client.Flipt().CreateDistribution(ctx, &flipt.CreateDistributionRequest{
					NamespaceKey: namespace,
					FlagKey:      "test",
					RuleId:       ruleTwo.Id,
					VariantId:    defaultVariant.Id,
					Rollout:      100,
				})

				variantMsg := fmt.Sprintf("rpc error: code = NotFound desc = variant %q/%q not found", namespace, defaultVariant.Id)
				require.EqualError(t, err, variantMsg)

				_, err = client.Flipt().CreateDistribution(ctx, &flipt.CreateDistributionRequest{
					FlagKey:   "test",
					RuleId:    ruleTwo.Id,
					VariantId: defaultVariant.Id,
					Rollout:   100,
				})

				ruleMsg := fmt.Sprintf("rpc error: code = NotFound desc = rule %q/%q not found", "default", ruleTwo.Id)
				require.EqualError(t, err, ruleMsg)

				_, err = client.Flipt().UpdateDistribution(ctx, &flipt.UpdateDistributionRequest{
					NamespaceKey: namespace,
					Id:           distribution.Id,
					FlagKey:      "test",
					VariantId:    defaultVariant.Id,
					RuleId:       ruleTwo.Id,
					Rollout:      50,
				})
				require.EqualError(t, err, variantMsg)

				_, err = client.Flipt().UpdateDistribution(ctx, &flipt.UpdateDistributionRequest{
					Id:        defaultDistribution.Id,
					FlagKey:   flagKey,
					RuleId:    ruleTwo.Id,
					VariantId: defaultVariant.Id,
					Rollout:   50,
				})
				require.EqualError(t, err, ruleMsg)

				err = client.Flipt().DeleteFlag(ctx, &flipt.DeleteFlagRequest{
					Key: defaultFlag.Key,
				})
				require.NoError(t, err)

				err = client.Flipt().DeleteVariant(ctx, &flipt.DeleteVariantRequest{
					Id:      defaultVariant.Id,
					FlagKey: flagKey,
				})
				require.NoError(t, err)

				err = client.Flipt().DeleteSegment(ctx, &flipt.DeleteSegmentRequest{
					Key: defaultSegment.Key,
				})
				require.NoError(t, err)

				err = client.Flipt().DeleteRule(ctx, &flipt.DeleteRuleRequest{
					Id:      defaultRule.Id,
					FlagKey: defaultRule.FlagKey,
				})
				require.NoError(t, err)

				err = client.Flipt().DeleteDistribution(ctx, &flipt.DeleteDistributionRequest{
					Id:        defaultDistribution.Id,
					FlagKey:   defaultFlag.Key,
					RuleId:    defaultRule.Id,
					VariantId: defaultVariant.Id,
				})
				require.NoError(t, err)
			}

			assert.Equal(t, ruleTwo.Id, distribution.RuleId)
			assert.Equal(t, float32(100), distribution.Rollout)
		})

		t.Run("Evaluation", func(t *testing.T) {
			t.Log(`Successful match.`)

			result, err := client.Flipt().Evaluate(ctx, &flipt.EvaluationRequest{
				NamespaceKey: namespace,
				FlagKey:      "test",
				EntityId:     uuid.Must(uuid.NewV4()).String(),
				Context: map[string]string{
					"foo":  "baz",
					"fizz": "bozz",
				},
			})
			require.NoError(t, err)

			require.True(t, result.Match, "Evaluation should have matched.")
			assert.Equal(t, "everyone", result.SegmentKey)
			assert.Equal(t, "one", result.Value)
			assert.Equal(t, flipt.EvaluationReason_MATCH_EVALUATION_REASON, result.Reason)

			t.Log(`Unsuccessful match.`)

			result, err = client.Flipt().Evaluate(ctx, &flipt.EvaluationRequest{
				NamespaceKey: namespace,
				FlagKey:      "test",
				EntityId:     uuid.Must(uuid.NewV4()).String(),
				Context: map[string]string{
					"fizz": "buzz",
				},
			})
			require.NoError(t, err)

			assert.False(t, result.Match, "Evaluation should not have matched.")
		})

		t.Run("Batch Evaluation", func(t *testing.T) {
			t.Log(`Successful match.`)

			results, err := client.Flipt().BatchEvaluate(ctx, &flipt.BatchEvaluationRequest{
				NamespaceKey: namespace,
				Requests: []*flipt.EvaluationRequest{
					{
						NamespaceKey: namespace,
						FlagKey:      "test",
						EntityId:     uuid.Must(uuid.NewV4()).String(),
						Context: map[string]string{
							"foo":  "baz",
							"fizz": "bozz",
						},
					},
				},
			})
			require.NoError(t, err)

			require.Len(t, results.Responses, 1)
			result := results.Responses[0]

			require.True(t, result.Match, "Evaluation should have matched.")
			assert.Equal(t, "everyone", result.SegmentKey)
			assert.Equal(t, "one", result.Value)
			assert.Equal(t, flipt.EvaluationReason_MATCH_EVALUATION_REASON, result.Reason)

			t.Log(`Unsuccessful match.`)

			results, err = client.Flipt().BatchEvaluate(ctx, &flipt.BatchEvaluationRequest{
				NamespaceKey: namespace,
				Requests: []*flipt.EvaluationRequest{
					{
						NamespaceKey: namespace,
						FlagKey:      "test",
						EntityId:     uuid.Must(uuid.NewV4()).String(),
						Context: map[string]string{
							"fizz": "buzz",
						},
					},
				},
			})
			require.NoError(t, err)

			require.Len(t, results.Responses, 1)
			result = results.Responses[0]

			assert.False(t, result.Match, "Evaluation should not have matched.")
		})

		t.Run("Delete", func(t *testing.T) {
			if namespace != "" {
				t.Log(`Namespace with flags fails.`)
				err := client.Flipt().DeleteNamespace(ctx, &flipt.DeleteNamespaceRequest{
					Key: namespace,
				})

				msg := fmt.Sprintf("rpc error: code = InvalidArgument desc = namespace %q cannot be deleted; flags must be deleted first", namespace)
				require.EqualError(t, err, msg)
			}

			t.Log(`Rules.`)
			rules, err := client.Flipt().ListRules(ctx, &flipt.ListRuleRequest{
				NamespaceKey: namespace,
				FlagKey:      "test",
			})
			require.NoError(t, err)

			for _, rule := range rules.Rules {
				client.Flipt().DeleteRule(ctx, &flipt.DeleteRuleRequest{
					NamespaceKey: namespace,
					FlagKey:      "test",
					Id:           rule.Id,
				})
			}

			t.Log(`Flag.`)
			err = client.Flipt().DeleteFlag(ctx, &flipt.DeleteFlagRequest{
				NamespaceKey: namespace,
				Key:          "test",
			})
			require.NoError(t, err)

			t.Log(`Segment.`)
			err = client.Flipt().DeleteSegment(ctx, &flipt.DeleteSegmentRequest{
				NamespaceKey: namespace,
				Key:          "everyone",
			})
			require.NoError(t, err)

			if namespace != "" {
				t.Log(`Namespace.`)
				err = client.Flipt().DeleteNamespace(ctx, &flipt.DeleteNamespaceRequest{
					Key: namespace,
				})
				require.NoError(t, err)
			}
		})

		t.Run("Meta", func(t *testing.T) {
			t.Log(`Info()`)

			info, err := client.Meta().GetInfo(ctx)
			require.NoError(t, err)

			assert.Equal(t, "application/json", info.ContentType)

			var infoMap map[string]any
			require.NoError(t, json.Unmarshal(info.Data, &infoMap))

			version, ok := infoMap["version"]
			assert.True(t, ok, "Missing version.")
			assert.NotEmpty(t, version)

			goVersion, ok := infoMap["goVersion"]
			assert.True(t, ok, "Missing Go version.")
			assert.NotEmpty(t, goVersion)

			t.Log(`Config()`)

			config, err := client.Meta().GetConfiguration(ctx)
			require.NoError(t, err)

			assert.Equal(t, "application/json", info.ContentType)

			var configMap map[string]any
			require.NoError(t, json.Unmarshal(config.Data, &configMap))

			for _, name := range []string{
				"log",
				"ui",
				"cache",
				"server",
				"db",
			} {
				field, ok := configMap[name]
				assert.True(t, ok, "Missing %s.", name)
				assert.NotEmpty(t, field)
			}
		})
	})
}

func Authenticated(t *testing.T, fn func(t *testing.T) (sdk.SDK, string)) {
	t.Run("Authentication Methods", func(t *testing.T) {
		client, _ := fn(t)

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
