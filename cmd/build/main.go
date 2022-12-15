package main

import (
	"context"
	"os"

	"dagger.io/dagger"
	"go.flipt.io/flipt/internal/build"
)

func main() {
	ctx := context.Background()
	client, err := dagger.Connect(ctx,
		dagger.WithWorkdir("."),
		dagger.WithLogOutput(os.Stdout),
	)
	if err != nil {
		panic(err)
	}
	defer client.Close()

	ui, err := build.UI(ctx, client)
	if err != nil {
		panic(err)
	}

	// write contents of container dist/ directory to the host
	_, err = ui.Directory("./dist").Export(ctx, "./ui/dist")
	if err != nil {
		panic(err)
	}

	flipt, err := build.Flipt(ctx, client, ui)
	if err != nil {
		panic(err)
	}

	_, err = flipt.Directory("./bin").Export(ctx, "./bin")
	if err != nil {
		panic(err)
	}
}
