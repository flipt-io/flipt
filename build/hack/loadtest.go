package hack

import (
	"context"

	"dagger.io/dagger"
	"github.com/google/uuid"
)

func LoadTest(ctx context.Context, client *dagger.Client, base, flipt *dagger.Container) error {
	flipt = flipt.WithEnvVariable("UNIQUE", uuid.New().String()).WithExposedPort(8080).WithExec(nil)

	const path = "/home/pyroscope/.local/share/pyroscope"

	// build the loadtest binary, and export it to the host
	_, err := base.
		WithWorkdir("build/hack").
		WithExec([]string{"go", "build", "-o", "./out/loadtest", "./cmd/loadtest/..."}).
		File("out/loadtest").
		Export(ctx, "build/hack/out/loadtest")
	if err != nil {
		return err
	}

	// run the loadtest binary from within the pyroscope container, attempting to mount the adhoc data
	// output to the host
	_, err = client.Container().
		From("pyroscope/pyroscope:latest").
		WithServiceBinding("flipt", flipt).
		WithMountedDirectory(path, client.Host().Directory("build/hack/out")).
		WithFile("loadtest", client.Host().Directory("build/hack/out").File("loadtest")).
		WithExec([]string{"adhoc", "--log-level", "debug", "--url", "flipt:8080", "./loadtest"}).
		ExitCode(ctx)

	return err
}
