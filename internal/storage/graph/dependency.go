package graph

import (
	"fmt"
	"sync"

	"go.flipt.io/flipt/internal/server/environments"
)

type Dependency struct {
	Resource   ResourceID
	Dependents []ResourceID
}

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
func (g *DependencyGraph) SetDependencies(deps []Dependency) {
	g.mu.Lock()
	defer g.mu.Unlock()

	for _, dep := range deps {
		// Remove old dependencies
		for d := range g.dependencies[dep.Resource] {
			delete(g.dependents[d], dep.Resource)
			if len(g.dependents[d]) == 0 {
				delete(g.dependents, d)
			}
		}
	}
	// Set new dependencies
	for _, dep := range deps {
		g.dependencies[dep.Resource] = make(map[ResourceID]struct{})
		for _, d := range dep.Dependents {
			g.dependencies[dep.Resource][d] = struct{}{}
			if g.dependents[d] == nil {
				g.dependents[d] = make(map[ResourceID]struct{})
			}
			g.dependents[d][dep.Resource] = struct{}{}
		}
	}
}

// AddDependency adds a dependency between two resources.
func (g *DependencyGraph) AddDependency(dep Dependency) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.dependencies[dep.Resource] == nil {
		g.dependencies[dep.Resource] = make(map[ResourceID]struct{})
	}

	for _, d := range dep.Dependents {
		if g.dependents[d] == nil {
			g.dependents[d] = make(map[ResourceID]struct{})
		}
		g.dependencies[dep.Resource][d] = struct{}{}
		g.dependents[d][dep.Resource] = struct{}{}
	}
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
