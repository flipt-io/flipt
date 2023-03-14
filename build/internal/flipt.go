package internal

import (
	"context"
	"fmt"
	"strings"

	"dagger.io/dagger"
	"github.com/containerd/containerd/platforms"
	specs "github.com/opencontainers/image-spec/specs-go/v1"
)

type FliptRequest struct {
	WorkDir     string
	ui          *dagger.Directory
	BuildTarget specs.Platform
	Target      specs.Platform
}

func (r FliptRequest) binary() string {
	return fmt.Sprintf("/bin/%s", platforms.Format(r.Target))
}

type Option func(*FliptRequest)

func WithWorkDir(dir string) Option {
	return func(r *FliptRequest) {
		r.WorkDir = dir
	}
}

func WithTarget(platform dagger.Platform) Option {
	return func(r *FliptRequest) {
		r.Target = platforms.MustParse(string(platform))
	}
}

func NewFliptRequest(ui *dagger.Directory, build dagger.Platform, opts ...Option) FliptRequest {
	platform := platforms.MustParse(string(build))
	req := FliptRequest{
		WorkDir:     ".",
		ui:          ui,
		BuildTarget: platform,
		// default target platform == build platform
		Target: platform,
	}

	for _, opt := range opts {
		opt(&req)
	}

	return req
}

func Base(ctx context.Context, client *dagger.Client, req FliptRequest) (*dagger.Container, error) {
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
		Platform: dagger.Platform(platforms.Format(req.BuildTarget)),
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

	golang = golang.WithEnvVariable("GOOS", req.Target.OS).
		WithEnvVariable("GOARCH", req.Target.Architecture).
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
			req.binary(),
			ldflags,
		)
	)

	// build the Flipt target binary
	return golang.
		WithMountedFile("./ui/embed.go", embed.File("./ui/embed.go")).
		WithMountedDirectory("./ui/dist", req.ui).
		WithExec([]string{"mkdir", "-p", req.binary()}).
		WithExec([]string{"sh", "-c", goBuildCmd}), nil
}

// Package copies the Flipt binaries built into the provoded flipt container
// into a thinner alpine distribution.
func Package(ctx context.Context, client *dagger.Client, flipt *dagger.Container, req FliptRequest) (*dagger.Container, error) {
	// build container with just Flipt + config
	return client.Container().From("alpine:3.16").
		WithExec([]string{"apk", "add", "--no-cache", "postgresql-client", "openssl", "ca-certificates"}).
		WithExec([]string{"mkdir", "-p", "/var/opt/flipt"}).
		WithFile("/bin/flipt",
			flipt.Directory(req.binary()).File("flipt")).
		WithFile("/etc/flipt/config/default.yml",
			flipt.Directory("/src/config").File("default.yml")).
		WithExec([]string{"addgroup", "flipt"}).
		WithExec([]string{"adduser", "-S", "-D", "-g", "''", "-G", "flipt", "-s", "/bin/sh", "flipt"}).
		WithExec([]string{"chown", "-R", "flipt:flipt", "/etc/flipt", "/var/opt/flipt"}).
		WithUser("flipt").
		WithDefaultArgs(dagger.ContainerWithDefaultArgsOpts{
			Args: []string{"/bin/flipt"},
		}), nil
}
