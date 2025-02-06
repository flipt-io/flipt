package internal

import (
	"context"

	"go.flipt.io/build/internal/dagger"
)

func UI(ctx context.Context, client *dagger.Client, source *dagger.Directory, chatwootToken *dagger.Secret) (*dagger.Container, error) {
	cache := client.CacheVolume("node-modules-cache")

	container := client.Container().From("node:18-bullseye").
		WithMountedDirectory("/src", source.
			WithoutDirectory("dist").
			WithoutDirectory("node_modules")).
		WithWorkdir("/src").
		WithMountedCache("/src/node_modules", cache)

	if chatwootToken != nil {
		plaintext, err := chatwootToken.Plaintext(ctx)
		if err != nil {
			return nil, err
		}
		if plaintext != "" {
			container = container.WithSecretVariable("REACT_APP_CHATWOOT_TOKEN", chatwootToken)
		}
	}

	container = container.WithExec([]string{"npm", "install"}).
		WithExec([]string{"npm", "run", "build"})

	return container, nil
}
