package ext

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestConstraint_MarshalYAML(t *testing.T) {
	tests := []struct {
		name       string
		constraint Constraint
		want       string
	}{
		{
			name: "isoneof string array expands to list",
			constraint: Constraint{
				Type:     "STRING_COMPARISON_TYPE",
				Property: "org",
				Operator: "isoneof",
				Value:    `["org-a","org-b","org-c"]`,
			},
			want: "type: STRING_COMPARISON_TYPE\nproperty: org\noperator: isoneof\nvalue:\n    - org-a\n    - org-b\n    - org-c\n",
		},
		{
			name: "isoneof number array expands to list",
			constraint: Constraint{
				Type:     "NUMBER_COMPARISON_TYPE",
				Property: "age",
				Operator: "isoneof",
				Value:    `[18,21,65]`,
			},
			want: "type: NUMBER_COMPARISON_TYPE\nproperty: age\noperator: isoneof\nvalue:\n    - 18\n    - 21\n    - 65\n",
		},
		{
			name: "eq operator stays as string",
			constraint: Constraint{
				Type:     "STRING_COMPARISON_TYPE",
				Property: "env",
				Operator: "eq",
				Value:    "production",
			},
			want: "type: STRING_COMPARISON_TYPE\nproperty: env\noperator: eq\nvalue: production\n",
		},
		{
			name: "isoneof with invalid JSON produces empty list",
			constraint: Constraint{
				Type:     "STRING_COMPARISON_TYPE",
				Property: "org",
				Operator: "isoneof",
				Value:    "not-json",
			},
			want: "type: STRING_COMPARISON_TYPE\nproperty: org\noperator: isoneof\nvalue: []\n",
		},
		{
			name: "isoneof with empty value produces empty list",
			constraint: Constraint{
				Type:     "STRING_COMPARISON_TYPE",
				Property: "org",
				Operator: "isoneof",
			},
			want: "type: STRING_COMPARISON_TYPE\nproperty: org\noperator: isoneof\nvalue: []\n",
		},
		{
			name: "isoneof with empty JSON array produces empty list",
			constraint: Constraint{
				Type:     "STRING_COMPARISON_TYPE",
				Property: "org",
				Operator: "isoneof",
				Value:    "[]",
			},
			want: "type: STRING_COMPARISON_TYPE\nproperty: org\noperator: isoneof\nvalue: []\n",
		},
		{
			name: "isoneof sorts values alphabetically",
			constraint: Constraint{
				Type:     "STRING_COMPARISON_TYPE",
				Property: "org",
				Operator: "isoneof",
				Value:    `["zebra","apple","mango"]`,
			},
			want: "type: STRING_COMPARISON_TYPE\nproperty: org\noperator: isoneof\nvalue:\n    - apple\n    - mango\n    - zebra\n",
		},
		{
			name: "isoneof with numeric-looking strings preserves quoting",
			constraint: Constraint{
				Type:     "STRING_COMPARISON_TYPE",
				Property: "code",
				Operator: "isoneof",
				Value:    `["123","456"]`,
			},
			want: "type: STRING_COMPARISON_TYPE\nproperty: code\noperator: isoneof\nvalue:\n    - \"123\"\n    - \"456\"\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := yaml.Marshal(&tt.constraint)
			require.NoError(t, err)
			assert.Equal(t, tt.want, string(data))
		})
	}
}

