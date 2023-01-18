//go:build assets
// +build assets

package ui

import "embed"

var (
	//go:embed dist/*
	UI    embed.FS
	Mount = "dist"
)
