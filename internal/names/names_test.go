package names

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRandom(t *testing.T) {
	name := Random()

	require.Regexp(t, regexp.MustCompile(`^[a-z]+-[a-z]+-[0-9a-f]{4}$`), name)
}
