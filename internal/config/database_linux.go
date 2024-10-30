//go:build linux
// +build linux

package config

import (
	"os"

	"github.com/spf13/afero"
)

func defaultDatabaseRoot() (string, error) {
	return findDatabaseRoot(afero.NewOsFs())
}

func findDatabaseRoot(fs afero.Fs) (string, error) {
	preferred := "/var/opt/flipt"
	if _, err := fs.Stat(preferred); os.IsNotExist(err) {
		// if /var/opt/flipt doesn't exist fallback to ~/.config/flipt.
		// It's the case when flipt runs locally in linux for testing with sqlite.
		return Dir()
	}
	return preferred, nil
}
