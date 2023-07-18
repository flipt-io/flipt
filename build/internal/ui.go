package internal

import (
	"context"
	"crypto/sha256"
	"fmt"

	"dagger.io/dagger"
)

func UI(ctx context.Context, client *dagger.Client) (*dagger.Container, error) {
	src := client.Host().Directory("./ui/", dagger.HostDirectoryOpts{
		Exclude: []string{
			"./dist/",
			"./node_modules/",
		},
	})

	contents, err := src.File("pnpm-lock.yaml").Contents(ctx)
	if err != nil {
		return nil, err
	}

	cache := client.CacheVolume(fmt.Sprintf("node-modules-%x", sha256.Sum256([]byte(contents))))

	return client.Container().From("node:18-bullseye").
		WithMountedDirectory("/src", src).WithWorkdir("/src").
		WithMountedCache("/src/node_modules", cache).
		WithExec([]string{"npm", "install", "-g", "pnpm"}).
		WithExec([]string{"pnpm", "install"}).
		WithExec([]string{"pnpm", "run", "build"}), nil
}
