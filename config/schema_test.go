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

// Test_SchemaSCMCredentialsOptional asserts both schemas accept an SCM block
// with or without credentials, matching the Go config.
func Test_SchemaSCMCredentialsOptional(t *testing.T) {
	tests := []struct {
		name string
		scm  map[string]any
	}{
		{name: "with credentials", scm: map[string]any{"type": "gitlab", "credentials": "gitlab"}},
		{name: "without credentials", scm: map[string]any{"type": "gitlab"}},
	}

	jsonSchemaBytes, err := os.ReadFile("flipt.schema.json")
	require.NoError(t, err)

	cueSchemaBytes, err := os.ReadFile("flipt.schema.cue")
	require.NoError(t, err)

	ctx := cuecontext.New()
	spec := ctx.CompileBytes(cueSchemaBytes).LookupPath(cue.MakePath(cue.Def("#FliptSpec")))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conf := defaultConfig(t)
			conf["environments"] = map[string]any{
				"default": map[string]any{
					"name":      "default",
					"default":   true,
					"storage":   "default",
					"directory": "",
					"scm":       tt.scm,
				},
			}

			res, err := gojsonschema.Validate(gojsonschema.NewBytesLoader(jsonSchemaBytes), gojsonschema.NewGoLoader(conf))
			require.NoError(t, err)
			assert.True(t, res.Valid(), "JSON schema: %v", res.Errors())

			assert.NoError(t, spec.Unify(ctx.Encode(conf)).Validate(cue.Concrete(true)), "CUE schema")
		})
	}
}

func defaultConfig(t *testing.T) (conf map[string]any) {
	dec, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		DecodeHook: mapstructure.ComposeDecodeHookFunc(config.DecodeHooks...),
		Result:     &conf,
	})
	config := config.Default()
	// hack to get around validation not being able to handle types that map[string]*struct
	config.Storage["default"].PollInterval = 0
	config.Authentication.Session.Storage.Redis.Mode = "single"
	require.NoError(t, err)
	require.NoError(t, dec.Decode(config))

	// adapt converts instances of time.Duration to their
	// string representation, which CUE is going to validate
	adapt(conf)

	return conf
}

func Test_SchemaSCMWebhook(t *testing.T) {
	tests := []struct {
		name    string
		webhook map[string]any
		valid   bool
	}{
		{name: "secret", webhook: map[string]any{"secret": "s3cr3t"}, valid: true},
		{name: "secret reference", webhook: map[string]any{"secret": "${secret:file:gitlab-webhook}"}, valid: true},
		{name: "missing secret", webhook: map[string]any{}},
		{name: "unknown field", webhook: map[string]any{"secret": "s3cr3t", "secret_ref": map[string]any{"provider": "file", "path": "webhooks", "key": "gitlab"}}},
	}

	jsonSchemaBytes, err := os.ReadFile("flipt.schema.json")
	require.NoError(t, err)

	cueSchemaBytes, err := os.ReadFile("flipt.schema.cue")
	require.NoError(t, err)

	ctx := cuecontext.New()
	spec := ctx.CompileBytes(cueSchemaBytes).LookupPath(cue.MakePath(cue.Def("#FliptSpec")))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conf := defaultConfig(t)
			conf["environments"] = map[string]any{
				"default": map[string]any{
					"name":      "default",
					"default":   true,
					"storage":   "default",
					"directory": "",
					"scm": map[string]any{
						"type":        "gitlab",
						"credentials": "git",
						"webhook":     tt.webhook,
					},
				},
			}

			res, err := gojsonschema.Validate(gojsonschema.NewBytesLoader(jsonSchemaBytes), gojsonschema.NewGoLoader(conf))
			require.NoError(t, err)
			assert.Equal(t, tt.valid, res.Valid(), "JSON schema: %v", res.Errors())

			cueErr := spec.Unify(ctx.Encode(conf)).Validate(cue.Concrete(true))
			if tt.valid {
				assert.NoError(t, cueErr, "CUE schema")
			} else {
				assert.Error(t, cueErr, "CUE schema")
			}
		})
	}
}
