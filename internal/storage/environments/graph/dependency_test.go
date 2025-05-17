package graph

import (
	"sync"
	"testing"

	"go.flipt.io/flipt/internal/server/environments"
	rpcenvironments "go.flipt.io/flipt/rpc/v2/environments"
)

type resource struct {
	NamespaceKey string
	Key          string
}

func (r resource) GetNamespaceKey() string {
	return r.NamespaceKey
}

func (r resource) GetKey() string {
	return r.Key
}

func makeResource(typ environments.ResourceType, ns, key string) environments.TypedResource {
	return environments.TypedResource{
		ResourceType: typ,
		Resource: &rpcenvironments.Resource{
			NamespaceKey: ns,
			Key:          key,
		},
	}
}

func TestDependencyGraph_SetAndGetDependents(t *testing.T) {
	g := NewResourceGraph()

	segA := makeResource(environments.SegmentResourceType, "ns1", "A")
	segB := makeResource(environments.SegmentResourceType, "ns1", "B")
	flag1 := makeResource(environments.FlagResourceType, "ns1", "flag1")
	flag2 := makeResource(environments.FlagResourceType, "ns1", "flag2")

	// Set flag1 depends on segA
	g.SetDependencies(flag1, []environments.TypedResource{segA})
	// Set flag2 depends on segA and segB
	g.SetDependencies(flag2, []environments.TypedResource{segA, segB})

	// segA should have flag1 and flag2 as dependents
	depsA := g.GetDependents(segA)
	if len(depsA) != 2 {
		t.Errorf("expected 2 dependents for segA, got %d", len(depsA))
	}
	// segB should have only flag2 as dependent
	depsB := g.GetDependents(segB)
	if len(depsB) != 1 || depsB[0] != "flipt.core.Flag/ns1/flag2" {
		t.Errorf("expected flag2 as only dependent for segB")
	}
}

func TestDependencyGraph_RemoveResource(t *testing.T) {
	g := NewResourceGraph()
	segA := makeResource(environments.SegmentResourceType, "ns1", "A")
	flag1 := makeResource(environments.FlagResourceType, "ns1", "flag1")
	g.SetDependencies(flag1, []environments.TypedResource{segA})

	// Remove flag1, segA should have no dependents
	g.RemoveResource(flag1)
	depsA := g.GetDependents(segA)
	if len(depsA) != 0 {
		t.Errorf("expected 0 dependents for segA after flag1 removed, got %d", len(depsA))
	}

	// Remove segA, should not panic
	g.RemoveResource(segA)
}

func TestDependencyGraph_SetDependencies_Overwrite(t *testing.T) {
	g := NewResourceGraph()
	segA := makeResource(environments.SegmentResourceType, "ns1", "A")
	segB := makeResource(environments.SegmentResourceType, "ns1", "B")
	flag1 := makeResource(environments.FlagResourceType, "ns1", "flag1")

	g.SetDependencies(flag1, []environments.TypedResource{segA})
	g.SetDependencies(flag1, []environments.TypedResource{segB})
	// should remove segA dependency

	depsA := g.GetDependents(segA)
	if len(depsA) != 0 {
		t.Errorf("expected 0 dependents for segA after overwrite, got %d", len(depsA))
	}
	depsB := g.GetDependents(segB)
	if len(depsB) != 1 || depsB[0] != "flipt.core.Flag/ns1/flag1" {
		t.Errorf("expected flag1 as only dependent for segB after overwrite")
	}
}

func TestDependencyGraph_AddDependency(t *testing.T) {
	g := NewResourceGraph()
	segA := makeResource(environments.SegmentResourceType, "ns1", "A")
	flag1 := makeResource(environments.FlagResourceType, "ns1", "flag1")
	g.AddDependency(flag1, segA)
	depsA := g.GetDependents(segA)
	if len(depsA) != 1 || depsA[0] != "flipt.core.Flag/ns1/flag1" {
		t.Errorf("expected flag1 as dependent for segA after AddDependency")
	}
}

func TestDependencyGraph_ThreadSafety(t *testing.T) {
	g := NewResourceGraph()
	segA := makeResource(environments.SegmentResourceType, "ns1", "A")
	flag1 := makeResource(environments.FlagResourceType, "ns1", "flag1")
	wg := sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			g.SetDependencies(flag1, []environments.TypedResource{segA})
			_ = g.GetDependents(segA)
			g.RemoveResource(flag1)
		}(i)
	}
	wg.Wait()
}
