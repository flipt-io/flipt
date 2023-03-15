package internal

import (
	"context"
	"fmt"
	"os"
	"strings"

	"dagger.io/dagger"
)

func parseUIRepoPath(ctx context.Context, client *dagger.Client, path string) (*dagger.Directory, error) {
	protocol, path, found := strings.Cut(path, "://")
	if !found {
		return nil, fmt.Errorf("protocol required: %q", path)
	}

	switch protocol {
	case "file":
		return client.Host().Directory(path, dagger.HostDirectoryOpts{
			Exclude: []string{
				"./dist/",
				"./node_modules/",
			},
		}), nil
	case "git", "http", "https":
	default:
		return nil, fmt.Errorf("unexpected protocol: %q", protocol)
	}

	path, ref, _ := strings.Cut(path, "#")
	if ref == "" {
		ref = "main"
	}

	var treeOpts dagger.GitRefTreeOpts
	if protocol == "git" {
		treeOpts.SSHAuthSocket = client.Host().UnixSocket(os.Getenv("SSH_AUTH_SOCK"))
	}

	return client.Git(protocol + "://" + path).
		Branch(ref).
		Tree(treeOpts), nil
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
