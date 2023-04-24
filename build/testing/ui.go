package testing

import (
	"context"
	"time"

	"dagger.io/dagger"
)

func UI(ctx context.Context, client *dagger.Client, ui, flipt *dagger.Container) error {
	_, err := ui.
		WithExec([]string{"npx", "playwright", "install", "chromium", "--with-deps"}).
		WithServiceBinding("flipt", flipt.
			WithEnvVariable("FLIPT_AUTHENTICATION_METHODS_TOKEN_ENABLED", "true").
			WithEnvVariable("UNIQUE", time.Now().String()).
			WithExec(nil)).
		WithEnvVariable("FLIPT_ADDRESS", "http://flipt:8080").
		WithExec([]string{"npx", "playwright", "test"}).
		ExitCode(ctx)
	return err
}
