package test

import (
	"context"
	"fmt"
	"log"

	"dagger.io/dagger"
	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"
)

func Integration(ctx context.Context, client *dagger.Client, base, flipt *dagger.Container) error {
	fmt.Println("Starting integration tests...")

	logs := client.CacheVolume(fmt.Sprintf("logs-%s", uuid.New()))
	_, err := flipt.WithUser("root").
		WithMountedCache("/logs", logs).
		WithExec([]string{"chown", "flipt:flipt", "/logs"}).
		ExitCode(ctx)
	if err != nil {
		return err
	}

	integration := func(name string, flipt *dagger.Container, flags ...string) func() error {
		logFile := fmt.Sprintf("/var/opt/flipt/logs/%s.log", name)

		flipt = flipt.
			// this ensures a unique instance of flipt is created per
			// integration test
			WithEnvVariable("UNIQUE", uuid.New().String()).
			WithEnvVariable("FLIPT_LOG_LEVEL", "debug").
			WithMountedCache("/var/opt/flipt/logs", logs).
			WithEnvVariable("FLIPT_LOG_FILE", logFile).
			WithExec(nil)

		return func() error {
			flags := append([]string{"-v", "-race"}, flags...)
			_, err := base.
				WithServiceBinding("flipt", flipt).
				WithWorkdir("hack/build/integration").
				WithExec(append([]string{"go", "test"}, append(flags, ".")...)).
				ExitCode(ctx)

			return err
		}
	}

	var g errgroup.Group

	for _, target := range []struct {
		protocol string
		address  string
	}{
		{"grpc", "flipt:9000"},
		{"http", "flipt:8080"},
	} {
		var (
			target = target
			addr   = fmt.Sprintf("%s://%s", target.protocol, target.address)
			name   = fmt.Sprintf("flipt-%s", target.protocol)
		)

		g.Go(integration(name, flipt, "-flipt-addr", addr))

		token := "abcdefghij"
		g.Go(integration(
			name+"-with-authentication",
			flipt.
				WithEnvVariable("FLIPT_AUTHENTICATION_REQUIRED", "true").
				WithEnvVariable("FLIPT_AUTHENTICATION_METHODS_TOKEN_ENABLED", "true").
				WithEnvVariable("FLIPT_AUTHENTICATION_METHODS_TOKEN_BOOTSTRAP_TOKEN", token),
			"-flipt-addr", addr,
			"-flipt-token", token,
		))
	}

	err = g.Wait()

	if _, lerr := flipt.
		WithMountedCache("/var/opt/flipt/logs", logs).
		WithExec([]string{"cp", "-r", "/var/opt/flipt/logs", "/var/opt/flipt/out"}).
		Directory("/var/opt/flipt/out").Export(ctx, "hack/build/logs"); lerr != nil {
		log.Println("Erroring copying logs", lerr)
	}

	return err
}
