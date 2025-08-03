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

// Returns a container with all the assets compiled and ready for testing and distribution
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

// Returns a container with coverage-enabled assets compiled and ready for testing and distribution
func (f *Flipt) BaseCoverage(ctx context.Context, source *dagger.Directory) (*dagger.Container, error) {
	platform, err := dag.DefaultPlatform(ctx)
	if err != nil {
		return nil, err
	}

	f.UIContainer, err = internal.UI(ctx, dag, source.Directory("ui"))
	if err != nil {
		return nil, err
	}

	f.BaseContainer, err = internal.BaseCoverage(ctx, dag, source, f.UIContainer.Directory("dist"), platforms.MustParse(string(platform)))
	return f.BaseContainer, err
}

// Return container with Flipt binaries in a thinner alpine distribution
func (f *Flipt) Build(ctx context.Context, source *dagger.Directory) (*dagger.Container, error) {
	base, err := f.Base(ctx, source)
	if err != nil {
		return nil, err
	}

	return internal.Package(ctx, dag, base)
}

// BuildCoverage returns a container with coverage-enabled Flipt binaries
func (f *Flipt) BuildCoverage(ctx context.Context, source *dagger.Directory) (*dagger.Container, error) {
	base, err := f.BaseCoverage(ctx, source)
	if err != nil {
		return nil, err
	}

	return internal.PackageCoverage(ctx, dag, base)
}

type Test struct {
	Source         *dagger.Directory
	BaseContainer  *dagger.Container
	UIContainer    *dagger.Container
	FliptContainer *dagger.Container
}

// Execute test specific by subcommand
// see all available subcommands with dagger call test --help
func (f *Flipt) Test(ctx context.Context, source *dagger.Directory) (*Test, error) {
	flipt, err := f.Build(ctx, source)
	if err != nil {
		return nil, err
	}

	return &Test{source, f.BaseContainer, f.UIContainer, flipt}, nil
}

// Execute test with coverage-enabled binaries
func (f *Flipt) TestCoverage(ctx context.Context, source *dagger.Directory) (*Test, error) {
	flipt, err := f.BuildCoverage(ctx, source)
	if err != nil {
		return nil, err
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

// Run all integration tests
func (t *Test) Integration(
	ctx context.Context,
	// +optional
	// +default="*"
	cases string,
) error {
	if cases == "list" {
		fmt.Println("Integration test cases:")
		for c := range testing.AllCases {
			fmt.Println("\t> ", c)
		}

		return nil
	}

	var opts []testing.IntegrationOptions
	if cases != "*" {
		opts = append(opts, testing.WithTestCases(strings.Split(cases, " ")...))
	}

	return testing.Integration(ctx, dag, t.BaseContainer, t.FliptContainer, opts...)
}

// Run all integration tests with coverage collection
func (t *Test) IntegrationCoverage(
	ctx context.Context,
	// +optional
	// +default="*"
	cases string,
) (*dagger.File, error) {
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

	return testing.IntegrationCoverage(ctx, dag, t.BaseContainer, t.FliptContainer, opts...)
}
