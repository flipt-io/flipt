package fs

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.flipt.io/flipt/core/validation"
	flipterrors "go.flipt.io/flipt/errors"
	"go.flipt.io/flipt/internal/ext"
	"go.flipt.io/flipt/internal/storage"
	"go.flipt.io/flipt/rpc/flipt/core"
	"go.uber.org/zap/zaptest"
)

//go:embed all:testdata
var testdata embed.FS

func TestSnapshotFromFS_Invalid(t *testing.T) {
	for _, test := range []struct {
		path string
		err  error
	}{
		{
			path: "testdata/invalid/extension",
			err:  errors.New("unexpected extension: \".unknown\""),
		},
		{
			path: "testdata/invalid/variant_flag_segment",
			err:  flipterrors.ErrInvalid("flag fruit/apple rule 1 references unknown segment \"unknown\""),
		},
		{
			path: "testdata/invalid/variant_flag_distribution",
			err:  flipterrors.ErrInvalid("flag fruit/apple rule 1 references unknown variant \"braeburn\""),
		},
		{
			path: "testdata/invalid/boolean_flag_segment",
			err:  flipterrors.ErrInvalid("flag fruit/apple rule 1 references unknown segment \"unknown\""),
		},
		{
			path: "testdata/invalid/namespace",
			err: errors.Join(
				validation.Error{Message: "namespace: 2 errors in empty disjunction:", Location: validation.Location{File: "features.json", Line: 1}},
				validation.Error{Message: "namespace: conflicting values 1 and string (mismatched types int and string)", Location: validation.Location{File: "features.json", Line: 1}},
				validation.Error{Message: "namespace: conflicting values 1 and {key:((string & =~\"^[-_,A-Za-z0-9]+$\")|*\"default\"),name?:(string & =~\"^.+$\"),description?:string} (mismatched types int and struct)", Location: validation.Location{File: "features.json", Line: 1}},
			),
		},
	} {
		t.Run(test.path, func(t *testing.T) {
			dir, err := fs.Sub(testdata, test.path)
			require.NoError(t, err)

			_, err = SnapshotFromFS(zaptest.NewLogger(t), dir)
			if !assert.Equal(t, test.err, err) {
				fmt.Println(err)
			}
		})
	}
}

func TestWalkDocuments(t *testing.T) {
	for _, test := range []struct {
		path  string
		count int
	}{
		{
			path:  "testdata/valid/explicit_index",
			count: 2,
		},
		{
			path:  "testdata/valid/implicit_index",
			count: 6,
		},
		{
			path:  "testdata/valid/exclude_index",
			count: 1,
		},
	} {
		t.Run(test.path, func(t *testing.T) {
			src, err := fs.Sub(testdata, test.path)
			require.NoError(t, err)

			var docs []*ext.Document
			require.NoError(t, WalkDocuments(zaptest.NewLogger(t), src, func(d *ext.Document) error {
				docs = append(docs, d)
				return nil
			}))

			assert.Len(t, docs, test.count)
		})
	}
}

func TestFSWithIndex(t *testing.T) {
	fwi, _ := fs.Sub(testdata, "testdata/valid/explicit_index")

	ss, err := SnapshotFromFS(zaptest.NewLogger(t), fwi)
	require.NoError(t, err)

	tfs := &FSIndexSuite{
		store: ss,
	}

	suite.Run(t, tfs)
}

type FSIndexSuite struct {
	suite.Suite
	store storage.ReadOnlyStore
}

func (fis *FSIndexSuite) TestCountFlag() {
	t := fis.T()

	flagCount, err := fis.store.CountFlags(context.TODO(), storage.NewNamespace("production"))
	require.NoError(t, err)

	assert.Equal(t, 12, int(flagCount))

	flagCount, err = fis.store.CountFlags(context.TODO(), storage.NewNamespace("sandbox"))
	require.NoError(t, err)
	assert.Equal(t, 12, int(flagCount))
}

