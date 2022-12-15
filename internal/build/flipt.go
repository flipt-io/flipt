package build

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"dagger.io/dagger"
)

func Flipt(ctx context.Context, client *dagger.Client, ui *dagger.Container) (*dagger.Container, error) {
	// add base dependencies to intialize project with
	src := client.Host().Directory(".", dagger.HostDirectoryOpts{
		Include: []string{
			"./go.mod",
			"./go.sum",
			"./_tools/",
			"./script/",
		},
	})

	golang := client.Container().From("golang:1.18-alpine3.16").
		WithMountedDirectory("/src", src).WithWorkdir("/src").
		WithExec([]string{"apk", "add", "bash", "gcc", "binutils-gold", "build-base"})
	if _, err := golang.ExitCode(ctx); err != nil {
		return nil, err
	}

	platform, err := golang.Platform(ctx)
	if err != nil {
		return nil, err
	}

	target := fmt.Sprintf("./bin/%s", platform)

	golang = golang.WithExec([]string{"go", "mod", "download"})
	if _, err := golang.ExitCode(ctx); err != nil {
		return nil, err
	}

	golang = golang.WithExec([]string{"./script/bootstrap"})
	if _, err := golang.ExitCode(ctx); err != nil {
		return nil, err
	}

	dirCMD := exec.Command("sh", "-c", "go list ./... | awk -F/ '{ print $3 }' | sort | uniq")
	if err != nil {
		return nil, err
	}

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

	goBuildCachePath = strings.TrimSpace(goBuildCachePath)
	goModCachePath = strings.TrimSpace(goModCachePath)

	cmd := fmt.Sprintf("go build -trimpath -tags assets -o %s ./...", target)
	golang = golang.
		WithMountedCache(goBuildCachePath, cacheGoBuild).
		WithMountedCache(goModCachePath, cacheGoMod).
		WithMountedDirectory(".", project).
		WithMountedFile("./ui/embed.go", embed.File("./ui/embed.go")).
		WithMountedDirectory("./ui/dist", ui.Directory("./dist")).
		WithExec([]string{"mkdir", "-p", target}).
		WithExec([]string{"sh", "-c", cmd})
	if _, err := golang.ExitCode(ctx); err != nil {
		return nil, err
	}

	return golang, nil
}
