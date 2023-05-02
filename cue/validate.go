package cue

import (
	_ "embed"
	"os"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/pkg/encoding/yaml"
	"github.com/hashicorp/go-multierror"
)

var (
	//go:embed flipt.cue
	cueFile []byte
)

// ValidateBytes takes a slice of bytes, and validates them against a cue definition.
func ValidateBytes(b []byte) error {
	ctx := cuecontext.New()

	v := ctx.CompileBytes(cueFile)
	if err := v.Err(); err != nil {
		return err
	}

	return validateHelper(b, v)
}

func validateHelper(b []byte, v cue.Value) error {
	_, err := yaml.Validate(b, v)
	if err != nil {
		return err
	}

	return nil
}

// ValidateFiles takes a slice of strings as filenames and validates them against
// our cue definition of features.
func ValidateFiles(files []string) error {
	ctx := cuecontext.New()

	v := ctx.CompileBytes(cueFile)
	if err := v.Err(); err != nil {
		return err
	}

	var merr error

	for _, f := range files {
		b, err := os.ReadFile(f)
		// Quit execution of the cue validating against the yaml
		// files upon failure to read file.
		if err != nil {
			return err
		}
		err = validateHelper(b, v)
		if err != nil {
			merr = multierror.Append(merr, err)
		}
	}

	return merr
}
