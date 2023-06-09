package ext

import (
	"bytes"
	"context"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.flipt.io/flipt/internal/storage"
	"go.flipt.io/flipt/rpc/flipt"
)

type mockLister struct {
	flags   []*flipt.Flag
	flagErr error

	segments   []*flipt.Segment
	segmentErr error

	rules   []*flipt.Rule
	ruleErr error
}

func (m mockLister) ListFlags(_ context.Context, _ *flipt.ListFlagRequest) (*flipt.FlagList, error) {
	return &flipt.FlagList{
		Flags: m.flags,
	}, m.flagErr
}

func (m mockLister) ListRules(_ context.Context, _ *flipt.ListRuleRequest) (*flipt.RuleList, error) {
	return &flipt.RuleList{
		Rules: m.rules,
	}, m.ruleErr
}

func (m mockLister) ListSegments(_ context.Context, _ *flipt.ListSegmentRequest) (*flipt.SegmentList, error) {
	return &flipt.SegmentList{
		Segments: m.segments,
	}, m.segmentErr
}

func TestExport(t *testing.T) {
	lister := mockLister{
		flags: []*flipt.Flag{
			{
				Key:         "flag1",
				Name:        "flag1",
				Description: "description",
				Enabled:     true,
				Variants: []*flipt.Variant{
					{
						Id:   "1",
						Key:  "variant1",
						Name: "variant1",
						Attachment: `{
							"pi": 3.141,
							"happy": true,
							"name": "Niels",
							"nothing": null,
							"answer": {
							  "everything": 42
							},
							"list": [1, 0, 2],
							"object": {
							  "currency": "USD",
							  "value": 42.99
							}
						  }`,
					},
					{
						Id:  "2",
						Key: "foo",
					},
				},
			},
		},
		segments: []*flipt.Segment{
			{
				Key:         "segment1",
				Name:        "segment1",
				Description: "description",
				MatchType:   flipt.MatchType_ANY_MATCH_TYPE,
				Constraints: []*flipt.Constraint{
					{
						Id:          "1",
						Type:        flipt.ComparisonType_STRING_COMPARISON_TYPE,
						Property:    "foo",
						Operator:    "eq",
						Value:       "baz",
						Description: "desc",
					},
					{
						Id:          "2",
						Type:        flipt.ComparisonType_STRING_COMPARISON_TYPE,
						Property:    "fizz",
						Operator:    "neq",
						Value:       "buzz",
						Description: "desc",
					},
				},
			},
		},
		rules: []*flipt.Rule{
			{
				Id:         "1",
				SegmentKey: "segment1",
				Rank:       1,
				Distributions: []*flipt.Distribution{
					{
						Id:        "1",
						VariantId: "1",
						RuleId:    "1",
						Rollout:   100,
					},
				},
			},
		},
	}

	var (
		exporter = NewExporter(lister, storage.DefaultNamespace)
		b        = new(bytes.Buffer)
	)

	err := exporter.Export(context.Background(), b)
	assert.NoError(t, err)

	in, err := ioutil.ReadFile("testdata/export.yml")
	assert.NoError(t, err)

	assert.YAMLEq(t, string(in), b.String())
}
