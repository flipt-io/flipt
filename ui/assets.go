// +build dev

package ui

import (
	"go/build"
	"net/http"
)

const path = "github.com/markphelps/flipt/ui/dist"

// Assets contains project assets.
var (
	p, _                   = build.Import(path, "", build.FindOnly)
	Assets http.FileSystem = http.Dir(p.Dir)
)
