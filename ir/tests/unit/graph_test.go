package graph

import (
	"flowforge/ir/internal/graph"
	"flowforge/ir/pkg"
	"testing"
)

func TestDAGNodes(t *testing.T) {
	spec := &ir.PipelineSpec{
		APIVersion: "flowforge.io/v1",
		Kind:       "Pipeline",
		Metadata:   ir.PipelineMetadata{Name: "test"},
		Tasks: map[string]*ir.Task{
			"task1": {Type: ir.TaskTypeSource, Handler: ir.Handler{}},
			"task2": {Type: ir.TaskTypeTransform, Handler: ir.Handler{}},
			"task3": {Type: ir.TaskTypeSink, Handler: ir.Handler{}},
		},
	}

	g := graph.NewDAG(spec)
	nodes := g.Nodes()
	if len(nodes) != 3 {
		t.Errorf("expected 3 nodes, got %d", len(nodes))
	}
}

func TestDAGTopologicalSort(t *testing.T) {
	spec := &ir.PipelineSpec{
		APIVersion: "flowforge.io/v1",
		Kind:       "Pipeline",
		Metadata:   ir.PipelineMetadata{Name: "test"},
		Tasks: map[string]*ir.Task{
			"extract":   {Type: ir.TaskTypeSource, Handler: ir.Handler{}},
			"transform": {Type: ir.TaskTypeTransform, Handler: ir.Handler{}},
			"load":      {Type: ir.TaskTypeSink, Handler: ir.Handler{}},
		},
		Edges: []ir.Edge{
			{From: ir.TaskPort{Task: "extract", Port: "out"}, To: ir.TaskPort{Task: "transform", Port: "in"}},
			{From: ir.TaskPort{Task: "transform", Port: "out"}, To: ir.TaskPort{Task: "load", Port: "in"}},
		},
	}

	g := graph.NewDAG(spec)
	sorted, err := g.TopologicalSort()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(sorted) != 3 {
		t.Errorf("expected 3 tasks, got %d", len(sorted))
	}

	// Verify order
	if sorted[0] != "extract" || sorted[1] != "transform" || sorted[2] != "load" {
		t.Errorf("unexpected order: %v", sorted)
	}
}

func TestDAGPredecessorsSuccessors(t *testing.T) {
	spec := &ir.PipelineSpec{
		APIVersion: "flowforge.io/v1",
		Kind:       "Pipeline",
		Metadata:   ir.PipelineMetadata{Name: "test"},
		Tasks: map[string]*ir.Task{
			"task1": {Type: ir.TaskTypeSource, Handler: ir.Handler{}},
			"task2": {Type: ir.TaskTypeTransform, Handler: ir.Handler{}},
			"task3": {Type: ir.TaskTypeSink, Handler: ir.Handler{}},
		},
		Edges: []ir.Edge{
			{From: ir.TaskPort{Task: "task1", Port: "out"}, To: ir.TaskPort{Task: "task2", Port: "in"}},
			{From: ir.TaskPort{Task: "task2", Port: "out"}, To: ir.TaskPort{Task: "task3", Port: "in"}},
		},
	}

	g := graph.NewDAG(spec)

	successors := g.Successors("task1")
	if len(successors) != 1 || successors[0] != "task2" {
		t.Errorf("expected [task2], got %v", successors)
	}

	predecessors := g.Predecessors("task3")
	if len(predecessors) != 1 || predecessors[0] != "task2" {
		t.Errorf("expected [task2], got %v", predecessors)
	}
}

func TestDAGHasCycle(t *testing.T) {
	spec := &ir.PipelineSpec{
		APIVersion: "flowforge.io/v1",
		Kind:       "Pipeline",
		Metadata:   ir.PipelineMetadata{Name: "test"},
		Tasks: map[string]*ir.Task{
			"task1": {Type: ir.TaskTypeSource, Handler: ir.Handler{}},
			"task2": {Type: ir.TaskTypeTransform, Handler: ir.Handler{}},
		},
		Edges: []ir.Edge{
			{From: ir.TaskPort{Task: "task1", Port: "out"}, To: ir.TaskPort{Task: "task2", Port: "in"}},
			{From: ir.TaskPort{Task: "task2", Port: "out"}, To: ir.TaskPort{Task: "task1", Port: "in"}},
		},
	}

	g := graph.NewDAG(spec)
	if !g.HasCycle() {
		t.Error("expected cycle, but hasCycle returned false")
	}

	cycle := g.GetCycle()
	if len(cycle) == 0 {
		t.Error("expected non-empty cycle")
	}
}

func TestDAGNoCycle(t *testing.T) {
	spec := &ir.PipelineSpec{
		APIVersion: "flowforge.io/v1",
		Kind:       "Pipeline",
		Metadata:   ir.PipelineMetadata{Name: "test"},
		Tasks: map[string]*ir.Task{
			"task1": {Type: ir.TaskTypeSource, Handler: ir.Handler{}},
			"task2": {Type: ir.TaskTypeTransform, Handler: ir.Handler{}},
		},
		Edges: []ir.Edge{
			{From: ir.TaskPort{Task: "task1", Port: "out"}, To: ir.TaskPort{Task: "task2", Port: "in"}},
		},
	}

	g := graph.NewDAG(spec)
	if g.HasCycle() {
		t.Error("expected no cycle")
	}

	cycle := g.GetCycle()
	if len(cycle) != 0 {
		t.Errorf("expected empty cycle, got %v", cycle)
	}
}

func TestDAGString(t *testing.T) {
	spec := &ir.PipelineSpec{
		APIVersion: "flowforge.io/v1",
		Kind:       "Pipeline",
		Metadata:   ir.PipelineMetadata{Name: "test"},
		Tasks: map[string]*ir.Task{
			"task1": {Type: ir.TaskTypeSource, Handler: ir.Handler{}},
			"task2": {Type: ir.TaskTypeTransform, Handler: ir.Handler{}},
		},
		Edges: []ir.Edge{
			{From: ir.TaskPort{Task: "task1", Port: "out"}, To: ir.TaskPort{Task: "task2", Port: "in"}},
		},
	}

	g := graph.NewDAG(spec)
	s := g.String()
	if len(s) == 0 {
		t.Error("expected non-empty string representation")
	}
}
