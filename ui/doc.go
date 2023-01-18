//go:build !assets
// +build !assets

package ui

import "embed"

//go:embed index.html
var UI embed.FS
var Mount = "."