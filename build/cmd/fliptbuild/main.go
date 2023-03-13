package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path"

	"dagger.io/dagger"
	"go.flipt.io/flipt/build/internal"
	"go.flipt.io/flipt/build/internal/publish"
	"golang.org/x/mod/modfile"
)

var (
	uiRepositoryPath string
	uiExport         bool
	publishImage     string
)

func main() {
	flag.StringVar(&uiRepositoryPath, "ui-path", "git://git@github.com:flipt-io/flipt-ui.git", "Path to UI V2 repository (file:// and git:// both supported)")
	flag.BoolVar(&uiExport, "ui-export", false, "Export the generated UI contents to ui/dist.")
	flag.StringVar(&publishImage, "publish", "", "Publish image to remote or local image repository (docker:// or docker-local://)")
	flag.Parse()

	curDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	mod, err := os.ReadFile(path.Join(curDir, "go.mod"))
	if err != nil {
		panic(err)
	}

	workDir := "."
	if modfile.ModulePath(mod) == "go.flipt.io/flipt/build" {
		workDir = ".."
	}

	ctx := context.Background()
	client, err := dagger.Connect(ctx,
		dagger.WithWorkdir(workDir),
		dagger.WithLogOutput(os.Stdout),
	)
	if err != nil {
		panic(err)
	}
	defer client.Close()

	ui, err := internal.UI(ctx, client, uiRepositoryPath)
	if err != nil {
		panic(err)
	}

	// write contents of container dist/ directory to the host
	dist := ui.Directory("./dist")
	if uiExport {
		_, err = dist.Export(ctx, "./ui/dist")
		if err != nil {
			panic(err)
		}
	}

	platform, err := client.DefaultPlatform(ctx)
	if err != nil {
		panic(err)
	}

	req := internal.NewFliptRequest(dist, platform, internal.WithWorkDir(workDir))
	flipt, err := internal.Flipt(ctx, client, req)
	if err != nil {
		panic(err)
	}

	if publishImage != "" {
		ref, err := publish.Publish(ctx, publishImage, flipt)
		if err != nil {
			panic(err)
		}

		fmt.Println("Image Published:", ref)
		return
	}

	if _, err := flipt.ExitCode(ctx); err != nil {
		panic(err)
	}
}
