package testing

import (
	"context"
	"encoding/json"
	"os"

	"dagger.io/dagger"
)

func Unit(ctx context.Context, client *dagger.Client, flipt *dagger.Container) error {
	// create Redis service container
	redisSrv := client.Container().
		From("redis:alpine").
		WithExposedPort(6379).
		WithExec(nil)

	gitea := client.Container().
		From("gitea/gitea:1.21.1").
		WithExposedPort(3000).
		WithExec(nil)

	flipt = flipt.
		WithServiceBinding("gitea", gitea.AsService()).
		WithExec([]string{"go", "run", "./build/internal/cmd/gitea/...", "-gitea-url", "http://gitea:3000", "-testdata-dir", "./internal/storage/fs/git/testdata"})

	out, err := flipt.Stdout(ctx)
	if err != nil {
		return err
	}

	var push map[string]string
	if err := json.Unmarshal([]byte(out), &push); err != nil {
		return err
	}

	minio := client.Container().
		From("quay.io/minio/minio:latest").
		WithExposedPort(9009).
		WithEnvVariable("MINIO_ROOT_USER", "user").
		WithEnvVariable("MINIO_ROOT_PASSWORD", "password").
		WithEnvVariable("MINIO_BROWSER", "off").
		WithExec([]string{"server", "/data", "--address", ":9009", "--quiet"})

	azurite := client.Container().
		From("mcr.microsoft.com/azure-storage/azurite").
		WithExposedPort(10000).
		WithExec([]string{"azurite-blob", "--blobHost", "0.0.0.0", "--silent"}).
		AsService()

	gcs := client.Container().
		From("fsouza/fake-gcs-server").
		WithExposedPort(4443).
		WithExec([]string{"-scheme", "http", "-public-host", "gcs:4443"}).
		AsService()

	// S3 unit testing

	flipt = flipt.
		WithServiceBinding("minio", minio.AsService()).
		WithEnvVariable("TEST_S3_ENDPOINT", "http://minio:9009").
		WithEnvVariable("AWS_ACCESS_KEY_ID", "user").
		WithEnvVariable("AWS_SECRET_ACCESS_KEY", "password")

	// GCS unit testing

	flipt = flipt.
		WithServiceBinding("gcs", gcs).
		WithEnvVariable("STORAGE_EMULATOR_HOST", "gcs:4443")

	// Azure unit testing

	flipt = flipt.
		WithServiceBinding("azurite", azurite).
		WithEnvVariable("TEST_AZURE_ENDPOINT", "http://azurite:10000/devstoreaccount1").
		WithEnvVariable("AZURE_STORAGE_ACCOUNT", "devstoreaccount1").
		WithEnvVariable("AZURE_STORAGE_KEY", "Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw==")

	if goFlags := os.Getenv("GOFLAGS"); goFlags != "" {
		flipt = flipt.WithEnvVariable("GOFLAGS", goFlags)
	}

	flipt, err = flipt.
		WithServiceBinding("redis", redisSrv.AsService()).
		WithEnvVariable("REDIS_HOST", "redis:6379").
		WithEnvVariable("TEST_GIT_REPO_URL", "http://gitea:3000/root/features.git").
		WithEnvVariable("TEST_GIT_REPO_HEAD", push["HEAD"]).
		WithEnvVariable("TEST_GIT_REPO_TAG", push["TAG"]).
		WithExec([]string{"go", "test", "-race", "-coverprofile=coverage.txt", "-covermode=atomic", "./..."}).
		Sync(ctx)
	if err != nil {
		return err
	}

	// attempt to export coverage if its exists
	_, _ = flipt.File("coverage.txt").Export(ctx, "coverage.txt")

	return nil
}
