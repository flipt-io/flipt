package graph

import (
	"fmt"
	"sync"

	"go.flipt.io/flipt/internal/server/environments"
)

// ResourceID uniquely identifies a resource in the system.
type ResourceID struct {
	Type      environments.ResourceType
	Namespace string
	Key       string
}

func (r ResourceID) String() string {
	return fmt.Sprintf("%s:%s:%s", r.Type, r.Namespace, r.Key)
}

// DependencyGraph tracks dependencies between resources in memory.
type DependencyGraph struct {
	// Map from a resource to the set of resources that depend on it
	dependents map[ResourceID]map[ResourceID]struct{}
	// Map from a resource to the set of resources it depends on
	dependencies map[ResourceID]map[ResourceID]struct{}
	mu           sync.RWMutex
}

// NewDependencyGraph creates a new, empty dependency graph.
func NewDependencyGraph() *DependencyGraph {
	return &DependencyGraph{
		dependents:   make(map[ResourceID]map[ResourceID]struct{}),
		dependencies: make(map[ResourceID]map[ResourceID]struct{}),
	}
}

// SetDependencies sets the dependencies for a resource, updating the graph accordingly.
// Note: It does not remove any existing dependencies for other resources
func (g *DependencyGraph) SetDependencies(resource ResourceID, dependencies []ResourceID) {
	g.mu.Lock()
	defer g.mu.Unlock()

	// Remove old dependencies for this resource
	for dep := range g.dependencies[resource] {
		delete(g.dependents[dep], resource)
		if len(g.dependents[dep]) == 0 {
			delete(g.dependents, dep)
		}
	}
	delete(g.dependencies, resource)

	// Set new dependencies
	g.dependencies[resource] = make(map[ResourceID]struct{})
	for _, dep := range dependencies {
		g.dependencies[resource][dep] = struct{}{}
		if g.dependents[dep] == nil {
			g.dependents[dep] = make(map[ResourceID]struct{})
		}
		g.dependents[dep][resource] = struct{}{}
	}
}

// AddDependency adds a dependency between two resources.
func (g *DependencyGraph) AddDependency(resource ResourceID, dependency ResourceID) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.dependencies[resource] == nil {
		g.dependencies[resource] = make(map[ResourceID]struct{})
	}

	if g.dependents[dependency] == nil {
		g.dependents[dependency] = make(map[ResourceID]struct{})
	}

	g.dependencies[resource][dependency] = struct{}{}
	g.dependents[dependency][resource] = struct{}{}
}

// RemoveResource removes a resource and all its dependency links from the graph.
func (g *DependencyGraph) RemoveResource(res ResourceID) {
	g.mu.Lock()
	defer g.mu.Unlock()
	// Remove from dependents
	for dep := range g.dependencies[res] {
		delete(g.dependents[dep], res)
		if len(g.dependents[dep]) == 0 {
			delete(g.dependents, dep)
		}
	}
	delete(g.dependencies, res)
	// Remove from dependencies
	for dep := range g.dependents[res] {
		delete(g.dependencies[dep], res)
		if len(g.dependencies[dep]) == 0 {
			delete(g.dependencies, dep)
		}
	}
	delete(g.dependents, res)
}

// GetDependents returns all resources that depend on the given resource.
func (g *DependencyGraph) GetDependents(res ResourceID) []ResourceID {
	g.mu.RLock()
	defer g.mu.RUnlock()
	var out []ResourceID
	for dep := range g.dependents[res] {
		out = append(out, dep)
	}
	return out
}
