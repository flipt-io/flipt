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

func AdditionalHeaders() map[string]string {
	return map[string]string{
		"X-Content-Type-Options":  "nosniff",
		"Content-Security-Policy": "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src * data:; frame-ancestors 'none'; connect-src 'self' https://app.formbricks.com; script-src-elem 'self' https://unpkg.com;",
	}
}
