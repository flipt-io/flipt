package graph

import (
	"fmt"
	"sync"

	"go.flipt.io/flipt/internal/server/environments"
)

// ResourceGraph is an in-memory dependency graph for any environments.Resource node type.
//
// The graph is thread-safe and can be used concurrently by multiple goroutines.
type ResourceGraph struct {
	// Map from a resource to the set of resources that depend on it
	dependents map[string]map[string]struct{}
	// Map from a resource to the set of resources it depends on
	dependencies map[string]map[string]struct{}
	mu           sync.RWMutex
}

// NewResourceGraph creates a new, empty dependency graph for environments.Resource.
func NewResourceGraph() *ResourceGraph {
	return &ResourceGraph{
		dependents:   make(map[string]map[string]struct{}),
		dependencies: make(map[string]map[string]struct{}),
	}
}

// SetDependencies sets the dependencies for a resource, updating the graph accordingly.
// It removes all previous dependencies for the resource and sets the new ones.
func (g *ResourceGraph) SetDependencies(resource environments.TypedResource, dependencies []environments.TypedResource) {
	g.mu.Lock()
	defer g.mu.Unlock()

	resourceKey := fmt.Sprintf("%s/%s/%s", resource.ResourceType.String(), resource.NamespaceKey, resource.Key)

	// Remove old dependencies for this resource
	for dep := range g.dependencies[resourceKey] {
		delete(g.dependents[dep], resourceKey)
		if len(g.dependents[dep]) == 0 {
			delete(g.dependents, dep)
		}
	}
	delete(g.dependencies, resourceKey)

	// Set new dependencies
	g.dependencies[resourceKey] = make(map[string]struct{})
	for _, dep := range dependencies {
		depKey := fmt.Sprintf("%s/%s/%s", dep.ResourceType.String(), dep.NamespaceKey, dep.Key)
		g.dependencies[resourceKey][depKey] = struct{}{}
		if g.dependents[depKey] == nil {
			g.dependents[depKey] = make(map[string]struct{})
		}
		g.dependents[depKey][resourceKey] = struct{}{}
	}
}

// AddDependency adds a dependency between two resources.
func (g *ResourceGraph) AddDependency(resource environments.TypedResource, dependency environments.TypedResource) {
	g.mu.Lock()
	defer g.mu.Unlock()

	resourceKey := fmt.Sprintf("%s/%s/%s", resource.ResourceType.String(), resource.NamespaceKey, resource.Key)
	dependencyKey := fmt.Sprintf("%s/%s/%s", dependency.ResourceType.String(), dependency.NamespaceKey, dependency.Key)

	if g.dependencies[resourceKey] == nil {
		g.dependencies[resourceKey] = make(map[string]struct{})
	}
	if g.dependents[dependencyKey] == nil {
		g.dependents[dependencyKey] = make(map[string]struct{})
	}
	g.dependencies[resourceKey][dependencyKey] = struct{}{}
	g.dependents[dependencyKey][resourceKey] = struct{}{}
}

// RemoveResource removes a resource and all its dependency links from the graph.
func (g *ResourceGraph) RemoveResource(resource environments.TypedResource) {
	g.mu.Lock()
	defer g.mu.Unlock()

	resourceKey := fmt.Sprintf("%s/%s/%s", resource.ResourceType.String(), resource.NamespaceKey, resource.Key)

	for dep := range g.dependencies[resourceKey] {
		delete(g.dependents[dep], resourceKey)
		if len(g.dependents[dep]) == 0 {
			delete(g.dependents, dep)
		}
	}
	delete(g.dependencies, resourceKey)
	for dep := range g.dependents[resourceKey] {
		delete(g.dependencies[dep], resourceKey)
		if len(g.dependencies[dep]) == 0 {
			delete(g.dependencies, dep)
		}
	}
	delete(g.dependents, resourceKey)
}

// GetDependents returns all resources that depend on the given resource.
func (g *ResourceGraph) GetDependents(resource environments.TypedResource) []string {
	g.mu.RLock()
	defer g.mu.RUnlock()

	resourceKey := fmt.Sprintf("%s/%s/%s", resource.ResourceType.String(), resource.NamespaceKey, resource.Key)

	var out []string
	for dep := range g.dependents[resourceKey] {
		out = append(out, dep)
	}
	return out
}
