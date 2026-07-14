package deployment

import (
	"context"
	"testing"
	"time"

	"flowforge/ir"
)

// TestPlanGeneration tests deployment plan generation
func TestPlanGeneration(t *testing.T) {
	engine := New(
		NewInMemoryStateManager(),
		&MockPlanGenerator{},
		&MockApplier{},
		NewMockVersionManager(),
	)

	spec := &ir.PipelineSpec{
		Metadata: map[string]interface{}{
			"name": "test_pipeline",
		},
		Tasks: map[string]*ir.Task{
			"task1": {
				Handler: &ir.Handler{Type: "python", Command: "python /scripts/task1.py"},
				Config:  &ir.Config{Image: "python:3.11"},
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	plan, err := engine.Plan(ctx, spec)
	if err != nil {
		t.Fatalf("plan generation failed: %v", err)
	}

	if plan == nil || plan.PipelineID == "" {
		t.Fatal("invalid plan returned")
	}

	if len(plan.Containers) == 0 {
		t.Fatal("no containers in plan")
	}

	t.Logf("Generated plan for %s with %d containers", plan.PipelineID, len(plan.Containers))
}

// TestDeploymentApply tests applying a deployment plan
func TestDeploymentApply(t *testing.T) {
	engine := New(
		NewInMemoryStateManager(),
		&MockPlanGenerator{},
		&MockApplier{},
		NewMockVersionManager(),
	)

	spec := &ir.PipelineSpec{
		Metadata: map[string]interface{}{
			"name":    "test_apply",
			"version": "1.0.0",
		},
		Tasks: map[string]*ir.Task{
			"task1": {
				Handler: &ir.Handler{Type: "bash", Command: "echo hello"},
				Config:  &ir.Config{Image: "bash:5.1"},
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := engine.Apply(ctx, spec)
	if err != nil {
		t.Fatalf("apply failed: %v", err)
	}

	if !result.Success {
		t.Fatal("deployment not successful")
	}

	if result.Created == 0 {
		t.Fatal("no resources created")
	}

	t.Logf("Applied deployment: %d resources created", result.Created)
}

// TestStateManagement tests state tracking
func TestStateManagement(t *testing.T) {
	engine := New(
		NewInMemoryStateManager(),
		&MockPlanGenerator{},
		&MockApplier{},
		NewMockVersionManager(),
	)

	spec := &ir.PipelineSpec{
		Metadata: map[string]interface{}{
			"name": "test_state",
		},
		Tasks: map[string]*ir.Task{
			"task1": {
				Handler: &ir.Handler{Type: "bash", Command: "echo test"},
				Config:  &ir.Config{Image: "bash:5.1"},
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Apply deployment
	_, err := engine.Apply(ctx, spec)
	if err != nil {
		t.Fatalf("apply failed: %v", err)
	}

	// Get status
	status, err := engine.GetStatus(ctx, "test_state")
	if err != nil {
		t.Fatalf("get status failed: %v", err)
	}

	if status == nil || status.PipelineID != "test_state" {
		t.Fatal("invalid status returned")
	}

	t.Logf("State retrieved: %s with status %s", status.PipelineID, status.Status)
}

// TestDeploymentHistory tests history tracking
func TestDeploymentHistory(t *testing.T) {
	stateMgr := NewInMemoryStateManager()
	engine := New(
		stateMgr,
		&MockPlanGenerator{},
		&MockApplier{},
		NewMockVersionManager(),
	)

	spec := &ir.PipelineSpec{
		Metadata: map[string]interface{}{
			"name": "test_history",
		},
		Tasks: map[string]*ir.Task{
			"task1": {
				Handler: &ir.Handler{Type: "bash", Command: "echo test"},
				Config:  &ir.Config{Image: "bash:5.1"},
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create multiple deployments
	for i := 0; i < 3; i++ {
		_, err := engine.Apply(ctx, spec)
		if err != nil {
			t.Fatalf("apply failed: %v", err)
		}
		time.Sleep(10 * time.Millisecond)
	}

	// Get history
	history, err := engine.GetHistory(ctx, "test_history", 10)
	if err != nil {
		t.Fatalf("get history failed: %v", err)
	}

	if len(history) == 0 {
		t.Fatal("no history entries")
	}

	t.Logf("Retrieved %d history entries", len(history))
}

// TestDryRun tests dry-run mode
func TestDryRun(t *testing.T) {
	engine := New(
		NewInMemoryStateManager(),
		&MockPlanGenerator{},
		&MockApplier{},
		NewMockVersionManager(),
	)

	spec := &ir.PipelineSpec{
		Metadata: map[string]interface{}{
			"name": "test_dry_run",
		},
		Tasks: map[string]*ir.Task{
			"task1": {
				Handler: &ir.Handler{Type: "bash", Command: "echo test"},
				Config:  &ir.Config{Image: "bash:5.1"},
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	plan, err := engine.DryRun(ctx, spec)
	if err != nil {
		t.Fatalf("dry run failed: %v", err)
	}

	if !plan.DryRun {
		t.Fatal("plan not marked as dry-run")
	}

	t.Logf("Dry-run successful for %s", plan.PipelineID)
}

// TestDestroy tests resource destruction
func TestDestroy(t *testing.T) {
	engine := New(
		NewInMemoryStateManager(),
		&MockPlanGenerator{},
		&MockApplier{},
		NewMockVersionManager(),
	)

	spec := &ir.PipelineSpec{
		Metadata: map[string]interface{}{
			"name": "test_destroy",
		},
		Tasks: map[string]*ir.Task{
			"task1": {
				Handler: &ir.Handler{Type: "bash", Command: "echo test"},
				Config:  &ir.Config{Image: "bash:5.1"},
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Deploy first
	_, err := engine.Apply(ctx, spec)
	if err != nil {
		t.Fatalf("apply failed: %v", err)
	}

	// Destroy
	err = engine.Destroy(ctx, "test_destroy")
	if err != nil {
		t.Fatalf("destroy failed: %v", err)
	}

	// Verify state is updated
	status, _ := engine.GetStatus(ctx, "test_destroy")
	if status.Status != "Destroyed" {
		t.Fatalf("expected Destroyed status, got %s", status.Status)
	}

	t.Logf("Successfully destroyed deployment: %s", "test_destroy")
}

// TestRollback tests deployment rollback
func TestRollback(t *testing.T) {
	engine := New(
		NewInMemoryStateManager(),
		&MockPlanGenerator{},
		&MockApplier{},
		NewMockVersionManager(),
	)

	spec := &ir.PipelineSpec{
		Metadata: map[string]interface{}{
			"name": "test_rollback",
		},
		Tasks: map[string]*ir.Task{
			"task1": {
				Handler: &ir.Handler{Type: "bash", Command: "echo test"},
				Config:  &ir.Config{Image: "bash:5.1"},
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Deploy
	result, err := engine.Apply(ctx, spec)
	if err != nil {
		t.Fatalf("apply failed: %v", err)
	}

	version1 := result.Version

	// Rollback to version 1
	err = engine.Rollback(ctx, "test_rollback", version1)
	if err != nil {
		t.Fatalf("rollback failed: %v", err)
	}

	// Verify state
	status, err := engine.GetStatus(ctx, "test_rollback")
	if err != nil {
		t.Fatalf("get status failed: %v", err)
	}

	if status.Status != "Rolled Back" {
		t.Fatalf("expected Rolled Back status, got %s", status.Status)
	}

	t.Logf("Successfully rolled back to version: %s", version1)
}

// TestVersionTracking tests version management
func TestVersionTracking(t *testing.T) {
	versionMgr := NewMockVersionManager()
	engine := New(
		NewInMemoryStateManager(),
		&MockPlanGenerator{},
		&MockApplier{},
		versionMgr,
	)

	pipelineID := "test_versions"

	// Add versions
	versionMgr.AddVersion(pipelineID, "1.0.0")
	versionMgr.AddVersion(pipelineID, "1.1.0")
	versionMgr.AddVersion(pipelineID, "1.2.0")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get current version
	current, err := engine.GetVersion(ctx, pipelineID)
	if err != nil {
		t.Fatalf("get version failed: %v", err)
	}

	if current != "1.2.0" {
		t.Fatalf("expected 1.2.0, got %s", current)
	}

	// List all versions
	versions, err := engine.ListVersions(ctx, pipelineID)
	if err != nil {
		t.Fatalf("list versions failed: %v", err)
	}

	if len(versions) != 3 {
		t.Fatalf("expected 3 versions, got %d", len(versions))
	}

	t.Logf("Version tracking: %d versions available", len(versions))
}

// TestTerraformGeneration tests Terraform code generation
func TestTerraformGeneration(t *testing.T) {
	gen := NewTerraformGenerator("kubernetes")

	spec := &ir.PipelineSpec{
		Metadata: map[string]interface{}{
			"name": "test_tf",
		},
		Tasks: map[string]*ir.Task{
			"task1": {
				Handler: &ir.Handler{Type: "bash", Command: "echo test"},
				Config:  &ir.Config{Image: "bash:5.1"},
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	terraform, err := gen.Generate(ctx, spec)
	if err != nil {
		t.Fatalf("terraform generation failed: %v", err)
	}

	if terraform == "" {
		t.Fatal("no terraform output")
	}

	if !contains(terraform, "kubernetes_namespace") {
		t.Fatal("missing namespace resource")
	}

	t.Logf("Generated %d lines of Terraform", countLines(terraform))
}

// TestHelmGeneration tests Helm configuration generation
func TestHelmGeneration(t *testing.T) {
	gen := NewHelmChartGenerator("flowforge-pipeline", "1.0.0")

	spec := &ir.PipelineSpec{
		Metadata: map[string]interface{}{
			"name":    "test_helm",
			"version": "1.0.0",
		},
		Tasks: map[string]*ir.Task{
			"task1": {
				Handler: &ir.Handler{Type: "bash", Command: "echo test"},
				Config:  &ir.Config{Image: "bash:5.1"},
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	chart, err := gen.GenerateChart(ctx, spec)
	if err != nil {
		t.Fatalf("helm chart generation failed: %v", err)
	}

	if chart["name"] != "flowforge-pipeline" {
		t.Fatal("invalid chart name")
	}

	values, err := gen.GenerateValues(ctx, spec)
	if err != nil {
		t.Fatalf("helm values generation failed: %v", err)
	}

	if values["namespace"] == nil {
		t.Fatal("missing namespace in values")
	}

	t.Logf("Generated Helm chart and values")
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

// TestInvalidSpec tests error handling for invalid specifications
func TestInvalidSpec(t *testing.T) {
	engine := New(
		NewInMemoryStateManager(),
		&MockPlanGenerator{},
		&MockApplier{},
		NewMockVersionManager(),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Test with nil spec
	_, err := engine.Apply(ctx, nil)
	if err == nil {
		t.Fatal("expected error for nil spec")
	}

	// Test with missing name
	spec := &ir.PipelineSpec{
		Metadata: map[string]interface{}{},
		Tasks:    map[string]*ir.Task{},
	}

	_, err = engine.Apply(ctx, spec)
	if err == nil {
		t.Fatal("expected error for missing name")
	}

	t.Logf("Correctly rejected invalid specifications")
}
