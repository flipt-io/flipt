// A generated module for Flipt functions
//
// This module has been generated via dagger init and serves as a reference to
// basic module structure as you get started with Dagger.
//
// Two functions have been pre-created. You can modify, delete, or add to them,
// as needed. They demonstrate usage of arguments and return types using simple
// echo and grep commands. The functions can be called from the dagger CLI or
// from one of the SDKs.
//
// The first line in this comment block is a short description line and the
// rest is a long description with more detail on the module's purpose or usage,
// if appropriate. All modules should have a short description.

package main

import (
	"context"
	"fmt"
	"strings"

	"dagger.io/dagger/dag"
	"github.com/containerd/platforms"
	"go.flipt.io/build/internal"
	"go.flipt.io/build/internal/dagger"
	"go.flipt.io/build/testing"
)

type Flipt struct {
	Source        *dagger.Directory
	BaseContainer *dagger.Container
	UIContainer   *dagger.Container
}

// Returns a container with all the assets compiled and ready for testing and distribution (with coverage enabled)
func (f *Flipt) Base(ctx context.Context, source *dagger.Directory) (*dagger.Container, error) {
	platform, err := dag.DefaultPlatform(ctx)
	if err != nil {
		return nil, err
	}

	f.UIContainer, err = internal.UI(ctx, dag, source.Directory("ui"))
	if err != nil {
		return nil, err
	}

	f.BaseContainer, err = internal.Base(ctx, dag, source, f.UIContainer.Directory("dist"), platforms.MustParse(string(platform)))
	return f.BaseContainer, err
}

// Return container with Flipt binaries in a thinner alpine distribution (with coverage enabled)
// If registryCache is provided, it will use cached layers and push the result to the registry
func (f *Flipt) Build(
	ctx context.Context,
	source *dagger.Directory,
	// +optional
	// +default=""
	// Registry to push cached image to (e.g., "ghcr.io/owner/repo-cache")
	registryCache string,
	// +optional
	// +default=""
	// Tag for the cached image (e.g., "base-abc123")
	cacheTag string,
) (*dagger.Container, error) {
	if registryCache != "" && cacheTag != "" {
		// Build with cache and push to registry
		platform, err := dag.DefaultPlatform(ctx)
		if err != nil {
			return nil, err
		}

		// Use the base container with registry cache (returns base without binary for cache builds)
		f.BaseContainer, err = internal.Base(ctx, dag, source, nil, platforms.MustParse(string(platform)), registryCache)
		if err != nil {
			return nil, err
		}

		// Build UI with cache
		f.UIContainer, err = internal.UI(ctx, dag, source.Directory("ui"), registryCache)
		if err != nil {
			return nil, err
		}

		// Build final container using cached layers and UI assets
		container, err := internal.PackageWithUIBuild(ctx, dag, f.BaseContainer, f.UIContainer.Directory("dist"), platforms.MustParse(string(platform)), registryCache)
		if err != nil {
			return nil, err
		}

		// Push to registry for caching
		imageRef := fmt.Sprintf("%s:%s", registryCache, cacheTag)
		_, err = container.Publish(ctx, imageRef)
		if err != nil {
			return nil, fmt.Errorf("failed to push cache image: %w", err)
		}

		return container, nil
	}

	// Regular build without cache
	base, err := f.Base(ctx, source)
	if err != nil {
		return nil, err
	}

	return internal.Package(ctx, dag, base)
}

type Test struct {
	Source         *dagger.Directory
	BaseContainer  *dagger.Container
	UIContainer    *dagger.Container
	FliptContainer *dagger.Container
}

// Execute test specific by subcommand
// see all available subcommands with dagger call test --help
func (f *Flipt) Test(
	ctx context.Context,
	source *dagger.Directory,
	// +optional
	// +default=""
	// Pre-built cached image to use for testing (e.g., "ghcr.io/owner/repo-cache:tag")
	cachedImage string,
) (*Test, error) {
	var flipt *dagger.Container
	var err error

	// If cached image is provided, use it directly
	if cachedImage != "" {
		flipt = dag.Container().From(cachedImage)
		// Use a lightweight base container for test execution when using cache
		// Lightweight test runner - just needs Go toolchain, not full build environment
		f.BaseContainer = dag.Container().
			From("golang:1.25-alpine3.21").
			WithExec([]string{"apk", "add", "--no-cache", "bash", "git"}).
			WithMountedDirectory("/src", source).
			WithWorkdir("/src")
	} else {
		// Fall back to normal build process
		flipt, err = f.Build(ctx, source, "", "")
		if err != nil {
			return nil, err
		}
		// Fall back to full base container for non-cached builds
		f.BaseContainer, err = f.Base(ctx, source)
		if err != nil {
			return nil, err
		}
	}

	return &Test{source, f.BaseContainer, f.UIContainer, flipt}, nil
}

// Run all ui tests
func (t *Test) UI(
	ctx context.Context,
	//+optional
	//+default=false
	trace bool,
) (*dagger.Container, error) {
	return testing.UI(ctx, dag, t.BaseContainer, t.FliptContainer, t.Source.Directory("ui"), trace)
}

// Run all unit tests
func (t *Test) Unit(ctx context.Context) (*dagger.File, error) {
	return testing.Unit(ctx, dag, t.BaseContainer)
}

// Run all integration tests (now with coverage collection by default)
func (t *Test) Integration(
	ctx context.Context,
	// +optional
	// +default="*"
	cases string,
	// +optional
	// +default=false
	outputCoverage bool,
) (*dagger.Directory, error) {
	if cases == "list" {
		fmt.Println("Integration test cases:")
		for c := range testing.AllCases {
			fmt.Println("\t> ", c)
		}

		return nil, nil
	}

	var opts []testing.IntegrationOptions
	if cases != "*" {
		opts = append(opts, testing.WithTestCases(strings.Split(cases, " ")...))
	}
	if outputCoverage {
		opts = append(opts, testing.WithCoverageOutput())
	}

	return testing.Integration(ctx, dag, t.BaseContainer, t.FliptContainer, opts...)
}

// CheckCacheExists checks if a cached image exists in the registry
func (f *Flipt) CheckCacheExists(
	ctx context.Context,
	// Registry and tag to check (e.g., "ghcr.io/owner/repo-cache:base-abc123")
	imageRef string,
) (bool, error) {
	// Try to pull the image to check if it exists
	container := dag.Container().From(imageRef)
	_, err := container.Sync(ctx)
	if err != nil {
		// Image doesn't exist or can't be pulled
		return false, nil
	}
	return true, nil
}
