package build

import (
	"context"
	"fmt"

	"dagger.io/dagger"
)

func UI(ctx context.Context, client *dagger.Client) (*dagger.Container, error) {
	// get reference to the local project
	src := client.Host().Directory("./ui", dagger.HostDirectoryOpts{
		Exclude: []string{
			"./dist/",
			"./node_modules/",
		},
	})

	id, err := src.File("package-lock.json").ID(ctx)
	if err != nil {
		return nil, err
	}

	cache := client.CacheVolume(fmt.Sprintf("node-modules-%s", id))

	node := client.Container().From("node:18-alpine3.16").
		WithMountedDirectory("/src", src).WithWorkdir("/src").
		WithMountedCache("./ui/node_modules", cache).
		WithExec([]string{"npm", "ci"})

	stdout, err := node.Stdout(ctx)
	if err != nil {
		return nil, err
	}
	fmt.Println(stdout)

	node = node.WithExec([]string{"npm", "run", "build"})
	stdout, err = node.Stdout(ctx)
	if err != nil {
		return nil, err
	}
	fmt.Println(stdout)

	return node, nil
}
