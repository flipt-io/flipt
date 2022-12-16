package test

import (
	"context"

	"dagger.io/dagger"
)

func Test(ctx context.Context, client *dagger.Client, flipt *dagger.Container) error {
	socket := client.Host().UnixSocket("/var/run/docker.sock")

	_, err := flipt.WithUnixSocket("/var/run/docker.sock", socket).
		WithExec([]string{"go", "test", "-tags", "assets", "./..."}).
		ExitCode(ctx)
	return err
}
