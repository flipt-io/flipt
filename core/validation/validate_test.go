package validation

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidate_V1_Success(t *testing.T) {
	const file = "testdata/valid_v1.yaml"
	f, err := os.Open(file)
	require.NoError(t, err)

	defer f.Close()

	v, err := NewFeaturesValidator()
	require.NoError(t, err)

	err = v.Validate(file, f)
	assert.NoError(t, err)
}

func TestValidate_Latest_Success(t *testing.T) {
	const file = "testdata/valid.yaml"
	f, err := os.Open(file)
	require.NoError(t, err)

	defer f.Close()

	v, err := NewFeaturesValidator()
	require.NoError(t, err)

	err = v.Validate(file, f)
	assert.NoError(t, err)
}

func TestValidate_Segments_V2(t *testing.T) {
	const file = "testdata/valid_segments_v2.yaml"
	f, err := os.Open(file)
	require.NoError(t, err)

	defer f.Close()

	v, err := NewFeaturesValidator()
	require.NoError(t, err)

	err = v.Validate(file, f)
	assert.NoError(t, err)
}

func TestValidate_ConstraintListValues(t *testing.T) {
	const file = "testdata/valid_constraint_list_values.yaml"
	f, err := os.Open(file)
	require.NoError(t, err)

	defer f.Close()

	v, err := NewFeaturesValidator()
	require.NoError(t, err)

	err = v.Validate(file, f)
	assert.NoError(t, err)
}

func TestValidate_ConstraintLTOperator(t *testing.T) {
	const file = "testdata/valid_constraint_lt_operator.yaml"
	f, err := os.Open(file)
	require.NoError(t, err)

	defer f.Close()

	v, err := NewFeaturesValidator()
	require.NoError(t, err)

	err = v.Validate(file, f)
	assert.NoError(t, err)
}

func TestValidate_DefaultVariant_V3(t *testing.T) {
	const file = "testdata/valid_default_variant_v3.yaml"
	f, err := os.Open(file)
	require.NoError(t, err)

	defer f.Close()

	v, err := NewFeaturesValidator()
	require.NoError(t, err)

	err = v.Validate(file, f)
	assert.NoError(t, err)
}

func TestValidate_Metadata_V3(t *testing.T) {
	const file = "testdata/valid_metadata_v3.yaml"
	f, err := os.Open(file)
	require.NoError(t, err)

	defer f.Close()

	v, err := NewFeaturesValidator()
	require.NoError(t, err)

	err = v.Validate(file, f)
	assert.NoError(t, err)
}

func TestValidate_NamespaceDetails_v4(t *testing.T) {
	const file = "testdata/valid_namespace_details_v4.yaml"
	f, err := os.Open(file)
	require.NoError(t, err)

	defer f.Close()

	v, err := NewFeaturesValidator()
	require.NoError(t, err)

	err = v.Validate(file, f)
	assert.NoError(t, err)
}

func TestValidate_ZeroValues_Success(t *testing.T) {
	const file = "testdata/valid_zero_values.yaml"
	f, err := os.Open(file)
	require.NoError(t, err)

	defer f.Close()

	v, err := NewFeaturesValidator()
	require.NoError(t, err)

	err = v.Validate(file, f)
	assert.NoError(t, err)
}

func TestValidate_YAML_Stream(t *testing.T) {
	const file = "testdata/valid_yaml_stream.yaml"
	f, err := os.Open(file)
	require.NoError(t, err)

	defer f.Close()

	v, err := NewFeaturesValidator()
	require.NoError(t, err)

	err = v.Validate(file, f)
	assert.NoError(t, err)
}

func TestValidate_Failure(t *testing.T) {
	const file = "testdata/invalid.yaml"
	f, err := os.Open(file)
	require.NoError(t, err)

	defer f.Close()

	v, err := NewFeaturesValidator()
	require.NoError(t, err)

	err = v.Validate(file, f)

	errs, ok := Unwrap(err)
	require.True(t, ok)

	var ferr Error
	require.True(t, errors.As(errs[0], &ferr))

	assert.Equal(t, "flags.0.rules.1.distributions.0.rollout: invalid value 110 (out of bound <=100)", ferr.Message)
	assert.Equal(t, file, ferr.Location.File)
	assert.Equal(t, 22, ferr.Location.Line)
}

