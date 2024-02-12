//go:build !assets
// +build !assets

package ui

import (
	_ "embed"
	"io/fs"
	"testing/fstest"
)

//go:embed index.dev.html
var ui []byte

func FS() (fs.FS, error) {
	// this allows us to serve 'index.dev.html' as 'index.html' so that it
	// is served correctly by http.FileServer
	return fstest.MapFS{
		"index.html": &fstest.MapFile{
			Data: ui,
		},
	}, nil
}

func AdditionalHeaders() map[string]string {
	return map[string]string{}
}
