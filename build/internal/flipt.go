package internal

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"path"
	"time"

	"dagger.io/dagger"
	"github.com/containerd/containerd/platforms"
	specs "github.com/opencontainers/image-spec/specs-go/v1"
	"golang.org/x/mod/modfile"
)

type FliptRequest struct {
	WorkDir     string
	UI          *dagger.Container
	Platform    dagger.Platform
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

func NewFliptRequest(ui *dagger.Container, build dagger.Platform, opts ...Option) FliptRequest {
	platform := platforms.MustParse(string(build))
	req := FliptRequest{
		WorkDir:     ".",
		UI:          ui,
		Platform:    build,
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
	var (
		goBuildCachePath = "/root/.cache/go-build"
		goModCachePath   = "/go/pkg/mod"
	)

	golang := client.Container(dagger.ContainerOpts{
		Platform: dagger.Platform(platforms.Format(req.BuildTarget)),
	}).
		From("golang:1.22-alpine3.19").
		WithEnvVariable("GOCACHE", goBuildCachePath).
		WithEnvVariable("GOMODCACHE", goModCachePath).
		WithExec([]string{"apk", "add", "bash", "gcc", "binutils-gold", "build-base", "git"})
	if _, err := golang.Sync(ctx); err != nil {
		return nil, err
	}

	workFilePath := path.Join(req.WorkDir, "go.work")
	work, err := os.ReadFile(workFilePath)
	if err != nil {
		return nil, err
	}

	workFile, err := modfile.ParseWork(workFilePath, work, nil)
	if err != nil {
		return nil, err
	}

	// infer mod and sum files from the contents of the work file.
	includes := []string{
		"./go.work",
		"./go.work.sum",
	}

	for _, use := range workFile.Use {
		includes = append(includes,
			path.Join(use.Path, "go.mod"),
			path.Join(use.Path, "go.sum"),
		)
	}

	// add base dependencies to initialize project with
	src := client.Host().Directory(".", dagger.HostDirectoryOpts{
		Include: includes,
	})

	contents, err := src.File("go.sum").Contents(ctx)
	if err != nil {
		return nil, err
	}

	sum := fmt.Sprintf("%x", sha256.Sum256([]byte(contents)))

	var (
		cacheGoBuild = client.CacheVolume(fmt.Sprintf("go-build-%s", sum))
		cacheGoMod   = client.CacheVolume(fmt.Sprintf("go-mod-%s", sum))
	)

	golang = golang.WithEnvVariable("GOOS", req.Target.OS).
		WithEnvVariable("GOARCH", req.Target.Architecture).
		WithMountedCache(goBuildCachePath, cacheGoBuild).
		WithMountedCache(goModCachePath, cacheGoMod).
		WithMountedDirectory("/src", src).
		WithWorkdir("/src")

	golang = golang.WithExec([]string{"go", "mod", "download"})
	if _, err := golang.Sync(ctx); err != nil {
		return nil, err
	}

	// fetch the rest of the project (- build & ui)
	project := client.Host().Directory(".", dagger.HostDirectoryOpts{
		Exclude: []string{
			// The UI contains mostly the JS frontend.
			// However, it does contain a single go package,
			// which is used to embed the built frontend
			// distribution directory.
			"./.build/",
			"./ui/",
			"./bin/",
			"./.git/",
		},
	})

	golang = golang.WithEnvVariable("CGO_ENABLED", "1").
		WithMountedDirectory(".", project)

	// fetch and add ui/embed.go on its own
	embed := client.Host().Directory(".", dagger.HostDirectoryOpts{
		Include: []string{
			"./ui/dev.go",
			"./ui/embed.go",
			"./ui/index.dev.html",
		},
	})

	// TODO(georgemac): wire in commit and version ldflags
	var (
		ldflags    = fmt.Sprintf("-s -w -linkmode external -extldflags -static -X main.date=%s", time.Now().UTC().Format(time.RFC3339))
		goBuildCmd = fmt.Sprintf(
			"go build -trimpath -tags assets,netgo -o %s -ldflags='%s' ./...",
			req.binary(),
			ldflags,
		)
	)

	// build the Flipt target binary
	return golang.
		WithMountedDirectory("./ui", embed.Directory("./ui")).
		WithMountedDirectory("./ui/dist", req.UI.Directory("./dist")).
		// see: https://github.com/golang/go/issues/60825
		// should be fixed in go 1.20.6
		WithEnvVariable("GOEXPERIMENT", "nocoverageredesign").
		WithExec([]string{"mkdir", "-p", req.binary()}).
		WithExec([]string{"sh", "-c", goBuildCmd}), nil
}

// Package copies the Flipt binaries built into the provided flipt container
// into a thinner alpine distribution.
func Package(ctx context.Context, client *dagger.Client, flipt *dagger.Container, req FliptRequest) (*dagger.Container, error) {
	// build container with just Flipt + config
	return client.Container().From("alpine:3.19").
		WithExec([]string{"apk", "add", "--no-cache", "postgresql-client", "openssl", "ca-certificates"}).
		WithExec([]string{"mkdir", "-p", "/var/opt/flipt"}).
		WithExec([]string{"mkdir", "-p", "/var/log/flipt"}).
		WithFile("/flipt",
			flipt.Directory(req.binary()).File("flipt")).
		WithFile("/etc/flipt/config/default.yml",
			flipt.Directory("/src/config").File("default.yml")).
		WithExec([]string{"addgroup", "flipt"}).
		WithExec([]string{"adduser", "-S", "-D", "-g", "''", "-G", "flipt", "-s", "/bin/sh", "flipt"}).
		WithExec([]string{"chown", "-R", "flipt:flipt", "/etc/flipt", "/var/opt/flipt", "/var/log/flipt"}).
		WithUser("flipt").
		WithDefaultArgs(dagger.ContainerWithDefaultArgsOpts{
			Args: []string{"/flipt"},
		}), nil
}
