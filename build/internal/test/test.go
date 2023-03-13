package test

import (
	"context"
	"time"

	"dagger.io/dagger"
)

func Test(ctx context.Context, client *dagger.Client, flipt *dagger.Container) error {
	// create Redis service container
	redisSrv := client.Container().
		From("redis").
		WithEnvVariable("CACHE_BUST", time.Now().String()).
		WithExposedPort(6379).
		WithExec(nil)

	_, err := flipt.
		WithServiceBinding("redis", redisSrv).
		WithEnvVariable("REDIS_HOST", "redis:6379").
		WithExec([]string{"go", "test", "-tags", "assets", "./..."}).
		ExitCode(ctx)
	return err
}
