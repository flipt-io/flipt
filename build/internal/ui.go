package internal

import (
	"context"

	"go.flipt.io/build/internal/dagger"
)

func UI(ctx context.Context, client *dagger.Client, source *dagger.Directory) (*dagger.Container, error) {
	cache := client.CacheVolume("node-modules-cache")

	return client.Container().From("node:18-bullseye").
		WithMountedDirectory("/src", source.
			WithoutDirectory("dist").
			WithoutDirectory("node_modules")).
		WithWorkdir("/src").
		WithMountedCache("/src/node_modules", cache).
		WithExec([]string{"npm", "install"}).
		WithExec([]string{"npm", "run", "build"}), nil
}
