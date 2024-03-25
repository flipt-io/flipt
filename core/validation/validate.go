package validation

import (
	_ "embed"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"strconv"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/ast"
	"cuelang.org/go/cue/cuecontext"
	cueerrors "cuelang.org/go/cue/errors"
	"cuelang.org/go/encoding/yaml"
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/util"
	"github.com/gobwas/glob"
	goyaml "gopkg.in/yaml.v3"
)

const (
	// IndexFileName is the name of the index file on disk
	IndexFileName = ".flipt.yml"
	indexVersion  = "1.0"
)

//go:embed flipt.cue
var cueFile []byte

// Location contains information about where an error has occurred during cue
// validation.
type Location struct {
	File string `json:"file,omitempty"`
	Line int    `json:"line"`
}

type unwrapable interface {
	Unwrap() []error
}

// Unwrap checks for the version of Unwrap which returns a slice
// see std errors package for details
func Unwrap(err error) ([]error, bool) {
	var u unwrapable
	if !errors.As(err, &u) {
		return nil, false
	}

	return u.Unwrap(), true
}

// Error is a collection of fields that represent positions in files where the user
// has made some kind of error.
type Error struct {
	Message  string   `json:"message"`
	Location Location `json:"location"`
}

func (e Error) Format(f fmt.State, verb rune) {
	if verb != 'v' {
		f.Write([]byte(e.Error()))
		return
	}

	fmt.Fprintf(f, `
- Message  : %s
  File     : %s
  Line     : %d
`, e.Message, e.Location.File, e.Location.Line)
}

func (e Error) Error() string {
	return fmt.Sprintf("%s (%s %d)", e.Message, e.Location.File, e.Location.Line)
}

type FeaturesValidator struct {
	cue *cue.Context
	v   cue.Value
}

type FeaturesValidatorOption func(*FeaturesValidator) error

func WithSchemaExtension(v []byte) FeaturesValidatorOption {
	return func(fv *FeaturesValidator) error {
		schema := fv.cue.CompileBytes(v)
		if err := schema.Err(); err != nil {
			return err
		}

		fv.v = fv.v.Unify(schema)
		return fv.v.Err()
	}
}

func NewFeaturesValidator(opts ...FeaturesValidatorOption) (*FeaturesValidator, error) {
	cctx := cuecontext.New()
	v := cctx.CompileBytes(cueFile)
	if v.Err() != nil {
		return nil, v.Err()
	}

	f := &FeaturesValidator{
		cue: cctx,
		v:   v,
	}

	for _, opt := range opts {
		if err := opt(f); err != nil {
			return nil, err
		}
	}

	return f, nil
}

func (v FeaturesValidator) validateSingleDocument(file string, f *ast.File, offset int) error {
	yv := v.cue.BuildFile(f)
	if err := yv.Err(); err != nil {
		return err
	}

	err := v.v.
		Unify(yv).
		Validate(cue.All(), cue.Concrete(true))

	var errs []error
OUTER:
	for _, e := range cueerrors.Errors(err) {
		rerr := Error{
			Message: e.Error(),
			Location: Location{
				File: file,
			},
		}

		// if the error has path segments we're going to use that
		// to select into the original document
		// we parse the slice of the path into selector
		selectors := []cue.Selector{}
		for _, p := range e.Path() {
			if i, err := strconv.ParseInt(p, 10, 64); err == nil {
				selectors = append(selectors, cue.Index(int(i)))
				continue
			}

			selectors = append(selectors, cue.Str(p))
		}

		// next we walk the selector back from the deapest path until
		// we select something that exists in the document
		for i := len(selectors); i > 0; i-- {
			selectors = selectors[:i]
			val := yv.LookupPath(cue.MakePath(selectors...))

			// if we manage to locate something then we use that
			// position in our error message
			if pos := val.Pos(); pos.IsValid() {
				rerr.Location.Line = pos.Line() + offset
				errs = append(errs, rerr)
				continue OUTER
			}
		}

		if pos := cueerrors.Positions(e); len(pos) > 0 {
			p := pos[len(pos)-1]
			rerr.Location.Line = p.Line() + offset
		}
		errs = append(errs, rerr)
	}

	return errors.Join(errs...)
}

// Validate validates a YAML file against our cue definition of features.
func (v FeaturesValidator) Validate(file string, reader io.Reader) error {
	decoder := goyaml.NewDecoder(reader)

	i := 0

	for {
		var node goyaml.Node

		if err := decoder.Decode(&node); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			return err
		}

		b, err := goyaml.Marshal(&node)
		if err != nil {
			return err
		}

		f, err := yaml.Extract("", b)
		if err != nil {
			return err
		}

		var offset = node.Line - 1
		if i > 0 {
			offset = node.Line
		}

		if err := v.validateSingleDocument(file, f, offset); err != nil {
			return err
		}

		i += 1
	}

	return nil
}