func TestConstraint_UnmarshalYAML(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  Constraint
	}{
		{
			name:  "YAML list of strings becomes JSON array",
			input: "type: STRING_COMPARISON_TYPE\nproperty: org\noperator: isoneof\nvalue:\n  - org-a\n  - org-b\n",
			want: Constraint{
				Type:     "STRING_COMPARISON_TYPE",
				Property: "org",
				Operator: "isoneof",
				Value:    `["org-a","org-b"]`,
			},
		},
		{
			name:  "YAML list of numbers becomes JSON array",
			input: "type: NUMBER_COMPARISON_TYPE\nproperty: age\noperator: isoneof\nvalue:\n  - 18\n  - 21\n",
			want: Constraint{
				Type:     "NUMBER_COMPARISON_TYPE",
				Property: "age",
				Operator: "isoneof",
				Value:    `[18,21]`,
			},
		},
		{
			name:  "plain string value preserved",
			input: "type: STRING_COMPARISON_TYPE\nproperty: env\noperator: eq\nvalue: production\n",
			want: Constraint{
				Type:     "STRING_COMPARISON_TYPE",
				Property: "env",
				Operator: "eq",
				Value:    "production",
			},
		},
		{
			name:  "old JSON string format still works",
			input: "type: STRING_COMPARISON_TYPE\nproperty: org\noperator: isoneof\nvalue: '[\"org-a\",\"org-b\"]'\n",
			want: Constraint{
				Type:     "STRING_COMPARISON_TYPE",
				Property: "org",
				Operator: "isoneof",
				Value:    `["org-a","org-b"]`,
			},
		},
		{
			name:  "missing value gives empty string",
			input: "type: STRING_COMPARISON_TYPE\nproperty: org\noperator: empty\n",
			want: Constraint{
				Type:     "STRING_COMPARISON_TYPE",
				Property: "org",
				Operator: "empty",
			},
		},
		{
			name:  "numeric scalar value",
			input: "type: NUMBER_COMPARISON_TYPE\nproperty: count\noperator: eq\nvalue: 42\n",
			want: Constraint{
				Type:     "NUMBER_COMPARISON_TYPE",
				Property: "count",
				Operator: "eq",
				Value:    "42",
			},
		},
		{
			name:  "boolean scalar value",
			input: "type: BOOLEAN_COMPARISON_TYPE\nproperty: active\noperator: true\nvalue: true\n",
			want: Constraint{
				Type:     "BOOLEAN_COMPARISON_TYPE",
				Property: "active",
				Operator: "true",
				Value:    "true",
			},
		},
		{
			name:  "YAML list of floats",
			input: "type: NUMBER_COMPARISON_TYPE\nproperty: score\noperator: isoneof\nvalue:\n  - 1.5\n  - 2.5\n",
			want: Constraint{
				Type:     "NUMBER_COMPARISON_TYPE",
				Property: "score",
				Operator: "isoneof",
				Value:    `[1.5,2.5]`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got Constraint
			err := yaml.Unmarshal([]byte(tt.input), &got)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestConstraint_YAML_RoundTrip(t *testing.T) {
	tests := []struct {
		name       string
		constraint Constraint
	}{
		{
			name: "isoneof string array",
			constraint: Constraint{
				Type:     "STRING_COMPARISON_TYPE",
				Property: "org",
				Operator: "isoneof",
				Value:    `["org-a","org-b","org-c"]`,
			},
		},
		{
			name: "isnotoneof string array",
			constraint: Constraint{
				Type:        "STRING_COMPARISON_TYPE",
				Property:    "country",
				Operator:    "isnotoneof",
				Value:       `["CA","UK","US"]`,
				Description: "excluded countries",
			},
		},
		{
			name: "isoneof number array",
			constraint: Constraint{
				Type:     "NUMBER_COMPARISON_TYPE",
				Property: "age",
				Operator: "isoneof",
				Value:    `[18,21,65]`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := yaml.Marshal(&tt.constraint)
			require.NoError(t, err)

			var got Constraint
			err = yaml.Unmarshal(data, &got)
			require.NoError(t, err)

			assert.Equal(t, tt.constraint, got)
		})
	}
}

func TestConstraint_InDocument_YAML_RoundTrip(t *testing.T) {
	doc := &Document{
		Version:   "1.5",
		Namespace: DefaultNamespace,
		Segments: []*Segment{
			{
				Key:       "test-segment",
				Name:      "Test Segment",
				MatchType: "ANY_MATCH_TYPE",
				Constraints: []*Constraint{
					{
						Type:     "STRING_COMPARISON_TYPE",
						Property: "org",
						Operator: "isoneof",
						Value:    `["org-a","org-b","org-c"]`,
					},
					{
						Type:     "STRING_COMPARISON_TYPE",
						Property: "env",
						Operator: "eq",
						Value:    "production",
					},
					{
						Type:     "NUMBER_COMPARISON_TYPE",
						Property: "age",
						Operator: "isoneof",
						Value:    `[18,21,65]`,
					},
				},
			},
		},
	}

	data, err := yaml.Marshal(doc)
	require.NoError(t, err)

	var got Document
	err = yaml.Unmarshal(data, &got)
	require.NoError(t, err)

	require.Len(t, got.Segments, 1)
	require.Len(t, got.Segments[0].Constraints, 3)

	assert.Equal(t, `["org-a","org-b","org-c"]`, got.Segments[0].Constraints[0].Value)
	assert.Equal(t, "isoneof", got.Segments[0].Constraints[0].Operator)

	assert.Equal(t, "production", got.Segments[0].Constraints[1].Value)
	assert.Equal(t, "eq", got.Segments[0].Constraints[1].Operator)

	assert.Equal(t, `[18,21,65]`, got.Segments[0].Constraints[2].Value)
	assert.Equal(t, "isoneof", got.Segments[0].Constraints[2].Operator)
}

func TestConstraint_OldFormatNormalizesToList_InDocument(t *testing.T) {
	old := `version: "1.5"
namespace: default
segments:
  - key: seg1
    name: Segment 1
    match_type: ANY_MATCH_TYPE
    constraints:
      - type: STRING_COMPARISON_TYPE
        property: org
        operator: isoneof
        value: '["org-a","org-b"]'
      - type: NUMBER_COMPARISON_TYPE
        property: age
        operator: isoneof
        value: '[18,21]'
      - type: STRING_COMPARISON_TYPE
        property: env
        operator: eq
        value: production
`
	var doc Document
	err := yaml.Unmarshal([]byte(old), &doc)
	require.NoError(t, err)

	data, err := yaml.Marshal(&doc)
	require.NoError(t, err)

	out := string(data)
	assert.Contains(t, out, "value:\n            - org-a\n            - org-b\n")
	assert.Contains(t, out, "value:\n            - 18\n            - 21\n")
	assert.Contains(t, out, "value: production\n")
	assert.NotContains(t, out, `'["org-a"`)
	assert.NotContains(t, out, `'[18`)
}

func TestConstraint_UnmarshalYAML_BareYAMLSpecialValues(t *testing.T) {
	input := `type: STRING_COMPARISON_TYPE
property: tag
operator: isoneof
value:
  - "NO"
  - "YES"
  - true
  - false
  - "null"
  - production
`
	var got Constraint
	err := yaml.Unmarshal([]byte(input), &got)
	require.NoError(t, err)
	assert.Equal(t, `["NO","YES","true","false","null","production"]`, got.Value)
}

func TestConstraint_UnmarshalYAML_MixedFormats(t *testing.T) {
	input := `- type: STRING_COMPARISON_TYPE
  property: org
  operator: isoneof
  value: '["org-a","org-b"]'
- type: STRING_COMPARISON_TYPE
  property: country
  operator: isnotoneof
  value:
    - US
    - UK
- type: ENTITY_ID_COMPARISON_TYPE
  property: user_id
  operator: isoneof
  value:
    - user-1
    - user-2
- type: STRING_COMPARISON_TYPE
  property: env
  operator: eq
  value: production
`
	var got []*Constraint
	err := yaml.Unmarshal([]byte(input), &got)
	require.NoError(t, err)
	require.Len(t, got, 4)

	assert.Equal(t, `["org-a","org-b"]`, got[0].Value, "old JSON-string format")
	assert.Equal(t, `["US","UK"]`, got[1].Value, "new YAML-list format")
	assert.Equal(t, `["user-1","user-2"]`, got[2].Value, "entity ID list format")
	assert.Equal(t, "production", got[3].Value, "plain string")
}

func TestDistribution_ZeroRollout_YAML(t *testing.T) {
	tests := []struct {
		name     string
		rollout  float32
		expected string
	}{
		{
			name:     "zero rollout",
			rollout:  0,
			expected: "variant: test\nrollout: 0\n",
		},
		{
			name:     "zero float rollout",
			rollout:  0.0,
			expected: "variant: test\nrollout: 0\n",
		},
		{
			name:     "non-zero rollout",
			rollout:  50.0,
			expected: "variant: test\nrollout: 50\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dist := &Distribution{
				VariantKey: "test",
				Rollout:    tt.rollout,
			}

			data, err := yaml.Marshal(dist)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, string(data))

			// Test unmarshaling back
			var unmarshaled Distribution
			err = yaml.Unmarshal(data, &unmarshaled)
			require.NoError(t, err)
			assert.Equal(t, dist.VariantKey, unmarshaled.VariantKey)
			assert.InDelta(t, dist.Rollout, unmarshaled.Rollout, 0.001)
		})
	}
}

