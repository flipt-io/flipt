package testing

import (
	"context"
	"fmt"

	"github.com/google/go-github/v53/github"
	"github.com/google/uuid"
	"go.flipt.io/build/internal/dagger"
)

func Migration(ctx context.Context, client *dagger.Client, base, flipt *dagger.Container) error {
	dir := client.CacheVolume("flipt-state")
	latest, err := client.Container().
		From("flipt/flipt:latest").
		// support Flipt instances without /var/log/flipt configured
		WithUser("root").
		WithExec([]string{"mkdir", "-p", "/var/log/flipt"}).
		WithExec([]string{"chown", "-R", "flipt:flipt", "/var/log/flipt"}).
		WithUser("flipt").
		WithEnvVariable("FLIPT_LOG_FILE", "/var/log/flipt/output.txt").
		WithMountedCache("/var/opt/flipt", dir, dagger.ContainerWithMountedCacheOpts{
			Owner: "flipt",
		}).
		WithExec([]string{"/flipt", "-v"}).
		Sync(ctx)
	if err != nil {
		return err
	}

	release, _, err := github.NewClient(nil).
		Repositories.
		GetLatestRelease(ctx, "flipt-io", "flipt")
	if err != nil {
		return err
	}

	// clone the last release so we can use the readonly test
	// suite defined here instead
	fliptDir := client.Git("https://github.com/flipt-io/flipt.git").
		Tag(*release.TagName).
		Tree()

	base = base.
		WithMountedDirectory(".", fliptDir)

	for _, namespace := range []string{"default", "production"} {
		fi, err := base.File(fmt.Sprintf("build/testing/integration/readonly/testdata/main/%s.yaml", namespace)).Sync(ctx)
		if err != nil {
			return err
		}

		// import testdata into latest Flipt instance (using latest image for import)
		_, err = latest.
			WithFile("import.yaml", fi).
			WithServiceBinding("flipt", latest.WithExec(nil).AsService()).
			WithExec([]string{"sh", "-c", "sleep 5 && /flipt import --address grpc://flipt:9000 import.yaml"}).
			Sync(ctx)
		if err != nil {
			return err
		}
	}

	// run migration with edge Flipt build
	flipt, err = flipt.
		// persist state between latest and new version
		WithMountedCache("/var/opt/flipt", dir, dagger.ContainerWithMountedCacheOpts{
			Owner: "flipt",
		}).
		WithEnvVariable("UNIQUE", uuid.New().String()).
		WithExec([]string{"/flipt", "-v"}).
		WithExec([]string{"/flipt", "migrate"}).
		Sync(ctx)
	if err != nil {
		return err
	}

	// ensure new edge Flipt build continues to work as expected
	_, err = base.
		WithServiceBinding("flipt", flipt.
			WithEnvVariable("FLIPT_AUTHENTICATION_REQUIRED", "true").
			WithEnvVariable("FLIPT_AUTHENTICATION_METHODS_TOKEN_ENABLED", "true").
			WithEnvVariable("FLIPT_AUTHENTICATION_METHODS_TOKEN_BOOTSTRAP_TOKEN", "secret").
			WithExec(nil).
			AsService()).
		WithWorkdir("build/testing/integration/readonly").
		WithEnvVariable("UNIQUE", uuid.New().String()).
		WithExec([]string{"go", "test", "-v", "-race",
			"--flipt-addr", "grpc://flipt:9000",
			"--flipt-token", "secret",
			"."}).
		Sync(ctx)

	return err
}
