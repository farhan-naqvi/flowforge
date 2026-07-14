package runtime

import (
	"context"
	"testing"
	"time"
)

// TestContainerization tests function containerization
func TestContainerization(t *testing.T) {
	rt := New(
		NewMockContainerManager(),
		NewMockRegistryClient(),
		NewMockExecutor(),
		NewMockVersionManager(),
		&MockLogCollector{},
	)

	functionCode := `
def transform(data):
    return [x * 2 for x in data]
`

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	image, err := rt.ContainerizeFunction(ctx, "func1", "transform", functionCode, []string{"numpy", "pandas"})
	if err != nil {
		t.Fatalf("containerization failed: %v", err)
	}

	if image == nil || image.ID == "" {
		t.Fatal("no image created")
	}

	if image.Name != "transform" {
		t.Fatalf("expected name 'transform', got %s", image.Name)
	}

	t.Logf("Containerized function: %s (size: %dMB)", image.ID, image.Size/(1024*1024))
}

// TestPushToRegistry tests pushing images to registry
func TestPushToRegistry(t *testing.T) {
	rt := New(
		NewMockContainerManager(),
		NewMockRegistryClient(),
		NewMockExecutor(),
		NewMockVersionManager(),
		&MockLogCollector{},
	)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	image, _ := rt.ContainerizeFunction(ctx, "func1", "test_func", "def test(): pass", nil)

	imageRef, err := rt.PushToRegistry(ctx, image)
	if err != nil {
		t.Fatalf("push failed: %v", err)
	}

	if imageRef == "" {
		t.Fatal("no image reference returned")
	}

	t.Logf("Pushed image: %s", imageRef)
}

// TestVersioning tests image versioning
func TestVersioning(t *testing.T) {
	rt := New(
		NewMockContainerManager(),
		NewMockRegistryClient(),
		NewMockExecutor(),
		NewMockVersionManager(),
		&MockLogCollector{},
	)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	image, _ := rt.ContainerizeFunction(ctx, "func1", "versioned_func", "def v1(): pass", nil)

	version, err := rt.TagVersion(ctx, "func1", image, "def v1(): pass", "Version 1.0")
	if err != nil {
		t.Fatalf("tag version failed: %v", err)
	}

	if version == nil || version.Tag == "" {
		t.Fatal("no version tag")
	}

	t.Logf("Created version: %s", version.Tag)
}

// TestExecution tests function execution
func TestExecution(t *testing.T) {
	rt := New(
		NewMockContainerManager(),
		NewMockRegistryClient(),
		NewMockExecutor(),
		NewMockVersionManager(),
		&MockLogCollector{},
	)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create and tag a function
	image, _ := rt.ContainerizeFunction(ctx, "func1", "exec_func", "def run(): return 42", nil)
	version, _ := rt.TagVersion(ctx, "func1", image, "def run(): return 42", "v1")

	// Execute
	opts := &ExecutionOptions{
		Timeout: 30 * time.Second,
		Memory:  512,
		Wait:    true,
	}

	result, err := rt.ExecuteFunction(ctx, "func1", version.Tag, opts)
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	if result.Status != "Succeeded" {
		t.Fatalf("expected Succeeded, got %s", result.Status)
	}

	if result.ExitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", result.ExitCode)
	}

	t.Logf("Executed function: %s (duration: %v)", result.ExecutionID, result.Duration)
}

// TestLogsCollection tests log collection
func TestLogsCollection(t *testing.T) {
	rt := New(
		NewMockContainerManager(),
		NewMockRegistryClient(),
		NewMockExecutor(),
		NewMockVersionManager(),
		&MockLogCollector{},
	)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	image, _ := rt.ContainerizeFunction(ctx, "func1", "log_func", "print('hello')", nil)
	version, _ := rt.TagVersion(ctx, "func1", image, "print('hello')", "v1")

	result, _ := rt.ExecuteFunction(ctx, "func1", version.Tag, &ExecutionOptions{})

	// Get logs
	logs, err := rt.GetExecutionLogs(ctx, result.ExecutionID)
	if err != nil {
		t.Fatalf("get logs failed: %v", err)
	}

	if logs == "" {
		t.Fatal("no logs returned")
	}

	t.Logf("Collected logs: %d bytes", len(logs))
}

// TestStreamingLogs tests log streaming
func TestStreamingLogs(t *testing.T) {
	rt := New(
		NewMockContainerManager(),
		NewMockRegistryClient(),
		NewMockExecutor(),
		NewMockVersionManager(),
		&MockLogCollector{},
	)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	image, _ := rt.ContainerizeFunction(ctx, "func1", "stream_func", "for i in range(3): print(i)", nil)
	version, _ := rt.TagVersion(ctx, "func1", image, "for i in range(3): print(i)", "v1")

	result, _ := rt.ExecuteFunction(ctx, "func1", version.Tag, &ExecutionOptions{})

	// Stream logs
	logCh, err := rt.StreamExecutionLogs(ctx, result.ExecutionID)
	if err != nil {
		t.Fatalf("stream logs failed: %v", err)
	}

	count := 0
	for line := range logCh {
		if line != "" {
			count++
		}
	}

	if count == 0 {
		t.Fatal("no log lines received")
	}

	t.Logf("Streamed %d log lines", count)
}

