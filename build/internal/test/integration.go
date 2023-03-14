package test

import (
	"context"

	"dagger.io/dagger"
)

func Integration(ctx context.Context, client *dagger.Client, base, flipt *dagger.Container) error {
	if _, err := flipt.ExitCode(ctx); err != nil {
		return err
	}

	_, err := base.
		WithServiceBinding("flipt", flipt.WithExec(nil)).
		WithWorkdir("build/integration").
		WithExec([]string{"go", "test", "-v", "-race", "-flipt-addr", "flipt:9000", "."}).
		ExitCode(ctx)
	return err
}
