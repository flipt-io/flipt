package internal

import (
	"context"
	"fmt"
	"path"
	"time"

	"github.com/containerd/platforms"
	"go.flipt.io/build/internal/dagger"
	"golang.org/x/mod/modfile"
)

func Base(ctx context.Context, dag *dagger.Client, source, uiDist *dagger.Directory, platform platforms.Platform, registryCache ...string) (*dagger.Container, error) {
	var (
		goBuildCachePath = "/root/.cache/go-build"
		goModCachePath   = "/go/pkg/mod"
	)

	golang := dag.Container(dagger.ContainerOpts{
		Platform: dagger.Platform(platforms.Format(platform)),
	})

	// Use cache if available
	if len(registryCache) > 0 && registryCache[0] != "" {
		baseImageRef := fmt.Sprintf("%s:golang-%s", registryCache[0], "base-1.25")
		// Try cached base first
		cachedBase := golang.From(baseImageRef)
		if _, err := cachedBase.Sync(ctx); err == nil {
			golang = cachedBase
		} else {
			// Build fresh base and cache it
			golang = golang.From("golang:1.25-alpine3.21").
				WithEnvVariable("GOCACHE", goBuildCachePath).
				WithEnvVariable("GOMODCACHE", goModCachePath).
				WithExec([]string{"apk", "add", "bash", "gcc", "binutils-gold", "build-base", "git"})
			// Cache this base layer
			golang.Publish(ctx, baseImageRef)
		}
	} else {
		// No cache - use regular build
		golang = golang.From("golang:1.25-alpine3.21").
			WithEnvVariable("GOCACHE", goBuildCachePath).
			WithEnvVariable("GOMODCACHE", goModCachePath).
			WithExec([]string{"apk", "add", "bash", "gcc", "binutils-gold", "build-base", "git"})
		if _, err := golang.Sync(ctx); err != nil {
			return nil, err
		}
	}

	goWork := source.File("go.work")
	work, err := goWork.Contents(ctx)
	if err != nil {
		return nil, err
	}

	workFile, err := modfile.ParseWork("go.work", []byte(work), nil)
	if err != nil {
		return nil, err
	}

	// infer mod and sum files from the contents of the work file.
	src := dag.Directory().
		WithFile("go.work", goWork).
		WithFile("go.work.sum", source.File("go.work.sum"))

	for _, use := range workFile.Use {
		mod := path.Join(use.Path, "go.mod")
		sum := path.Join(use.Path, "go.sum")
		src = src.
			WithFile(mod, source.File(mod)).
			WithFile(sum, source.File(sum))
	}

	// Use cache volumes for Go modules
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
		WithoutDirectory("./bin/")

	golang = golang.WithMountedDirectory(".", project)

	// fetch and add ui/embed.go on its own
	embed := dag.Directory().WithFiles("./ui", []*dagger.File{
		source.File("./ui/dev.go"),
		source.File("./ui/embed.go"),
		source.File("./ui/index.dev.html"),
	})

	// For cache builds, return the base container without building the binary
	if len(registryCache) > 0 && registryCache[0] != "" {
		return golang.WithMountedDirectory("./ui", embed.Directory("./ui")), nil
	}

	gitCommit, err := golang.WithExec([]string{"git", "rev-parse", "HEAD"}).Stdout(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting git commit: %w", err)
	}

	// TODO(georgemac): wire in version ldflag
	var (
		ldflags = fmt.Sprintf("-s -w -linkmode external -extldflags -static -X main.date=%s -X main.commit=%s", time.Now().UTC().Format(time.RFC3339), gitCommit)
		path    = path.Join("/bin", platforms.Format(platform))
		// Note: -cover with -coverpkg=go.flipt.io/flipt/... ensures all packages in the module are instrumented
		goBuildCmd = fmt.Sprintf(
			"go build -cover -coverpkg=./... -trimpath -tags assets,netgo -o %s/flipt -ldflags='%s' ./cmd/flipt",
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
// into a thinner alpine distribution with coverage support.
// If registryCache is provided, it will try to use cached runtime base layers.
func Package(ctx context.Context, client *dagger.Client, flipt *dagger.Container, registryCache ...string) (*dagger.Container, error) {
	platform, err := flipt.Platform(ctx)
	if err != nil {
		return nil, err
	}

	var runtime *dagger.Container
	// Use cache if available
	if len(registryCache) > 0 && registryCache[0] != "" {
		runtimeImageRef := fmt.Sprintf("%s:runtime-base", registryCache[0])
		runtime = client.Container().From(runtimeImageRef)
		if _, err := runtime.Sync(ctx); err != nil {
			// Build fresh runtime base and cache it
			runtime = client.Container().From("alpine:3.21").
				WithExec([]string{"apk", "add", "--no-cache", "openssl", "ca-certificates"}).
				WithExec([]string{"mkdir", "-p", "/var/log/flipt"}).
				WithExec([]string{"addgroup", "flipt"}).
				WithExec([]string{"adduser", "-S", "-D", "-g", "''", "-G", "flipt", "-s", "/bin/sh", "flipt"}).
				WithExec([]string{"mkdir", "-p", "/tmp/coverage"}).
				WithExec([]string{"chown", "flipt:flipt", "/tmp/coverage"})
			// Cache this runtime base
			runtime.Publish(ctx, runtimeImageRef)
		}
	} else {
		// No cache - use regular build
		runtime = client.Container().From("alpine:3.21").
			WithExec([]string{"apk", "add", "--no-cache", "openssl", "ca-certificates"}).
			WithExec([]string{"mkdir", "-p", "/var/log/flipt"}).
			WithExec([]string{"addgroup", "flipt"}).
			WithExec([]string{"adduser", "-S", "-D", "-g", "''", "-G", "flipt", "-s", "/bin/sh", "flipt"}).
			WithExec([]string{"mkdir", "-p", "/tmp/coverage"}).
			WithExec([]string{"chown", "flipt:flipt", "/tmp/coverage"}) // Ensure flipt user can write to coverage dir
	}

	// Add the binary to the runtime container
	return runtime.WithFile("/flipt",
		flipt.Directory(path.Join("/bin", platforms.Format(platforms.MustParse(string(platform))))).File("flipt")).
		WithUser("flipt").
		WithDefaultArgs([]string{"/flipt", "server"}), nil
}

// PackageWithUIBuild builds the final container with UI assets and optional caching
// This function is used when we have a separate base container and need to add UI assets and build the binary
func PackageWithUIBuild(ctx context.Context, client *dagger.Client, base *dagger.Container, uiDist *dagger.Directory, platform platforms.Platform, registryCache ...string) (*dagger.Container, error) {
	// Get git commit for build
	gitCommit, err := base.WithExec([]string{"git", "rev-parse", "HEAD"}).Stdout(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting git commit: %w", err)
	}

	// Mount UI dist and build
	var (
		ldflags    = fmt.Sprintf("-s -w -linkmode external -extldflags -static -X main.date=%s -X main.commit=%s", time.Now().UTC().Format(time.RFC3339), gitCommit)
		binPath    = path.Join("/bin", platforms.Format(platform))
		goBuildCmd = fmt.Sprintf(
			"go build -cover -coverpkg=./... -trimpath -tags assets,netgo -o %s/flipt -ldflags='%s' ./cmd/flipt",
			binPath,
			ldflags,
		)
	)

	// Build the Flipt binaries with UI assets
	buildContainer := base.
		WithMountedDirectory("./ui/dist", uiDist).
		WithExec([]string{"mkdir", "-p", binPath}).
		WithExec([]string{"sh", "-c", goBuildCmd})

	var runtime *dagger.Container
	// Use cache if available
	if len(registryCache) > 0 && registryCache[0] != "" {
		runtimeImageRef := fmt.Sprintf("%s:runtime-base", registryCache[0])
		runtime = client.Container().From(runtimeImageRef)
		if _, err := runtime.Sync(ctx); err != nil {
			// Build fresh runtime base and cache it
			runtime = client.Container().From("alpine:3.21").
				WithExec([]string{"apk", "add", "--no-cache", "openssl", "ca-certificates"}).
				WithExec([]string{"mkdir", "-p", "/var/log/flipt"}).
				WithExec([]string{"addgroup", "flipt"}).
				WithExec([]string{"adduser", "-S", "-D", "-g", "''", "-G", "flipt", "-s", "/bin/sh", "flipt"}).
				WithExec([]string{"mkdir", "-p", "/tmp/coverage"}).
				WithExec([]string{"chown", "flipt:flipt", "/tmp/coverage"})
			// Cache this runtime base
			runtime.Publish(ctx, runtimeImageRef)
		}
	} else {
		// No cache - use regular build
		runtime = client.Container().From("alpine:3.21").
			WithExec([]string{"apk", "add", "--no-cache", "openssl", "ca-certificates"}).
			WithExec([]string{"mkdir", "-p", "/var/log/flipt"}).
			WithExec([]string{"addgroup", "flipt"}).
			WithExec([]string{"adduser", "-S", "-D", "-g", "''", "-G", "flipt", "-s", "/bin/sh", "flipt"}).
			WithExec([]string{"mkdir", "-p", "/tmp/coverage"}).
			WithExec([]string{"chown", "flipt:flipt", "/tmp/coverage"})
	}

	// Add the binary to the runtime container
	return runtime.WithFile("/flipt",
		buildContainer.Directory(binPath).File("flipt")).
		WithUser("flipt").
		WithDefaultArgs([]string{"/flipt", "server"}), nil
}
