package fs

import (
	"io"
	"os"
	"path"

	"go.flipt.io/flipt/errors"
)

const (
	DefaultFeaturesFilename = "features.yaml"
)

// TryOpenFeaturesFile attempts to open a features file with both .yaml and .yml extensions
// It returns the file reader, the filename that was successfully opened, and any error
func TryOpenFeaturesFile(fs Filesystem, dir string) (io.ReadCloser, string, error) {
	// Try .yaml first for backward compatibility
	for _, filename := range []string{"features.yaml", "features.yml"} {
		filePath := path.Join(dir, filename)
		fi, err := fs.OpenFile(filePath, os.O_RDONLY, 0644)
		if err == nil {
			return fi, filename, nil
		}
		if !errors.Is(err, os.ErrNotExist) {
			return nil, "", err
		}
	}
	return nil, "", os.ErrNotExist
}

// FindFeaturesFilename determines which features file exists in a directory
// It returns the existing filename, or the default filename for new files
func FindFeaturesFilename(fs Filesystem, dir string) (string, error) {
	for _, filename := range []string{"features.yaml", "features.yml"} {
		filePath := path.Join(dir, filename)
		if _, err := fs.Stat(filePath); err == nil {
			return filename, nil
		} else if !errors.Is(err, os.ErrNotExist) {
			return "", err
		}
	}
	return DefaultFeaturesFilename, nil // default to .yaml for new files
}
