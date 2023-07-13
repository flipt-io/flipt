package testing

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"time"

	"dagger.io/dagger"
)

func UI(ctx context.Context, client *dagger.Client, ui, flipt *dagger.Container) error {
	test, err := buildUI(ctx, ui, flipt)
	if err != nil {
		return err
	}

	_, err = test.
		WithExec([]string{"npx", "playwright", "install", "chromium", "--with-deps"}).
		WithExec([]string{"npx", "playwright", "test"}).
		Directory("playwright-report").
		Export(ctx, "playwright-report")

	return err
}

func Screenshots(ctx context.Context, client *dagger.Client, flipt *dagger.Container) error {
	src := client.Host().Directory("./ui/", dagger.HostDirectoryOpts{
		Include: []string{
			"./package.json",
			"./package-lock.json",
			"./playwright.config.ts",
		},
	})

	contents, err := src.File("package-lock.json").Contents(ctx)
	if err != nil {
		return err
	}

	cache := client.CacheVolume(fmt.Sprintf("node-modules-new-%x", sha256.Sum256([]byte(contents))))

	ui, err := client.Container().From("node:18-bullseye").
		WithMountedDirectory("/src", src).WithWorkdir("/src").
		WithMountedCache("/src/node_modules", cache).
		WithExec([]string{"npm", "install"}).
		WithExec([]string{"npx", "playwright", "install", "chromium", "--with-deps"}).
		Sync(ctx)
	if err != nil {
		return err
	}

	src = client.Host().Directory("./ui/", dagger.HostDirectoryOpts{
		Exclude: []string{
			"./dist/",
			"./node_modules/",
		},
	})

	// remount entire directory with module cache
	ui, err = ui.WithMountedDirectory("/src", src).
		WithMountedCache("/src/node_modules", cache).
		WithExec([]string{"npm", "install"}).
		Sync(ctx)
	if err != nil {
		return err
	}

	test, err := buildUI(ctx, ui, flipt)
	if err != nil {
		return err
	}

	_, err = test.
		WithExec([]string{"node", "screenshot.js"}).
		Directory("screenshots").
		Export(ctx, "screenshots")

	return err
}

func buildUI(ctx context.Context, ui, flipt *dagger.Container) (_ *dagger.Container, err error) {
	flipt, err = flipt.Sync(ctx)
	if err != nil {
		return nil, err
	}

	ui, err = ui.Sync(ctx)
	if err != nil {
		return nil, err
	}

	return ui.
		WithServiceBinding("flipt", flipt.
			WithEnvVariable("CI", os.Getenv("CI")).
			WithEnvVariable("FLIPT_AUTHENTICATION_METHODS_TOKEN_ENABLED", "true").
			WithEnvVariable("UNIQUE", time.Now().String()).
			WithExec(nil)).
		WithEnvVariable("FLIPT_ADDRESS", "http://flipt:8080"), nil
}