func TestDistribution_ZeroRollout_JSON(t *testing.T) {
	tests := []struct {
		name     string
		rollout  float32
		expected string
	}{
		{
			name:     "zero rollout",
			rollout:  0,
			expected: `{"variant":"test","rollout":0}`,
		},
		{
			name:     "zero float rollout",
			rollout:  0.0,
			expected: `{"variant":"test","rollout":0}`,
		},
		{
			name:     "non-zero rollout",
			rollout:  50.0,
			expected: `{"variant":"test","rollout":50}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dist := &Distribution{
				VariantKey: "test",
				Rollout:    tt.rollout,
			}

			data, err := json.Marshal(dist)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, string(data))

			// Test unmarshaling back
			var unmarshaled Distribution
			err = json.Unmarshal(data, &unmarshaled)
			require.NoError(t, err)
			assert.Equal(t, dist.VariantKey, unmarshaled.VariantKey)
			assert.InDelta(t, dist.Rollout, unmarshaled.Rollout, 0.001)
		})
	}
}

func TestThresholdRule_ZeroPercentage_YAML(t *testing.T) {
	tests := []struct {
		name       string
		percentage float32
		value      bool
		expected   string
	}{
		{
			name:       "zero percentage with true value",
			percentage: 0,
			value:      true,
			expected:   "percentage: 0\nvalue: true\n",
		},
		{
			name:       "zero float percentage with false value",
			percentage: 0.0,
			value:      false,
			expected:   "percentage: 0\n",
		},
		{
			name:       "non-zero percentage",
			percentage: 50.0,
			value:      true,
			expected:   "percentage: 50\nvalue: true\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			threshold := &ThresholdRule{
				Percentage: tt.percentage,
				Value:      tt.value,
			}

			data, err := yaml.Marshal(threshold)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, string(data))

			// Test unmarshaling back
			var unmarshaled ThresholdRule
			err = yaml.Unmarshal(data, &unmarshaled)
			require.NoError(t, err)
			assert.InDelta(t, threshold.Percentage, unmarshaled.Percentage, 0.001)
			assert.Equal(t, threshold.Value, unmarshaled.Value)
		})
	}
}

func TestConstraint_MarshalJSON(t *testing.T) {
	tests := []struct {
		name       string
		constraint Constraint
		want       string
	}{
		{
			name: "isoneof string array expands to list",
			constraint: Constraint{
				Type:     "STRING_COMPARISON_TYPE",
				Property: "org",
				Operator: "isoneof",
				Value:    `["org-a","org-b","org-c"]`,
			},
			want: `{"type":"STRING_COMPARISON_TYPE","property":"org","operator":"isoneof","value":["org-a","org-b","org-c"]}`,
		},
		{
			name: "isoneof number array expands to list",
			constraint: Constraint{
				Type:     "NUMBER_COMPARISON_TYPE",
				Property: "age",
				Operator: "isoneof",
				Value:    `[18,21,65]`,
			},
			want: `{"type":"NUMBER_COMPARISON_TYPE","property":"age","operator":"isoneof","value":[18,21,65]}`,
		},
		{
			name: "eq operator stays as string",
			constraint: Constraint{
				Type:     "STRING_COMPARISON_TYPE",
				Property: "env",
				Operator: "eq",
				Value:    "production",
			},
			want: `{"type":"STRING_COMPARISON_TYPE","property":"env","operator":"eq","value":"production"}`,
		},
		{
			name: "isoneof with invalid JSON produces empty list",
			constraint: Constraint{
				Type:     "STRING_COMPARISON_TYPE",
				Property: "org",
				Operator: "isoneof",
				Value:    "not-json",
			},
			want: `{"type":"STRING_COMPARISON_TYPE","property":"org","operator":"isoneof","value":[]}`,
		},
		{
			name: "isoneof with empty value produces empty list",
			constraint: Constraint{
				Type:     "STRING_COMPARISON_TYPE",
				Property: "org",
				Operator: "isoneof",
			},
			want: `{"type":"STRING_COMPARISON_TYPE","property":"org","operator":"isoneof","value":[]}`,
		},
		{
			name: "isoneof with empty JSON array produces empty list",
			constraint: Constraint{
				Type:     "STRING_COMPARISON_TYPE",
				Property: "org",
				Operator: "isoneof",
				Value:    "[]",
			},
			want: `{"type":"STRING_COMPARISON_TYPE","property":"org","operator":"isoneof","value":[]}`,
		},
		{
			name: "isoneof sorts values alphabetically",
			constraint: Constraint{
				Type:     "STRING_COMPARISON_TYPE",
				Property: "org",
				Operator: "isoneof",
				Value:    `["zebra","apple","mango"]`,
			},
			want: `{"type":"STRING_COMPARISON_TYPE","property":"org","operator":"isoneof","value":["apple","mango","zebra"]}`,
		},
		{
			name: "isoneof with numeric-looking strings preserves quoting",
			constraint: Constraint{
				Type:     "STRING_COMPARISON_TYPE",
				Property: "code",
				Operator: "isoneof",
				Value:    `["123","456"]`,
			},
			want: `{"type":"STRING_COMPARISON_TYPE","property":"code","operator":"isoneof","value":["123","456"]}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(&tt.constraint)
			require.NoError(t, err)
			assert.Equal(t, tt.want, string(data))
		})
	}
}

