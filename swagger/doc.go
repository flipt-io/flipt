// Package swagger provides the assets via a virtual filesystem.
package swagger

import (
	// appease the linter
	_ "github.com/shurcooL/vfsgen"
)

//go:generate go run assets_generate.go
