//go:build assets
// +build assets

package ui

import "embed"

//go:embed dist/*
var UI embed.FS
var Mount = "/dist"