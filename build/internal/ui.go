package internal

import (
	"context"
	"fmt"

	"go.flipt.io/build/internal/dagger"
)

func UI(ctx context.Context, client *dagger.Client, source *dagger.Directory, registryCache ...string) (*dagger.Container, error) {
	var nodeBase *dagger.Container

	// Use cache if available
	if len(registryCache) > 0 && registryCache[0] != "" {
		nodeBaseRef := fmt.Sprintf("%s:node-base", registryCache[0])
		nodeBase = client.Container().From(nodeBaseRef)
		if _, err := nodeBase.Sync(ctx); err != nil {
			// Build fresh Node.js base and cache it
			nodeBase = client.Container().From("node:22-bullseye-slim")
			// Cache this Node.js base layer
			nodeBase.Publish(ctx, nodeBaseRef)
		}
	} else {
		// No cache - use regular build
		nodeBase = client.Container().From("node:22-bullseye-slim")
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
