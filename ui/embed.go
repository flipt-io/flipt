//go:build assets
// +build assets

package ui

import "embed"

//go:embed all:dist/*
var UI embed.FS
