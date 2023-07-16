package hack

import (
	"context"

	"dagger.io/dagger"
	"github.com/google/uuid"
)

func LoadTest(ctx context.Context, client *dagger.Client, base, flipt *dagger.Container) error {
	seed := base.File("build/testing/integration/readonly/testdata/default.yaml")
	importCmd := []string{"/flipt", "import", "--create-namespace", "import.yaml"}

	// import some test data
	flipt, err := flipt.WithEnvVariable("UNIQUE", uuid.New().String()).
		WithFile("import.yaml", seed).
		WithExec(importCmd).
		Sync(ctx)

	if err != nil {
		return err
	}

	flipt = flipt.WithEnvVariable("UNIQUE", uuid.New().String()).
		WithExposedPort(8080).
		WithExec(nil)

	// build the loadtest binary
	loadtest := base.
		WithWorkdir("build/hack").
		WithExec([]string{"go", "build", "-o", "./out/loadtest", "./cmd/loadtest/..."}).
		File("out/loadtest")

	// run the loadtest binary from within the pyroscope container and export the adhoc data
	// output to the host
	_, err = client.Container().
		From("pyroscope/pyroscope:latest").
		WithFile("loadtest", loadtest).
		WithServiceBinding("flipt", flipt).
		WithFile("loadtest", client.Host().Directory("build/hack/out").File("loadtest")).
		WithExec([]string{"adhoc", "--log-level", "debug", "--url", "flipt:8080", "./loadtest", "-duration", "60s"}).
		Directory("/home/pyroscope/.local/share/pyroscope").
		Export(ctx, "build/hack/out/profiles")

	return err
}
