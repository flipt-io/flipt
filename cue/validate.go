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

	return validateHelper(b, cctx)
}

func validateHelper(b []byte, cctx *cue.Context) error {
	v := cctx.CompileBytes(cueFile)

	f, err := yaml.Extract("", b)
	if err != nil {
		return err
	}

	yv := cctx.BuildFile(f, cue.Scope(v))
	yv = v.Unify(yv)

	return yv.Validate()
}

// Location contains information about where an error has occurred during cue
// validation.
type Location struct {
	File   string `json:"file,omitempty"`
	Line   int    `json:"line"`
	Column int    `json:"column"`
}

// CueError is a collection of fields that represent positions in files where the user
// has made some kind of error.
type CueError struct {
	Message  string   `json:"message"`
	Location Location `json:"location"`
}

func getCueErrors(err error, file string) []CueError {
	cerrs := make([]CueError, 0)

	ce := cueerror.Errors(err)

	for _, m := range ce {
		ips := m.InputPositions()
		if len(ips) > 0 {
			fp := ips[0]
			format, args := m.Msg()

			cerrs = append(cerrs, CueError{
				Message: fmt.Sprintf(format, args...),
				Location: Location{
					File:   file,
					Line:   fp.Line(),
					Column: fp.Column(),
				},
			})
		}
	}

	return cerrs
}

// ValidateFiles takes a slice of strings as filenames and validates them against
// our cue definition of features.
func ValidateFiles(files []string) error {
	cctx := cuecontext.New()

	cerrs := make([]CueError, 0)

	for _, f := range files {
		b, err := os.ReadFile(f)
		// Quit execution of the cue validating against the yaml
		// files upon failure to read file.
		if err != nil {
			return err
		}
		err = validateHelper(b, cctx)
		if err != nil {
			cerrs = append(cerrs, getCueErrors(err, f)...)
		}
	}

	if len(cerrs) > 0 {
		allErrors := struct {
			Errors []CueError `json:"errors"`
		}{
			Errors: cerrs,
		}

		// Write out the json output to stdout upon error detection.
		if err := json.NewEncoder(os.Stdout).Encode(allErrors); err != nil {
			return err
		}

		return errors.New("validation error")
	}

	return nil
}
