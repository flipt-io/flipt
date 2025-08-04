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

// UIWithCache builds the UI with registry cache support for layers
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

	// Try cached dependencies layer - hash based on package files
	depsRef := fmt.Sprintf("%s:ui-deps", registryCache)
	
	// Mount source without dist and node_modules
	sourceClean := source.WithoutDirectory("dist").WithoutDirectory("node_modules") 
	container := nodeBase.
		WithMountedDirectory("/src", sourceClean).
		WithWorkdir("/src")

	// Always need the cache volume for node_modules access
	cache := client.CacheVolume("node-modules-cache")
	
	// Check if we have cached UI dependencies
	cachedDeps := client.Container().From(depsRef)
	if _, err := cachedDeps.Sync(ctx); err == nil {
		container = cachedDeps.
			WithMountedDirectory("/src", sourceClean).
			WithMountedCache("/src/node_modules", cache)
	} else {
		// Install dependencies and cache the layer
		container = container.
			WithMountedCache("/src/node_modules", cache).
			WithExec([]string{"npm", "install"})
		
		// Cache this dependencies layer  
		container.Publish(ctx, depsRef)
	}

	// Build the UI
	return container.WithExec([]string{"npm", "run", "build"}), nil
}
