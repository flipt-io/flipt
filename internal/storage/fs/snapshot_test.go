package fs

import (
	"context"
	"embed"
	"io"
	"io/fs"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.flipt.io/flipt/internal/storage"
	"go.flipt.io/flipt/rpc/flipt"
	"go.uber.org/zap"
)

//go:embed all:fixtures
var testdata embed.FS

func TestFSWithIndex(t *testing.T) {
	fwi, _ := fs.Sub(testdata, "fixtures/fswithindex")

	filenames, err := buildSnapshotHelper(zap.NewNop(), fwi)
	assert.NoError(t, err)

	expected := []string{
		"prod/prod.features.yml",
		"sandbox/sandbox.features.yaml",
	}
	assert.Len(t, filenames, 2)
	assert.ElementsMatch(t, filenames, expected)

	readers := make([]io.Reader, 0, 2)

	for _, f := range filenames {
		fr, err := fwi.Open(f)
		assert.NoError(t, err)

		readers = append(readers, fr)
	}

	ss, err := snapshotFromFS(readers...)
	assert.NoError(t, err)

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
	assert.NoError(t, err)

	assert.Equal(t, 11, int(flagCount))

	flagCount, err = fis.store.CountFlags(context.TODO(), "sandbox")
	assert.NoError(t, err)
	assert.Equal(t, 11, int(flagCount))
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
			assert.NoError(t, err)

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
	}{
		{
			name:      "Production",
			namespace: "production",
		},
		{
			name:      "Sandbox",
			namespace: "sandbox",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			flags, err := fis.store.ListFlags(context.TODO(), tc.namespace, storage.WithLimit(5), storage.WithPageToken("0"))
			assert.NoError(t, err)

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
			name:         "Production",
			namespaceKey: "production",
			fliptNs: &flipt.Namespace{
				Key:  "production",
				Name: "Production",
			},
		},
		{
			name:         "Sandbox",
			namespaceKey: "sandbox",
			fliptNs: &flipt.Namespace{
				Key:  "sandbox",
				Name: "Sandbox",
			},
		},
	}

	for _, tc := range testCases {
		ns, err := fis.store.GetNamespace(context.TODO(), tc.namespaceKey)
		assert.NoError(t, err)

		assert.Equal(t, tc.fliptNs.Key, ns.Key)
		assert.Equal(t, tc.fliptNs.Name, ns.Name)
	}
}

func (fis *FSIndexSuite) TestCountNamespaces() {
	t := fis.T()

	namespacesCount, err := fis.store.CountNamespaces(context.TODO())
	assert.NoError(t, err)

	assert.Equal(t, 3, int(namespacesCount))
}

func (fis *FSIndexSuite) TestListNamespaces() {
	t := fis.T()

	namespaces, err := fis.store.ListNamespaces(context.TODO(), storage.WithLimit(2), storage.WithPageToken("0"))
	assert.NoError(t, err)

	assert.Len(t, namespaces.Results, 2)
	assert.Equal(t, "2", namespaces.NextPageToken)
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
						NamespaceKey: "production",
					},
					{
						SegmentKey:   "segment1",
						Property:     "fizz",
						Operator:     "neq",
						Value:        "buzz",
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
						NamespaceKey: "sandbox",
					},
					{
						SegmentKey:   "segment1",
						Property:     "fizz",
						Operator:     "neq",
						Value:        "buzz",
						NamespaceKey: "sandbox",
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sgmt, err := fis.store.GetSegment(context.TODO(), tc.namespace, tc.segmentKey)
			assert.NoError(t, err)

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
	}{
		{
			name:      "Production",
			namespace: "production",
		},
		{
			name:      "Sandbox",
			namespace: "sandbox",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			flags, err := fis.store.ListSegments(context.TODO(), tc.namespace, storage.WithLimit(5), storage.WithPageToken("0"))
			assert.NoError(t, err)

			assert.Len(t, flags.Results, 5)
			assert.Equal(t, "5", flags.NextPageToken)
		})
	}
}

func (fis *FSIndexSuite) TestCountSegment() {
	t := fis.T()

	segmentCount, err := fis.store.CountSegments(context.TODO(), "production")
	assert.NoError(t, err)

	assert.Equal(t, 11, int(segmentCount))

	segmentCount, err = fis.store.CountSegments(context.TODO(), "sandbox")
	assert.NoError(t, err)
	assert.Equal(t, 11, int(segmentCount))
}

func (fis *FSIndexSuite) TestGetEvaluationRules() {
	t := fis.T()

	testCases := []struct {
		name      string
		namespace string
		flagKey   string
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
	}

	for _, tc := range testCases {
		dist, err := fis.store.GetEvaluationRules(context.TODO(), tc.namespace, tc.flagKey)
		assert.NoError(t, err)

		assert.Len(t, dist, 1)
	}
}

type FSWithoutIndexSuite struct {
	suite.Suite
	store storage.Store
}

func TestFSWithoutIndex(t *testing.T) {
	fwoi, _ := fs.Sub(testdata, "fixtures/fswithoutindex")
	filenames, err := buildSnapshotHelper(zap.NewNop(), fwoi)
	assert.NoError(t, err)

	expected := []string{
		"prod/prod.features.yaml",
		"prod/features.yml",
		"sandbox/sandbox.features.yml",
		"sandbox/features.yaml",
		"staging/staging.features.yaml",
		"staging/features.yml",
	}
	assert.Len(t, filenames, 6)
	assert.ElementsMatch(t, filenames, expected)

	readers := make([]io.Reader, 0, 6)

	for _, f := range filenames {
		fr, err := fwoi.Open(f)
		assert.NoError(t, err)

		readers = append(readers, fr)
	}

	ss, err := snapshotFromFS(readers...)
	assert.NoError(t, err)

	tfs := &FSWithoutIndexSuite{
		store: ss,
	}

	suite.Run(t, tfs)
}

