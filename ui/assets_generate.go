// +build ignore

package main

import (
	"fmt"
	"os"

	"github.com/markphelps/flipt/internal/fs"
	"github.com/markphelps/flipt/ui"
	"github.com/shurcooL/vfsgen"
)

func main() {
	// Override all file mod times to be zero using ModTimeFS.
	inputFS := fs.NewModTimeFS(ui.Assets)

	err := vfsgen.Generate(inputFS, vfsgen.Options{
		PackageName:  "ui",
		BuildTags:    "!dev",
		VariableName: "Assets",
	})

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
