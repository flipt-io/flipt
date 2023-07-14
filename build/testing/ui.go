package testing

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"path"
	"time"

	"dagger.io/dagger"
	"golang.org/x/sync/errgroup"
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
			"/screenshots/",
		},
	})

	contents, err := src.File("package-lock.json").Contents(ctx)
	if err != nil {
		return err
	}

	cache := client.CacheVolume(fmt.Sprintf("node-modules-screenshot-%x", sha256.Sum256([]byte(contents))))

	ui, err := client.Container().From("node:18-bullseye").
		WithMountedDirectory("/src", src).WithWorkdir("/src").
		WithMountedCache("/src/node_modules", cache).
		WithExec([]string{"npm", "install"}).
		WithExec([]string{"npx", "playwright@1.36.0", "install", "chromium", "--with-deps"}).
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

	entries, err := ui.Directory("screenshot").Entries(ctx)
	if err != nil {
		return err
	}

	var (
		g          errgroup.Group
		containers = make(chan *dagger.Container, 0)
	)

	go func() {
		g.Wait()
		close(containers)
	}()

	for _, entry := range entries {
		entry := entry
		g.Go(func() error {
			test, err := buildUI(ctx, ui, flipt)
			if err != nil {
				return err
			}

			if ext := path.Ext(entry); ext != ".js" {
				return nil
			}

			c, err := test.WithExec([]string{"node", path.Join("screenshot", entry)}).Sync(ctx)
			if err != nil {
				return err
			}

			containers <- c

			return err
		})
	}

	for c := range containers {
		fmt.Println("exporting")
		if _, err := c.Directory("screenshots").
			Export(ctx, "screenshots"); err != nil {
			return err
		}
	}

	return g.Wait()
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
		WithFile("/usr/bin/flipt", flipt.File("/flipt")).
		WithEnvVariable("FLIPT_ADDRESS", "http://flipt:8080"), nil
}
