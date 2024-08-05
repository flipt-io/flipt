package validation

import (
	"errors"
	"os"
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
	// location of the start of the boolean flag
	// which lacks a description
	assert.Equal(t, 33, ferr.Location.Line)
}
