package cue

import (
	"os"
	"testing"

	"cuelang.org/go/cue/cuecontext"
	"github.com/stretchr/testify/require"
)

func TestValidateHelper_Success(t *testing.T) {
	b, err := os.ReadFile("fixtures/valid.yaml")
	require.NoError(t, err)
	cctx := cuecontext.New()

	err = validateHelper(b, cctx)

	require.NoError(t, err)
}

func TestValidateHelper_Failure(t *testing.T) {
	b, err := os.ReadFile("fixtures/invalid.yaml")
	require.NoError(t, err)

	cctx := cuecontext.New()

	err = validateHelper(b, cctx)
	require.EqualError(t, err, "flags.0.rules.0.distributions.0.rollout: invalid value 110 (out of bound <=100)")
}
