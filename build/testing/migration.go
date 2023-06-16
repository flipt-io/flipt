package testing

import (
	"context"
	"fmt"
	"time"

	"dagger.io/dagger"
	"github.com/google/uuid"
)

func Migration(ctx context.Context, client *dagger.Client, base, flipt *dagger.Container) error {
	dir := client.CacheVolume(fmt.Sprintf("flipt-state-%s", time.Now()))
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

	// import testdata into latest Flipt instance
	_, err = flipt.
		WithFile("import.yaml", base.File("build/testing/integration/readonly/testdata/default.yaml")).
		WithServiceBinding("flipt", latest.WithExec(nil)).
		WithExec([]string{"/flipt", "import", "--address", "grpc://flipt:9000", "import.yaml"}).
		Sync(ctx)
	if err != nil {
		return err
	}

	// run migration with edge Flipt build
	flipt, err = flipt.
		// WithEnvVariable("FLIPT_LOG_FILE", "/var/log/flipt/output.txt").
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
		WithServiceBinding("flipt", flipt.WithExec(nil)).
		WithWorkdir("build/testing/integration/readonly").
		WithEnvVariable("UNIQUE", uuid.New().String()).
		WithExec([]string{"go", "test", "-v", "-race", "--flipt-addr", "grpc://flipt:9000", "."}).
		Sync(ctx)

	return err
}
