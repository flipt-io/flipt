//go:build mage
// +build mage

package main

import (
	"context"
	"fmt"
	"os"
	"path"

	"dagger.io/dagger"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"go.flipt.io/flipt/build/internal"
	"go.flipt.io/flipt/build/internal/publish"
	"go.flipt.io/flipt/build/internal/test"
	"go.flipt.io/flipt/build/release"
	"golang.org/x/mod/modfile"
	"golang.org/x/sync/errgroup"
)

// Build is a collection of targets which build Flipt into target Docker images.
type Build mg.Namespace

// Flipt builds a development version of Flipt as a Docker image and loads it into a local Docker instance.
func (b Build) Flipt(ctx context.Context) error {
	client, err := daggerClient(ctx)
	if err != nil {
		return err
	}

	defer client.Close()

	platform, err := client.DefaultPlatform(ctx)
	if err != nil {
		return err
	}

	req, err := newRequest(ctx, client, platform)
	if err != nil {
		return err
	}

	base, err := internal.Base(ctx, client, req)
	if err != nil {
		return err
	}

	flipt, err := internal.Package(ctx, client, base, req)
	if err != nil {
		return err
	}

	out, err := sh.Output("git", "rev-parse", "HEAD")
	if err != nil {
		return err
	}

	ref, err := publish.Publish(ctx, publish.PublishSpec{
		TargetType: publish.LocalTargetType,
		Target:     "flipt:dev-" + out[:7],
	}, client, publish.Variants{
		platform: flipt,
	})
	if err != nil {
		return err
	}

	fmt.Println("Successfully Built Flipt:", ref)

	return nil
}

// Base builds Flipts base image via Dagger and buildkit.
// This can be used for debugging or cache warming.
// There is no resulting artefact (only buildkit cache state).
func (b Build) Base(ctx context.Context) error {
	client, err := daggerClient(ctx)
	if err != nil {
		return err
	}

	defer client.Close()

	platform, err := client.DefaultPlatform(ctx)
	if err != nil {
		return err
	}

	req, err := newRequest(ctx, client, platform)
	if err != nil {
		return err
	}

	base, err := internal.Base(ctx, client, req)
	if err != nil {
		return err
	}

	_, err = base.ExitCode(ctx)
	return err
}

// Test contains all the targets used to test the Flipt base container.
type Test mg.Namespace

// All runs Flipt's unit test suite against all the databases Flipt supports.
func (t Test) All(ctx context.Context) error {
	var g errgroup.Group

	for db := range test.All {
		db := db
		g.Go(func() error {
			return t.Database(ctx, db)
		})
	}

	return g.Wait()
}

// Unit runs the base suite of tests for all of Flipt.
// It uses SQLite as the default database.
func (t Test) Unit(ctx context.Context) error {
	client, err := daggerClient(ctx)
	if err != nil {
		return err
	}

	defer client.Close()

	platform, err := client.DefaultPlatform(ctx)
	if err != nil {
		return err
	}

	req, err := newRequest(ctx, client, platform)
	if err != nil {
		return err
	}

	base, err := internal.Base(ctx, client, req)
	if err != nil {
		return err
	}

	return test.Test(ctx, client, base)
}

// Database runs the unit test suite against the desired database (one of ["sqlite" "postgres" "mysql" "cockroach"]).
func (t Test) Database(ctx context.Context, db string) error {
	client, err := daggerClient(ctx)
	if err != nil {
		return err
	}

	defer client.Close()

	platform, err := client.DefaultPlatform(ctx)
	if err != nil {
		return err
	}

	req, err := newRequest(ctx, client, platform)
	if err != nil {
		return err
	}

	base, err := internal.Base(ctx, client, req)
	if err != nil {
		return err
	}

	return test.Test(test.All[db](ctx, client, base))
}

// Integration runs the entire integration test suite.
// The suite runs a number of operations via the Go SDK against Flipt
// in various configurations using both HTTP and GRPC.
func (t Test) Integration(ctx context.Context) error {
	client, err := daggerClient(ctx)
	if err != nil {
		return err
	}

	defer client.Close()

	platform, err := client.DefaultPlatform(ctx)
	if err != nil {
		return err
	}

	req, err := newRequest(ctx, client, platform)
	if err != nil {
		return err
	}

	base, err := internal.Base(ctx, client, req)
	if err != nil {
		return err
	}

	flipt, err := internal.Package(ctx, client, base, req)
	if err != nil {
		return err
	}

	return test.Integration(ctx, client, base, flipt)
}

type Release mg.Namespace

func (r Release) Submodules(ctx context.Context, tag string) error {
	client, err := daggerClient(ctx)
	if err != nil {
		return err
	}

	return release.Submodules(ctx, client, tag)
}

func daggerClient(ctx context.Context) (*dagger.Client, error) {
	return dagger.Connect(ctx,
		dagger.WithWorkdir(workDir()),
		dagger.WithLogOutput(os.Stdout),
	)
}

func newRequest(ctx context.Context, client *dagger.Client, platform dagger.Platform) (internal.FliptRequest, error) {
	ui, err := internal.UI(ctx, client, uiRepositoryPath())
	if err != nil {
		return internal.FliptRequest{}, err
	}

	// write contents of container dist/ directory to the host
	dist := ui.Directory("./dist")

	return internal.NewFliptRequest(dist, platform, internal.WithWorkDir(workDir())), nil
}

func uiRepositoryPath() string {
	if path := os.Getenv("FLIPT_UI_PATH"); path != "" {
		return path
	}

	return "https://github.com/flipt-io/flipt-ui.git"
}

func workDir() string {
	curDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	mod, err := os.ReadFile(path.Join(curDir, "go.mod"))
	if err != nil {
		panic(err)
	}

	if modfile.ModulePath(mod) == "go.flipt.io/flipt/build" {
		return "../.."
	}

	return "."
}
