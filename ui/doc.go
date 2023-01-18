//go:build !assets
// +build !assets

package ui

import "embed"

var (
	//go:embed index.html
	UI    embed.FS
	Mount = "."
)
