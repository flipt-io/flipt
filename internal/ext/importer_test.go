package ext

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/rpc/flipt"
)

type mockCreator struct {
	getNSReqs []*flipt.GetNamespaceRequest
	getNSErr  error

	createNSReqs []*flipt.CreateNamespaceRequest
	createNSErr  error

	flagReqs []*flipt.CreateFlagRequest
	flagErr  error

	variantReqs []*flipt.CreateVariantRequest
	variantErr  error

	segmentReqs []*flipt.CreateSegmentRequest
	segmentErr  error

	constraintReqs []*flipt.CreateConstraintRequest
	constraintErr  error

	ruleReqs []*flipt.CreateRuleRequest
	ruleErr  error

	distributionReqs []*flipt.CreateDistributionRequest
	distributionErr  error

	rolloutReqs []*flipt.CreateRolloutRequest
	rolloutErr  error
}

func (m *mockCreator) GetNamespace(ctx context.Context, r *flipt.GetNamespaceRequest) (*flipt.Namespace, error) {
	m.getNSReqs = append(m.getNSReqs, r)
	return &flipt.Namespace{Key: "default"}, m.getNSErr
}

func (m *mockCreator) CreateNamespace(ctx context.Context, r *flipt.CreateNamespaceRequest) (*flipt.Namespace, error) {
	m.createNSReqs = append(m.createNSReqs, r)
	return &flipt.Namespace{Key: "default"}, m.createNSErr
}

func (m *mockCreator) CreateFlag(ctx context.Context, r *flipt.CreateFlagRequest) (*flipt.Flag, error) {
	m.flagReqs = append(m.flagReqs, r)
	if m.flagErr != nil {
		return nil, m.flagErr
	}
	return &flipt.Flag{
		Key:         r.Key,
		Name:        r.Name,
		Description: r.Description,
		Type:        r.Type,
		Enabled:     r.Enabled,
	}, nil
}

func (m *mockCreator) CreateVariant(ctx context.Context, r *flipt.CreateVariantRequest) (*flipt.Variant, error) {
	m.variantReqs = append(m.variantReqs, r)
	if m.variantErr != nil {
		return nil, m.variantErr
	}
	return &flipt.Variant{
		Id:          uuid.Must(uuid.NewV4()).String(),
		FlagKey:     r.FlagKey,
		Key:         r.Key,
		Name:        r.Name,
		Description: r.Description,
		Attachment:  r.Attachment,
	}, nil
}

func (m *mockCreator) CreateSegment(ctx context.Context, r *flipt.CreateSegmentRequest) (*flipt.Segment, error) {
	m.segmentReqs = append(m.segmentReqs, r)
	if m.segmentErr != nil {
		return nil, m.segmentErr
	}
	return &flipt.Segment{
		Key:         r.Key,
		Name:        r.Name,
		Description: r.Description,
		MatchType:   r.MatchType,
	}, nil
}

func (m *mockCreator) CreateConstraint(ctx context.Context, r *flipt.CreateConstraintRequest) (*flipt.Constraint, error) {
	m.constraintReqs = append(m.constraintReqs, r)
	if m.constraintErr != nil {
		return nil, m.constraintErr
	}
	return &flipt.Constraint{
		Id:         uuid.Must(uuid.NewV4()).String(),
		SegmentKey: r.SegmentKey,
		Type:       r.Type,
		Property:   r.Property,
		Operator:   r.Operator,
		Value:      r.Value,
	}, nil
}

func (m *mockCreator) CreateRule(ctx context.Context, r *flipt.CreateRuleRequest) (*flipt.Rule, error) {
	m.ruleReqs = append(m.ruleReqs, r)
	if m.ruleErr != nil {
		return nil, m.ruleErr
	}
	return &flipt.Rule{
		Id:         uuid.Must(uuid.NewV4()).String(),
		FlagKey:    r.FlagKey,
		SegmentKey: r.SegmentKey,
		Rank:       r.Rank,
	}, nil
}

func (m *mockCreator) CreateDistribution(ctx context.Context, r *flipt.CreateDistributionRequest) (*flipt.Distribution, error) {
	m.distributionReqs = append(m.distributionReqs, r)
	if m.distributionErr != nil {
		return nil, m.distributionErr
	}
	return &flipt.Distribution{
		Id:        uuid.Must(uuid.NewV4()).String(),
		RuleId:    r.RuleId,
		VariantId: r.VariantId,
		Rollout:   r.Rollout,
	}, nil
}

