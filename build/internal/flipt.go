package internal

import (
	"context"
	"fmt"
	"path"
	"time"

	"github.com/containerd/platforms"
	"go.flipt.io/build/internal/dagger"
)

func Base(ctx context.Context, dag *dagger.Client, source, uiDist *dagger.Directory, platform platforms.Platform) (*dagger.Container, error) {
	var (
		goBuildCachePath = "/root/.cache/go-build"
		goModCachePath   = "/go/pkg/mod"
	)

	golang := dag.Container(dagger.ContainerOpts{
		Platform: dagger.Platform(platforms.Format(platform)),
	}).
		From("golang:1.25-alpine3.21").
		WithEnvVariable("GOCACHE", goBuildCachePath).
		WithEnvVariable("GOMODCACHE", goModCachePath).
		WithExec([]string{"apk", "add", "bash", "gcc", "binutils-gold", "build-base", "git"})
	if _, err := golang.Sync(ctx); err != nil {
		return nil, err
	}

	// Mount the main module's go.mod and go.sum
	src := dag.Directory().
		WithFile("go.mod", source.File("go.mod")).
		WithFile("go.sum", source.File("go.sum"))

	// Mount submodule dependency files referenced in replace directives
	submodules := []string{
		"core",
		"errors",
		"rpc/flipt",
		"sdk/go",
	}

	for _, submodule := range submodules {
		mod := path.Join(submodule, "go.mod")
		sum := path.Join(submodule, "go.sum")
		src = src.
			WithFile(mod, source.File(mod)).
			WithFile(sum, source.File(sum))
	}

	var (
		cacheGoBuild = dag.CacheVolume("go-build-cache")
		cacheGoMod   = dag.CacheVolume("go-mod-cache")
	)

	golang = golang.WithEnvVariable("GOOS", platform.OS).
		WithEnvVariable("GOARCH", platform.Architecture).
		WithMountedCache(goBuildCachePath, cacheGoBuild).
		WithMountedCache(goModCachePath, cacheGoMod).
		WithMountedDirectory("/src", src).
		WithWorkdir("/src")

	golang = golang.WithExec([]string{"go", "mod", "download"})
	if _, err := golang.Sync(ctx); err != nil {
		return nil, err
	}

	// fetch the rest of the project (- build & ui)
	project := source.
		WithoutDirectory("./.build/").
		WithoutDirectory("./ui/").
		WithoutDirectory("./bin/").
		WithoutDirectory("./.git/")

	golang = golang.WithEnvVariable("CGO_ENABLED", "1").
		WithMountedDirectory(".", project)

	// fetch and add ui/embed.go on its own
	embed := dag.Directory().WithFiles("./ui", []*dagger.File{
		source.File("./ui/dev.go"),
		source.File("./ui/embed.go"),
		source.File("./ui/index.dev.html"),
	})

	// TODO(georgemac): wire in commit and version ldflags
	var (
		ldflags    = fmt.Sprintf("-s -w -linkmode external -extldflags -static -X main.date=%s", time.Now().UTC().Format(time.RFC3339))
		path       = path.Join("/bin", platforms.Format(platform))
		goBuildCmd = fmt.Sprintf(
			"go build -trimpath -tags assets,netgo -o %s -ldflags='%s' ./...",
			path,
			ldflags,
		)
	)

	// build the Flipt target binary
	return golang.
		WithMountedDirectory("./ui", embed.Directory("./ui")).
		WithMountedDirectory("./ui/dist", uiDist).
		WithExec([]string{"mkdir", "-p", path}).
		WithExec([]string{"sh", "-c", goBuildCmd}), nil
}

// Package copies the Flipt binaries built into the provided flipt container
// into a thinner alpine distribution.
func Package(ctx context.Context, client *dagger.Client, flipt *dagger.Container) (*dagger.Container, error) {
	platform, err := flipt.Platform(ctx)
	if err != nil {
		return nil, err
	}

	// build container with just Flipt + config
	return client.Container().From("alpine:3.19").
		WithExec([]string{"apk", "add", "--no-cache", "postgresql-client", "openssl", "ca-certificates"}).
		WithExec([]string{"mkdir", "-p", "/var/opt/flipt"}).
		WithExec([]string{"mkdir", "-p", "/var/log/flipt"}).
		WithFile("/flipt",
			flipt.Directory(path.Join("/bin", platforms.Format(platforms.MustParse(string(platform))))).File("flipt")).
		WithFile("/etc/flipt/config/default.yml",
			flipt.Directory("/src/config").File("default.yml")).
		WithExec([]string{"addgroup", "flipt"}).
		WithExec([]string{"adduser", "-S", "-D", "-g", "''", "-G", "flipt", "-s", "/bin/sh", "flipt"}).
		WithExec([]string{"chown", "-R", "flipt:flipt", "/etc/flipt", "/var/opt/flipt", "/var/log/flipt"}).
		WithUser("flipt").
		WithDefaultArgs([]string{"/flipt"}), nil
}
