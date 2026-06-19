package argo

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"flowforge/ir"
)

// ArgoWorkflowExecutor handles execution of pipelines on Argo Workflows
type ArgoWorkflowExecutor struct {
	client       ArgoClient
	namespace    string
	watchTimeout time.Duration
}

// ArgoClient is the interface for communicating with Argo
type ArgoClient interface {
	Submit(ctx context.Context, workflow *ArgoWorkflow) (string, error)
	GetStatus(ctx context.Context, workflowName string) (*WorkflowStatus, error)
	GetLogs(ctx context.Context, workflowName, podName string) (string, error)
	Delete(ctx context.Context, workflowName string) error
	Watch(ctx context.Context, workflowName string) (<-chan WorkflowStatus, error)
}

// ExecutionResult contains the result of a pipeline execution
type ExecutionResult struct {
	WorkflowName string
	Status       string
	StartTime    time.Time
	EndTime      time.Time
	Duration     time.Duration
	Succeeded    bool
	Failed       bool
	Message      string
	TaskResults  map[string]TaskResult
}

// TaskResult contains the result of a single task
type TaskResult struct {
	TaskID    string
	Status    string
	ExitCode  int
	Logs      string
	Artifacts []ArtifactRef
	StartTime time.Time
	EndTime   time.Time
	Duration  time.Duration
}

// ArtifactRef references an artifact produced by a task
type ArtifactRef struct {
	Name string
	Path string
	Mode string
	Size int64
	Hash string
}

// New creates a new Argo executor
func New(client ArgoClient, namespace string) *ArgoWorkflowExecutor {
	return &ArgoWorkflowExecutor{
		client:       client,
		namespace:    namespace,
		watchTimeout: 30 * time.Minute,
	}
}

// ExecuteWithOptions executes a pipeline on Argo Workflows
func (e *ArgoWorkflowExecutor) ExecuteWithOptions(ctx context.Context, spec *ir.PipelineSpec, opts ExecuteOptions) (*ExecutionResult, error) {
	// Validate the spec
	validator := ir.NewValidator()
	if errs := validator.ValidatePipelineSpec(spec); len(errs) > 0 {
		return nil, fmt.Errorf("invalid pipeline specification: %v", errs)
	}

	// Generate Argo Workflow
	compiler := NewCompiler()
	workflow, err := compiler.Compile(ctx, spec, opts.CompilerOptions)
	if err != nil {
		return nil, fmt.Errorf("compilation failed: %w", err)
	}

	// Set namespace
	workflow.Metadata.Namespace = e.namespace

	// Apply overrides
	if opts.Namespace != "" {
		workflow.Metadata.Namespace = opts.Namespace
	}
	if opts.Parallelism > 0 {
		workflow.Spec.Parallelism = opts.Parallelism
	}
	if opts.TTL > 0 {
		workflow.Spec.TTLSecondsAfterFinished = int32(opts.TTL.Seconds())
	}

	// Submit to Argo
	workflowName, err := e.client.Submit(ctx, workflow)
	if err != nil {
		return nil, fmt.Errorf("submission failed: %w", err)
	}

	// Watch execution
	result := &ExecutionResult{
		WorkflowName: workflowName,
		StartTime:    time.Now(),
		TaskResults:  make(map[string]TaskResult),
	}

	if opts.Wait {
		statusChan, err := e.client.Watch(ctx, workflowName)
		if err != nil {
			return nil, fmt.Errorf("watch failed: %w", err)
		}

		watchCtx, cancel := context.WithTimeout(ctx, e.watchTimeout)
		defer cancel()

		for {
			select {
			case status := <-statusChan:
				if status.Phase == "Succeeded" || status.Phase == "Failed" {
					result.Status = status.Phase
					result.Succeeded = status.Phase == "Succeeded"
					result.Failed = status.Phase == "Failed"
					result.EndTime = time.Now()
					result.Duration = result.EndTime.Sub(result.StartTime)
					result.Message = status.Message

					// Collect task results
					for taskID, taskStatus := range status.TaskStatuses {
						result.TaskResults[taskID] = TaskResult{
							TaskID:    taskID,
							Status:    taskStatus.Phase,
							ExitCode:  taskStatus.ExitCode,
							StartTime: taskStatus.StartedAt,
							EndTime:   taskStatus.FinishedAt,
						}
					}

					return result, nil
				}
				result.Status = status.Phase
			case <-watchCtx.Done():
				return nil, fmt.Errorf("watch timeout exceeded")
			}
		}
	}

	// Non-blocking return
	result.Status = "Submitted"
	return result, nil
}

