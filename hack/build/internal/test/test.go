package test

import (
	"context"

	"dagger.io/dagger"
)

func Test(ctx context.Context, client *dagger.Client, flipt *dagger.Container) error {
	// create Redis service container
	redisSrv := client.Container().
		From("redis").
		WithExposedPort(6379).
		WithExec(nil)

	_, err := flipt.
		WithServiceBinding("redis", redisSrv).
		WithEnvVariable("REDIS_HOST", "redis:6379").
		WithExec([]string{"go", "test", "-race", "-p", "1", "./..."}).
		ExitCode(ctx)
	return err
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
