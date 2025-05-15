package graph

import (
	"sync"
	"testing"

	"go.flipt.io/flipt/internal/server/environments"
)

func makeID(typ, ns, key string) ResourceID {
	return ResourceID{
		Type:      environments.NewResourceType("flipt.core", typ),
		Namespace: ns,
		Key:       key,
	}
}

func TestDependencyGraph_SetAndGetDependents(t *testing.T) {
	g := NewDependencyGraph()

	segA := makeID("Segment", "ns1", "A")
	segB := makeID("Segment", "ns1", "B")
	rule1 := makeID("Rule", "ns1", "rule1")
	rule2 := makeID("Rule", "ns1", "rule2")

	// Set rule1 depends on segA
	g.SetDependencies(rule1, []ResourceID{segA})
	// Set rule2 depends on segA and segB
	g.SetDependencies(rule2, []ResourceID{segA, segB})

	// segA should have rule1 and rule2 as dependents
	depsA := g.GetDependents(segA)
	if len(depsA) != 2 {
		t.Errorf("expected 2 dependents for segA, got %d", len(depsA))
	}
	// segB should have only rule2 as dependent
	depsB := g.GetDependents(segB)
	if len(depsB) != 1 || depsB[0] != rule2 {
		t.Errorf("expected rule2 as only dependent for segB")
	}
}

func TestDependencyGraph_RemoveResource(t *testing.T) {
	g := NewDependencyGraph()
	segA := makeID("Segment", "ns1", "A")
	rule1 := makeID("Rule", "ns1", "rule1")
	g.SetDependencies(rule1, []ResourceID{segA})

	// Remove rule1, segA should have no dependents
	g.RemoveResource(rule1)
	depsA := g.GetDependents(segA)
	if len(depsA) != 0 {
		t.Errorf("expected 0 dependents for segA after rule1 removed, got %d", len(depsA))
	}

	// Remove segA, should not panic
	g.RemoveResource(segA)
}

func TestDependencyGraph_SetDependencies_Overwrite(t *testing.T) {
	g := NewDependencyGraph()
	segA := makeID("Segment", "ns1", "A")
	segB := makeID("Segment", "ns1", "B")
	rule1 := makeID("Rule", "ns1", "rule1")

	g.SetDependencies(rule1, []ResourceID{segA})
	g.SetDependencies(rule1, []ResourceID{segB})
	// should remove segA dependency

	depsA := g.GetDependents(segA)
	if len(depsA) != 0 {
		t.Errorf("expected 0 dependents for segA after overwrite, got %d", len(depsA))
	}
	depsB := g.GetDependents(segB)
	if len(depsB) != 1 || depsB[0] != rule1 {
		t.Errorf("expected rule1 as only dependent for segB after overwrite")
	}
}

func TestDependencyGraph_AddDependency(t *testing.T) {
	g := NewDependencyGraph()
	segA := makeID("Segment", "ns1", "A")
	rule1 := makeID("Rule", "ns1", "rule1")
	g.AddDependency(rule1, segA)
	depsA := g.GetDependents(segA)
	if len(depsA) != 1 || depsA[0] != rule1 {
		t.Errorf("expected rule1 as dependent for segA after AddDependency")
	}
}

func TestDependencyGraph_ThreadSafety(t *testing.T) {
	g := NewDependencyGraph()
	segA := makeID("Segment", "ns1", "A")
	rule1 := makeID("Rule", "ns1", "rule1")
	wg := sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			g.SetDependencies(rule1, []ResourceID{segA})
			_ = g.GetDependents(segA)
			g.RemoveResource(rule1)
		}(i)
	}
	wg.Wait()
}
