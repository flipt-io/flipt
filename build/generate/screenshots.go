package generate

import (
	"context"
	"log"
	"os"
	"path"
	"time"

	"go.flipt.io/build/internal/dagger"
	"golang.org/x/sync/errgroup"
)

func Screenshots(ctx context.Context, client *dagger.Client, source *dagger.Directory, flipt *dagger.Container) error {
	if err := os.RemoveAll("./tmp/screenshots"); err != nil {
		return err
	}

	src := client.Directory().WithFiles("ui", []*dagger.File{
		source.File("ui/package.json"),
		source.File("ui/package-lock.json"),
		source.File("ui/playwright.config.ts"),
	}).WithDirectory("ui/screenshot", source.Directory("ui/screenshot"))

	cache := client.CacheVolume("node-modules-screenshot")

	ui, err := client.Container().From("node:18-bullseye").
		WithMountedDirectory("/src", src).WithWorkdir("/src").
		WithMountedCache("/src/node_modules", cache).
		WithExec([]string{"npm", "install"}).
		WithExec([]string{"npx", "playwright", "install", "chromium", "--with-deps"}).
		Sync(ctx)
	if err != nil {
		return err
	}

	src = source.Directory("./ui/").
		WithoutDirectory("./dist/").
		WithoutDirectory("./node_modules/")

	// remount entire directory with module cache
	ui, err = ui.WithMountedDirectory("/src", src).
		WithMountedCache("/src/node_modules", cache).
		WithExec([]string{"npm", "install"}).
		Sync(ctx)
	if err != nil {
		return err
	}

	dirs := []string{
		"getting_started", "concepts", "configuration", "extra",
	}

	for _, dir := range dirs {
		var (
			g          errgroup.Group
			containers = make(chan *dagger.Container)
			dir        = dir
		)

		entries, err := ui.Directory("screenshot/" + dir).Entries(ctx)
		if err != nil {
			// skip if directory does not exist
			continue
		}

		go func() {
			_ = g.Wait()
			close(containers)
		}()

		for _, theme := range []string{"", "dark"} {
			theme := theme
			for _, entry := range entries {
				entry := entry
				g.Go(func() error {
					test, err := buildUI(ctx, ui, flipt, theme)
					if err != nil {
						return err
					}

					if ext := path.Ext(entry); ext != ".js" {
						return nil
					}

					c, err := test.WithExec([]string{"node", path.Join("screenshot", dir, entry)}).Sync(ctx)
					if err != nil {
						return err
					}

					containers <- c
					log.Printf("Generating screenshot for %s %s/%s\n", theme, dir, entry)

					return err
				})
			}
		}

		for c := range containers {
			if _, err := c.Directory("screenshots").
				Export(ctx, "./tmp/screenshots"); err != nil {
				return err
			}
		}
	}

	return err
}

func buildUI(ctx context.Context, ui, flipt *dagger.Container, theme string) (_ *dagger.Container, err error) {
	flipt, err = flipt.Sync(ctx)
	if err != nil {
		return nil, err
	}

	ui, err = ui.Sync(ctx)
	if err != nil {
		return nil, err
	}

	flipt = flipt.
		WithEnvVariable("CI", os.Getenv("CI")).
		WithEnvVariable("FLIPT_AUTHENTICATION_METHODS_TOKEN_ENABLED", "true").
		WithEnvVariable("UNIQUE", time.Now().String()).
		WithExposedPort(8080)

	if theme != "" {
		flipt = flipt.WithEnvVariable("FLIPT_UI_DEFAULT_THEME", theme)
	}

	return ui.
		WithServiceBinding("flipt", flipt.WithExec(nil).AsService()).WithFile("/usr/bin/flipt", flipt.File("/flipt")).
		WithEnvVariable("FLIPT_ADDRESS", "http://flipt:8080"), nil
}