func TestValidate_Failure_YAML_Stream(t *testing.T) {
	const file = "testdata/invalid_yaml_stream.yaml"
	f, err := os.Open(file)
	require.NoError(t, err)

	defer f.Close()

	v, err := NewFeaturesValidator()
	require.NoError(t, err)

	err = v.Validate(file, f)

	errs, ok := Unwrap(err)
	require.True(t, ok)

	var ferr Error
	require.True(t, errors.As(errs[0], &ferr))

	assert.Equal(t, "flags.0.rules.1.distributions.0.rollout: invalid value 110 (out of bound <=100)", ferr.Message)
	assert.Equal(t, file, ferr.Location.File)
	assert.Equal(t, 59, ferr.Location.Line)
}

func TestValidate_Extended(t *testing.T) {
	const file = "testdata/valid.yaml"
	f, err := os.Open(file)
	require.NoError(t, err)

	defer f.Close()

	extended, err := os.ReadFile("extended.cue")
	require.NoError(t, err)

	v, err := NewFeaturesValidator(WithSchemaExtension(extended))
	require.NoError(t, err)

	err = v.Validate(file, f)

	errs, ok := Unwrap(err)
	require.True(t, ok)

	var ferr Error
	require.True(t, errors.As(errs[0], &ferr))

	assert.Equal(t, `flags.1.description: incomplete value =~"^.+$"`, ferr.Message)
	assert.Equal(t, file, ferr.Location.File)
}

// TestValidate_NullDescriptions tests that null descriptions are allowed
// in all entities (namespace, flags, variants, segments, constraints, rollouts).
// This test would have caught the bug where CUE schema was rejecting null descriptions.
func TestValidate_NullDescriptions(t *testing.T) {
	const file = "testdata/valid_null_descriptions.yaml"
	f, err := os.Open(file)
	require.NoError(t, err)

	defer f.Close()

	v, err := NewFeaturesValidator()
	require.NoError(t, err)

	err = v.Validate(file, f)
	assert.NoError(t, err)
}

// TestValidate_VariantFlagWithoutRollouts tests that variant flags
// do not require the rollouts field. This test would have caught the bug
// where the CUE schema incorrectly required rollouts on all flags due to
// the unconditional #FlagBoolean | *{} pattern.
func TestValidate_VariantFlagWithoutRollouts(t *testing.T) {
	const file = "testdata/valid_variant_flag_without_rollouts.yaml"
	f, err := os.Open(file)
	require.NoError(t, err)

	defer f.Close()

	v, err := NewFeaturesValidator()
	require.NoError(t, err)

	err = v.Validate(file, f)
	assert.NoError(t, err)
}

