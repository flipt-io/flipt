package main

import (
	"context"
	"os"

	"dagger.io/dagger"
	"go.flipt.io/flipt/build/internal"
)

func main() {
	ctx := context.Background()
	client, err := dagger.Connect(ctx,
		dagger.WithWorkdir(".."),
		dagger.WithLogOutput(os.Stdout),
	)
	if err != nil {
		panic(err)
	}
	defer client.Close()

	ui, err := internal.UI(ctx, client)
	if err != nil {
		panic(err)
	}

	// write contents of container dist/ directory to the host
	dist := ui.Directory("./dist")
	_, err = dist.Export(ctx, "./ui/dist")
	if err != nil {
		panic(err)
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

	_, err = flipt.Directory("./bin").Export(ctx, "./bin")
	if err != nil {
		panic(err)
	}
}
