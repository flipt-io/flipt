package test

import (
	"context"

	"dagger.io/dagger"
	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"
)

func Integration(ctx context.Context, client *dagger.Client, base, flipt *dagger.Container) error {
	if _, err := flipt.ExitCode(ctx); err != nil {
		return err
	}

	integration := func(flipt *dagger.Container, flags ...string) func() error {
		return func() error {
			flags := append([]string{"-v", "-race"}, flags...)
			_, err := base.
				WithServiceBinding("flipt", flipt.
					// this ensures a unique instance of flipt is created per
					// integration test
					WithEnvVariable("UNIQUE", uuid.New().String()).
					WithEnvVariable("FLIPT_LOG_LEVEL", "debug").
					WithExec(nil)).
				WithWorkdir("hack/build/integration").
				WithExec(append([]string{"go", "test"}, append(flags, ".")...)).
				ExitCode(ctx)

			return err
		}
	}

	var g errgroup.Group

	for _, addr := range []string{
		"grpc://flipt:9000",
		"http://flipt:8080",
	} {
		addr := addr
		g.Go(integration(flipt, "-flipt-addr", addr))

		token := "abcdefghij"
		g.Go(integration(
			flipt.
				WithEnvVariable("FLIPT_AUTHENTICATION_REQUIRED", "true").
				WithEnvVariable("FLIPT_AUTHENTICATION_METHODS_TOKEN_ENABLED", "true").
				WithEnvVariable("FLIPT_AUTHENTICATION_METHODS_TOKEN_BOOTSTRAP_TOKEN", token),
			"-flipt-addr", addr,
			"-flipt-token", token,
		))
	}

	return g.Wait()
}