// TestVersionHistory tests version tracking
func TestVersionHistory(t *testing.T) {
	rt := New(
		NewMockContainerManager(),
		NewMockRegistryClient(),
		NewMockExecutor(),
		NewMockVersionManager(),
		&MockLogCollector{},
	)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create multiple versions
	for i := 1; i <= 3; i++ {
		code := "def run(): pass # v" + string(rune('0'+i))
		image, _ := rt.ContainerizeFunction(ctx, "func1", "versioned_func", code, nil)
		rt.TagVersion(ctx, "func1", image, code, "")
	}

	// List versions
	versions, err := rt.ListVersions(ctx, "func1")
	if err != nil {
		t.Fatalf("list versions failed: %v", err)
	}

	if len(versions) != 3 {
		t.Fatalf("expected 3 versions, got %d", len(versions))
	}

	t.Logf("Created %d versions", len(versions))
}

// TestRollback tests version rollback
func TestRollback(t *testing.T) {
	rt := New(
		NewMockContainerManager(),
		NewMockRegistryClient(),
		NewMockExecutor(),
		NewMockVersionManager(),
		&MockLogCollector{},
	)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create and tag a version
	image, _ := rt.ContainerizeFunction(ctx, "func1", "rollback_func", "def run(): return 1", nil)
	version, _ := rt.TagVersion(ctx, "func1", image, "def run(): return 1", "v1")

	// Rollback to that version
	rollbackVersion, err := rt.RollbackVersion(ctx, "func1", version.Tag)
	if err != nil {
		t.Fatalf("rollback failed: %v", err)
	}

	if !rollbackVersion.Rollback {
		t.Fatal("rollback flag not set")
	}

	t.Logf("Rolled back to version: %s", rollbackVersion.Tag)
}

// TestGetVersion tests retrieving specific versions
func TestGetVersion(t *testing.T) {
	rt := New(
		NewMockContainerManager(),
		NewMockRegistryClient(),
		NewMockExecutor(),
		NewMockVersionManager(),
		&MockLogCollector{},
	)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create a version
	image, _ := rt.ContainerizeFunction(ctx, "func1", "get_func", "def run(): pass", nil)
	version, _ := rt.TagVersion(ctx, "func1", image, "def run(): pass", "v1")

	// Retrieve it
	retrieved, err := rt.GetVersion(ctx, "func1", version.Tag)
	if err != nil {
		t.Fatalf("get version failed: %v", err)
	}

	if retrieved == nil || retrieved.Tag != version.Tag {
		t.Fatal("version mismatch")
	}

	t.Logf("Retrieved version: %s", retrieved.Tag)
}

// TestStopExecution tests stopping an execution
func TestStopExecution(t *testing.T) {
	rt := New(
		NewMockContainerManager(),
		NewMockRegistryClient(),
		NewMockExecutor(),
		NewMockVersionManager(),
		&MockLogCollector{},
	)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	image, _ := rt.ContainerizeFunction(ctx, "func1", "stop_func", "def run(): pass", nil)
	version, _ := rt.TagVersion(ctx, "func1", image, "def run(): pass", "v1")

	result, _ := rt.ExecuteFunction(ctx, "func1", version.Tag, &ExecutionOptions{})

	// Stop execution
	err := rt.StopExecution(ctx, result.ExecutionID)
	if err != nil {
		t.Fatalf("stop execution failed: %v", err)
	}

	t.Logf("Stopped execution: %s", result.ExecutionID)
}

// TestGetExecutionStatus tests status retrieval
func TestGetExecutionStatus(t *testing.T) {
	rt := New(
		NewMockContainerManager(),
		NewMockRegistryClient(),
		NewMockExecutor(),
		NewMockVersionManager(),
		&MockLogCollector{},
	)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	image, _ := rt.ContainerizeFunction(ctx, "func1", "status_func", "def run(): pass", nil)
	version, _ := rt.TagVersion(ctx, "func1", image, "def run(): pass", "v1")

	result, _ := rt.ExecuteFunction(ctx, "func1", version.Tag, &ExecutionOptions{})

	// Get status
	status, err := rt.GetExecutionStatus(ctx, result.ExecutionID)
	if err != nil {
		t.Fatalf("get status failed: %v", err)
	}

	if status == nil || status.Status != "Succeeded" {
		t.Fatal("invalid status")
	}

	t.Logf("Execution status: %s (%d%% complete)", status.Status, status.Progress)
}

// TestMultipleFunctions tests managing multiple functions
func TestMultipleFunctions(t *testing.T) {
	rt := New(
		NewMockContainerManager(),
		NewMockRegistryClient(),
		NewMockExecutor(),
		NewMockVersionManager(),
		&MockLogCollector{},
	)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create multiple functions
	functions := []struct {
		id   string
		name string
		code string
	}{
		{"func1", "func_a", "def a(): pass"},
		{"func2", "func_b", "def b(): pass"},
		{"func3", "func_c", "def c(): pass"},
	}

	for _, f := range functions {
		image, _ := rt.ContainerizeFunction(ctx, f.id, f.name, f.code, nil)
		if image == nil {
			t.Fatalf("failed to containerize %s", f.id)
		}
	}

	t.Logf("Created %d containerized functions", len(functions))
}

// TestErrorHandling tests error handling
func TestErrorHandling(t *testing.T) {
	rt := New(
		NewMockContainerManager(),
		NewMockRegistryClient(),
		NewMockExecutor(),
		NewMockVersionManager(),
		&MockLogCollector{},
	)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Try to execute non-existent version
	_, err := rt.ExecuteFunction(ctx, "non_existent", "v1", &ExecutionOptions{})
	if err == nil {
		t.Fatal("expected error for non-existent function")
	}

	// Try to rollback non-existent version
	_, err = rt.RollbackVersion(ctx, "non_existent", "v1")
	if err == nil {
		t.Fatal("expected error for non-existent version")
	}

	t.Logf("Error handling working correctly")
}
