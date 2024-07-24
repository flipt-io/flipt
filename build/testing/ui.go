package testing

import (
	"context"
	"os"
	"time"

	"go.flipt.io/build/internal/dagger"
)

func UI(ctx context.Context, client *dagger.Client, ui, flipt *dagger.Container) error {
	test, err := buildUI(ctx, ui, flipt)
	if err != nil {
		return err
	}

	_, err = test.
		WithExec([]string{"npx", "playwright", "install", "chromium", "--with-deps"}).
		WithExec([]string{"npx", "playwright", "test"}).
		Sync(ctx)
	if err != nil {
		return err
	}

	_, _ = test.Directory("playwright-report").
		Export(ctx, "playwright-report")

	return nil
}

func buildUI(ctx context.Context, ui, flipt *dagger.Container) (_ *dagger.Container, err error) {
	flipt, err = flipt.Sync(ctx)
	if err != nil {
		return nil, err
	}

	ui, err = ui.Sync(ctx)
	if err != nil {
		return nil, err
	}

	return ui.
		WithServiceBinding("flipt", flipt.
			WithEnvVariable("CI", os.Getenv("CI")).
			WithEnvVariable("FLIPT_LOG_LEVEL", "WARN").
			WithEnvVariable("FLIPT_AUTHENTICATION_METHODS_TOKEN_ENABLED", "true").
			WithEnvVariable("UNIQUE", time.Now().String()).
			WithExec(nil).
			AsService()).
		WithFile("/usr/bin/flipt", flipt.File("/flipt")).
		WithEnvVariable("FLIPT_ADDRESS", "http://flipt:8080"), nil
}