func TestValidate_VersionCompatibility(t *testing.T) {
	v, err := NewFeaturesValidator()
	require.NoError(t, err)

	tests := []struct {
		name    string
		file    string
		doc     string
		wantErr bool
	}{
		{
			name: "versionless YAML STRING isoneof JSON string",
			file: "versionless_string_isoneof.yaml",
			doc:  membershipYAML("", "STRING_COMPARISON_TYPE", "isoneof", `'["org-a","org-b"]'`),
		},
		{
			name: "versionless YAML STRING isnotoneof JSON string",
			file: "versionless_string_isnotoneof.yaml",
			doc:  membershipYAML("", "STRING_COMPARISON_TYPE", "isnotoneof", `'["US","UK"]'`),
		},
		{
			name: "versionless YAML ENTITY_ID isoneof JSON string",
			file: "versionless_entity_isoneof.yaml",
			doc:  membershipYAML("", "ENTITY_ID_COMPARISON_TYPE", "isoneof", `'["user-1","user-2"]'`),
		},
		{
			name: "versionless YAML ENTITY_ID isnotoneof JSON string",
			file: "versionless_entity_isnotoneof.yaml",
			doc:  membershipYAML("", "ENTITY_ID_COMPARISON_TYPE", "isnotoneof", `'["user-9"]'`),
		},
		{
			name: "versionless YAML NUMBER isoneof JSON string",
			file: "versionless_number_isoneof.yaml",
			doc:  membershipYAML("", "NUMBER_COMPARISON_TYPE", "isoneof", `'[18,21,65]'`),
		},
		{
			name: "versionless YAML NUMBER isnotoneof JSON string",
			file: "versionless_number_isnotoneof.yaml",
			doc:  membershipYAML("", "NUMBER_COMPARISON_TYPE", "isnotoneof", `'[0,1]'`),
		},
		{
			name: "versionless YAML STRING isoneof typed list",
			file: "versionless_string_list.yaml",
			doc:  membershipYAML("", "STRING_COMPARISON_TYPE", "isoneof", "\n      - org-a\n      - org-b"),
		},
		{
			name: "versionless YAML NUMBER isoneof typed list",
			file: "versionless_number_list.yaml",
			doc:  membershipYAML("", "NUMBER_COMPARISON_TYPE", "isoneof", "\n      - 18\n      - 21"),
		},
		{
			name: "versionless YAML ENTITY_ID isoneof typed list",
			file: "versionless_entity_list.yaml",
			doc:  membershipYAML("", "ENTITY_ID_COMPARISON_TYPE", "isoneof", "\n      - user-1\n      - user-2"),
		},
		{
			name: "explicit 1.5 YAML STRING isoneof JSON string",
			file: "explicit_1_5_string.yaml",
			doc:  membershipYAML("1.5", "STRING_COMPARISON_TYPE", "isoneof", `'["org-a","org-b"]'`),
		},
		{
			name: "explicit 1.5 YAML NUMBER isoneof JSON string",
			file: "explicit_1_5_number.yaml",
			doc:  membershipYAML("1.5", "NUMBER_COMPARISON_TYPE", "isoneof", `'[18,21]'`),
		},
		{
			name: "explicit 1.5 YAML ENTITY_ID isoneof JSON string",
			file: "explicit_1_5_entity.yaml",
			doc:  membershipYAML("1.5", "ENTITY_ID_COMPARISON_TYPE", "isoneof", `'["user-1"]'`),
		},
		{
			name: "explicit 1.6 YAML STRING isoneof typed list",
			file: "explicit_1_6_string.yaml",
			doc:  membershipYAML("1.6", "STRING_COMPARISON_TYPE", "isoneof", "\n      - org-a\n      - org-b"),
		},
		{
			name: "explicit 1.6 YAML NUMBER isoneof typed list",
			file: "explicit_1_6_number.yaml",
			doc:  membershipYAML("1.6", "NUMBER_COMPARISON_TYPE", "isoneof", "\n      - 18\n      - 21"),
		},
		{
			name: "explicit 1.6 YAML ENTITY_ID isoneof typed list",
			file: "explicit_1_6_entity.yaml",
			doc:  membershipYAML("1.6", "ENTITY_ID_COMPARISON_TYPE", "isoneof", "\n      - user-1\n      - user-2"),
		},
		{
			name:    "explicit 1.6 YAML STRING isoneof JSON string fails",
			file:    "explicit_1_6_string_legacy.yaml",
			doc:     membershipYAML("1.6", "STRING_COMPARISON_TYPE", "isoneof", `'["org-a","org-b"]'`),
			wantErr: true,
		},
		{
			name:    "explicit 1.6 YAML NUMBER isoneof JSON string fails",
			file:    "explicit_1_6_number_legacy.yaml",
			doc:     membershipYAML("1.6", "NUMBER_COMPARISON_TYPE", "isoneof", `'[18,21]'`),
			wantErr: true,
		},
		{
			name:    "explicit 1.6 YAML ENTITY_ID isoneof JSON string fails",
			file:    "explicit_1_6_entity_legacy.yaml",
			doc:     membershipYAML("1.6", "ENTITY_ID_COMPARISON_TYPE", "isoneof", `'["user-1"]'`),
			wantErr: true,
		},
		{
			name:    "null version does not retry legacy schema",
			file:    "null_version.yaml",
			doc:     membershipYAML("null", "STRING_COMPARISON_TYPE", "isoneof", `'["org-a"]'`),
			wantErr: true,
		},
		{
			name:    "empty version does not retry legacy schema",
			file:    "empty_version.yaml",
			doc:     membershipYAML(`""`, "STRING_COMPARISON_TYPE", "isoneof", `'["org-a"]'`),
			wantErr: true,
		},
		{
			name:    "unknown version does not retry legacy schema",
			file:    "unknown_version.yaml",
			doc:     membershipYAML(`"9.9"`, "STRING_COMPARISON_TYPE", "isoneof", `'["org-a"]'`),
			wantErr: true,
		},
		{
			name:    "invalid constraint operator remains invalid",
			file:    "invalid_operator.yaml",
			doc:     membershipYAML("", "STRING_COMPARISON_TYPE", "not-a-real-operator", `'["org-a"]'`),
			wantErr: true,
		},
		{
			name:    "malformed YAML remains invalid",
			file:    "malformed.yaml",
			doc:     "flags: [\n",
			wantErr: true,
		},
		{
			name: "missing required property remains invalid",
			file: "missing_property.yaml",
			doc: `namespace: default
flags: []
segments:
- key: segment1
  name: Segment 1
  match_type: ANY_MATCH_TYPE
  constraints:
  - type: STRING_COMPARISON_TYPE
    operator: isoneof
    value: '["org-a"]'
`,
			wantErr: true,
		},
		{
			name: "versionless JSON STRING isoneof JSON string",
			file: "versionless_string_isoneof.json",
			doc:  membershipJSON("", "STRING_COMPARISON_TYPE", "isoneof", `"[\"org-a\",\"org-b\"]"`),
		},
		{
			name: "versionless JSON STRING isnotoneof JSON string",
			file: "versionless_string_isnotoneof.json",
			doc:  membershipJSON("", "STRING_COMPARISON_TYPE", "isnotoneof", `"[\"US\",\"UK\"]"`),
		},
		{
			name: "versionless JSON STRING isoneof typed list",
			file: "versionless_string_list.json",
			doc:  membershipJSON("", "STRING_COMPARISON_TYPE", "isoneof", `["org-a","org-b"]`),
		},
		{
			name: "versionless JSON NUMBER isoneof JSON string",
			file: "versionless_number_isoneof.json",
			doc:  membershipJSON("", "NUMBER_COMPARISON_TYPE", "isoneof", `"[18,21]"`),
		},
		{
			name: "versionless JSON ENTITY_ID isoneof JSON string",
			file: "versionless_entity_isoneof.json",
			doc:  membershipJSON("", "ENTITY_ID_COMPARISON_TYPE", "isoneof", `"[\"user-1\"]"`),
		},
		{
			name: "explicit 1.5 JSON STRING isoneof JSON string",
			file: "explicit_1_5_string.json",
			doc:  membershipJSON("1.5", "STRING_COMPARISON_TYPE", "isoneof", `"[\"org-a\"]"`),
		},
		{
			name: "explicit 1.6 JSON STRING isoneof typed list",
			file: "explicit_1_6_string.json",
			doc:  membershipJSON("1.6", "STRING_COMPARISON_TYPE", "isoneof", `["org-a"]`),
		},
		{
			name:    "explicit 1.6 JSON STRING isoneof JSON string fails",
			file:    "explicit_1_6_string_legacy.json",
			doc:     membershipJSON("1.6", "STRING_COMPARISON_TYPE", "isoneof", `"[\"org-a\"]"`),
			wantErr: true,
		},
		{
			name:    "null JSON version does not retry legacy schema",
			file:    "null_version.json",
			doc:     membershipJSON("null", "STRING_COMPARISON_TYPE", "isoneof", `"[\"org-a\"]"`),
			wantErr: true,
		},
		{
			name:    "empty JSON version does not retry legacy schema",
			file:    "empty_version.json",
			doc:     `{"version":"","namespace":"default","flags":[],"segments":[{"key":"segment1","name":"Segment 1","match_type":"ANY_MATCH_TYPE","constraints":[{"type":"STRING_COMPARISON_TYPE","property":"prop","operator":"isoneof","value":"[\"org-a\"]"}]}]}`,
			wantErr: true,
		},
		{
			name:    "unknown JSON version does not retry legacy schema",
			file:    "unknown_version.json",
			doc:     membershipJSON("9.9", "STRING_COMPARISON_TYPE", "isoneof", `"[\"org-a\"]"`),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.Validate(tt.file, strings.NewReader(tt.doc))
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestValidate_VersionlessFallbackDiagnostics(t *testing.T) {
	v, err := NewFeaturesValidator()
	require.NoError(t, err)

	const file = "invalid_versionless.yaml"
	const doc = `namespace: default
flags:
- key: flag1
  name: Flag 1
  enabled: true
  variants:
  - key: on
  rules:
  - segment: segment1
    distributions:
    - variant: on
      rollout: 110
segments:
- key: segment1
  name: Segment 1
  match_type: ANY_MATCH_TYPE
  constraints:
  - type: STRING_COMPARISON_TYPE
    property: org
    operator: isoneof
    value: '["org-a"]'
`

	err = v.Validate(file, strings.NewReader(doc))
	require.Error(t, err)

	errs, ok := Unwrap(err)
	require.True(t, ok)
	require.NotEmpty(t, errs)

	var ferr Error
	require.ErrorAs(t, errs[0], &ferr)
	assert.Equal(t, file, ferr.Location.File)
	assert.Positive(t, ferr.Location.Line)
	assert.NotEmpty(t, ferr.Message)
	assert.Contains(t, ferr.Message, "rollout")
}

func TestValidate_VersionlessFallbackDiagnostics_YAMLStream(t *testing.T) {
	v, err := NewFeaturesValidator()
	require.NoError(t, err)

	const file = "invalid_versionless_stream.yaml"
	const doc = `namespace: first
flags: []
segments:
- key: segment1
  name: Segment 1
  match_type: ANY_MATCH_TYPE
  constraints:
  - type: STRING_COMPARISON_TYPE
    property: org
    operator: isoneof
    value: '["org-a"]'
---
namespace: second
flags:
- key: flag2
  name: Flag 2
  enabled: true
  variants:
  - key: on
  rules:
  - segment: segment2
    distributions:
    - variant: on
      rollout: 110
segments:
- key: segment2
  name: Segment 2
  match_type: ANY_MATCH_TYPE
  constraints:
  - type: STRING_COMPARISON_TYPE
    property: org
    operator: isoneof
    value: '["org-b"]'
`

	err = v.Validate(file, strings.NewReader(doc))
	require.Error(t, err)

	errs, ok := Unwrap(err)
	require.True(t, ok)
	require.NotEmpty(t, errs)

	var ferr Error
	require.ErrorAs(t, errs[0], &ferr)
	assert.Equal(t, file, ferr.Location.File)
	assert.Greater(t, ferr.Location.Line, 11)
	assert.NotEmpty(t, ferr.Message)
}

func TestValidate_VersionlessThenCurrentYAMLStream(t *testing.T) {
	v, err := NewFeaturesValidator()
	require.NoError(t, err)

	const stream = `namespace: first
flags: []
segments:
- key: segment1
  name: Segment 1
  match_type: ANY_MATCH_TYPE
  constraints:
  - type: STRING_COMPARISON_TYPE
    property: org
    operator: isoneof
    value: '["org-a","org-b"]'
---
namespace: second
flags: []
segments:
- key: segment2
  name: Segment 2
  match_type: ANY_MATCH_TYPE
  constraints:
  - type: STRING_COMPARISON_TYPE
    property: org
    operator: isoneof
    value:
      - org-c
      - org-d
`

	require.NoError(t, v.Validate("mixed_stream.yaml", strings.NewReader(stream)))

	const later = `namespace: third
flags: []
segments:
- key: segment3
  name: Segment 3
  match_type: ANY_MATCH_TYPE
  constraints:
  - type: NUMBER_COMPARISON_TYPE
    property: age
    operator: isoneof
    value:
      - 18
      - 21
`

	require.NoError(t, v.Validate("later.yaml", strings.NewReader(later)))
}

func TestValidate_VersionlessSchemaExtension(t *testing.T) {
	extended, err := os.ReadFile("extended.cue")
	require.NoError(t, err)

	v, err := NewFeaturesValidator(WithSchemaExtension(extended))
	require.NoError(t, err)

	const withoutDescription = `namespace: default
flags:
- key: flag1
  name: Flag 1
  enabled: true
segments:
- key: segment1
  name: Segment 1
  match_type: ANY_MATCH_TYPE
  constraints:
  - type: STRING_COMPARISON_TYPE
    property: org
    operator: isoneof
    value: '["org-a"]'
`

	err = v.Validate("missing_description.yaml", strings.NewReader(withoutDescription))
	require.Error(t, err)

	errs, ok := Unwrap(err)
	require.True(t, ok)
	require.NotEmpty(t, errs)

	var ferr Error
	require.ErrorAs(t, errs[0], &ferr)
	assert.Equal(t, "missing_description.yaml", ferr.Location.File)
	assert.NotEmpty(t, ferr.Message)

	const withDescription = `namespace: default
flags:
- key: flag1
  name: Flag 1
  description: required by extension
  enabled: true
segments:
- key: segment1
  name: Segment 1
  match_type: ANY_MATCH_TYPE
  constraints:
  - type: STRING_COMPARISON_TYPE
    property: org
    operator: isoneof
    value: '["org-a"]'
`

	require.NoError(t, v.Validate("with_description.yaml", strings.NewReader(withDescription)))
}

func membershipYAML(version, comparisonType, operator, value string) string {
	var b strings.Builder
	if version != "" {
		b.WriteString("version: ")
		if version == "null" || strings.HasPrefix(version, `"`) {
			b.WriteString(version)
		} else {
			b.WriteString(`"`)
			b.WriteString(version)
			b.WriteString(`"`)
		}
		b.WriteString("\n")
	}
	b.WriteString("namespace: default\n")
	b.WriteString("flags: []\n")
	b.WriteString("segments:\n")
	b.WriteString("- key: segment1\n")
	b.WriteString("  name: Segment 1\n")
	b.WriteString("  match_type: ANY_MATCH_TYPE\n")
	b.WriteString("  constraints:\n")
	b.WriteString("  - type: ")
	b.WriteString(comparisonType)
	b.WriteString("\n    property: prop\n")
	b.WriteString("    operator: ")
	b.WriteString(operator)
	b.WriteString("\n    value: ")
	b.WriteString(value)
	b.WriteString("\n")
	return b.String()
}

func membershipJSON(version, comparisonType, operator, value string) string {
	var b strings.Builder
	b.WriteString("{")
	if version != "" {
		if version == "null" {
			b.WriteString(`"version":null,`)
		} else {
			b.WriteString(`"version":"`)
			b.WriteString(version)
			b.WriteString(`",`)
		}
	}
	b.WriteString(`"namespace":"default","flags":[],"segments":[{`)
	b.WriteString(`"key":"segment1","name":"Segment 1","match_type":"ANY_MATCH_TYPE",`)
	b.WriteString(`"constraints":[{"type":"`)
	b.WriteString(comparisonType)
	b.WriteString(`","property":"prop","operator":"`)
	b.WriteString(operator)
	b.WriteString(`","value":`)
	b.WriteString(value)
	b.WriteString("}]}]}")
	return b.String()
}