// Execute executes a pipeline with default options
func (e *ArgoWorkflowExecutor) Execute(ctx context.Context, spec *ir.PipelineSpec) (*ExecutionResult, error) {
	return e.ExecuteWithOptions(ctx, spec, DefaultExecuteOptions())
}

// GetStatus retrieves the status of a workflow
func (e *ArgoWorkflowExecutor) GetStatus(ctx context.Context, workflowName string) (*ExecutionResult, error) {
	status, err := e.client.GetStatus(ctx, workflowName)
	if err != nil {
		return nil, err
	}

	result := &ExecutionResult{
		WorkflowName: workflowName,
		Status:       status.Phase,
		Succeeded:    status.Phase == "Succeeded",
		Failed:       status.Phase == "Failed",
		Message:      status.Message,
		TaskResults:  make(map[string]TaskResult),
		StartTime:    status.StartedAt,
		EndTime:      status.FinishedAt,
	}

	if !result.EndTime.IsZero() && !result.StartTime.IsZero() {
		result.Duration = result.EndTime.Sub(result.StartTime)
	}

	// Collect task results
	for taskID, taskStatus := range status.TaskStatuses {
		result.TaskResults[taskID] = TaskResult{
			TaskID:    taskID,
			Status:    taskStatus.Phase,
			ExitCode:  taskStatus.ExitCode,
			StartTime: taskStatus.StartedAt,
			EndTime:   taskStatus.FinishedAt,
		}
		if !taskStatus.FinishedAt.IsZero() && !taskStatus.StartedAt.IsZero() {
			result.TaskResults[taskID].Duration = taskStatus.FinishedAt.Sub(taskStatus.StartedAt)
		}
	}

	return result, nil
}

// GetLogs retrieves logs for a task
func (e *ArgoWorkflowExecutor) GetLogs(ctx context.Context, workflowName, taskID string) (string, error) {
	return e.client.GetLogs(ctx, workflowName, taskID)
}

// Delete removes a workflow
func (e *ArgoWorkflowExecutor) Delete(ctx context.Context, workflowName string) error {
	return e.client.Delete(ctx, workflowName)
}

// WorkflowStatus represents the status of a workflow
type WorkflowStatus struct {
	Name         string
	Phase        string
	Message      string
	StartedAt    time.Time
	FinishedAt   time.Time
	TaskStatuses map[string]TaskStatus
}

// TaskStatus represents the status of a task
type TaskStatus struct {
	Phase      string
	ExitCode   int
	Message    string
	StartedAt  time.Time
	FinishedAt time.Time
	Logs       string
}

// ExecuteOptions provides options for execution
type ExecuteOptions struct {
	Wait            bool
	Namespace       string
	Parallelism     int32
	TTL             time.Duration
	CompilerOptions CompilerOptions
}

// DefaultExecuteOptions returns default execution options
func DefaultExecuteOptions() ExecuteOptions {
	return ExecuteOptions{
		Wait:        true,
		Namespace:   "default",
		Parallelism: 0,
		TTL:         24 * time.Hour,
		CompilerOptions: CompilerOptions{
			ServiceAccount:  "default",
			ImagePullPolicy: "IfNotPresent",
		},
	}
}

// String returns a JSON representation of the execution result
func (r *ExecutionResult) String() string {
	data, _ := json.MarshalIndent(r, "", "  ")
	return string(data)
}
