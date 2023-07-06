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
	cue *cue.Context
	v   cue.Value
}

func NewFeaturesValidator() (*FeaturesValidator, error) {
	cctx := cuecontext.New()
	v := cctx.CompileBytes(cueFile)
	if v.Err() != nil {
		return nil, v.Err()
	}

	return &FeaturesValidator{
		cue: cctx,
		v:   v,
	}, nil
}

// Validate validates a YAML file against our cue definition of features.
func (v FeaturesValidator) Validate(file string, b []byte) (Result, error) {
	var result Result

	f, err := yaml.Extract("", b)
	if err != nil {
		return result, err
	}

	yv := v.cue.BuildFile(f)
	if err := yv.Err(); err != nil {
		return Result{}, err
	}

	err = v.v.
		Unify(yv).
		Validate(cue.All(), cue.Concrete(true))

	for _, e := range cueerrors.Errors(err) {
		rerr := Error{
			Message: e.Error(),
			Location: Location{
				File: file,
			},
		}

		if pos := cueerrors.Positions(e); len(pos) > 0 {
			p := pos[len(pos)-1]
			rerr.Location.Line = p.Line()
			rerr.Location.Column = p.Column()
		}

		result.Errors = append(result.Errors, rerr)
	}

	if len(result.Errors) > 0 {
		return result, ErrValidationFailed
	}

	return result, nil
}
