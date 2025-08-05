package internal

import (
	"context"
	"fmt"

	"go.flipt.io/build/internal/dagger"
)

func UI(ctx context.Context, client *dagger.Client, source *dagger.Directory) (*dagger.Container, error) {
	cache := client.CacheVolume("node-modules-cache")

	return client.Container().From("node:22-bullseye-slim").
		WithMountedDirectory("/src", source.
			WithoutDirectory("dist").
			WithoutDirectory("node_modules")).
		WithWorkdir("/src").
		WithMountedCache("/src/node_modules", cache).
		WithExec([]string{"npm", "install"}).
		WithExec([]string{"npm", "run", "build"}), nil
}

// UIWithCache builds the UI with registry cache support for base layer only
// Dependencies are handled via cache volumes which are more efficient than registry caching
func UIWithCache(ctx context.Context, client *dagger.Client, source *dagger.Directory, registryCache string) (*dagger.Container, error) {
	// Try cached Node.js base
	nodeBaseRef := fmt.Sprintf("%s:node-base", registryCache)
	nodeBase := client.Container().From(nodeBaseRef)
	
	if _, err := nodeBase.Sync(ctx); err != nil {
		// Build fresh Node.js base and cache it
		nodeBase = client.Container().From("node:22-bullseye-slim")
		
		// Cache this Node.js base layer
		nodeBase.Publish(ctx, nodeBaseRef)
	}

	// Use cache volume for dependencies (more reliable than registry caching for npm)
	cache := client.CacheVolume("node-modules-cache")
	
	// Mount source without dist and node_modules
	sourceClean := source.WithoutDirectory("dist").WithoutDirectory("node_modules")
	container := nodeBase.
		WithMountedDirectory("/src", sourceClean).
		WithWorkdir("/src").
		WithMountedCache("/src/node_modules", cache).
		WithExec([]string{"npm", "install"})

	// Build the UI
	return container.WithExec([]string{"npm", "run", "build"}), nil
}
