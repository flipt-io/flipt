package testing

import (
	"context"
	"time"

	"dagger.io/dagger"
)

func UI(ctx context.Context, client *dagger.Client, ui, flipt *dagger.Container) error {
	test := ui.
		WithExec([]string{"npx", "playwright", "install", "chromium", "--with-deps"}).
		WithServiceBinding("flipt", flipt.
			WithEnvVariable("FLIPT_AUTHENTICATION_METHODS_TOKEN_ENABLED", "true").
			WithEnvVariable("UNIQUE", time.Now().String()).
			WithExec(nil)).
		WithEnvVariable("FLIPT_ADDRESS", "http://flipt:8080").
		WithExec([]string{"npx", "playwright", "test"})
	_, err := test.ExitCode(ctx)
	if err != nil {
		return err
	}

	if _, err := test.
		Directory("playwright-report").
		Export(ctx, "playwright-report"); err != nil {
		return err
	}

	return nil
}