func (m *mockCreator) CreateRollout(ctx context.Context, r *flipt.CreateRolloutRequest) (*flipt.Rollout, error) {
	m.rolloutReqs = append(m.rolloutReqs, r)
	if m.rolloutErr != nil {
		return nil, m.rolloutErr
	}

	rollout := &flipt.Rollout{
		Id:           uuid.Must(uuid.NewV4()).String(),
		NamespaceKey: r.NamespaceKey,
		FlagKey:      r.FlagKey,
		Description:  r.Description,
		Rank:         r.Rank,
	}

	switch rule := r.Rule.(type) {
	case *flipt.CreateRolloutRequest_Threshold:
		rollout.Rule = &flipt.Rollout_Threshold{
			Threshold: rule.Threshold,
		}
	case *flipt.CreateRolloutRequest_Segment:
		rollout.Rule = &flipt.Rollout_Segment{
			Segment: rule.Segment,
		}
	default:
		return nil, errors.New("unexpected rollout rule type")
	}

	return rollout, nil
}

func TestImport(t *testing.T) {
	tests := []struct {
		name          string
		path          string
		hasAttachment bool
	}{
		{
			name:          "import with attachment",
			path:          "testdata/import.yml",
			hasAttachment: true,
		},
		{
			name:          "import without attachment",
			path:          "testdata/import_no_attachment.yml",
			hasAttachment: false,
		},
		{
			name:          "import with implicit rule ranks",
			path:          "testdata/import_implicit_rule_rank.yml",
			hasAttachment: true,
		},
		{
			name:          "import with multiple segments",
			path:          "testdata/import_rule_multiple_segments.yml",
			hasAttachment: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			var (
				creator  = &mockCreator{}
				importer = NewImporter(creator)
			)

			in, err := os.Open(tc.path)
			assert.NoError(t, err)
			defer in.Close()

			err = importer.Import(context.Background(), in)
			assert.NoError(t, err)

			require.Len(t, creator.flagReqs, 2)

			flag := creator.flagReqs[0]
			assert.Equal(t, "flag1", flag.Key)
			assert.Equal(t, "flag1", flag.Name)
			assert.Equal(t, "description", flag.Description)
			assert.Equal(t, flipt.FlagType_VARIANT_FLAG_TYPE, flag.Type)
			assert.Equal(t, true, flag.Enabled)

			assert.NotEmpty(t, creator.variantReqs)
			assert.Equal(t, 1, len(creator.variantReqs))
			variant := creator.variantReqs[0]
			assert.Equal(t, "variant1", variant.Key)
			assert.Equal(t, "variant1", variant.Name)
			assert.Equal(t, "variant description", variant.Description)

			if tc.hasAttachment {
				attachment := `{
					"pi": 3.141,
					"happy": true,
					"name": "Niels",
					"answer": {
					  "everything": 42
					},
					"list": [1, 0, 2],
					"object": {
					  "currency": "USD",
					  "value": 42.99
					}
				  }`

				assert.JSONEq(t, attachment, variant.Attachment)
			} else {
				assert.Empty(t, variant.Attachment)
			}

			boolFlag := creator.flagReqs[1]
			assert.Equal(t, "flag2", boolFlag.Key)
			assert.Equal(t, "flag2", boolFlag.Name)
			assert.Equal(t, "a boolean flag", boolFlag.Description)
			assert.Equal(t, flipt.FlagType_BOOLEAN_FLAG_TYPE, boolFlag.Type)
			assert.Equal(t, false, boolFlag.Enabled)

			require.Len(t, creator.segmentReqs, 1)
			segment := creator.segmentReqs[0]
			assert.Equal(t, "segment1", segment.Key)
			assert.Equal(t, "segment1", segment.Name)
			assert.Equal(t, "description", segment.Description)
			assert.Equal(t, flipt.MatchType_ANY_MATCH_TYPE, segment.MatchType)

			require.Len(t, creator.constraintReqs, 1)
			constraint := creator.constraintReqs[0]
			assert.Equal(t, flipt.ComparisonType_STRING_COMPARISON_TYPE, constraint.Type)
			assert.Equal(t, "fizz", constraint.Property)
			assert.Equal(t, "neq", constraint.Operator)
			assert.Equal(t, "buzz", constraint.Value)

			require.Len(t, creator.ruleReqs, 1)
			rule := creator.ruleReqs[0]
			if rule.SegmentKey != "" {
				assert.Equal(t, "segment1", rule.SegmentKey)
			} else {
				assert.Len(t, rule.SegmentKeys, 1)
				assert.Equal(t, "segment1", rule.SegmentKeys[0])
			}
			assert.Equal(t, int32(1), rule.Rank)

			require.Len(t, creator.distributionReqs, 1)
			distribution := creator.distributionReqs[0]
			assert.Equal(t, "flag1", distribution.FlagKey)
			assert.NotEmpty(t, distribution.VariantId)
			assert.NotEmpty(t, distribution.RuleId)
			assert.Equal(t, float32(100), distribution.Rollout)

			require.Len(t, creator.rolloutReqs, 2)
			segmentRollout := creator.rolloutReqs[0]
			assert.Equal(t, "flag2", segmentRollout.FlagKey)
			assert.Equal(t, "enabled for internal users", segmentRollout.Description)
			assert.Equal(t, int32(1), segmentRollout.Rank)
			require.NotNil(t, segmentRollout.GetSegment())
			assert.Equal(t, "internal_users", segmentRollout.GetSegment().SegmentKey)
			assert.Equal(t, true, segmentRollout.GetSegment().Value)

			percentageRollout := creator.rolloutReqs[1]
			assert.Equal(t, "flag2", percentageRollout.FlagKey)
			assert.Equal(t, "enabled for 50%", percentageRollout.Description)
			assert.Equal(t, int32(2), percentageRollout.Rank)
			require.NotNil(t, percentageRollout.GetThreshold())
			assert.Equal(t, float32(50.0), percentageRollout.GetThreshold().Percentage)
			assert.Equal(t, true, percentageRollout.GetThreshold().Value)
		})
	}
}

