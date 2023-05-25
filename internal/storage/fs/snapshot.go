package fs

import (
	iofs "io/fs"

	"github.com/gobwas/glob"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

const (
	indexFile = ".flipt.yml"
)

var (
	// globDefaults will be used if index file does not exist
	// or can not be successfully read (malformed).
	globDefaults []string = []string{"**/features.yml", "**/features.yaml", "**/*.features.yml", "**/*.features.yaml"}
)

// FliptIndex represents the structure of a well-known file ".flipt.yml"
// at the root of an FS.
type FliptIndex struct {
	Version string   `yaml:"version,omitempty"`
	Include []string `yaml:"include,omitempty"`
	Exclude []string `yaml:"exclude,omitempty"`
}

// Snapshot represents the flag state from an FS.
type Snapshot struct{}

func buildSnapshotHelper(logger *zap.Logger, source iofs.FS) ([]string, []glob.Glob, error) {
	var (
		exclusions []string
		inclusions = globDefaults
	)

	// Read index file
	inFile, err := source.Open(indexFile)
	if err == nil {
		var fi FliptIndex
		if err := yaml.NewDecoder(inFile).Decode(&fi); err == nil {
			inclusions = fi.Include
			exclusions = fi.Exclude
		} else {
			logger.Debug("error in index file structure, defaulting...", zap.Error(err))
		}
	} else {
		logger.Debug("error reading index file, defaulting...", zap.String("file", indexFile), zap.Error(err))
	}

	filenames := make([]string, 0)
	excludeGlobs := make([]glob.Glob, 0)

	for _, g := range inclusions {
		f, err := iofs.Glob(source, g)
		if err != nil {
			logger.Error("malformed glob pattern for included files", zap.String("glob", g), zap.Error(err))
			return nil, nil, err
		}

		filenames = append(filenames, f...)
	}

	for _, e := range exclusions {
		g, err := glob.Compile(e)
		if err != nil {
			logger.Error("malformed glob pattern for excluded files", zap.String("glob", e), zap.Error(err))
			return nil, nil, err
		}
		excludeGlobs = append(excludeGlobs, g)
	}

	return filenames, excludeGlobs, nil
}

// BuildSnapshot will take in an FS implementation, and build the Snapshot of the flag state from the files
// necessary within the FS.
func BuildSnapshot(logger *zap.Logger, source iofs.FS) (*Snapshot, error) {
	_, _, err := buildSnapshotHelper(logger, source)
	if err != nil {
		return nil, err
	}

	return &Snapshot{}, nil
}
