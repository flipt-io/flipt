package test

import (
	"context"

	"dagger.io/dagger"
	"golang.org/x/sync/errgroup"
)

func Integration(ctx context.Context, client *dagger.Client, base, flipt *dagger.Container) error {
	if _, err := flipt.ExitCode(ctx); err != nil {
		return err
	}

	integration := func(flipt *dagger.Container, flags ...string) func() error {
		return func() error {
			flags := append([]string{"-v", "-race", "-flipt-addr", "flipt:9000"}, flags...)
			_, err := base.
				WithServiceBinding("flipt", flipt.
					WithEnvVariable("FLIPT_LOG_LEVEL", "debug").
					WithExec(nil)).
				WithWorkdir("build/integration").
				WithExec(append([]string{"go", "test"}, append(flags, ".")...)).
				ExitCode(ctx)

			return err
		}
	}

	var g errgroup.Group

	g.Go(integration(flipt))

	token := "abcdefghij"

	g.Go(integration(
		flipt.
			WithEnvVariable("FLIPT_AUTHENTICATION_REQUIRED", "true").
			WithEnvVariable("FLIPT_AUTHENTICATION_METHODS_TOKEN_BOOTSTRAP_TOKEN", token),
		"-flipt-token", token,
	))

	return g.Wait()
}
