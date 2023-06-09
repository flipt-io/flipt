package testing

import (
	"context"
	"encoding/json"

	"dagger.io/dagger"
)

func Unit(ctx context.Context, client *dagger.Client, flipt *dagger.Container) error {
	// create Redis service container
	redisSrv := client.Container().
		From("redis").
		WithExposedPort(6379).
		WithExec(nil)

	gitea := client.Container().
		From("gitea/gitea:latest").
		WithExposedPort(3000).
		WithExec(nil)

	flipt = flipt.
		WithServiceBinding("gitea", gitea).
		WithExec([]string{"go", "run", "./build/internal/cmd/gitea/...", "-gitea-url", "http://gitea:3000", "-testdata-dir", "./internal/storage/fs/git/testdata"})

	out, err := flipt.Stdout(ctx)
	if err != nil {
		return err
	}

	var push map[string]string
	if err := json.Unmarshal([]byte(out), &push); err != nil {
		return err
	}

	flipt, err = flipt.
		WithServiceBinding("redis", redisSrv).
		WithEnvVariable("REDIS_HOST", "redis:6379").
		WithEnvVariable("TEST_GIT_REPO_URL", "http://gitea:3000/root/features.git").
		WithEnvVariable("TEST_GIT_REPO_HEAD", push["HEAD"]).
		WithExec([]string{"go", "test", "-race", "-p", "1", "-coverprofile=coverage.txt", "-covermode=atomic", "./..."}).
		Sync(ctx)
	if err != nil {
		return err
	}

	// attempt to export coverage if its exists
	_, _ = flipt.File("coverage.txt").Export(ctx, "coverage.txt")

	return nil
}

var All = map[string]Wrapper{
	"sqlite":    WithSQLite,
	"postgres":  WithPostgres,
	"mysql":     WithMySQL,
	"cockroach": WithCockroach,
}

type Wrapper func(context.Context, *dagger.Client, *dagger.Container) (context.Context, *dagger.Client, *dagger.Container)

func WithSQLite(ctx context.Context, client *dagger.Client, container *dagger.Container) (context.Context, *dagger.Client, *dagger.Container) {
	return ctx, client, container
}

func WithPostgres(ctx context.Context, client *dagger.Client, flipt *dagger.Container) (context.Context, *dagger.Client, *dagger.Container) {
	return ctx, client, flipt.
		WithEnvVariable("FLIPT_TEST_DB_URL", "postgres://postgres:password@postgres:5432").
		WithServiceBinding("postgres", client.Container().
			From("postgres").
			WithEnvVariable("POSTGRES_PASSWORD", "password").
			WithExposedPort(5432).
			WithExec(nil))
}

func WithMySQL(ctx context.Context, client *dagger.Client, flipt *dagger.Container) (context.Context, *dagger.Client, *dagger.Container) {
	return ctx, client, flipt.
		WithEnvVariable(
			"FLIPT_TEST_DB_URL",
			"mysql://flipt:password@mysql:3306/flipt_test?multiStatements=true",
		).
		WithServiceBinding("mysql", client.Container().
			From("mysql:8").
			WithEnvVariable("MYSQL_USER", "flipt").
			WithEnvVariable("MYSQL_PASSWORD", "password").
			WithEnvVariable("MYSQL_DATABASE", "flipt_test").
			WithEnvVariable("MYSQL_ALLOW_EMPTY_PASSWORD", "true").
			WithExposedPort(3306).
			WithExec(nil))
}

func WithCockroach(ctx context.Context, client *dagger.Client, flipt *dagger.Container) (context.Context, *dagger.Client, *dagger.Container) {
	return ctx, client, flipt.
		WithEnvVariable("FLIPT_TEST_DB_URL", "cockroachdb://root@cockroach:26257/defaultdb?sslmode=disable").
		WithServiceBinding("cockroach", client.Container().
			From("cockroachdb/cockroach:latest-v21.2").
			WithEnvVariable("COCKROACH_USER", "root").
			WithEnvVariable("COCKROACH_DATABASE", "defaultdb").
			WithExposedPort(26257).
			WithExec([]string{"start-single-node", "--insecure"}))
}
