package fs

import (
	"errors"
	"fmt"
	"io"
	"io/fs"

	"github.com/gobwas/glob"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

const (
	// IndexFileName is the name of the index file on disk
	IndexFileName = ".flipt.yml"
	indexVersion  = "1.0"
)

// FliptIndex is a set of glob include and exclude patterns
// which can be used to filter a set of provided paths.
type FliptIndex struct {
	includes []glob.Glob
	excludes []glob.Glob
}

// DefaultFliptIndex returns the default value for the FliptIndex.
// It used when a .flipt.yml can not be located.
func DefaultFliptIndex() (*FliptIndex, error) {
	return newFliptIndex(defaultFliptIndex())
}

// OpenFliptIndex attempts to retrieve a FliptIndex from the provided source
// fs.FS implementation. If the index cannot be found then it returns the
// instance returned by defaultFliptIndex().
func OpenFliptIndex(logger *zap.Logger, src fs.FS) (*FliptIndex, error) {
	// Read index file
	fi, err := src.Open(IndexFileName)
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			return nil, err
		}

		logger.Debug("using default index", zap.String("file", IndexFileName), zap.Error(err))

		return DefaultFliptIndex()
	}
	defer fi.Close()

	return ParseFliptIndex(fi)
}

// ParseFliptIndex reads and parses the provided io.Reader and returns
// a *FliptIndex it deemed valid.
func ParseFliptIndex(r io.Reader) (*FliptIndex, error) {
	idx, err := parseFliptIndex(r)
	if err != nil {
		return nil, err
	}

	return newFliptIndex(idx)
}

// Match returns true if the path should be kept because it matches
// the underlying index filter
func (i *FliptIndex) Match(path string) bool {
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
func newFliptIndex(index fliptIndex) (*FliptIndex, error) {
	filter := &FliptIndex{}

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

// fliptIndex represents the structure of a well-known file ".flipt.yml"
// at the root of an FS.
type fliptIndex struct {
	Version string   `yaml:"version,omitempty"`
	Include []string `yaml:"include,omitempty"`
	Exclude []string `yaml:"exclude,omitempty"`
}

func defaultFliptIndex() fliptIndex {
	return fliptIndex{
		Version: indexVersion,
		Include: []string{
			"**features.yml", "**features.yaml", "**.features.yml", "**.features.yaml",
		},
	}
}

// parseFliptIndex parses the provided Reader into a FliptIndex.
// It sets the version to the default version if not explicitly set
// in the reader contents.
func parseFliptIndex(r io.Reader) (fliptIndex, error) {
	idx := fliptIndex{Version: indexVersion}
	if derr := yaml.NewDecoder(r).Decode(&idx); derr != nil {
		return idx, derr
	}

	return idx, nil
}
