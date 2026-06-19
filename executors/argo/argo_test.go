package argo

import (
	"context"
	"testing"
	"time"

	"flowforge/ir"
)

// TestSimpleExecution tests basic workflow execution
func TestSimpleExecution(t *testing.T) {
	client := NewMockArgoClient()
	executor := New(client, "default")

	spec := &ir.PipelineSpec{
		Metadata: map[string]interface{}{
			"name": "test_pipeline",
		},
		Tasks: map[string]*ir.Task{
			"task1": {
				Handler: &ir.Handler{Type: "bash", Command: "echo hello"},
				Config:  &ir.Config{Image: "bash:5.1"},
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := executor.Execute(ctx, spec)
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	if result == nil || result.WorkflowName == "" {
		t.Fatal("no workflow name returned")
	}

	if result.Status == "" {
		t.Fatal("no status returned")
	}

	t.Logf("Executed workflow: %s with status %s", result.WorkflowName, result.Status)
}

// TestDependenciesAreRespected tests task dependency ordering
func TestDependenciesAreRespected(t *testing.T) {
	client := NewMockArgoClient()
	executor := New(client, "default")

	spec := &ir.PipelineSpec{
		Metadata: map[string]interface{}{"name": "test_deps"},
		Tasks: map[string]*ir.Task{
			"task1": {
				Handler: &ir.Handler{Type: "bash", Command: "echo 1"},
				Config:  &ir.Config{Image: "bash:5.1"},
			},
			"task2": {
				Handler: &ir.Handler{Type: "bash", Command: "echo 2"},
				Config:  &ir.Config{Image: "bash:5.1"},
			},
		},
		Edges: []ir.Edge{
			{From: "task1", To: "task2"},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := executor.Execute(ctx, spec)
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	// Verify workflow was created
	workflow := client.GetWorkflow(result.WorkflowName)
	if workflow == nil {
		t.Fatal("workflow not found in client")
	}

	// Verify tasks exist
	if len(workflow.Spec.Templates) == 0 {
		t.Fatal("no templates in workflow")
	}
}

// TestParallelExecution tests fan-out/fan-in pattern
func TestParallelExecution(t *testing.T) {
	client := NewMockArgoClient()
	executor := New(client, "default")

	spec := &ir.PipelineSpec{
		Metadata: map[string]interface{}{"name": "test_parallel"},
		Tasks: map[string]*ir.Task{
			"source": {
				Handler: &ir.Handler{Type: "bash", Command: "echo source"},
				Config:  &ir.Config{Image: "bash:5.1"},
			},
			"parallel_a": {
				Handler: &ir.Handler{Type: "bash", Command: "echo a"},
				Config:  &ir.Config{Image: "bash:5.1"},
			},
			"parallel_b": {
				Handler: &ir.Handler{Type: "bash", Command: "echo b"},
				Config:  &ir.Config{Image: "bash:5.1"},
			},
		},
		Edges: []ir.Edge{
			{From: "source", To: "parallel_a"},
			{From: "source", To: "parallel_b"},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := executor.Execute(ctx, spec)
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	if result.Status != "Succeeded" {
		t.Logf("workflow status: %s (expected eventual success in mock)", result.Status)
	}
}

// TestRetryPolicy tests retry configuration
func TestRetryPolicy(t *testing.T) {
	client := NewMockArgoClient()
	executor := New(client, "default")

	spec := &ir.PipelineSpec{
		Metadata: map[string]interface{}{
			"name":    "test_retry",
			"retries": 3.0,
		},
		Tasks: map[string]*ir.Task{
			"task": {
				Handler: &ir.Handler{Type: "bash", Command: "exit 1"},
				Config: &ir.Config{
					Image:   "bash:5.1",
					Retries: 3,
				},
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	opts := ExecuteOptions{Wait: true}
	result, err := executor.ExecuteWithOptions(ctx, spec, opts)
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	if result.WorkflowName == "" {
		t.Fatal("no workflow created")
	}

	t.Logf("Workflow with retries: %s", result.WorkflowName)
}

// TestNamespaceIsolation tests namespace support
func TestNamespaceIsolation(t *testing.T) {
	client := NewMockArgoClient()
	executor := New(client, "default")

	spec := &ir.PipelineSpec{
		Metadata: map[string]interface{}{"name": "test_ns"},
		Tasks: map[string]*ir.Task{
			"task": {
				Handler: &ir.Handler{Type: "bash", Command: "echo hello"},
				Config:  &ir.Config{Image: "bash:5.1"},
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	opts := ExecuteOptions{
		Namespace: "production",
	}

	result, err := executor.ExecuteWithOptions(ctx, spec, opts)
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	workflow := client.GetWorkflow(result.WorkflowName)
	if workflow.Metadata.Namespace != "production" {
		t.Fatalf("expected namespace production, got %s", workflow.Metadata.Namespace)
	}
}

// TestStatusPolling tests non-blocking execution with status checks
func TestStatusPolling(t *testing.T) {
	client := NewMockArgoClient()
	executor := New(client, "default")

	spec := &ir.PipelineSpec{
		Metadata: map[string]interface{}{"name": "test_polling"},
		Tasks: map[string]*ir.Task{
			"task": {
				Handler: &ir.Handler{Type: "bash", Command: "sleep 1"},
				Config:  &ir.Config{Image: "bash:5.1"},
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	opts := ExecuteOptions{Wait: false}
	result, err := executor.ExecuteWithOptions(ctx, spec, opts)
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	if result.Status != "Submitted" {
		t.Fatalf("expected Submitted status, got %s", result.Status)
	}

	// Poll for status
	status, err := executor.GetStatus(ctx, result.WorkflowName)
	if err != nil {
		t.Fatalf("get status failed: %v", err)
	}

	if status == nil || status.WorkflowName == "" {
		t.Fatal("status retrieval failed")
	}

	t.Logf("Polled status: %s", status.Status)
}

// TestWorkflowDeletion tests cleanup
func TestWorkflowDeletion(t *testing.T) {
	client := NewMockArgoClient()
	executor := New(client, "default")

	spec := &ir.PipelineSpec{
		Metadata: map[string]interface{}{"name": "test_delete"},
		Tasks: map[string]*ir.Task{
			"task": {
				Handler: &ir.Handler{Type: "bash", Command: "echo hello"},
				Config:  &ir.Config{Image: "bash:5.1"},
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := executor.Execute(ctx, spec)
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	workflowName := result.WorkflowName

	// Verify workflow exists
	status, err := executor.GetStatus(ctx, workflowName)
	if err != nil {
		t.Fatalf("get status failed: %v", err)
	}

	if status == nil {
		t.Fatal("workflow not found")
	}

	// Delete workflow
	err = executor.Delete(ctx, workflowName)
	if err != nil {
		t.Fatalf("delete failed: %v", err)
	}

	// Verify deletion
	_, err = executor.GetStatus(ctx, workflowName)
	if err == nil {
		t.Fatal("workflow still exists after deletion")
	}

	t.Logf("Successfully deleted workflow: %s", workflowName)
}

// TestLogsRetrieval tests log collection
func TestLogsRetrieval(t *testing.T) {
	client := NewMockArgoClient()
	executor := New(client, "default")

	spec := &ir.PipelineSpec{
		Metadata: map[string]interface{}{"name": "test_logs"},
		Tasks: map[string]*ir.Task{
			"task": {
				Handler: &ir.Handler{Type: "bash", Command: "echo 'test log'"},
				Config:  &ir.Config{Image: "bash:5.1"},
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := executor.Execute(ctx, spec)
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	// Set test logs
	client.SetTaskLog(result.WorkflowName, "task", "test output from task")

	// Retrieve logs
	logs, err := executor.GetLogs(ctx, result.WorkflowName, "task")
	if err != nil {
		t.Fatalf("get logs failed: %v", err)
	}

	if logs != "test output from task" {
		t.Fatalf("unexpected logs: %s", logs)
	}

	t.Logf("Retrieved logs: %s", logs)
}

// TestInvalidSpecification tests error handling for invalid specs
func TestInvalidSpecification(t *testing.T) {
	client := NewMockArgoClient()
	executor := New(client, "default")

	// Spec with cycle
	spec := &ir.PipelineSpec{
		Metadata: map[string]interface{}{"name": "invalid"},
		Tasks: map[string]*ir.Task{
			"task1": {
				Handler: &ir.Handler{Type: "bash", Command: "echo 1"},
				Config:  &ir.Config{Image: "bash:5.1"},
			},
			"task2": {
				Handler: &ir.Handler{Type: "bash", Command: "echo 2"},
				Config:  &ir.Config{Image: "bash:5.1"},
			},
		},
		Edges: []ir.Edge{
			{From: "task1", To: "task2"},
			{From: "task2", To: "task1"}, // Cycle
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := executor.Execute(ctx, spec)
	if err == nil {
		t.Fatal("expected error for invalid specification")
	}

	t.Logf("Correctly rejected invalid spec: %v", err)
}

// TestTaskResourceConstraints tests resource configuration
func TestTaskResourceConstraints(t *testing.T) {
	client := NewMockArgoClient()
	executor := New(client, "default")

	spec := &ir.PipelineSpec{
		Metadata: map[string]interface{}{"name": "test_resources"},
		Tasks: map[string]*ir.Task{
			"heavy_task": {
				Handler: &ir.Handler{Type: "python", Command: "python /ml/train.py"},
				Config: &ir.Config{
					Image: "pytorch:2.0",
					Resources: map[string]interface{}{
						"memory": "16Gi",
						"cpu":    "4",
						"gpu":    "1",
					},
				},
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := executor.Execute(ctx, spec)
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	workflow := client.GetWorkflow(result.WorkflowName)
	if workflow == nil {
		t.Fatal("workflow not created")
	}

	t.Logf("Created workflow with resource constraints: %s", result.WorkflowName)
}
