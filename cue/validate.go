package cue

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	cueerror "cuelang.org/go/cue/errors"
	"cuelang.org/go/encoding/yaml"
)

var (
	//go:embed flipt.cue
	cueFile []byte
)

// ValidateBytes takes a slice of bytes, and validates them against a cue definition.
func ValidateBytes(b []byte) error {
	cctx := cuecontext.New()

	v := cctx.CompileBytes(cueFile)
	if err := v.Err(); err != nil {
		return err
	}

	return validateHelper(b, v)
}

func validateHelper(b []byte, v cue.Value) error {
	err := yaml.Validate(b, v)
	if err != nil {
		return err
	}

	return nil
}

// CueError is a collection of fields that represent positions in files where the user
// has made some kind of error.
type CueError struct {
	Filename string `json:"filename"`
	Line     int    `json:"line"`
	Column   int    `json:"column"`
	Msg      string `json:"msg"`
}

// ValidateFiles takes a slice of strings as filenames and validates them against
// our cue definition of features.
func ValidateFiles(files []string) error {
	cctx := cuecontext.New()

	v := cctx.CompileBytes(cueFile)
	if err := v.Err(); err != nil {
		return err
	}

	cerrs := make([]CueError, 0)

	for _, f := range files {
		b, err := os.ReadFile(f)
		// Quit execution of the cue validating against the yaml
		// files upon failure to read file.
		if err != nil {
			return err
		}
		err = validateHelper(b, v)
		if err != nil {
			var cerr cueerror.Error
			if errors.As(err, &cerr) {
				ip := cerr.InputPositions()
				for _, i := range ip {
					if i.Filename() != "" {
						cerrs = append(cerrs, CueError{
							Filename: f,
							Line:     i.Line(),
							Column:   i.Column(),
						})
					}
				}
			}
		}
	}

	if len(cerrs) > 0 {
		b, err := json.Marshal(cerrs)
		if err != nil {
			return err
		}

		// Write out the json output to stdout upon error detection.
		fmt.Fprintln(os.Stdout, string(b))
		return fmt.Errorf("validation error")
	}

	return nil
}