func (fis *FSWithoutIndexSuite) TestCountFlag() {
	t := fis.T()

	flagCount, err := fis.store.CountFlags(context.TODO(), "production")
	assert.NoError(t, err)

	assert.Equal(t, 12, int(flagCount))

	flagCount, err = fis.store.CountFlags(context.TODO(), "sandbox")
	assert.NoError(t, err)
	assert.Equal(t, 12, int(flagCount))

	flagCount, err = fis.store.CountFlags(context.TODO(), "staging")
	assert.NoError(t, err)
	assert.Equal(t, 12, int(flagCount))
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
			assert.NoError(t, err)

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
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			flags, err := fis.store.ListFlags(context.TODO(), tc.namespace, storage.WithLimit(5), storage.WithPageToken("0"))
			assert.NoError(t, err)

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
			name:         "Production",
			namespaceKey: "production",
			fliptNs: &flipt.Namespace{
				Key:  "production",
				Name: "Production",
			},
		},
		{
			name:         "Sandbox",
			namespaceKey: "sandbox",
			fliptNs: &flipt.Namespace{
				Key:  "sandbox",
				Name: "Sandbox",
			},
		},
		{
			name:         "Staging",
			namespaceKey: "staging",
			fliptNs: &flipt.Namespace{
				Key:  "staging",
				Name: "Staging",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ns, err := fis.store.GetNamespace(context.TODO(), tc.namespaceKey)
			assert.NoError(t, err)

			assert.Equal(t, tc.fliptNs.Key, ns.Key)
			assert.Equal(t, tc.fliptNs.Name, ns.Name)
		})
	}
}

func (fis *FSWithoutIndexSuite) TestCountNamespaces() {
	t := fis.T()

	namespacesCount, err := fis.store.CountNamespaces(context.TODO())
	assert.NoError(t, err)

	assert.Equal(t, 4, int(namespacesCount))
}

func (fis *FSWithoutIndexSuite) TestListNamespaces() {
	t := fis.T()

	namespaces, err := fis.store.ListNamespaces(context.TODO(), storage.WithLimit(3), storage.WithPageToken("0"))
	assert.NoError(t, err)

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
						NamespaceKey: "production",
					},
					{
						SegmentKey:   "segment1",
						Property:     "fizz",
						Operator:     "neq",
						Value:        "buzz",
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
						NamespaceKey: "production",
					},
					{
						SegmentKey:   "segment2",
						Property:     "fizz",
						Operator:     "neq",
						Value:        "buzz",
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
						NamespaceKey: "sandbox",
					},
					{
						SegmentKey:   "segment1",
						Property:     "fizz",
						Operator:     "neq",
						Value:        "buzz",
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
						NamespaceKey: "sandbox",
					},
					{
						SegmentKey:   "segment2",
						Property:     "fizz",
						Operator:     "neq",
						Value:        "buzz",
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
						NamespaceKey: "staging",
					},
					{
						SegmentKey:   "segment1",
						Property:     "fizz",
						Operator:     "neq",
						Value:        "buzz",
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
						NamespaceKey: "staging",
					},
					{
						SegmentKey:   "segment2",
						Property:     "fizz",
						Operator:     "neq",
						Value:        "buzz",
						NamespaceKey: "staging",
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sgmt, err := fis.store.GetSegment(context.TODO(), tc.namespace, tc.segmentKey)
			assert.NoError(t, err)

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
				assert.Equal(t, c.NamespaceKey, fc.NamespaceKey)
			}
		})
	}
}

func (fis *FSWithoutIndexSuite) TestCountSegment() {
	t := fis.T()

	segmentCount, err := fis.store.CountSegments(context.TODO(), "production")
	assert.NoError(t, err)

	assert.Equal(t, 12, int(segmentCount))

	segmentCount, err = fis.store.CountSegments(context.TODO(), "sandbox")
	assert.NoError(t, err)
	assert.Equal(t, 12, int(segmentCount))

	segmentCount, err = fis.store.CountSegments(context.TODO(), "staging")
	assert.NoError(t, err)
	assert.Equal(t, 12, int(segmentCount))
}

func (fis *FSWithoutIndexSuite) TestListSegments() {
	t := fis.T()

	testCases := []struct {
		name      string
		namespace string
		pageToken string
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
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			flags, err := fis.store.ListSegments(context.TODO(), tc.namespace, storage.WithLimit(5), storage.WithPageToken("0"))
			assert.NoError(t, err)

			assert.Len(t, flags.Results, 5)
			assert.Equal(t, "5", flags.NextPageToken)
		})
	}
}

func (fis *FSWithoutIndexSuite) TestGetEvaluationRules() {
	t := fis.T()

	testCases := []struct {
		name      string
		namespace string
		flagKey   string
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
			name:      "Sandbox",
			namespace: "staging",
			flagKey:   "staging-flag",
		},
	}

	for _, tc := range testCases {
		dist, err := fis.store.GetEvaluationRules(context.TODO(), tc.namespace, tc.flagKey)
		assert.NoError(t, err)

		assert.Len(t, dist, 1)
	}
}