func TestConstraint_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  Constraint
	}{
		{
			name:  "JSON array of strings becomes JSON array",
			input: `{"type":"STRING_COMPARISON_TYPE","property":"org","operator":"isoneof","value":["org-a","org-b"]}`,
			want: Constraint{
				Type:     "STRING_COMPARISON_TYPE",
				Property: "org",
				Operator: "isoneof",
				Value:    `["org-a","org-b"]`,
			},
		},
		{
			name:  "JSON array of numbers becomes JSON array",
			input: `{"type":"NUMBER_COMPARISON_TYPE","property":"age","operator":"isoneof","value":[18,21]}`,
			want: Constraint{
				Type:     "NUMBER_COMPARISON_TYPE",
				Property: "age",
				Operator: "isoneof",
				Value:    `[18,21]`,
			},
		},
		{
			name:  "plain string value preserved",
			input: `{"type":"STRING_COMPARISON_TYPE","property":"env","operator":"eq","value":"production"}`,
			want: Constraint{
				Type:     "STRING_COMPARISON_TYPE",
				Property: "env",
				Operator: "eq",
				Value:    "production",
			},
		},
		{
			name:  "old JSON string format still works",
			input: `{"type":"STRING_COMPARISON_TYPE","property":"org","operator":"isoneof","value":"[\"org-a\",\"org-b\"]"}`,
			want: Constraint{
				Type:     "STRING_COMPARISON_TYPE",
				Property: "org",
				Operator: "isoneof",
				Value:    `["org-a","org-b"]`,
			},
		},
		{
			name:  "missing value gives empty string",
			input: `{"type":"STRING_COMPARISON_TYPE","property":"org","operator":"empty"}`,
			want: Constraint{
				Type:     "STRING_COMPARISON_TYPE",
				Property: "org",
				Operator: "empty",
			},
		},
		{
			name:  "numeric scalar value",
			input: `{"type":"NUMBER_COMPARISON_TYPE","property":"count","operator":"eq","value":42}`,
			want: Constraint{
				Type:     "NUMBER_COMPARISON_TYPE",
				Property: "count",
				Operator: "eq",
				Value:    "42",
			},
		},
		{
			name:  "boolean scalar value",
			input: `{"type":"BOOLEAN_COMPARISON_TYPE","property":"active","operator":"true","value":true}`,
			want: Constraint{
				Type:     "BOOLEAN_COMPARISON_TYPE",
				Property: "active",
				Operator: "true",
				Value:    "true",
			},
		},
		{
			name:  "null value gives empty string",
			input: `{"type":"STRING_COMPARISON_TYPE","property":"tag","operator":"empty","value":null}`,
			want: Constraint{
				Type:     "STRING_COMPARISON_TYPE",
				Property: "tag",
				Operator: "empty",
			},
		},
		{
			name:  "JSON array of floats",
			input: `{"type":"NUMBER_COMPARISON_TYPE","property":"score","operator":"isoneof","value":[1.5,2.5]}`,
			want: Constraint{
				Type:     "NUMBER_COMPARISON_TYPE",
				Property: "score",
				Operator: "isoneof",
				Value:    `[1.5,2.5]`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got Constraint
			err := json.Unmarshal([]byte(tt.input), &got)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestConstraint_JSON_RoundTrip(t *testing.T) {
	tests := []struct {
		name       string
		constraint Constraint
	}{
		{
			name: "isoneof string array",
			constraint: Constraint{
				Type:     "STRING_COMPARISON_TYPE",
				Property: "org",
				Operator: "isoneof",
				Value:    `["org-a","org-b","org-c"]`,
			},
		},
		{
			name: "isnotoneof string array",
			constraint: Constraint{
				Type:        "STRING_COMPARISON_TYPE",
				Property:    "country",
				Operator:    "isnotoneof",
				Value:       `["CA","UK","US"]`,
				Description: "excluded countries",
			},
		},
		{
			name: "isoneof number array",
			constraint: Constraint{
				Type:     "NUMBER_COMPARISON_TYPE",
				Property: "age",
				Operator: "isoneof",
				Value:    `[18,21,65]`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(&tt.constraint)
			require.NoError(t, err)

			var got Constraint
			err = json.Unmarshal(data, &got)
			require.NoError(t, err)

			assert.Equal(t, tt.constraint, got)
		})
	}
}

func TestLatestVersionString(t *testing.T) {
	got := LatestVersionString()
	want := "1.6"
	assert.Equal(t, want, got, "LatestVersionString should return %q", want)
}

func TestThresholdRule_ZeroPercentage_JSON(t *testing.T) {
	tests := []struct {
		name       string
		percentage float32
		value      bool
		expected   string
	}{
		{
			name:       "zero percentage with true value",
			percentage: 0,
			value:      true,
			expected:   `{"percentage":0,"value":true}`,
		},
		{
			name:       "zero float percentage with false value",
			percentage: 0.0,
			value:      false,
			expected:   `{"percentage":0}`,
		},
		{
			name:       "non-zero percentage",
			percentage: 50.0,
			value:      true,
			expected:   `{"percentage":50,"value":true}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			threshold := &ThresholdRule{
				Percentage: tt.percentage,
				Value:      tt.value,
			}

			data, err := json.Marshal(threshold)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, string(data))

			// Test unmarshaling back
			var unmarshaled ThresholdRule
			err = json.Unmarshal(data, &unmarshaled)
			require.NoError(t, err)
			assert.InDelta(t, threshold.Percentage, unmarshaled.Percentage, 0.001)
			assert.Equal(t, threshold.Value, unmarshaled.Value)
		})
	}
}
