package internal

import (
	"context"
	"crypto/sha256"
	"fmt"

	"go.flipt.io/build/internal/dagger"
)

func UI(ctx context.Context, client *dagger.Client, source *dagger.Directory) (*dagger.Container, error) {
	contents, err := source.File("package-lock.json").Contents(ctx)
	if err != nil {
		return nil, err
	}

	cache := client.CacheVolume(fmt.Sprintf("node-modules-%x", sha256.Sum256([]byte(contents))))

	return client.Container().From("node:18-bullseye").
		WithMountedDirectory("/src", source.
			WithoutDirectory("dist").
			WithoutDirectory("node_modules")).
		WithWorkdir("/src").
		WithMountedCache("/src/node_modules", cache).
		WithExec([]string{"npm", "install"}).
		WithExec([]string{"npm", "run", "build"}), nil
}
