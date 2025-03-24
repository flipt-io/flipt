package testing

import (
	"context"
	"encoding/json"
	"os"

	"go.flipt.io/build/internal/dagger"
)

func Unit(ctx context.Context, client *dagger.Client, flipt *dagger.Container) (*dagger.File, error) {
	// create Redis service container
	redisSrv := client.Container().
		From("redis:alpine").
		WithExposedPort(6379)

	gitea := client.Container().
		From("gitea/gitea:1.21.1").
		WithExposedPort(3000)

	flipt = flipt.
		WithServiceBinding("gitea", gitea.AsService()).
		WithExec([]string{"go", "run", "./build/internal/cmd/gitea/...", "-gitea-url", "http://gitea:3000", "-testdata-dir", "./internal/storage/fs/git/testdata"})

	out, err := flipt.Stdout(ctx)
	if err != nil {
		return nil, err
	}

	var push map[string]string
	if err := json.Unmarshal([]byte(out), &push); err != nil {
		return nil, err
	}

	if goFlags := os.Getenv("GOFLAGS"); goFlags != "" {
		flipt = flipt.WithEnvVariable("GOFLAGS", goFlags)
	}

	flipt, err = flipt.
		WithServiceBinding("redis", redisSrv.AsService()).
		WithEnvVariable("REDIS_HOST", "redis:6379").
		WithEnvVariable("TEST_GIT_REPO_URL", "http://gitea:3000/root/features.git").
		WithEnvVariable("TEST_GIT_REPO_HEAD", push["HEAD"]).
		WithEnvVariable("TEST_GIT_REPO_TAG", push["TAG"]).
		WithExec([]string{"go", "test", "-race", "-coverprofile=coverage.txt", "-covermode=atomic", "-coverpkg=./...", "./..."}).
		Sync(ctx)
	if err != nil {
		return nil, err
	}

	// attempt to export coverage if its exists
	return flipt.File("coverage.txt"), nil
}
