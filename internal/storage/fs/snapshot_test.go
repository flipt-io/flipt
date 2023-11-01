package fs

import (
	"context"
	"embed"
	"errors"
	"io/fs"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.flipt.io/flipt/internal/storage"
	"go.flipt.io/flipt/rpc/flipt"
	"go.uber.org/zap/zaptest"
)

//go:embed all:fixtures
var testdata embed.FS

func TestFSWithIndex(t *testing.T) {
	fwi, _ := fs.Sub(testdata, "fixtures/fswithindex")

	ss, err := SnapshotFromFS(zaptest.NewLogger(t), fwi)
	require.NoError(t, err)

	tfs := &FSIndexSuite{
		store: ss,
	}

	suite.Run(t, tfs)
}

type FSIndexSuite struct {
	suite.Suite
	store storage.Store
}

func (fis *FSIndexSuite) TestCountFlag() {
	t := fis.T()

	flagCount, err := fis.store.CountFlags(context.TODO(), "production")
	require.NoError(t, err)

	assert.Equal(t, 12, int(flagCount))

	flagCount, err = fis.store.CountFlags(context.TODO(), "sandbox")
	require.NoError(t, err)
	assert.Equal(t, 12, int(flagCount))
}

func (fis *FSIndexSuite) TestGetFlag() {
	t := fis.T()

	testCases := []struct {
		name      string
		namespace string
		flagKey   string
		flag      *flipt.Flag
	}{
		{
			name:      "Production",
			namespace: "production",
			flagKey:   "prod-flag",
			flag: &flipt.Flag{
				NamespaceKey: "production",
				Key:          "prod-flag",
				Name:         "Prod Flag",
				Description:  "description",
				Enabled:      true,
				Variants: []*flipt.Variant{
					{
						Key:          "prod-variant",
						Name:         "Prod Variant",
						NamespaceKey: "production",
					},
					{
						Key:          "foo",
						NamespaceKey: "production",
					},
				},
			},
		},
		{
			name:      "Sandbox",
			namespace: "sandbox",
			flagKey:   "sandbox-flag",
			flag: &flipt.Flag{
				NamespaceKey: "sandbox",
				Key:          "sandbox-flag",
				Name:         "Sandbox Flag",
				Description:  "description",
				Enabled:      true,
				Variants: []*flipt.Variant{
					{
						Key:          "sandbox-variant",
						Name:         "Sandbox Variant",
						NamespaceKey: "sandbox",
					},
					{
						Key:          "foo",
						NamespaceKey: "sandbox",
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			flag, err := fis.store.GetFlag(context.TODO(), tc.namespace, tc.flagKey)
			require.NoError(t, err)

			assert.Equal(t, tc.flag.Key, flag.Key)
			assert.Equal(t, tc.flag.NamespaceKey, flag.NamespaceKey)
			assert.Equal(t, tc.flag.Name, flag.Name)
			assert.Equal(t, tc.flag.Description, flag.Description)

			for i := 0; i < len(flag.Variants); i++ {
				v := tc.flag.Variants[i]
				fv := flag.Variants[i]
				assert.Equal(t, v.NamespaceKey, fv.NamespaceKey)
				assert.Equal(t, v.Key, fv.Key)
				assert.Equal(t, v.Name, fv.Name)
				assert.Equal(t, v.Description, fv.Description)
			}
		})
	}
}

func (fis *FSIndexSuite) TestListFlags() {
	t := fis.T()

	testCases := []struct {
		name      string
		namespace string
		pageToken string
		listError error
	}{
		{
			name:      "Production",
			namespace: "production",
		},
		{
			name:      "Sandbox",
			namespace: "sandbox",
		},
		{
			name:      "Page Token Invalid",
			namespace: "production",
			pageToken: "foo",
			listError: errors.New("pageToken is not valid: \"foo\""),
		},
		{
			name:      "Invalid Offset",
			namespace: "production",
			pageToken: "60000",
			listError: errors.New("invalid offset: 60000"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			flags, err := fis.store.ListFlags(context.TODO(), tc.namespace, storage.WithLimit(5), storage.WithPageToken(tc.pageToken))
			if tc.listError != nil {
				assert.EqualError(t, err, tc.listError.Error())
				return
			}

			require.NoError(t, err)
			assert.Len(t, flags.Results, 5)
			assert.Equal(t, "5", flags.NextPageToken)
		})
	}
}

func (fis *FSIndexSuite) TestGetNamespace() {
	t := fis.T()

	testCases := []struct {
		name         string
		namespaceKey string
		fliptNs      *flipt.Namespace
	}{
		{
			name:         "production",
			namespaceKey: "production",
			fliptNs: &flipt.Namespace{
				Key:  "production",
				Name: "production",
			},
		},
		{
			name:         "sandbox",
			namespaceKey: "sandbox",
			fliptNs: &flipt.Namespace{
				Key:  "sandbox",
				Name: "sandbox",
			},
		},
	}

	for _, tc := range testCases {
		ns, err := fis.store.GetNamespace(context.TODO(), tc.namespaceKey)
		require.NoError(t, err)

		assert.Equal(t, tc.fliptNs.Key, ns.Key)
		assert.Equal(t, tc.fliptNs.Name, ns.Name)
		assert.NotZero(t, ns.CreatedAt)
		assert.NotZero(t, ns.UpdatedAt)
	}
}

func (fis *FSIndexSuite) TestCountNamespaces() {
	t := fis.T()

	namespacesCount, err := fis.store.CountNamespaces(context.TODO())
	require.NoError(t, err)

	assert.Equal(t, 3, int(namespacesCount))
}

func (fis *FSIndexSuite) TestListNamespaces() {
	t := fis.T()

	namespaces, err := fis.store.ListNamespaces(context.TODO(), storage.WithLimit(2), storage.WithPageToken("0"))
	require.NoError(t, err)

	assert.Len(t, namespaces.Results, 2)
	assert.Equal(t, "2", namespaces.NextPageToken)

	for _, ns := range namespaces.Results {
		assert.NotZero(t, ns.CreatedAt)
		assert.NotZero(t, ns.UpdatedAt)
	}
}

func (fis *FSIndexSuite) TestGetSegment() {
	t := fis.T()

	testCases := []struct {
		name       string
		namespace  string
		segmentKey string
		segment    *flipt.Segment
	}{
		{
			name:       "Production",
			namespace:  "production",
			segmentKey: "segment1",
			segment: &flipt.Segment{
				Key:          "segment1",
				Name:         "segment1",
				MatchType:    flipt.MatchType_ANY_MATCH_TYPE,
				NamespaceKey: "production",
				Constraints: []*flipt.Constraint{
					{
						SegmentKey:   "segment1",
						Property:     "foo",
						Operator:     "eq",
						Value:        "baz",
						Description:  "desc",
						NamespaceKey: "production",
					},
					{
						SegmentKey:   "segment1",
						Property:     "fizz",
						Operator:     "neq",
						Value:        "buzz",
						Description:  "desc",
						NamespaceKey: "production",
					},
				},
			},
		},
		{
			name:       "Sandbox",
			namespace:  "sandbox",
			segmentKey: "segment1",
			segment: &flipt.Segment{
				Key:          "segment1",
				Name:         "segment1",
				MatchType:    flipt.MatchType_ANY_MATCH_TYPE,
				NamespaceKey: "sandbox",
				Constraints: []*flipt.Constraint{
					{
						SegmentKey:   "segment1",
						Property:     "foo",
						Operator:     "eq",
						Value:        "baz",
						Description:  "desc",
						NamespaceKey: "sandbox",
					},
					{
						SegmentKey:   "segment1",
						Property:     "fizz",
						Operator:     "neq",
						Value:        "buzz",
						Description:  "desc",
						NamespaceKey: "sandbox",
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sgmt, err := fis.store.GetSegment(context.TODO(), tc.namespace, tc.segmentKey)
			require.NoError(t, err)

			assert.Equal(t, tc.segment.Key, sgmt.Key)
			assert.Equal(t, tc.segment.Name, sgmt.Name)
			assert.Equal(t, tc.segment.NamespaceKey, sgmt.NamespaceKey)

			for i := 0; i < len(tc.segment.Constraints); i++ {
				c := tc.segment.Constraints[i]
				fc := sgmt.Constraints[i]
				assert.Equal(t, c.SegmentKey, fc.SegmentKey)
				assert.Equal(t, c.Property, fc.Property)
				assert.Equal(t, c.Operator, fc.Operator)
				assert.Equal(t, c.Value, fc.Value)
				assert.Equal(t, c.Description, fc.Description)
				assert.Equal(t, c.NamespaceKey, fc.NamespaceKey)
			}
		})
	}
}

func (fis *FSIndexSuite) TestListSegments() {
	t := fis.T()

	testCases := []struct {
		name      string
		namespace string
		pageToken string
		listError error
	}{
		{
			name:      "Production",
			namespace: "production",
			pageToken: "0",
		},
		{
			name:      "Sandbox",
			namespace: "sandbox",
			pageToken: "0",
		},
		{
			name:      "Page Token Invalid",
			namespace: "production",
			pageToken: "foo",
			listError: errors.New("pageToken is not valid: \"foo\""),
		},
		{
			name:      "Invalid Offset",
			namespace: "production",
			pageToken: "60000",
			listError: errors.New("invalid offset: 60000"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			flags, err := fis.store.ListSegments(context.TODO(), tc.namespace, storage.WithLimit(5), storage.WithPageToken(tc.pageToken))
			if tc.listError != nil {
				assert.EqualError(t, err, tc.listError.Error())
				return
			}

			require.NoError(t, err)
			assert.Len(t, flags.Results, 5)
			assert.Equal(t, "5", flags.NextPageToken)
		})
	}
}

func (fis *FSIndexSuite) TestCountSegment() {
	t := fis.T()

	segmentCount, err := fis.store.CountSegments(context.TODO(), "production")
	require.NoError(t, err)

	assert.Equal(t, 11, int(segmentCount))

	segmentCount, err = fis.store.CountSegments(context.TODO(), "sandbox")
	require.NoError(t, err)
	assert.Equal(t, 11, int(segmentCount))
}

func (fis *FSIndexSuite) TestGetEvaluationRollouts() {
	t := fis.T()

	testCases := []struct {
		name      string
		namespace string
		flagKey   string
	}{
		{
			name:      "Production",
			namespace: "production",
			flagKey:   "flag_boolean",
		},
		{
			name:      "Sandbox",
			namespace: "sandbox",
			flagKey:   "flag_boolean",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rollouts, err := fis.store.GetEvaluationRollouts(context.TODO(), tc.namespace, tc.flagKey)
			require.NoError(t, err)

			assert.Len(t, rollouts, 2)

			assert.Equal(t, tc.namespace, rollouts[0].NamespaceKey)
			assert.Equal(t, int32(1), rollouts[0].Rank)

			require.NotNil(t, rollouts[0].Segment)
			assert.Contains(t, rollouts[0].Segment.Segments, "segment1")
			assert.True(t, rollouts[0].Segment.Value, "segment value should be true")

			assert.Equal(t, tc.namespace, rollouts[1].NamespaceKey)
			assert.Equal(t, int32(2), rollouts[1].Rank)

			require.NotNil(t, rollouts[1].Threshold)
			assert.Equal(t, float32(50), rollouts[1].Threshold.Percentage)
			assert.True(t, rollouts[1].Threshold.Value, "threshold value should be true")
		})
	}
}

func (fis *FSIndexSuite) TestGetEvaluationRules() {
	t := fis.T()

	testCases := []struct {
		name        string
		namespace   string
		flagKey     string
		constraints []*flipt.Constraint
	}{
		{
			name:      "Production",
			namespace: "production",
			flagKey:   "prod-flag",
			constraints: []*flipt.Constraint{
				{
					SegmentKey:   "segment1",
					Property:     "foo",
					Operator:     "eq",
					Value:        "baz",
					Type:         flipt.ComparisonType_STRING_COMPARISON_TYPE,
					NamespaceKey: "production",
				},
				{
					SegmentKey:   "segment1",
					Property:     "fizz",
					Operator:     "neq",
					Value:        "buzz",
					Type:         flipt.ComparisonType_STRING_COMPARISON_TYPE,
					NamespaceKey: "production",
				},
			},
		},
		{
			name:      "Sandbox",
			namespace: "sandbox",
			flagKey:   "sandbox-flag",
			constraints: []*flipt.Constraint{
				{
					SegmentKey:   "segment1",
					Property:     "foo",
					Operator:     "eq",
					Value:        "baz",
					Type:         flipt.ComparisonType_STRING_COMPARISON_TYPE,
					NamespaceKey: "sandbox",
				},
				{
					SegmentKey:   "segment1",
					Property:     "fizz",
					Operator:     "neq",
					Value:        "buzz",
					Type:         flipt.ComparisonType_STRING_COMPARISON_TYPE,
					NamespaceKey: "sandbox",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rules, err := fis.store.GetEvaluationRules(context.TODO(), tc.namespace, tc.flagKey)
			require.NoError(t, err)

			assert.Len(t, rules, 1)

			assert.Equal(t, tc.namespace, rules[0].NamespaceKey)
			assert.Equal(t, tc.flagKey, rules[0].FlagKey)
			assert.Equal(t, int32(1), rules[0].Rank)
			assert.Contains(t, rules[0].Segments, "segment1")

			for i := 0; i < len(tc.constraints); i++ {
				fc := tc.constraints[i]
				c := rules[0].Segments[fc.SegmentKey].Constraints[i]

				assert.Equal(t, fc.Type, c.Type)
				assert.Equal(t, fc.Property, c.Property)
				assert.Equal(t, fc.Operator, c.Operator)
				assert.Equal(t, fc.Value, c.Value)
			}
		})
	}
}

func (fis *FSIndexSuite) TestGetEvaluationDistributions() {
	t := fis.T()

	testCases := []struct {
		name                string
		namespace           string
		flagKey             string
		expectedVariantName string
	}{
		{
			name:                "Sandbox",
			namespace:           "sandbox",
			flagKey:             "sandbox-flag",
			expectedVariantName: "sandbox-variant",
		},
		{
			name:                "Production",
			namespace:           "production",
			flagKey:             "prod-flag",
			expectedVariantName: "prod-variant",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rules, err := fis.store.ListRules(context.TODO(), tc.namespace, tc.flagKey)
			require.NoError(t, err)
			assert.Len(t, rules.Results, 1)

			dist, err := fis.store.GetEvaluationDistributions(context.TODO(), rules.Results[0].Id)

			require.NoError(t, err)

			assert.Equal(t, tc.expectedVariantName, dist[0].VariantKey)
			assert.Equal(t, float32(100), dist[0].Rollout)
		})
	}
}

func (fis *FSIndexSuite) TestCountRollouts() {
	t := fis.T()

	rolloutsCount, err := fis.store.CountRollouts(context.TODO(), "production", "flag_boolean")
	require.NoError(t, err)
	assert.Equal(t, 2, int(rolloutsCount))

	rolloutsCount, err = fis.store.CountRollouts(context.TODO(), "sandbox", "flag_boolean")
	require.NoError(t, err)
	assert.Equal(t, 2, int(rolloutsCount))
}

func (fis *FSIndexSuite) TestListAndGetRollouts() {
	t := fis.T()

	testCases := []struct {
		name      string
		namespace string
		flagKey   string
		pageToken string
		listError error
	}{
		{
			name:      "Production",
			namespace: "production",
			flagKey:   "flag_boolean",
		},
		{
			name:      "Sandbox",
			namespace: "sandbox",
			flagKey:   "flag_boolean",
		},
		{
			name:      "Page Token Invalid",
			namespace: "production",
			flagKey:   "flag_boolean",
			pageToken: "foo",
			listError: errors.New("pageToken is not valid: \"foo\""),
		},
		{
			name:      "Invalid Offset",
			namespace: "production",
			flagKey:   "flag_boolean",
			pageToken: "60000",
			listError: errors.New("invalid offset: 60000"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rollouts, err := fis.store.ListRollouts(context.TODO(), tc.namespace, tc.flagKey, storage.WithPageToken(tc.pageToken))
			if tc.listError != nil {
				assert.EqualError(t, err, tc.listError.Error())
				return
			}

			require.NoError(t, err)

			for _, rollout := range rollouts.Results {
				r, err := fis.store.GetRollout(context.TODO(), tc.namespace, rollout.Id)
				require.NoError(t, err)

				assert.Equal(t, r, rollout)
			}
		})
	}
}

func (fis *FSIndexSuite) TestListAndGetRules() {
	t := fis.T()

	testCases := []struct {
		name      string
		namespace string
		flagKey   string
		pageToken string
		listError error
	}{
		{
			name:      "Production",
			namespace: "production",
			flagKey:   "prod-flag",
		},
		{
			name:      "Sandbox",
			namespace: "sandbox",
			flagKey:   "sandbox-flag",
		},
		{
			name:      "Page Token Invalid",
			namespace: "production",
			flagKey:   "prod-flag",
			pageToken: "foo",
			listError: errors.New("pageToken is not valid: \"foo\""),
		},
		{
			name:      "Invalid Offset",
			namespace: "production",
			flagKey:   "prod-flag",
			pageToken: "60000",
			listError: errors.New("invalid offset: 60000"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rules, err := fis.store.ListRules(context.TODO(), tc.namespace, tc.flagKey, storage.WithPageToken(tc.pageToken))
			if tc.listError != nil {
				assert.EqualError(t, err, tc.listError.Error())
				return
			}

			require.NoError(t, err)

			for _, rule := range rules.Results {
				r, err := fis.store.GetRule(context.TODO(), tc.namespace, rule.Id)
				require.NoError(t, err)

				assert.Equal(t, r, rule)
			}
		})
	}
}

type FSWithoutIndexSuite struct {
	suite.Suite
	store storage.Store
}

func TestFSWithoutIndex(t *testing.T) {
	fwoi, _ := fs.Sub(testdata, "fixtures/fswithoutindex")

	ss, err := SnapshotFromFS(zaptest.NewLogger(t), fwoi)
	require.NoError(t, err)

	tfs := &FSWithoutIndexSuite{
		store: ss,
	}

	suite.Run(t, tfs)
}

func (fis *FSWithoutIndexSuite) TestCountFlag() {
	t := fis.T()

	flagCount, err := fis.store.CountFlags(context.TODO(), "production")
	require.NoError(t, err)

	assert.Equal(t, 14, int(flagCount))

	flagCount, err = fis.store.CountFlags(context.TODO(), "sandbox")
	require.NoError(t, err)
	assert.Equal(t, 14, int(flagCount))

	flagCount, err = fis.store.CountFlags(context.TODO(), "staging")
	require.NoError(t, err)
	assert.Equal(t, 14, int(flagCount))
}

func (fis *FSWithoutIndexSuite) TestGetFlag() {
	t := fis.T()

	testCases := []struct {
		name      string
		namespace string
		flagKey   string
		flag      *flipt.Flag
	}{
		{
			name:      "Production",
			namespace: "production",
			flagKey:   "prod-flag",
			flag: &flipt.Flag{
				NamespaceKey: "production",
				Key:          "prod-flag",
				Name:         "Prod Flag",
				Description:  "description",
				Enabled:      true,
				Variants: []*flipt.Variant{
					{
						Key:          "prod-variant",
						Name:         "Prod Variant",
						NamespaceKey: "production",
					},
					{
						Key:          "foo",
						NamespaceKey: "production",
					},
				},
			},
		},
		{
			name:      "Production One",
			namespace: "production",
			flagKey:   "prod-flag-1",
			flag: &flipt.Flag{
				NamespaceKey: "production",
				Key:          "prod-flag-1",
				Name:         "Prod Flag 1",
				Description:  "description",
				Enabled:      true,
				Variants: []*flipt.Variant{
					{
						Key:          "prod-variant",
						Name:         "Prod Variant",
						NamespaceKey: "production",
					},
					{
						Key:          "foo",
						NamespaceKey: "production",
					},
				},
			},
		},
		{
			name:      "Sandbox",
			namespace: "sandbox",
			flagKey:   "sandbox-flag",
			flag: &flipt.Flag{
				NamespaceKey: "sandbox",
				Key:          "sandbox-flag",
				Name:         "Sandbox Flag",
				Description:  "description",
				Enabled:      true,
				Variants: []*flipt.Variant{
					{
						Key:          "sandbox-variant",
						Name:         "Sandbox Variant",
						NamespaceKey: "sandbox",
					},
					{
						Key:          "foo",
						NamespaceKey: "sandbox",
					},
				},
			},
		},
		{
			name:      "Sandbox One",
			namespace: "sandbox",
			flagKey:   "sandbox-flag-1",
			flag: &flipt.Flag{
				NamespaceKey: "sandbox",
				Key:          "sandbox-flag-1",
				Name:         "Sandbox Flag 1",
				Description:  "description",
				Enabled:      true,
				Variants: []*flipt.Variant{
					{
						Key:          "sandbox-variant",
						Name:         "Sandbox Variant",
						NamespaceKey: "sandbox",
					},
					{
						Key:          "foo",
						NamespaceKey: "sandbox",
					},
				},
			},
		},
		{
			name:      "Staging",
			namespace: "staging",
			flagKey:   "staging-flag",
			flag: &flipt.Flag{
				NamespaceKey: "staging",
				Key:          "staging-flag",
				Name:         "Staging Flag",
				Description:  "description",
				Enabled:      true,
				Variants: []*flipt.Variant{
					{
						Key:          "staging-variant",
						Name:         "Staging Variant",
						NamespaceKey: "staging",
					},
					{
						Key:          "foo",
						NamespaceKey: "staging",
					},
				},
			},
		},
		{
			name:      "Staging One",
			namespace: "staging",
			flagKey:   "staging-flag-1",
			flag: &flipt.Flag{
				NamespaceKey: "staging",
				Key:          "staging-flag-1",
				Name:         "Staging Flag 1",
				Description:  "description",
				Enabled:      true,
				Variants: []*flipt.Variant{
					{
						Key:          "staging-variant",
						Name:         "Staging Variant",
						NamespaceKey: "staging",
					},
					{
						Key:          "foo",
						NamespaceKey: "staging",
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			flag, err := fis.store.GetFlag(context.TODO(), tc.namespace, tc.flagKey)
			require.NoError(t, err)

			assert.Equal(t, tc.flag.Key, flag.Key)
			assert.Equal(t, tc.flag.NamespaceKey, flag.NamespaceKey)
			assert.Equal(t, tc.flag.Name, flag.Name)
			assert.Equal(t, tc.flag.Description, flag.Description)

			for i := 0; i < len(flag.Variants); i++ {
				v := tc.flag.Variants[i]
				fv := flag.Variants[i]

				assert.Equal(t, v.NamespaceKey, fv.NamespaceKey)
				assert.Equal(t, v.Key, fv.Key)
				assert.Equal(t, v.Name, fv.Name)
				assert.Equal(t, v.Description, fv.Description)
			}
		})
	}
}

func (fis *FSWithoutIndexSuite) TestListFlags() {
	t := fis.T()

	testCases := []struct {
		name      string
		namespace string
		pageToken string
		listError error
	}{
		{
			name:      "Production",
			namespace: "production",
		},
		{
			name:      "Sandbox",
			namespace: "sandbox",
		},
		{
			name:      "Staging",
			namespace: "staging",
		},
		{
			name:      "Page Token Invalid",
			namespace: "production",
			pageToken: "foo",
			listError: errors.New("pageToken is not valid: \"foo\""),
		},
		{
			name:      "Invalid Offset",
			namespace: "production",
			pageToken: "60000",
			listError: errors.New("invalid offset: 60000"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			flags, err := fis.store.ListFlags(context.TODO(), tc.namespace, storage.WithLimit(5), storage.WithPageToken(tc.pageToken))
			if tc.listError != nil {
				assert.EqualError(t, err, tc.listError.Error())
				return
			}

			require.NoError(t, err)
			assert.Len(t, flags.Results, 5)
			assert.Equal(t, "5", flags.NextPageToken)
		})
	}
}

func (fis *FSWithoutIndexSuite) TestGetNamespace() {
	t := fis.T()

	testCases := []struct {
		name         string
		namespaceKey string
		fliptNs      *flipt.Namespace
	}{
		{
			name:         "production",
			namespaceKey: "production",
			fliptNs: &flipt.Namespace{
				Key:  "production",
				Name: "production",
			},
		},
		{
			name:         "sandbox",
			namespaceKey: "sandbox",
			fliptNs: &flipt.Namespace{
				Key:  "sandbox",
				Name: "sandbox",
			},
		},
		{
			name:         "staging",
			namespaceKey: "staging",
			fliptNs: &flipt.Namespace{
				Key:  "staging",
				Name: "staging",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ns, err := fis.store.GetNamespace(context.TODO(), tc.namespaceKey)
			require.NoError(t, err)

			assert.Equal(t, tc.fliptNs.Key, ns.Key)
			assert.Equal(t, tc.fliptNs.Name, ns.Name)
		})
	}
}

func (fis *FSWithoutIndexSuite) TestCountNamespaces() {
	t := fis.T()

	namespacesCount, err := fis.store.CountNamespaces(context.TODO())
	require.NoError(t, err)

	assert.Equal(t, 4, int(namespacesCount))
}

func (fis *FSWithoutIndexSuite) TestListNamespaces() {
	t := fis.T()

	namespaces, err := fis.store.ListNamespaces(context.TODO(), storage.WithLimit(3), storage.WithPageToken("0"))
	require.NoError(t, err)

	assert.Len(t, namespaces.Results, 3)
	assert.Equal(t, "3", namespaces.NextPageToken)
}

func (fis *FSWithoutIndexSuite) TestGetSegment() {
	t := fis.T()

	testCases := []struct {
		name       string
		namespace  string
		segmentKey string
		segment    *flipt.Segment
	}{
		{
			name:       "Production",
			namespace:  "production",
			segmentKey: "segment1",
			segment: &flipt.Segment{
				Key:          "segment1",
				Name:         "segment1",
				MatchType:    flipt.MatchType_ANY_MATCH_TYPE,
				NamespaceKey: "production",
				Constraints: []*flipt.Constraint{
					{
						SegmentKey:   "segment1",
						Property:     "foo",
						Operator:     "eq",
						Value:        "baz",
						Description:  "desc",
						NamespaceKey: "production",
					},
					{
						SegmentKey:   "segment1",
						Property:     "fizz",
						Operator:     "neq",
						Value:        "buzz",
						Description:  "desc",
						NamespaceKey: "production",
					},
				},
			},
		},
		{
			name:       "Production One",
			namespace:  "production",
			segmentKey: "segment2",
			segment: &flipt.Segment{
				Key:          "segment2",
				Name:         "segment2",
				MatchType:    flipt.MatchType_ANY_MATCH_TYPE,
				NamespaceKey: "production",
				Constraints: []*flipt.Constraint{
					{
						SegmentKey:   "segment2",
						Property:     "foo",
						Operator:     "eq",
						Value:        "baz",
						Description:  "desc",
						NamespaceKey: "production",
					},
					{
						SegmentKey:   "segment2",
						Property:     "fizz",
						Operator:     "neq",
						Value:        "buzz",
						Description:  "desc",
						NamespaceKey: "production",
					},
				},
			},
		},
		{
			name:       "Sandbox",
			namespace:  "sandbox",
			segmentKey: "segment1",
			segment: &flipt.Segment{
				Key:          "segment1",
				Name:         "segment1",
				MatchType:    flipt.MatchType_ANY_MATCH_TYPE,
				NamespaceKey: "sandbox",
				Constraints: []*flipt.Constraint{
					{
						SegmentKey:   "segment1",
						Property:     "foo",
						Operator:     "eq",
						Value:        "baz",
						Description:  "desc",
						NamespaceKey: "sandbox",
					},
					{
						SegmentKey:   "segment1",
						Property:     "fizz",
						Operator:     "neq",
						Value:        "buzz",
						Description:  "desc",
						NamespaceKey: "sandbox",
					},
				},
			},
		},
		{
			name:       "Sandbox One",
			namespace:  "sandbox",
			segmentKey: "segment2",
			segment: &flipt.Segment{
				Key:          "segment2",
				Name:         "segment2",
				MatchType:    flipt.MatchType_ANY_MATCH_TYPE,
				NamespaceKey: "sandbox",
				Constraints: []*flipt.Constraint{
					{
						SegmentKey:   "segment2",
						Property:     "foo",
						Operator:     "eq",
						Value:        "baz",
						Description:  "desc",
						NamespaceKey: "sandbox",
					},
					{
						SegmentKey:   "segment2",
						Property:     "fizz",
						Operator:     "neq",
						Value:        "buzz",
						Description:  "desc",
						NamespaceKey: "sandbox",
					},
				},
			},
		},
		{
			name:       "Staging",
			namespace:  "staging",
			segmentKey: "segment1",
			segment: &flipt.Segment{
				Key:          "segment1",
				Name:         "segment1",
				MatchType:    flipt.MatchType_ANY_MATCH_TYPE,
				NamespaceKey: "staging",
				Constraints: []*flipt.Constraint{
					{
						SegmentKey:   "segment1",
						Property:     "foo",
						Operator:     "eq",
						Value:        "baz",
						Description:  "desc",
						NamespaceKey: "staging",
					},
					{
						SegmentKey:   "segment1",
						Property:     "fizz",
						Operator:     "neq",
						Value:        "buzz",
						Description:  "desc",
						NamespaceKey: "staging",
					},
				},
			},
		},
		{
			name:       "Staging One",
			namespace:  "staging",
			segmentKey: "segment2",
			segment: &flipt.Segment{
				Key:          "segment2",
				Name:         "segment2",
				MatchType:    flipt.MatchType_ANY_MATCH_TYPE,
				NamespaceKey: "staging",
				Constraints: []*flipt.Constraint{
					{
						SegmentKey:   "segment2",
						Property:     "foo",
						Operator:     "eq",
						Value:        "baz",
						Description:  "desc",
						NamespaceKey: "staging",
					},
					{
						SegmentKey:   "segment2",
						Property:     "fizz",
						Operator:     "neq",
						Value:        "buzz",
						Description:  "desc",
						NamespaceKey: "staging",
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sgmt, err := fis.store.GetSegment(context.TODO(), tc.namespace, tc.segmentKey)
			require.NoError(t, err)

			assert.Equal(t, tc.segment.Key, sgmt.Key)
			assert.Equal(t, tc.segment.Name, sgmt.Name)
			assert.Equal(t, tc.segment.NamespaceKey, sgmt.NamespaceKey)

			for i := 0; i < len(tc.segment.Constraints); i++ {
				c := tc.segment.Constraints[i]
				fc := sgmt.Constraints[i]
				assert.Equal(t, c.SegmentKey, fc.SegmentKey)
				assert.Equal(t, c.Property, fc.Property)
				assert.Equal(t, c.Operator, fc.Operator)
				assert.Equal(t, c.Value, fc.Value)
				assert.Equal(t, c.Description, fc.Description)
				assert.Equal(t, c.NamespaceKey, fc.NamespaceKey)
			}
		})
	}
}

func (fis *FSWithoutIndexSuite) TestCountSegment() {
	t := fis.T()

	segmentCount, err := fis.store.CountSegments(context.TODO(), "production")
	require.NoError(t, err)

	assert.Equal(t, 12, int(segmentCount))

	segmentCount, err = fis.store.CountSegments(context.TODO(), "sandbox")
	require.NoError(t, err)
	assert.Equal(t, 12, int(segmentCount))

	segmentCount, err = fis.store.CountSegments(context.TODO(), "staging")
	require.NoError(t, err)
	assert.Equal(t, 12, int(segmentCount))
}

func (fis *FSWithoutIndexSuite) TestListSegments() {
	t := fis.T()

	testCases := []struct {
		name      string
		namespace string
		pageToken string
		listError error
	}{
		{
			name:      "Production",
			namespace: "production",
		},
		{
			name:      "Sandbox",
			namespace: "sandbox",
		},
		{
			name:      "Staging",
			namespace: "staging",
		},
		{
			name:      "Page Token Invalid",
			namespace: "production",
			pageToken: "foo",
			listError: errors.New("pageToken is not valid: \"foo\""),
		},
		{
			name:      "Invalid Offset",
			namespace: "production",
			pageToken: "60000",
			listError: errors.New("invalid offset: 60000"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			flags, err := fis.store.ListSegments(context.TODO(), tc.namespace, storage.WithLimit(5), storage.WithPageToken(tc.pageToken))
			if tc.listError != nil {
				assert.EqualError(t, err, tc.listError.Error())
				return
			}

			require.NoError(t, err)
			assert.Len(t, flags.Results, 5)
			assert.Equal(t, "5", flags.NextPageToken)
		})
	}
}

func (fis *FSWithoutIndexSuite) TestGetEvaluationRollouts() {
	t := fis.T()

	testCases := []struct {
		name      string
		namespace string
		flagKey   string
	}{
		{
			name:      "Production",
			namespace: "production",
			flagKey:   "flag_boolean",
		},
		{
			name:      "Sandbox",
			namespace: "sandbox",
			flagKey:   "flag_boolean",
		},
		{
			name:      "Staging",
			namespace: "staging",
			flagKey:   "flag_boolean",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rollouts, err := fis.store.GetEvaluationRollouts(context.TODO(), tc.namespace, tc.flagKey)
			require.NoError(t, err)

			assert.Len(t, rollouts, 2)

			assert.Equal(t, tc.namespace, rollouts[0].NamespaceKey)
			assert.Equal(t, int32(1), rollouts[0].Rank)

			require.NotNil(t, rollouts[0].Segment)
			assert.Contains(t, rollouts[0].Segment.Segments, "segment1")
			assert.True(t, rollouts[0].Segment.Value, "segment value should be true")

			assert.Equal(t, tc.namespace, rollouts[1].NamespaceKey)
			assert.Equal(t, int32(2), rollouts[1].Rank)

			require.NotNil(t, rollouts[1].Threshold)
			assert.Equal(t, float32(50), rollouts[1].Threshold.Percentage)
			assert.True(t, rollouts[1].Threshold.Value, "threshold value should be true")
		})
	}
}

func (fis *FSWithoutIndexSuite) TestGetEvaluationRules() {
	t := fis.T()

	testCases := []struct {
		name        string
		namespace   string
		flagKey     string
		constraints []*flipt.Constraint
	}{
		{
			name:      "Production",
			namespace: "production",
			flagKey:   "prod-flag",
			constraints: []*flipt.Constraint{
				{
					SegmentKey:   "segment1",
					Property:     "foo",
					Operator:     "eq",
					Value:        "baz",
					Type:         flipt.ComparisonType_STRING_COMPARISON_TYPE,
					NamespaceKey: "production",
				},
				{
					SegmentKey:   "segment1",
					Property:     "fizz",
					Operator:     "neq",
					Value:        "buzz",
					Type:         flipt.ComparisonType_STRING_COMPARISON_TYPE,
					NamespaceKey: "production",
				},
			},
		},
		{
			name:      "Sandbox",
			namespace: "sandbox",
			flagKey:   "sandbox-flag",
			constraints: []*flipt.Constraint{
				{
					SegmentKey:   "segment1",
					Property:     "foo",
					Operator:     "eq",
					Value:        "baz",
					Type:         flipt.ComparisonType_STRING_COMPARISON_TYPE,
					NamespaceKey: "sandbox",
				},
				{
					SegmentKey:   "segment1",
					Property:     "fizz",
					Operator:     "neq",
					Value:        "buzz",
					Type:         flipt.ComparisonType_STRING_COMPARISON_TYPE,
					NamespaceKey: "sandbox",
				},
			},
		},
		{
			name:      "Staging",
			namespace: "staging",
			flagKey:   "staging-flag",
			constraints: []*flipt.Constraint{
				{
					SegmentKey:   "segment1",
					Property:     "foo",
					Operator:     "eq",
					Value:        "baz",
					Type:         flipt.ComparisonType_STRING_COMPARISON_TYPE,
					NamespaceKey: "staging",
				},
				{
					SegmentKey:   "segment1",
					Property:     "fizz",
					Operator:     "neq",
					Value:        "buzz",
					Type:         flipt.ComparisonType_STRING_COMPARISON_TYPE,
					NamespaceKey: "staging",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rules, err := fis.store.GetEvaluationRules(context.TODO(), tc.namespace, tc.flagKey)
			require.NoError(t, err)

			assert.Len(t, rules, 1)

			assert.Equal(t, tc.namespace, rules[0].NamespaceKey)
			assert.Equal(t, tc.flagKey, rules[0].FlagKey)
			assert.Equal(t, int32(1), rules[0].Rank)
			assert.Contains(t, rules[0].Segments, "segment1")

			for i := 0; i < len(tc.constraints); i++ {
				fc := tc.constraints[i]
				c := rules[0].Segments[fc.SegmentKey].Constraints[i]

				assert.Equal(t, fc.Type, c.Type)
				assert.Equal(t, fc.Property, c.Property)
				assert.Equal(t, fc.Operator, c.Operator)
				assert.Equal(t, fc.Value, c.Value)
			}
		})
	}
}

func (fis *FSWithoutIndexSuite) TestGetEvaluationDistributions() {
	t := fis.T()

	testCases := []struct {
		name                string
		namespace           string
		flagKey             string
		expectedVariantName string
	}{
		{
			name:                "Production",
			namespace:           "production",
			flagKey:             "prod-flag",
			expectedVariantName: "prod-variant",
		},
		{
			name:                "Sandbox",
			namespace:           "sandbox",
			flagKey:             "sandbox-flag",
			expectedVariantName: "sandbox-variant",
		},
		{
			name:                "Staging",
			namespace:           "staging",
			flagKey:             "staging-flag",
			expectedVariantName: "staging-variant",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rules, err := fis.store.ListRules(context.TODO(), tc.namespace, tc.flagKey)
			require.NoError(t, err)
			assert.Len(t, rules.Results, 1)

			dist, err := fis.store.GetEvaluationDistributions(context.TODO(), rules.Results[0].Id)

			require.NoError(t, err)

			assert.Equal(t, tc.expectedVariantName, dist[0].VariantKey)
			assert.Equal(t, float32(100), dist[0].Rollout)
		})
	}
}

func (fis *FSWithoutIndexSuite) TestCountRollouts() {
	t := fis.T()

	rolloutsCount, err := fis.store.CountRollouts(context.TODO(), "production", "flag_boolean")
	require.NoError(t, err)

	assert.Equal(t, 2, int(rolloutsCount))

	rolloutsCount, err = fis.store.CountRollouts(context.TODO(), "sandbox", "flag_boolean")
	require.NoError(t, err)

	assert.Equal(t, 2, int(rolloutsCount))

	rolloutsCount, err = fis.store.CountRollouts(context.TODO(), "staging", "flag_boolean")
	require.NoError(t, err)

	assert.Equal(t, 2, int(rolloutsCount))
}

func (fis *FSWithoutIndexSuite) TestListAndGetRollouts() {
	t := fis.T()

	testCases := []struct {
		name      string
		namespace string
		flagKey   string
		pageToken string
		listError error
	}{
		{
			name:      "Production",
			namespace: "production",
			flagKey:   "flag_boolean",
		},
		{
			name:      "Sandbox",
			namespace: "sandbox",
			flagKey:   "flag_boolean",
		},
		{
			name:      "Staging",
			namespace: "staging",
			flagKey:   "flag_boolean",
		},
		{
			name:      "Page Token Invalid",
			namespace: "production",
			flagKey:   "flag_boolean",
			pageToken: "foo",
			listError: errors.New("pageToken is not valid: \"foo\""),
		},
		{
			name:      "Invalid Offset",
			namespace: "production",
			flagKey:   "flag_boolean",
			pageToken: "60000",
			listError: errors.New("invalid offset: 60000"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rollouts, err := fis.store.ListRollouts(context.TODO(), tc.namespace, tc.flagKey, storage.WithPageToken(tc.pageToken))
			if tc.listError != nil {
				assert.EqualError(t, err, tc.listError.Error())
				return
			}

			require.NoError(t, err)

			for _, rollout := range rollouts.Results {
				r, err := fis.store.GetRollout(context.TODO(), tc.namespace, rollout.Id)
				require.NoError(t, err)

				assert.Equal(t, r, rollout)
			}
		})
	}
}

func (fis *FSWithoutIndexSuite) TestListAndGetRules() {
	t := fis.T()

	testCases := []struct {
		name      string
		namespace string
		flagKey   string
		pageToken string
		listError error
	}{
		{
			name:      "Production",
			namespace: "production",
			flagKey:   "prod-flag",
		},
		{
			name:      "Sandbox",
			namespace: "sandbox",
			flagKey:   "sandbox-flag",
		},
		{
			name:      "Staging",
			namespace: "staging",
			flagKey:   "staging-flag",
		},
		{
			name:      "Page Token Invalid",
			namespace: "production",
			flagKey:   "prod-flag",
			pageToken: "foo",
			listError: errors.New("pageToken is not valid: \"foo\""),
		},
		{
			name:      "Invalid Offset",
			namespace: "production",
			flagKey:   "prod-flag",
			pageToken: "60000",
			listError: errors.New("invalid offset: 60000"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rules, err := fis.store.ListRules(context.TODO(), tc.namespace, tc.flagKey, storage.WithPageToken(tc.pageToken))
			if tc.listError != nil {
				assert.EqualError(t, err, tc.listError.Error())
				return
			}

			require.NoError(t, err)

			for _, rule := range rules.Results {
				r, err := fis.store.GetRule(context.TODO(), tc.namespace, rule.Id)
				require.NoError(t, err)

				assert.Equal(t, r, rule)
			}
		})
	}
}

func TestFS_Invalid_VariantFlag_Segment(t *testing.T) {
	fs, _ := fs.Sub(testdata, "fixtures/invalid_variant_flag_segment")
	_, err := SnapshotFromFS(zaptest.NewLogger(t), fs)
	require.EqualError(t, err, "flag fruit/apple rule 1 references unknown segment \"unknown\"")
}

func TestFS_Invalid_VariantFlag_Distribution(t *testing.T) {
	fs, _ := fs.Sub(testdata, "fixtures/invalid_variant_flag_distribution")
	_, err := SnapshotFromFS(zaptest.NewLogger(t), fs)
	require.EqualError(t, err, "flag fruit/apple rule 1 references unknown variant \"braeburn\"")
}

func TestFS_Invalid_BooleanFlag_Segment(t *testing.T) {
	fs, _ := fs.Sub(testdata, "fixtures/invalid_boolean_flag_segment")
	_, err := SnapshotFromFS(zaptest.NewLogger(t), fs)
	require.EqualError(t, err, "flag fruit/apple rule 1 references unknown segment \"unknown\"")
}

func TestFS_Empty_Features_File(t *testing.T) {
	fs, _ := fs.Sub(testdata, "fixtures/empty_features")
	ss, err := SnapshotFromFS(zaptest.NewLogger(t), fs)
	require.NoError(t, err)

	_, err = ss.ListFlags(context.TODO(), flipt.DefaultNamespace)
	require.NoError(t, err)

	_, err = ss.ListSegments(context.TODO(), flipt.DefaultNamespace)
	require.NoError(t, err)

	_, err = ss.ListRollouts(context.TODO(), flipt.DefaultNamespace, "no-flag")
	require.NoError(t, err)

	_, err = ss.ListRules(context.TODO(), flipt.DefaultNamespace, "no-flag")
	require.NoError(t, err)
}

func TestFS_YAML_Stream(t *testing.T) {
	fs, _ := fs.Sub(testdata, "fixtures/yaml_stream")
	ss, err := SnapshotFromFS(zaptest.NewLogger(t), fs)
	require.NoError(t, err)

	ns, err := ss.ListNamespaces(context.TODO())
	require.NoError(t, err)

	// 3 namespaces including default
	assert.Len(t, ns.Results, 3)
	var namespaces = make([]string, 0, len(ns.Results))

	for _, n := range ns.Results {
		namespaces = append(namespaces, n.Key)
	}

	for _, namespace := range []string{"default", "football", "fruit"} {
		assert.Contains(t, namespaces, namespace)
	}

	fflags, err := ss.ListFlags(context.TODO(), "football")
	require.NoError(t, err)

	assert.Len(t, fflags.Results, 1)
	assert.Equal(t, "team", fflags.Results[0].Key)

	frflags, err := ss.ListFlags(context.TODO(), "fruit")
	require.NoError(t, err)

	assert.Len(t, frflags.Results, 1)
	assert.Equal(t, "apple", frflags.Results[0].Key)

	fsegments, err := ss.ListSegments(context.TODO(), "football")
	require.NoError(t, err)

	assert.Len(t, fsegments.Results, 1)
	assert.Equal(t, "internal", fsegments.Results[0].Key)

	frsegments, err := ss.ListSegments(context.TODO(), "fruit")
	require.NoError(t, err)

	assert.Len(t, frsegments.Results, 1)
	assert.Equal(t, "internal", frsegments.Results[0].Key)
}
