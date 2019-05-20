// +build ignore

package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/markphelps/flipt/internal/fs"
	"github.com/shurcooL/vfsgen"
)

func main() {
	source := http.Dir("dist")

	// Override all file mod times to be zero using ModTimeFS.
	err := vfsgen.Generate(fs.NewModTimeFS(source), vfsgen.Options{
		PackageName:  "ui",
		VariableName: "Assets",
	})

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
