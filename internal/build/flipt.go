package build

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"dagger.io/dagger"
	"github.com/containerd/containerd/platforms"
	specs "github.com/opencontainers/image-spec/specs-go/v1"
)

type FliptRequest struct {
	ui     *dagger.Directory
	build  specs.Platform
	target specs.Platform
}

type Option func(*FliptRequest)

func WithTarget(platform dagger.Platform) Option {
	return func(r *FliptRequest) {
		r.target = platforms.MustParse(string(platform))
	}
}

func NewFliptRequest(ui *dagger.Directory, build dagger.Platform, opts ...Option) FliptRequest {
	platform := platforms.MustParse(string(build))
	req := FliptRequest{
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
			"./go.mod",
			"./go.sum",
			"./_tools/",
			"./script/",
		},
	})

	golang := client.Container(dagger.ContainerOpts{
		Platform: dagger.Platform(platforms.Format(req.build)),
	}).
		From("golang:1.18-alpine3.16").
		WithMountedDirectory("/src", src).
		WithWorkdir("/src").
		WithExec([]string{"apk", "add", "bash", "gcc", "binutils-gold", "build-base"})
	if _, err := golang.ExitCode(ctx); err != nil {
		return nil, err
	}

	target := fmt.Sprintf("./bin/%s", platforms.Format(req.target))

	goBuildCachePath, err := golang.WithExec([]string{"go", "env", "GOCACHE"}).Stdout(ctx)
	if err != nil {
		return nil, err
	}

	goModCachePath, err := golang.WithExec([]string{"go", "env", "GOMODCACHE"}).Stdout(ctx)
	if err != nil {
		return nil, err
	}

	sumID, err := src.File("go.sum").ID(ctx)
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
	golang = golang.WithExec([]string{"./script/bootstrap"})
	if _, err := golang.ExitCode(ctx); err != nil {
		return nil, err
	}

	// use go list to get the minimal subset of dirs needed to build Flipt.
	dirCMD := exec.Command("sh", "-c", "go list ./... | awk -F/ '{ print $3 }' | sort | uniq")
	out, err := dirCMD.Output()
	if err != nil {
		return nil, err
	}

	var dirs []string
	for _, dir := range strings.Split(string(out), "\n") {
		if dir != "" && dir != "ui" {
			dirs = append(dirs, dir)
		}
	}

	// fetch the rest of the project (- build & ui)
	project := client.Host().Directory(".", dagger.HostDirectoryOpts{
		Include: append(dirs,
			"go.mod",
			"go.sum",
		),
		Exclude: []string{
			// exclude the build project
			"./cmd/build/",
			"./internal/build/",
			"./ui/",
			"./bin/",
		},
	})

	// fetch ui/embed.go on its own
	embed := client.Host().Directory(".", dagger.HostDirectoryOpts{
		Include: []string{
			// exclude the build project
			"./ui/embed.go",
		},
	})

	cmd := fmt.Sprintf("go build -trimpath -tags assets,netgo -o %s -ldflags='-s -w -linkmode external -extldflags -static' ./...", target)
	golang = golang.
		WithEnvVariable("CGO_ENABLED", "1").
		WithMountedDirectory(".", project).
		WithMountedFile("./ui/embed.go", embed.File("./ui/embed.go")).
		WithMountedDirectory("./ui/dist", req.ui).
		WithExec([]string{"mkdir", "-p", target}).
		WithExec([]string{"sh", "-c", cmd})
	if _, err := golang.ExitCode(ctx); err != nil {
		return nil, err
	}

	return golang, nil
}
