package airflow

import (
	"context"
	"testing"
	"time"

	"flowforge/ir"
)

// TestDAGCompilation tests DAG compilation
func TestDAGCompilation(t *testing.T) {
	spec := &ir.PipelineSpec{
		Metadata: map[string]interface{}{
			"name":    "test_dag",
			"version": "1.0.0",
		},
		Tasks: map[string]*ir.Task{
			"task1": {
				Handler: &ir.Handler{Type: "bash", Command: "echo hello"},
				Config:  &ir.Config{Image: "bash:5.1"},
			},
		},
		Edges: []*ir.Edge{},
	}

	executor := New(NewMockAirflowClient(), "/dags")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	dag, err := executor.Compile(ctx, spec, nil)
	if err != nil {
		t.Fatalf("compilation failed: %v", err)
	}

	if dag == "" {
		t.Fatal("no DAG generated")
	}

	if !contains(dag, "test_dag") {
		t.Fatal("DAG name not in output")
	}

	if !contains(dag, "task1") {
		t.Fatal("task not in output")
	}

	t.Logf("Compiled DAG: %d lines", countLines(dag))
}

// TestDAGExecution tests DAG execution
func TestDAGExecution(t *testing.T) {
	spec := &ir.PipelineSpec{
		Metadata: map[string]interface{}{
			"name": "exec_dag",
		},
		Tasks: map[string]*ir.Task{
			"task1": {
				Handler: &ir.Handler{Type: "bash", Command: "echo test"},
				Config:  &ir.Config{Image: "bash:5.1"},
			},
		},
		Edges: []*ir.Edge{},
	}

	executor := New(NewMockAirflowClient(), "/dags")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := executor.Execute(ctx, spec)
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	if result.DAGID == "" {
		t.Fatal("no DAG ID returned")
	}

	if result.Status == "" {
		t.Fatal("no status returned")
	}

	t.Logf("Executed DAG: %s", result.DAGID)
}

// TestDAGStatus tests status retrieval
func TestDAGStatus(t *testing.T) {
	executor := New(NewMockAirflowClient(), "/dags")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	status, err := executor.GetStatus(ctx, "dag-123")
	if err != nil {
		t.Fatalf("get status failed: %v", err)
	}

	if status == "" {
		t.Fatal("no status returned")
	}

	t.Logf("DAG status: %s", status)
}

// TestTaskStatus tests task status
func TestTaskStatus(t *testing.T) {
	executor := New(NewMockAirflowClient(), "/dags")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	status, err := executor.GetTaskStatus(ctx, "dag-123", "task1")
	if err != nil {
		t.Fatalf("get task status failed: %v", err)
	}

	if status == "" {
		t.Fatal("no status returned")
	}

	t.Logf("Task status: %s", status)
}

// TestTaskLogs tests log retrieval
func TestTaskLogs(t *testing.T) {
	executor := New(NewMockAirflowClient(), "/dags")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	logs, err := executor.GetLogs(ctx, "dag-123", "task1", "2024-01-01T00:00:00Z")
	if err != nil {
		t.Fatalf("get logs failed: %v", err)
	}

	if logs == "" {
		t.Fatal("no logs returned")
	}

	t.Logf("Task logs: %s", logs)
}

// TestDAGDeletion tests DAG deletion
func TestDAGDeletion(t *testing.T) {
	executor := New(NewMockAirflowClient(), "/dags")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := executor.Delete(ctx, "dag-123")
	if err != nil {
		t.Fatalf("delete failed: %v", err)
	}

	t.Logf("Deleted DAG: dag-123")
}

// TestComplexDAG tests complex DAG with dependencies
func TestComplexDAG(t *testing.T) {
	spec := &ir.PipelineSpec{
		Metadata: map[string]interface{}{
			"name":    "complex_dag",
			"version": "1.0.0",
		},
		Tasks: map[string]*ir.Task{
			"extract": {
				Handler: &ir.Handler{Type: "bash", Command: "extract.sh"},
				Config:  &ir.Config{Image: "python:3.11"},
			},
			"transform": {
				Handler: &ir.Handler{Type: "python", Command: "transform.py"},
				Config:  &ir.Config{Image: "python:3.11"},
			},
			"load": {
				Handler: &ir.Handler{Type: "bash", Command: "load.sh"},
				Config:  &ir.Config{Image: "python:3.11"},
			},
		},
		Edges: []*ir.Edge{
			{From: "extract", To: "transform"},
			{From: "transform", To: "load"},
		},
	}

	executor := New(NewMockAirflowClient(), "/dags")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := executor.Execute(ctx, spec)
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	if result.DAGID == "" {
		t.Fatal("no DAG ID")
	}

	t.Logf("Executed complex ETL DAG: %s", result.DAGID)
}

// TestInvalidSpec tests error handling
func TestInvalidSpec(t *testing.T) {
	executor := New(NewMockAirflowClient(), "/dags")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Test nil spec
	_, err := executor.Execute(ctx, nil)
	if err == nil {
		t.Fatal("expected error for nil spec")
	}

	// Test missing name
	spec := &ir.PipelineSpec{
		Metadata: map[string]interface{}{},
		Tasks:    map[string]*ir.Task{},
	}

	_, err = executor.Execute(ctx, spec)
	if err == nil {
		t.Fatal("expected error for missing name")
	}

	t.Logf("Error handling working correctly")
}

// Helper functions
func contains(s, substr string) bool {
	for i := 0; i < len(s)-len(substr)+1; i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func countLines(s string) int {
	count := 0
	for _, c := range s {
		if c == '\n' {
			count++
		}
	}
	return count
}
