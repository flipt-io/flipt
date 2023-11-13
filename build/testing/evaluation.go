package testing

import (
	"context"

	"dagger.io/dagger"
	"github.com/google/uuid"
)

func Evaluation(ctx context.Context, client *dagger.Client, base, flipt *dagger.Container) error {
	seed := base.File("build/testing/integration/evaluation/testdata/evaluations.yaml")
	importCmd := []string{"/flipt", "import", "import.yaml"}

	// import some test data
	flipt, err := flipt.WithEnvVariable("UNIQUE", uuid.New().String()).
		WithFile("import.yaml", seed).
		WithExec(importCmd).
		Sync(ctx)
	if err != nil {
		return err
	}

	flipt = flipt.WithEnvVariable("UNIQUE", uuid.New().String()).WithExposedPort(8080)

	flipt = flipt.WithExec(nil)

	_, err = base.WithWorkdir("build/testing/integration/evaluation").
		WithEnvVariable("UNIQUE", uuid.New().String()).
		WithServiceBinding("flipt", flipt).
		WithExec([]string{"sh", "-c", "go test -v -timeout=1m -race --flipt-addr=http://flipt:8080 ."}).Sync(ctx)

	return err
}
