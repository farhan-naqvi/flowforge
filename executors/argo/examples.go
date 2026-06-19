package argo

import (
	"context"
	"fmt"
	"time"

	"flowforge/ir"
)

// Example functions demonstrating Argo executor usage

// ExampleETLExecution demonstrates a simple ETL pipeline execution
func ExampleETLExecution(client ArgoClient) error {
	executor := New(client, "default")

	// Create IR specification
	spec := &ir.PipelineSpec{
		Metadata: map[string]interface{}{
			"name":        "example_etl",
			"version":     "1.0.0",
			"description": "Example ETL pipeline",
		},
		Tasks: map[string]*ir.Task{
			"extract": {
				Handler: &ir.Handler{
					Type:    "python",
					Command: "python /scripts/extract.py",
				},
				Config: &ir.Config{
					Image: "python:3.11",
					Resources: map[string]interface{}{
						"memory": "512Mi",
						"cpu":    "100m",
					},
					Timeout: 300,
					Retries: 2,
				},
			},
			"transform": {
				Handler: &ir.Handler{
					Type:    "python",
					Command: "python /scripts/transform.py",
				},
				Config: &ir.Config{
					Image: "python:3.11",
					Resources: map[string]interface{}{
						"memory": "1Gi",
						"cpu":    "500m",
					},
					Timeout: 600,
					Retries: 1,
				},
			},
			"load": {
				Handler: &ir.Handler{
					Type:    "bash",
					Command: "bash /scripts/load.sh",
				},
				Config: &ir.Config{
					Image: "bash:5.1",
					Resources: map[string]interface{}{
						"memory": "256Mi",
						"cpu":    "100m",
					},
					Timeout: 300,
					Retries: 3,
				},
			},
		},
		Edges: []ir.Edge{
			{From: "extract", To: "transform", FromPort: "output", ToPort: "input"},
			{From: "transform", To: "load", FromPort: "result", ToPort: "data"},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	opts := ExecuteOptions{
		Wait:        true,
		Namespace:   "default",
		Parallelism: 0,
	}

	result, err := executor.ExecuteWithOptions(ctx, spec, opts)
	if err != nil {
		return fmt.Errorf("execution failed: %w", err)
	}

	fmt.Printf("Workflow: %s\n", result.WorkflowName)
	fmt.Printf("Status: %s\n", result.Status)
	fmt.Printf("Duration: %v\n", result.Duration)
	fmt.Printf("Succeeded: %v\n", result.Succeeded)

	return nil
}

// ExampleFanOutFanInExecution demonstrates parallel execution
func ExampleFanOutFanInExecution(client ArgoClient) error {
	executor := New(client, "default")

	spec := &ir.PipelineSpec{
		Metadata: map[string]interface{}{
			"name":    "fan_out_fan_in",
			"version": "1.0.0",
		},
		Tasks: map[string]*ir.Task{
			"source": {
				Handler: &ir.Handler{
					Type:    "python",
					Command: "python /scripts/source.py",
				},
				Config: &ir.Config{Image: "python:3.11"},
			},
			"process_a": {
				Handler: &ir.Handler{
					Type:    "python",
					Command: "python /scripts/process_a.py",
				},
				Config: &ir.Config{Image: "python:3.11"},
			},
			"process_b": {
				Handler: &ir.Handler{
					Type:    "python",
					Command: "python /scripts/process_b.py",
				},
				Config: &ir.Config{Image: "python:3.11"},
			},
			"merge": {
				Handler: &ir.Handler{
					Type:    "python",
					Command: "python /scripts/merge.py",
				},
				Config: &ir.Config{Image: "python:3.11"},
			},
		},
		Edges: []ir.Edge{
			{From: "source", To: "process_a"},
			{From: "source", To: "process_b"},
			{From: "process_a", To: "merge"},
			{From: "process_b", To: "merge"},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	result, err := executor.Execute(ctx, spec)
	if err != nil {
		return err
	}

	fmt.Printf("Fan-out/fan-in workflow: %s\n", result.WorkflowName)
	fmt.Printf("Tasks: %d\n", len(result.TaskResults))

	for taskID, taskResult := range result.TaskResults {
		fmt.Printf("  %s: %s (%v)\n", taskID, taskResult.Status, taskResult.Duration)
	}

	return nil
}

// ExampleLongRunningExecution demonstrates long-running pipeline monitoring
func ExampleLongRunningExecution(client ArgoClient) error {
	executor := New(client, "default")

	spec := &ir.PipelineSpec{
		Metadata: map[string]interface{}{
			"name":    "long_running",
			"version": "1.0.0",
		},
		Tasks: map[string]*ir.Task{
			"train": {
				Handler: &ir.Handler{
					Type:    "python",
					Command: "python /ml/train.py",
				},
				Config: &ir.Config{
					Image: "pytorch:2.0",
					Resources: map[string]interface{}{
						"memory": "16Gi",
						"cpu":    "4",
					},
					Timeout: 3600,
				},
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Hour)
	defer cancel()

	// Start without waiting
	opts := ExecuteOptions{
		Wait: false,
	}

	result, err := executor.ExecuteWithOptions(ctx, spec, opts)
	if err != nil {
		return err
	}

	fmt.Printf("Submitted: %s\n", result.WorkflowName)

	// Poll for status periodically
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			status, err := executor.GetStatus(ctx, result.WorkflowName)
			if err != nil {
				return err
			}

			fmt.Printf("Status: %s (Duration: %v)\n", status.Status, status.Duration)

			if status.Succeeded || status.Failed {
				return nil
			}

		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// ExampleWithArtifacts demonstrates pipeline with artifact handling
func ExampleWithArtifacts(client ArgoClient) error {
	executor := New(client, "default")

	spec := &ir.PipelineSpec{
		Metadata: map[string]interface{}{
			"name":    "artifact_pipeline",
			"version": "1.0.0",
		},
		Tasks: map[string]*ir.Task{
			"generate": {
				Handler: &ir.Handler{
					Type:    "python",
					Command: "python /scripts/generate.py > /output/data.csv",
				},
				Config: &ir.Config{
					Image: "python:3.11",
				},
			},
			"process": {
				Handler: &ir.Handler{
					Type:    "python",
					Command: "python /scripts/process.py",
				},
				Config: &ir.Config{
					Image: "python:3.11",
				},
			},
		},
		Edges: []ir.Edge{
			{From: "generate", To: "process"},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	result, err := executor.Execute(ctx, spec)
	if err != nil {
		return err
	}

	fmt.Printf("Execution: %s\n", result.WorkflowName)
	fmt.Printf("Result: %s\n", result.Status)

	// Retrieve artifacts and logs
	for taskID := range spec.Tasks {
		logs, err := executor.GetLogs(ctx, result.WorkflowName, taskID)
		if err == nil && logs != "" {
			fmt.Printf("Logs for %s:\n%s\n", taskID, logs)
		}
	}

	return nil
}

// ExampleScheduledPipeline demonstrates scheduled pipeline execution
func ExampleScheduledPipeline(client ArgoClient) error {
	executor := New(client, "default")

	spec := &ir.PipelineSpec{
		Metadata: map[string]interface{}{
			"name":     "scheduled_job",
			"version":  "1.0.0",
			"schedule": "0 2 * * *", // Daily at 2 AM
		},
		Tasks: map[string]*ir.Task{
			"daily_report": {
				Handler: &ir.Handler{
					Type:    "python",
					Command: "python /scripts/daily_report.py",
				},
				Config: &ir.Config{
					Image:   "python:3.11",
					Timeout: 1800,
					Retries: 2,
				},
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Hour)
	defer cancel()

	result, err := executor.Execute(ctx, spec)
	if err != nil {
		return err
	}

	fmt.Printf("Scheduled workflow: %s\n", result.WorkflowName)
	return nil
}

// ExampleRetryLogic demonstrates retry policy configuration
func ExampleRetryLogic(client ArgoClient) error {
	executor := New(client, "default")

	spec := &ir.PipelineSpec{
		Metadata: map[string]interface{}{
			"name":    "retry_example",
			"version": "1.0.0",
			"retries": 5.0, // Attempt up to 5 times
		},
		Tasks: map[string]*ir.Task{
			"unreliable_task": {
				Handler: &ir.Handler{
					Type:    "bash",
					Command: "bash /scripts/flaky.sh",
				},
				Config: &ir.Config{
					Image:   "bash:5.1",
					Retries: 5,
					Timeout: 600,
				},
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	opts := ExecuteOptions{
		Wait: true,
	}

	result, err := executor.ExecuteWithOptions(ctx, spec, opts)
	if err != nil {
		return err
	}

	fmt.Printf("Execution with retries: %s\n", result.WorkflowName)
	fmt.Printf("Final status: %s\n", result.Status)

	return nil
}

// ExampleCleanupOnCompletion demonstrates workflow cleanup
func ExampleCleanupOnCompletion(client ArgoClient) error {
	executor := New(client, "default")

	spec := &ir.PipelineSpec{
		Metadata: map[string]interface{}{
			"name":    "cleanup_example",
			"version": "1.0.0",
			"ttl":     "1h", // Delete after 1 hour
		},
		Tasks: map[string]*ir.Task{
			"main_task": {
				Handler: &ir.Handler{
					Type:    "python",
					Command: "python /scripts/main.py",
				},
				Config: &ir.Config{Image: "python:3.11"},
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	opts := ExecuteOptions{
		TTL: 1 * time.Hour,
	}

	result, err := executor.ExecuteWithOptions(ctx, spec, opts)
	if err != nil {
		return err
	}

	fmt.Printf("Execution: %s (cleanup after 1h)\n", result.WorkflowName)
	return nil
}
