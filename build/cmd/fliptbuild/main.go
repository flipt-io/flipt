package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"dagger.io/dagger"
	"go.flipt.io/flipt/build/internal"
)

var (
	uiVersion        int
	uiRepositoryPath string
	uiExport         bool
)

func main() {
	flag.IntVar(&uiVersion, "ui-version", 1, "Version of the UI to build")
	flag.StringVar(&uiRepositoryPath, "ui-path", "git://git@github.com:flipt-io/flipt-ui.git", "Path to UI V2 repository (file:// and git:// both supported)")
	flag.BoolVar(&uiExport, "ui-export", false, "Export the generated UI contents to ui/dist.")
	flag.Parse()

	ctx := context.Background()
	client, err := dagger.Connect(ctx,
		dagger.WithWorkdir(".."),
		dagger.WithLogOutput(os.Stdout),
	)
	if err != nil {
		panic(err)
	}
	defer client.Close()

	var ui *dagger.Container
	switch uiVersion {
	case 1:
		ui, err = internal.UI(ctx, client)
		if err != nil {
			panic(err)
		}
	case 2:
		ui, err = internal.UIV2(ctx, client, uiRepositoryPath)
		if err != nil {
			panic(err)
		}
	default:
		panic("unexpected UI version")
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

	req := internal.NewFliptRequest(dist, platform)
	flipt, err := internal.Flipt(ctx, client, req)
	if err != nil {
		panic(err)
	}

	out := fmt.Sprintf("flipt-%s.tar", strings.ReplaceAll(string(platform), "/", "-"))
	if _, err := flipt.Export(ctx, out); err != nil {
		panic(err)
	}
}