func (fis *FSIndexSuite) TestGetFlag() {
	t := fis.T()

	testCases := []struct {
		name      string
		namespace string
		flagKey   string
		flag      *core.Flag
	}{
		{
			name:      "Production",
			namespace: "production",
			flagKey:   "prod-flag",
			flag: &core.Flag{
				Key:         "prod-flag",
				Name:        "Prod Flag",
				Description: "description",
				Enabled:     true,
				Variants: []*core.Variant{
					{
						Key:  "prod-variant",
						Name: "Prod Variant",
					},
					{
						Key: "foo",
					},
				},
			},
		},
		{
			name:      "Sandbox",
			namespace: "sandbox",
			flagKey:   "sandbox-flag",
			flag: &core.Flag{
				Key:         "sandbox-flag",
				Name:        "Sandbox Flag",
				Description: "description",
				Enabled:     true,
				Variants: []*core.Variant{
					{
						Key:  "sandbox-variant",
						Name: "Sandbox Variant",
					},
					{
						Key: "foo",
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			flag, err := fis.store.GetFlag(context.TODO(), storage.NewResource(tc.namespace, tc.flagKey))
			require.NoError(t, err)

			assert.Equal(t, tc.flag.Key, flag.Key)
			assert.Equal(t, tc.flag.Name, flag.Name)
			assert.Equal(t, tc.flag.Description, flag.Description)

			for i := 0; i < len(flag.Variants); i++ {
				v := tc.flag.Variants[i]
				fv := flag.Variants[i]
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
			flags, err := fis.store.ListFlags(context.TODO(), storage.ListWithOptions(
				storage.NewNamespace(tc.namespace),
				storage.ListWithQueryParamOptions[storage.NamespaceRequest](storage.WithLimit(5), storage.WithPageToken(tc.pageToken)),
			))
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
			rollouts, err := fis.store.GetEvaluationRollouts(context.TODO(), storage.NewResource(tc.namespace, tc.flagKey))
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
			assert.InDelta(t, float32(50), rollouts[1].Threshold.Percentage, 0)
			assert.True(t, rollouts[1].Threshold.Value, "threshold value should be true")
		})
	}
}

func (fis *FSIndexSuite) TestGetEvaluationRules() {
	t := fis.T()

	testCases := []struct {
		name      string
		namespace string
		flagKey   string
		segments  map[string]*storage.EvaluationSegment
	}{
		{
			name:      "Production",
			namespace: "production",
			flagKey:   "prod-flag",
			segments: map[string]*storage.EvaluationSegment{
				"segment1": {
					SegmentKey: "segment1",
					MatchType:  core.MatchType_ANY_MATCH_TYPE,
					Constraints: []storage.EvaluationConstraint{
						{
							Property: "foo",
							Operator: "eq",
							Value:    "baz",
							Type:     core.ComparisonType_STRING_COMPARISON_TYPE,
						},
						{
							Property: "fizz",
							Operator: "neq",
							Value:    "buzz",
							Type:     core.ComparisonType_STRING_COMPARISON_TYPE,
						},
					},
				},
			},
		},
		{
			name:      "Sandbox",
			namespace: "sandbox",
			flagKey:   "sandbox-flag",
			segments: map[string]*storage.EvaluationSegment{
				"segment1": {
					SegmentKey: "segment1",
					MatchType:  core.MatchType_ANY_MATCH_TYPE,
					Constraints: []storage.EvaluationConstraint{
						{
							Property: "foo",
							Operator: "eq",
							Value:    "baz",
							Type:     core.ComparisonType_STRING_COMPARISON_TYPE,
						},
						{
							Property: "fizz",
							Operator: "neq",
							Value:    "buzz",
							Type:     core.ComparisonType_STRING_COMPARISON_TYPE,
						},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rules, err := fis.store.GetEvaluationRules(context.TODO(), storage.NewResource(tc.namespace, tc.flagKey))
			require.NoError(t, err)

			assert.Len(t, rules, 1)

			assert.Equal(t, tc.namespace, rules[0].NamespaceKey)
			assert.Equal(t, tc.flagKey, rules[0].FlagKey)
			assert.Equal(t, int32(1), rules[0].Rank)
			assert.Contains(t, rules[0].Segments, "segment1")
			assert.Equal(t, tc.segments, rules[0].Segments)
		})
	}
}

func (fis *FSIndexSuite) TestGetEvaluationDistributions() {
	t := fis.T()

	testCases := []struct {
		name      string
		namespace string
		flagKey   string
		count     int
	}{
		{
			name:      "Sandbox",
			namespace: "sandbox",
			flagKey:   "sandbox-flag",
			count:     1,
		},
		{
			name:      "Production",
			namespace: "production",
			flagKey:   "prod-flag",
			count:     1,
		},
		{
			name:      "Production No Distributions",
			namespace: "production",
			flagKey:   "no-distributions",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rules, err := fis.store.GetEvaluationRules(context.TODO(), storage.NewResource(tc.namespace, tc.flagKey))
			require.NoError(t, err)
			assert.Len(t, rules, 1)

			dist, err := fis.store.GetEvaluationDistributions(context.TODO(), storage.NewResource(tc.namespace, tc.flagKey), storage.NewID(rules[0].ID))

			require.NoError(t, err)
			assert.Len(t, dist, tc.count)
		})
	}
}

type FSWithoutIndexSuite struct {
	suite.Suite
	store storage.ReadOnlyStore
}

func TestFSWithoutIndex(t *testing.T) {
	fwoi, _ := fs.Sub(testdata, "testdata/valid/implicit_index")

	ss, err := SnapshotFromFS(zaptest.NewLogger(t), fwoi)
	require.NoError(t, err)

	tfs := &FSWithoutIndexSuite{
		store: ss,
	}

	suite.Run(t, tfs)
}

func (fis *FSWithoutIndexSuite) TestCountFlag() {
	t := fis.T()

	flagCount, err := fis.store.CountFlags(context.TODO(), storage.NewNamespace("production"))
	require.NoError(t, err)

	assert.Equal(t, 14, int(flagCount))

	flagCount, err = fis.store.CountFlags(context.TODO(), storage.NewNamespace("sandbox"))
	require.NoError(t, err)
	assert.Equal(t, 14, int(flagCount))

	flagCount, err = fis.store.CountFlags(context.TODO(), storage.NewNamespace("staging"))
	require.NoError(t, err)
	assert.Equal(t, 14, int(flagCount))
}

func (fis *FSWithoutIndexSuite) TestGetFlag() {
	t := fis.T()

	testCases := []struct {
		name      string
		namespace string
		flagKey   string
		flag      *core.Flag
	}{
		{
			name:      "Production",
			namespace: "production",
			flagKey:   "prod-flag",
			flag: &core.Flag{
				Key:         "prod-flag",
				Name:        "Prod Flag",
				Description: "description",
				Enabled:     true,
				Variants: []*core.Variant{
					{
						Key:  "prod-variant",
						Name: "Prod Variant",
					},
					{
						Key: "foo",
					},
				},
			},
		},
		{
			name:      "Production One",
			namespace: "production",
			flagKey:   "prod-flag-1",
			flag: &core.Flag{
				Key:         "prod-flag-1",
				Name:        "Prod Flag 1",
				Description: "description",
				Enabled:     true,
				Variants: []*core.Variant{
					{
						Key:  "prod-variant",
						Name: "Prod Variant",
					},
					{
						Key: "foo",
					},
				},
			},
		},
		{
			name:      "Sandbox",
			namespace: "sandbox",
			flagKey:   "sandbox-flag",
			flag: &core.Flag{
				Key:         "sandbox-flag",
				Name:        "Sandbox Flag",
				Description: "description",
				Enabled:     true,
				Variants: []*core.Variant{
					{
						Key:  "sandbox-variant",
						Name: "Sandbox Variant",
					},
					{
						Key: "foo",
					},
				},
			},
		},
		{
			name:      "Sandbox One",
			namespace: "sandbox",
			flagKey:   "sandbox-flag-1",
			flag: &core.Flag{
				Key:         "sandbox-flag-1",
				Name:        "Sandbox Flag 1",
				Description: "description",
				Enabled:     true,
				Variants: []*core.Variant{
					{
						Key:  "sandbox-variant",
						Name: "Sandbox Variant",
					},
					{
						Key: "foo",
					},
				},
			},
		},
		{
			name:      "Staging",
			namespace: "staging",
			flagKey:   "staging-flag",
			flag: &core.Flag{
				Key:         "staging-flag",
				Name:        "Staging Flag",
				Description: "description",
				Enabled:     true,
				Variants: []*core.Variant{
					{
						Key:  "staging-variant",
						Name: "Staging Variant",
					},
					{
						Key: "foo",
					},
				},
			},
		},
		{
			name:      "Staging One",
			namespace: "staging",
			flagKey:   "staging-flag-1",
			flag: &core.Flag{
				Key:         "staging-flag-1",
				Name:        "Staging Flag 1",
				Description: "description",
				Enabled:     true,
				Variants: []*core.Variant{
					{
						Key:  "staging-variant",
						Name: "Staging Variant",
					},
					{
						Key: "foo",
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			flag, err := fis.store.GetFlag(context.TODO(), storage.NewResource(tc.namespace, tc.flagKey))
			require.NoError(t, err)

			assert.Equal(t, tc.flag.Key, flag.Key)
			assert.Equal(t, tc.flag.Name, flag.Name)
			assert.Equal(t, tc.flag.Description, flag.Description)

			for i := 0; i < len(flag.Variants); i++ {
				v := tc.flag.Variants[i]
				fv := flag.Variants[i]

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
			flags, err := fis.store.ListFlags(context.TODO(), storage.ListWithOptions(storage.NewNamespace(tc.namespace),
				storage.ListWithQueryParamOptions[storage.NamespaceRequest](storage.WithLimit(5), storage.WithPageToken(tc.pageToken))))
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
			rollouts, err := fis.store.GetEvaluationRollouts(context.TODO(), storage.NewResource(tc.namespace, tc.flagKey))
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
			assert.InDelta(t, float32(50), rollouts[1].Threshold.Percentage, 0.0)
			assert.True(t, rollouts[1].Threshold.Value, "threshold value should be true")
		})
	}
}

func (fis *FSWithoutIndexSuite) TestGetEvaluationRules() {
	t := fis.T()

	testCases := []struct {
		name      string
		namespace string
		flagKey   string
		segments  map[string]*storage.EvaluationSegment
	}{
		{
			name:      "Production",
			namespace: "production",
			flagKey:   "prod-flag",
			segments: map[string]*storage.EvaluationSegment{
				"segment1": {
					SegmentKey: "segment1",
					MatchType:  core.MatchType_ANY_MATCH_TYPE,
					Constraints: []storage.EvaluationConstraint{
						{
							Property: "foo",
							Operator: "eq",
							Value:    "baz",
							Type:     core.ComparisonType_STRING_COMPARISON_TYPE,
						},
						{
							Property: "fizz",
							Operator: "neq",
							Value:    "buzz",
							Type:     core.ComparisonType_STRING_COMPARISON_TYPE,
						},
					},
				},
			},
		},
		{
			name:      "Sandbox",
			namespace: "sandbox",
			flagKey:   "sandbox-flag",
			segments: map[string]*storage.EvaluationSegment{
				"segment1": {
					SegmentKey: "segment1",
					MatchType:  core.MatchType_ANY_MATCH_TYPE,
					Constraints: []storage.EvaluationConstraint{
						{
							Property: "foo",
							Operator: "eq",
							Value:    "baz",
							Type:     core.ComparisonType_STRING_COMPARISON_TYPE,
						},
						{
							Property: "fizz",
							Operator: "neq",
							Value:    "buzz",
							Type:     core.ComparisonType_STRING_COMPARISON_TYPE,
						},
					},
				},
			},
		},
		{
			name:      "Staging",
			namespace: "staging",
			flagKey:   "staging-flag",
			segments: map[string]*storage.EvaluationSegment{
				"segment1": {
					SegmentKey: "segment1",
					MatchType:  core.MatchType_ANY_MATCH_TYPE,
					Constraints: []storage.EvaluationConstraint{
						{
							Property: "foo",
							Operator: "eq",
							Value:    "baz",
							Type:     core.ComparisonType_STRING_COMPARISON_TYPE,
						},
						{
							Property: "fizz",
							Operator: "neq",
							Value:    "buzz",
							Type:     core.ComparisonType_STRING_COMPARISON_TYPE,
						},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rules, err := fis.store.GetEvaluationRules(context.TODO(), storage.NewResource(tc.namespace, tc.flagKey))
			require.NoError(t, err)

			assert.Len(t, rules, 1)

			assert.Equal(t, tc.namespace, rules[0].NamespaceKey)
			assert.Equal(t, tc.flagKey, rules[0].FlagKey)
			assert.Equal(t, int32(1), rules[0].Rank)
			assert.Equal(t, tc.segments, rules[0].Segments)
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
			rules, err := fis.store.GetEvaluationRules(context.TODO(), storage.NewResource(tc.namespace, tc.flagKey))
			require.NoError(t, err)
			assert.Len(t, rules, 1)

			dist, err := fis.store.GetEvaluationDistributions(context.TODO(), storage.NewResource(tc.namespace, tc.flagKey), storage.NewID(rules[0].ID))

			require.NoError(t, err)

			assert.Equal(t, tc.expectedVariantName, dist[0].VariantKey)
			assert.InDelta(t, float32(100), dist[0].Rollout, 0)
		})
	}
}

func TestFS_Empty_Features_File(t *testing.T) {
	fs, _ := fs.Sub(testdata, "testdata/valid/empty_features")
	ss, err := SnapshotFromFS(zaptest.NewLogger(t), fs)
	require.NoError(t, err)

	defaultList := storage.ListWithOptions(storage.NewNamespace(storage.DefaultNamespace))
	_, err = ss.ListFlags(context.TODO(), defaultList)
	require.NoError(t, err)
}

func TestFS_YAML_Stream(t *testing.T) {
	fs, _ := fs.Sub(testdata, "testdata/valid/yaml_stream")
	ss, err := SnapshotFromFS(zaptest.NewLogger(t), fs)
	require.NoError(t, err)

	var (
		listFootball = storage.ListWithOptions(storage.NewNamespace("football"))
		listFruit    = storage.ListWithOptions(storage.NewNamespace("fruit"))
	)
	fflags, err := ss.ListFlags(context.TODO(), listFootball)
	require.NoError(t, err)

	assert.Len(t, fflags.Results, 1)
	assert.Equal(t, "team", fflags.Results[0].Key)

	frflags, err := ss.ListFlags(context.TODO(), listFruit)
	require.NoError(t, err)

	assert.Len(t, frflags.Results, 1)
	assert.Equal(t, "apple", frflags.Results[0].Key)
}
