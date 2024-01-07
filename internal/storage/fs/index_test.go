package fs

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseFliptIndex(t *testing.T) {
	buf := bytes.NewBufferString(`
version: "1.0"
include:
  - "*.features.yml"
  - b.features.json
exclude:
  - c.features.yml
`)
	index, err := ParseFliptIndex(buf)
	require.NoError(t, err)
	require.Equal(t, 2, len(index.includes))
	require.Equal(t, 1, len(index.excludes))
	require.True(t, index.Match("a.features.yml"))
	require.True(t, index.Match("b.features.json"))
	require.False(t, index.Match("c.features.yml"))
}

func TestParseFliptIndexParsingError(t *testing.T) {
	_, err := ParseFliptIndex(bytes.NewBufferString("version: &"))
	require.Error(t, err)
	require.EqualError(t, err, "yaml: did not find expected alphabetic or numeric character")

}