func TestImport_Export(t *testing.T) {
	var (
		creator  = &mockCreator{}
		importer = NewImporter(creator)
	)

	in, err := os.Open("testdata/export.yml")
	assert.NoError(t, err)
	defer in.Close()

	err = importer.Import(context.Background(), in)
	require.NoError(t, err)
	assert.Equal(t, "default", creator.flagReqs[0].NamespaceKey)
}

func TestImport_InvalidVersion(t *testing.T) {
	var (
		creator  = &mockCreator{}
		importer = NewImporter(creator)
	)

	in, err := os.Open("testdata/import_invalid_version.yml")
	assert.NoError(t, err)
	defer in.Close()

	err = importer.Import(context.Background(), in)
	assert.EqualError(t, err, "unsupported version: 5.0")
}

func TestImport_FlagType_LTVersion1_1(t *testing.T) {
	var (
		creator  = &mockCreator{}
		importer = NewImporter(creator)
	)

	in, err := os.Open("testdata/import_v1_flag_type_not_supported.yml")
	assert.NoError(t, err)
	defer in.Close()

	err = importer.Import(context.Background(), in)
	assert.EqualError(t, err, "flag.type is supported in version >=1.1, found 1.0")
}

func TestImport_Rollouts_LTVersion1_1(t *testing.T) {
	var (
		creator  = &mockCreator{}
		importer = NewImporter(creator)
	)

	in, err := os.Open("testdata/import_v1_rollouts_not_supported.yml")
	assert.NoError(t, err)
	defer in.Close()

	err = importer.Import(context.Background(), in)
	assert.EqualError(t, err, "flag.rollouts is supported in version >=1.1, found 1.0")
}

func TestImport_Namespaces(t *testing.T) {
	tests := []struct {
		name         string
		docNamespace string
		cliNamespace string
		wantError    bool
	}{
		{
			name:         "import with namespace",
			docNamespace: "namespace1",
			cliNamespace: "namespace1",
		}, {
			name:         "import without doc namespace",
			docNamespace: "",
			cliNamespace: "namespace1",
		}, {
			name:         "import without cli namespace",
			docNamespace: "namespace1",
			cliNamespace: "",
		}, {
			name:         "import with different namespaces",
			docNamespace: "namespace1",
			cliNamespace: "namespace2",
			wantError:    true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {

			var (
				creator  = &mockCreator{}
				importer = NewImporter(creator, WithNamespace(tc.cliNamespace))
			)

			doc := fmt.Sprintf("version: 1.0\nnamespace: %s", tc.docNamespace)
			err := importer.Import(context.Background(), strings.NewReader(doc))
			if tc.wantError {
				msg := fmt.Sprintf("namespace mismatch: namespaces must match in file and args if both provided: %s != %s", tc.docNamespace, tc.cliNamespace)
				assert.EqualError(t, err, msg)
				return
			}

			assert.NoError(t, err)
		})
	}

}
