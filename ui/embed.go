//go:build assets
// +build assets

package ui

import (
	"embed"
	"fmt"
	"io/fs"
)

//go:embed dist/*
var ui embed.FS

func FS() (fs.FS, error) {
	u, err := fs.Sub(ui, "dist")
	if err != nil {
		return nil, fmt.Errorf("embedding file: %w", err)
	}

	return u, nil
}
