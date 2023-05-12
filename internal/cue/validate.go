package cue

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	cueerror "cuelang.org/go/cue/errors"
	"cuelang.org/go/encoding/yaml"
)

var (
	//go:embed flipt.cue
	cueFile []byte

	jsonFormat = "json"
	textFormat = "text"
)

var (
	ErrValidationFailed = errors.New("validation failed")
)

// ValidateBytes takes a slice of bytes, and validates them against a cue definition.
func ValidateBytes(b []byte) error {
	cctx := cuecontext.New()

	return validate(b, cctx)
}

func validate(b []byte, cctx *cue.Context) error {
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

// Error is a collection of fields that represent positions in files where the user
// has made some kind of error.
type Error struct {
	Message  string   `json:"message"`
	Location Location `json:"location"`
}

func writeErrorDetails(format string, cerrs []Error, w io.Writer) error {
	if format == jsonFormat {
		allErrors := struct {
			Errors []Error `json:"errors"`
		}{
			Errors: cerrs,
		}

		if err := json.NewEncoder(os.Stdout).Encode(allErrors); err != nil {
			fmt.Fprintln(w, "Internal error.")
			return err
		}

		return nil
	}

	if format != textFormat {
		fmt.Fprintln(w, "Invalid format chosen, defaulting to \"text\" format...")
	}

	var buffer bytes.Buffer

	buffer.WriteString("❌ Validation failure!\n\n")

	for i := 0; i < len(cerrs); i++ {
		errString := fmt.Sprintf(`
- Message: %s
  File   : %s
  Line   : %d
  Column : %d
`, cerrs[i].Message, cerrs[i].Location.File, cerrs[i].Location.Line, cerrs[i].Location.Column)

		if i < len(cerrs)-1 {
			errString += "\n"
		}

		buffer.WriteString(errString)
	}

	fmt.Fprintln(w, buffer.String())

	return nil
}

func writeSuccessDetails(format string, w io.Writer) error {
	if format == jsonFormat {
		success := struct {
			Validate string `json:"validate"`
		}{
			Validate: "success",
		}
		if err := json.NewEncoder(w).Encode(success); err != nil {
			fmt.Fprintln(w, "Internal error.")
			return err
		}

		return nil
	}

	if format != textFormat {
		fmt.Fprintln(w, "Invalid format chosen, defaulting to text...")
	}

	successMessage := "✅ Validation success!"

	fmt.Fprintln(w, successMessage)

	return nil
}

// ValidateFiles takes a slice of strings as filenames and validates them against
// our cue definition of features.
func ValidateFiles(dst io.Writer, files []string, format string) error {
	cctx := cuecontext.New()

	cerrs := make([]Error, 0)

	for _, f := range files {
		b, err := os.ReadFile(f)
		// Quit execution of the cue validating against the yaml
		// files upon failure to read file.
		if err != nil {
			var buffer bytes.Buffer

			buffer.WriteString("❌ Validation failure!\n")
			buffer.WriteString(fmt.Sprintf("Failed to read file %s", f))

			fmt.Fprintln(dst, buffer.String())

			return ErrValidationFailed
		}
		err = validate(b, cctx)
		if err != nil {

			ce := cueerror.Errors(err)

			for _, m := range ce {
				ips := m.InputPositions()
				if len(ips) > 0 {
					fp := ips[0]
					format, args := m.Msg()

					cerrs = append(cerrs, Error{
						Message: fmt.Sprintf(format, args...),
						Location: Location{
							File:   f,
							Line:   fp.Line(),
							Column: fp.Column(),
						},
					})
				}
			}
		}
	}

	if len(cerrs) > 0 {
		if err := writeErrorDetails(format, cerrs, dst); err != nil {
			return err
		}

		return ErrValidationFailed
	} else {
		if err := writeSuccessDetails(format, dst); err != nil {
			return err
		}
	}

	return nil
}
