//go:build !linux
// +build !linux

package config

func defaultDatabaseRoot() (string, error) {
	return Dir()
}
