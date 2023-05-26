package fs

import (
	"errors"
	iofs "io/fs"
	"path"

	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

const (
	indexFile = ".flipt.yml"
)

// FliptIndex represents the structure of a well-known file ".flipt.yml"
// at the root of an FS.
type FliptIndex struct {
	Version    string   `yaml:"version,omitempty"`
	Inclusions []string `yaml:"inclusions,omitempty"`
	Exclusions []string `yaml:"exclusions,omitempty"`
}

func buildSnapshotHelper(logger *zap.Logger, source iofs.FS) ([]string, error) {
	// This is the default variable + value for the FliptIndex. It will preserve its value if
	// a .flipt.yml can not be read for whatever reason.
	idx := FliptIndex{
		Version: "1.0",
		Inclusions: []string{
			"**/features.yml", "**/features.yaml", "**/*.features.yml", "**/*.features.yaml",
		},
	}

	// Read index file
	inFile, err := source.Open(indexFile)
	if err == nil {
		if derr := yaml.NewDecoder(inFile).Decode(&idx); derr != nil {
			return nil, derr
		}
	}

	if err != nil && !errors.Is(err, iofs.ErrNotExist) {
		return nil, err
	} else {
		logger.Debug("index file does not exist, defaulting...", zap.String("file", indexFile), zap.Error(err))
	}

	filenames := make([]string, 0)

	for _, g := range idx.Inclusions {
		f, err := iofs.Glob(source, g)
		if err != nil {
			logger.Error("malformed glob pattern for included files", zap.String("glob", g), zap.Error(err))
			return nil, err
		}

		filenames = append(filenames, f...)
	}

	if len(idx.Exclusions) > 0 {
		for i := range filenames {
			anyMatch := false
			for _, e := range idx.Exclusions {
				match, _ := path.Match(e, filenames[i])
				anyMatch = anyMatch || match
			}

			if anyMatch {
				filenames = append(filenames[:i], filenames[i+1:]...)
			}
		}
	}

	return filenames, nil
}
