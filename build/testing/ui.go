package testing

import (
	"context"
	"os"
	"time"

	"go.flipt.io/build/internal/dagger"
	"go.flipt.io/stew/config"
	"gopkg.in/yaml.v3"
)

func UI(ctx context.Context, client *dagger.Client, base, flipt *dagger.Container, ui *dagger.Directory, trace bool) (*dagger.Container, error) {
	// create unique instance for test case
	gitea := client.Container().
		From("gitea/gitea:1.21.1").
		WithExposedPort(3000).
		WithEnvVariable("UNIQUE", time.Now().String()).
		AsService()

	contents, err := yaml.Marshal(&config.Config{
		URL: "http://gitea:3000",
		Admin: struct {
			Username string "json:\"username\""
			Email    string "json:\"email\""
			Password string "json:\"password\""
		}{
			Username: "root",
			Password: "password",
			Email:    "dev@flipt.io",
		},
		Repositories: []config.Repository{
			{
				Name: "features",
				Contents: []config.Content{
					{
						Branch:  "main",
						Path:    "/work/base",
						Message: "feat: add directory contents",
					},
				},
			},
		},
	})
	if err != nil {
		return nil, err
	}

	// Wait for Gitea to be ready before running stew
	_, err = client.Container().
		From("ghcr.io/flipt-io/stew:latest").
		WithWorkdir("/work").
		WithDirectory("/work/base", base.Directory(environmentsTestdataDir)).
		WithNewFile("/etc/stew/config.yml", string(contents)).
		WithServiceBinding("gitea", gitea).
		WithExec([]string{"sh", "-c", `
			# Install wget if not available
			if ! command -v wget >/dev/null 2>&1; then
				apk add --no-cache wget >/dev/null 2>&1 || true
			fi
			# Wait for Gitea to be ready (up to 60 seconds)
			for i in $(seq 1 60); do
				if wget -q --spider http://gitea:3000 2>/dev/null || nc -z gitea 3000 2>/dev/null; then
					echo "Gitea is ready"
					break
				fi
				echo "Waiting for Gitea... ($i/60)"
				sleep 1
			done
		`}).
		WithExec([]string{"/usr/local/bin/stew", "-config", "/etc/stew/config.yml"}).
		Sync(ctx)
	if err != nil {
		return nil, err
	}

	flipt = flipt.
		WithServiceBinding("gitea", gitea).
		WithEnvVariable("FLIPT_LOG_LEVEL", "WARN").
		WithEnvVariable("FLIPT_ENVIRONMENTS_PRODUCTION_HOST", "flipt").
		WithEnvVariable("FLIPT_ENVIRONMENTS_PRODUCTION_ORGANIZATION", "myorg").
		WithEnvVariable("FLIPT_ENVIRONMENTS_PRODUCTION_SOURCE", "gitea").
		WithEnvVariable("FLIPT_SOURCES_GITEA_TYPE", "git").
		WithEnvVariable("FLIPT_SOURCES_GITEA_GIT_REPOSITORY", "http://gitea:3000/root/features.git").
		WithEnvVariable("FLIPT_SOURCES_GITEA_GIT_AUTHENTICATION_BASIC_USERNAME", "root").
		WithEnvVariable("FLIPT_SOURCES_GITEA_GIT_AUTHENTICATION_BASIC_PASSWORD", "password")

	test, err := buildUI(ctx, client, flipt, ui)
	if err != nil {
		return nil, err
	}

	if trace {
		return test.WithExec([]string{"sh", "-c", "npx playwright test --workers=1 --trace on || exit 0"}), nil
	}

	return test.WithExec([]string{"npx", "playwright", "test", "--workers=1"}), nil
}

func buildUI(ctx context.Context, client *dagger.Client, flipt *dagger.Container, source *dagger.Directory) (_ *dagger.Container, err error) {
	cache := client.CacheVolume("node-modules-cache")

	ui, err := client.Container().From("node:22-bullseye-slim").
		// initially mount only the package json
		WithMountedDirectory("/src", client.Directory().
			WithFile("package.json", source.File("package.json")).
			WithFile("package-lock.json", source.File("package-lock.json"))).
		WithWorkdir("/src").
		WithMountedCache("/src/node_modules", cache).
		// install dependencies for build and test
		WithExec([]string{"npm", "install"}).
		WithExec([]string{"npx", "playwright", "install", "chromium", "--only-shell", "--with-deps"}).
		// mount the rest of the project
		WithMountedDirectory("/src", source.
			WithoutDirectory("dist").
			WithoutDirectory("node_modules")).
		// build assets
		WithMountedCache("/src/node_modules", cache).
		WithExec([]string{"npm", "run", "build"}).
		Sync(ctx)
	if err != nil {
		return nil, err
	}

	flipt, err = flipt.Sync(ctx)
	if err != nil {
		return nil, err
	}

	return ui.
		WithServiceBinding("flipt", flipt.
			WithEnvVariable("CI", os.Getenv("CI")).
			WithEnvVariable("UNIQUE", time.Now().String()).
			AsService()).
		WithFile("/usr/bin/flipt", flipt.File("/flipt")).
		WithEnvVariable("FLIPT_ADDRESS", "http://flipt:8080").
		// Wait for Flipt service to be ready before running tests
		WithExec([]string{"sh", "-c", `
			# Install wget if not available (node container should have it or curl)
			if ! command -v wget >/dev/null 2>&1 && ! command -v curl >/dev/null 2>&1; then
				apt-get update -qq && apt-get install -qq -y wget >/dev/null 2>&1 || true
			fi
			# Wait for Flipt to be ready (up to 60 seconds)
			for i in $(seq 1 60); do
				if wget -q --spider http://flipt:8080/health 2>/dev/null || curl -f -s http://flipt:8080/health >/dev/null 2>&1 || wget -q --spider http://flipt:8080 2>/dev/null || curl -f -s http://flipt:8080 >/dev/null 2>&1; then
					echo "Flipt is ready"
					break
				fi
				echo "Waiting for Flipt... ($i/60)"
				sleep 1
			done
		`}), nil
}
