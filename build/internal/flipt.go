package internal

import (
	"context"
	"fmt"
	"strings"

	"dagger.io/dagger"
	"github.com/containerd/containerd/platforms"
	specs "github.com/opencontainers/image-spec/specs-go/v1"
	"go.flipt.io/flipt/build/internal/test"
)

type FliptRequest struct {
	dir    string
	ui     *dagger.Directory
	build  specs.Platform
	target specs.Platform
}

type Option func(*FliptRequest)

func WithWorkDir(dir string) Option {
	return func(r *FliptRequest) {
		r.dir = dir
	}
}

func WithTarget(platform dagger.Platform) Option {
	return func(r *FliptRequest) {
		r.target = platforms.MustParse(string(platform))
	}
}

func NewFliptRequest(ui *dagger.Directory, build dagger.Platform, opts ...Option) FliptRequest {
	platform := platforms.MustParse(string(build))
	req := FliptRequest{
		dir:   ".",
		ui:    ui,
		build: platform,
		// default target platform == build platform
		target: platform,
	}

	for _, opt := range opts {
		opt(&req)
	}

	return req
}

func Flipt(ctx context.Context, client *dagger.Client, req FliptRequest) (*dagger.Container, error) {
	// add base dependencies to intialize project with
	src := client.Host().Directory(".", dagger.HostDirectoryOpts{
		Include: []string{
			"./go.work",
			"./go.work.sum",
			"./go.mod",
			"./go.sum",
			"./build/go.mod",
			"./build/go.sum",
			"./rpc/flipt/go.mod",
			"./rpc/flipt/go.sum",
			"./errors/go.mod",
			"./errors/go.sum",
			"./_tools/",
			"./magefile.go",
		},
	})

	golang := client.Container(dagger.ContainerOpts{
		Platform: dagger.Platform(platforms.Format(req.build)),
	}).
		From("golang:1.18-alpine3.16").
		WithExec([]string{"apk", "add", "bash", "gcc", "binutils-gold", "build-base", "git"})
	if _, err := golang.ExitCode(ctx); err != nil {
		return nil, err
	}

	golang = golang.
		WithWorkdir("/deps").
		WithExec([]string{"git", "clone", "https://github.com/magefile/mage"}).
		WithWorkdir("/deps/mage").
		WithExec([]string{"go", "run", "bootstrap.go"}).
		WithMountedDirectory("/src", src).
		WithWorkdir("/src")

	target := fmt.Sprintf("/bin/%s", platforms.Format(req.target))

	goBuildCachePath, err := golang.WithExec([]string{"go", "env", "GOCACHE"}).Stdout(ctx)
	if err != nil {
		return nil, err
	}

	goModCachePath, err := golang.WithExec([]string{"go", "env", "GOMODCACHE"}).Stdout(ctx)
	if err != nil {
		return nil, err
	}

	sumID, err := src.File("go.work.sum").ID(ctx)
	if err != nil {
		return nil, err
	}

	var (
		cacheGoBuild = client.CacheVolume(fmt.Sprintf("go-build-%s", sumID))
		cacheGoMod   = client.CacheVolume(fmt.Sprintf("go-mod-%s", sumID))
	)

	golang = golang.WithEnvVariable("GOOS", req.target.OS).
		WithEnvVariable("GOARCH", req.target.Architecture).
		// sanitize output as it returns with a \n on the end
		// and that breaks the mount silently
		WithMountedCache(strings.TrimSpace(goBuildCachePath), cacheGoBuild).
		WithMountedCache(strings.TrimSpace(goModCachePath), cacheGoMod)

	golang = golang.WithExec([]string{"go", "mod", "download"})
	if _, err := golang.ExitCode(ctx); err != nil {
		return nil, err
	}

	golang = golang.WithExec([]string{"mage", "bootstrap"})
	if _, err := golang.ExitCode(ctx); err != nil {
		return nil, err
	}

	// fetch the rest of the project (- build & ui)
	project := client.Host().Directory(".", dagger.HostDirectoryOpts{
		Exclude: []string{
			// The UI contains mostly the JS frontend.
			// However, it does contain a single go package,
			// which is used to embed the built frontend
			// distribution directory.
			"./ui/",
			"./.build/",
			"./bin/",
			"./.git/",
		},
	})

	golang = golang.WithEnvVariable("CGO_ENABLED", "1").
		WithMountedDirectory(".", project)

	// fetch and add ui/embed.go on its own
	embed := client.Host().Directory(".", dagger.HostDirectoryOpts{
		Include: []string{
			"./ui/embed.go",
		},
	})

	var (
		ldflags    = "-s -w -linkmode external -extldflags -static"
		goBuildCmd = fmt.Sprintf(
			"go build -trimpath -tags assets,netgo -o %s -ldflags='%s' ./...",
			target,
			ldflags,
		)
	)

	// build the Flipt target binary
	golang = golang.
		WithMountedFile("./ui/embed.go", embed.File("./ui/embed.go")).
		WithMountedDirectory("./ui/dist", req.ui).
		WithExec([]string{"mkdir", "-p", target}).
		WithExec([]string{"sh", "-c", goBuildCmd})

	if _, err := golang.ExitCode(ctx); err != nil {
		return nil, err
	}

	if err := test.Test(ctx, client, golang); err != nil {
		return nil, err
	}

	// build container with just Flipt + config
	return client.Container().From("alpine:3.16").
		WithExec([]string{"apk", "add", "--no-cache", "postgresql-client", "openssl", "ca-certificates"}).
		WithExec([]string{"mkdir", "-p", "/var/opt/flipt"}).
		WithFile("/bin/flipt",
			golang.Directory(target).File("flipt")).
		WithFile("/etc/flipt/config/default.yml",
			golang.Directory("/src/config").File("default.yml")).
		WithExec([]string{"addgroup", "flipt"}).
		WithExec([]string{"adduser", "-S", "-D", "-g", "''", "-G", "flipt", "-s", "/bin/sh", "flipt"}).
		WithExec([]string{"chown", "-R", "flipt:flipt", "/etc/flipt", "/var/opt/flipt"}).
		WithUser("flipt").
		WithDefaultArgs(dagger.ContainerWithDefaultArgsOpts{
			Args: []string{"/bin/flipt"},
		}), nil
}
