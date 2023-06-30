package cue

import (
	_ "embed"
	"errors"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	cueerrors "cuelang.org/go/cue/errors"
	"cuelang.org/go/encoding/yaml"
)

var (
	//go:embed flipt.cue
	cueFile             []byte
	ErrValidationFailed = errors.New("validation failed")
)

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

// Result is a collection of errors that occurred during validation.
type Result struct {
	Errors []Error `json:"errors"`
}

type FeaturesValidator struct {
	Format string
	cue    *cue.Context
	v      cue.Value
}

func NewFeaturesValidator(format string) (*FeaturesValidator, error) {
	cctx := cuecontext.New()
	v := cctx.CompileBytes(cueFile)
	if v.Err() != nil {
		return nil, v.Err()
	}

	return &FeaturesValidator{
		Format: format,
		cue:    cctx,
		v:      v,
	}, nil
}

// Validate validates a YAML file against our cue definition of features.
func (v FeaturesValidator) Validate(file string, b []byte) (Result, error) {
	var result Result

	f, err := yaml.Extract("", b)
	if err != nil {
		return result, err
	}

	yv := v.cue.BuildFile(f, cue.Scope(v.v))
	yv = v.v.Unify(yv)
	err = yv.Validate()

	for _, e := range cueerrors.Errors(err) {
		pos := cueerrors.Positions(e)
		p := pos[len(pos)-1]

		result.Errors = append(result.Errors, Error{
			Message: e.Error(),
			Location: Location{
				File:   file,
				Line:   p.Line(),
				Column: p.Column(),
			},
		})
	}

	if len(result.Errors) > 0 {
		return result, ErrValidationFailed
	}

	return result, nil
}
