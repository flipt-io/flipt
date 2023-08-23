//go:build !linux
// +build !linux

package config

import (
	"os"
)

func defaultDatabaseRoot() (string, error) {
	return os.UserConfigDir()
}
