//go:build !assets
// +build !assets

package ui

import "embed"

var (
	//go:embed dev.html
	UI    embed.FS
	Mount = "."
)