// ValidateFilesFromBillyFS will walk an billy.Filesystem and find relevant files to
// validate.
func (v FeaturesValidator) ValidateFilesFromBillyFS(src billy.Filesystem) error {
	idx, err := newFliptIndex(defaultFliptIndex())
	if err != nil {
		return err
	}

	fi, err := src.Open(IndexFileName)
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			return err
		}
	} else {
		parsed, err := parseFliptIndex(fi)
		if err != nil {
			return err
		}

		idx, err = newFliptIndex(parsed)
		if err != nil {
			return err
		}
		defer fi.Close()
	}

	if err := util.Walk(src, ".", func(path string, fi fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if fi.IsDir() {
			return nil
		}

		if idx.match(path) {
			fi, err := src.Open(path)
			if err != nil {
				return err
			}

			err = v.Validate(path, fi)
			if err != nil {
				return err
			}
		}

		return nil

	}); err != nil {
		return err
	}

	return nil
}

// ValidateFilesFromFS will walk an FS and find relevant files to
// validate.
func (v FeaturesValidator) ValidateFilesFromFS(src fs.FS) error {
	idx, err := openFliptIndex(src)
	if err != nil {
		return err
	}

	if err := fs.WalkDir(src, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		if idx.match(path) {
			fi, err := src.Open(path)
			if err != nil {
				return err
			}

			err = v.Validate(path, fi)
			if err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}

// fliptIndex is a set of glob include and exclude patterns
// which can be used to filter a set of provided paths.
type fliptIndex struct {
	includes []glob.Glob
	excludes []glob.Glob
}

// fliptIndexSource represents the structure of a well-known file ".flipt.yml"
// at the root of an FS.
type fliptIndexSource struct {
	Version string   `yaml:"version,omitempty"`
	Include []string `yaml:"include,omitempty"`
	Exclude []string `yaml:"exclude,omitempty"`
}

// openFliptIndex attempts to retrieve a FliptIndex from the provided source
// fs.FS implementation. If the index cannot be found then it returns the
// instance returned by defaultFliptIndex().
func openFliptIndex(src fs.FS) (*fliptIndex, error) {
	// Read index file
	fi, err := src.Open(IndexFileName)
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			return nil, err
		}

		return newFliptIndex(defaultFliptIndex())
	}
	defer fi.Close()

	idx, err := parseFliptIndex(fi)
	if err != nil {
		return nil, err
	}

	return newFliptIndex(idx)
}

func defaultFliptIndex() fliptIndexSource {
	return fliptIndexSource{
		Version: indexVersion,
		Include: []string{
			"**features.yml", "**features.yaml", "**.features.yml", "**.features.yaml",
		},
	}
}

// parseFliptIndex parses the provided Reader into a FliptIndex.
// It sets the version to the default version if not explicitly set
// in the reader contents.
func parseFliptIndex(r io.Reader) (fliptIndexSource, error) {
	idx := fliptIndexSource{Version: indexVersion}
	if derr := goyaml.NewDecoder(r).Decode(&idx); derr != nil {
		return idx, derr
	}

	return idx, nil
}

// match returns true if the path should be kept because it matches
// the underlying index filter
func (i *fliptIndex) match(path string) bool {
	for _, include := range i.includes {
		if include.Match(path) {
			var excluded bool
			for _, exclude := range i.excludes {
				if excluded = exclude.Match(path); excluded {
					break
				}
			}

			if !excluded {
				return true
			}
		}
	}

	return false
}

// newFliptIndex takes an index definition, parses the defined globs
// and returns an instance of FliptIndex.
// It returns err != nil if any of the defined globs cannot be parsed.
func newFliptIndex(index fliptIndexSource) (*fliptIndex, error) {
	filter := &fliptIndex{}

	for _, g := range index.Include {
		glob, err := glob.Compile(g)
		if err != nil {
			return nil, fmt.Errorf("compiling include glob: %w", err)
		}

		filter.includes = append(filter.includes, glob)
	}

	for _, g := range index.Exclude {
		glob, err := glob.Compile(g)
		if err != nil {
			return nil, fmt.Errorf("compiling exclude glob: %w", err)
		}

		filter.excludes = append(filter.excludes, glob)
	}

	return filter, nil
}
