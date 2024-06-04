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

	"github.com/containerd/containerd/platforms"
	"go.flipt.io/build/generate"
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
func (f *Flipt) Base(ctx context.Context, source *dagger.Directory) (*Container, error) {
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

func (f *Flipt) Build(ctx context.Context, source *dagger.Directory) (*Container, error) {
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

func (f *Flipt) Test(ctx context.Context, source *dagger.Directory) (*Test, error) {
	flipt, err := f.Build(ctx, source)
	if err != nil {
		return nil, err
	}

	return &Test{source, f.BaseContainer, f.UIContainer, flipt}, nil
}

func (t *Test) UI(ctx context.Context) error {
	return testing.UI(ctx, dag, t.UIContainer, t.FliptContainer)
}

func (t *Test) Unit(ctx context.Context) error {
	return testing.Unit(ctx, dag, t.BaseContainer)
}

func (t *Test) CLI(ctx context.Context) error {
	return testing.CLI(ctx, dag, t.Source, t.FliptContainer)
}

func (t *Test) Migration(ctx context.Context) error {
	return testing.Migration(ctx, dag, t.BaseContainer, t.FliptContainer)
}

func (t *Test) Load(ctx context.Context) error {
	return testing.LoadTest(ctx, dag, t.BaseContainer, t.FliptContainer)
}

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

	var tests []string
	if cases != "*" {
		tests = strings.Split(cases, " ")
	}

	return testing.Integration(ctx, dag, t.BaseContainer, t.FliptContainer, tests...)
}

type Generate struct {
	Source         *dagger.Directory
	FliptContainer *dagger.Container
}

func (f *Flipt) Generate(ctx context.Context, source *dagger.Directory) (*Generate, error) {
	flipt, err := f.Build(ctx, source)
	if err != nil {
		return nil, err
	}

	return &Generate{source, flipt}, nil
}

func (g *Generate) Screenshots(ctx context.Context) error {
	return generate.Screenshots(ctx, dag, g.Source, g.FliptContainer)
}
