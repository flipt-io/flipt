package internal

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"dagger.io/dagger"
)

func parseUIRepoPath(ctx context.Context, client *dagger.Client, path string) (*dagger.Directory, error) {
	if !strings.HasPrefix(path, "git://") {
		if !strings.HasPrefix(path, "file://") {
			return nil, errors.New("unexpected ui repo path scheme")
		}

		path = path[len("file://"):]
		return client.Host().Directory(path, dagger.HostDirectoryOpts{
			Exclude: []string{
				"./dist/",
				"./node_modules/",
			},
		}), nil
	}

	path, ref, _ := strings.Cut(path[len("git://"):], "#")
	if ref == "" {
		ref = "main"
	}

	socket := client.Host().UnixSocket(os.Getenv("SSH_AUTH_SOCK"))
	return client.Git(path).
		Branch(ref).
		Tree(dagger.GitRefTreeOpts{
			SSHAuthSocket: socket,
		}), nil
}

func UI(ctx context.Context, client *dagger.Client, path string) (*dagger.Container, error) {
	src, err := parseUIRepoPath(ctx, client, path)
	if err != nil {
		return nil, err
	}

	id, err := src.File("package-lock.json").ID(ctx)
	if err != nil {
		return nil, err
	}

	cache := client.CacheVolume(fmt.Sprintf("node-modules-%s", id))

	return client.Container().From("node:18-alpine3.16").
		WithMountedDirectory("/src", src).WithWorkdir("/src").
		WithMountedCache("./ui/node_modules", cache).
		WithExec([]string{"npm", "ci"}).
		WithExec([]string{"npm", "run", "build"}), nil
}
