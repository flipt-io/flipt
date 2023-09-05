package config

import (
	"os"
	"testing"
	"time"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/errors"
	"github.com/mitchellh/mapstructure"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xeipuuv/gojsonschema"
	"go.flipt.io/flipt/internal/config"
)

func Test_CUE(t *testing.T) {
	ctx := cuecontext.New()

	schemaBytes, err := os.ReadFile("flipt.schema.cue")
	require.NoError(t, err)

	v := ctx.CompileBytes(schemaBytes)

	conf := defaultConfig(t)

	dflt := ctx.Encode(conf)

	err = v.LookupPath(cue.MakePath(cue.Def("#FliptSpec"))).Unify(dflt).Validate(
		cue.Concrete(true),
	)

	if errs := errors.Errors(err); len(errs) > 0 {
		for _, err := range errs {
			t.Log(err)
		}
		t.Fatal("Errors validating CUE schema against default configuration")
	}
}

func adapt(m map[string]any) {
	for k, v := range m {
		switch t := v.(type) {
		case map[string]any:
			adapt(t)
		case time.Duration:
			m[k] = t.String()
		}
	}
}

func Test_JSONSchema(t *testing.T) {
	schemaBytes, err := os.ReadFile("flipt.schema.json")
	require.NoError(t, err)

	schema := gojsonschema.NewBytesLoader(schemaBytes)

	conf := defaultConfig(t)
	res, err := gojsonschema.Validate(schema, gojsonschema.NewGoLoader(conf))
	require.NoError(t, err)

	if !assert.True(t, res.Valid(), "Schema is invalid") {
		for _, err := range res.Errors() {
			t.Log(err)
		}
	}
}

func defaultConfig(t *testing.T) (conf map[string]any) {
	dec, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		DecodeHook: mapstructure.ComposeDecodeHookFunc(config.DecodeHooks...),
		Result:     &conf,
	})
	require.NoError(t, err)
	require.NoError(t, dec.Decode(config.Default()))

	// adapt converts instances of time.Duration to their
	// string representation, which CUE is going to validate
	adapt(conf)

	return
}
