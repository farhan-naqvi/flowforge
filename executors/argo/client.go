package argo

import (
	"context"
	"fmt"
	"sync"
	"time"

	"flowforge/ir"
)

// MockArgoClient is a mock implementation for testing
type MockArgoClient struct {
	mu              sync.Mutex
	workflows       map[string]*ArgoWorkflow
	statuses        map[string]*WorkflowStatus
	logs            map[string]map[string]string
	submitCallback  func(*ArgoWorkflow) error
	statusCallbacks map[string]func() *WorkflowStatus
}

// NewMockArgoClient creates a new mock client
func NewMockArgoClient() *MockArgoClient {
	return &MockArgoClient{
		workflows:       make(map[string]*ArgoWorkflow),
		statuses:        make(map[string]*WorkflowStatus),
		logs:            make(map[string]map[string]string),
		statusCallbacks: make(map[string]func() *WorkflowStatus),
	}
}

// Submit submits a workflow
func (m *MockArgoClient) Submit(ctx context.Context, workflow *ArgoWorkflow) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.submitCallback != nil {
		if err := m.submitCallback(workflow); err != nil {
			return "", err
		}
	}

	name := workflow.Metadata.Name
	m.workflows[name] = workflow
	m.statuses[name] = &WorkflowStatus{
		Name:      name,
		Phase:     "Pending",
		StartedAt: time.Now(),
	}
	m.logs[name] = make(map[string]string)

	return name, nil
}

// GetStatus returns the status of a workflow
func (m *MockArgoClient) GetStatus(ctx context.Context, workflowName string) (*WorkflowStatus, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if callback, ok := m.statusCallbacks[workflowName]; ok {
		return callback(), nil
	}

	status, ok := m.statuses[workflowName]
	if !ok {
		return nil, fmt.Errorf("workflow not found: %s", workflowName)
	}

	return status, nil
}

// GetLogs returns logs for a task
func (m *MockArgoClient) GetLogs(ctx context.Context, workflowName, podName string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if logs, ok := m.logs[workflowName]; ok {
		if taskLogs, ok := logs[podName]; ok {
			return taskLogs, nil
		}
	}

	return "", nil
}

// Delete removes a workflow
func (m *MockArgoClient) Delete(ctx context.Context, workflowName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.workflows, workflowName)
	delete(m.statuses, workflowName)
	delete(m.logs, workflowName)

	return nil
}

// Watch watches a workflow for status changes
func (m *MockArgoClient) Watch(ctx context.Context, workflowName string) (<-chan WorkflowStatus, error) {
	m.mu.Lock()
	status, ok := m.statuses[workflowName]
	m.mu.Unlock()

	if !ok {
		return nil, fmt.Errorf("workflow not found: %s", workflowName)
	}

	ch := make(chan WorkflowStatus, 10)

	go func() {
		defer close(ch)

		// Simulate workflow progression
		phases := []string{"Pending", "Running", "Succeeded"}
		for i, phase := range phases {
			m.mu.Lock()
			status.Phase = phase
			status.FinishedAt = time.Now()
			m.statuses[workflowName] = status
			m.mu.Unlock()

			select {
			case ch <- *status:
			case <-ctx.Done():
				return
			}

			if phase != "Running" {
				break
			}
			time.Sleep(100 * time.Millisecond)
		}
	}()

	return ch, nil
}

// SetStatusCallback sets a callback for GetStatus
func (m *MockArgoClient) SetStatusCallback(workflowName string, callback func() *WorkflowStatus) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.statusCallbacks[workflowName] = callback
}

// SetTaskLog sets logs for a task
func (m *MockArgoClient) SetTaskLog(workflowName, taskID, log string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.logs[workflowName]; !ok {
		m.logs[workflowName] = make(map[string]string)
	}
	m.logs[workflowName][taskID] = log
}

// SetWorkflowStatus sets the status of a workflow
func (m *MockArgoClient) SetWorkflowStatus(workflowName string, status *WorkflowStatus) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.statuses[workflowName] = status
}

// GetWorkflow returns a submitted workflow
func (m *MockArgoClient) GetWorkflow(name string) *ArgoWorkflow {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.workflows[name]
}

// CompilerOptions provides options for compilation
type CompilerOptions struct {
	ServiceAccount  string
	ImagePullPolicy string
	TTL             int32
}

// RetryPolicy defines retry behavior
type RetryPolicy struct {
	Limit      int32
	Backoff    *Backoff
	RetryOn    string
	Expression string
}

// Backoff defines exponential backoff
type Backoff struct {
	InitialDuration time.Duration
	MaxDuration     time.Duration
	Multiplier      float64
}

// SchedulePolicy defines scheduling
type SchedulePolicy struct {
	Priority          int32
	PriorityClassName string
	NodeSelector      map[string]string
	Affinity          map[string]interface{}
	Toleration        map[string]interface{}
}

// ArtifactS3 defines S3 artifact configuration
type ArtifactS3 struct {
	Bucket string
	Key    string
	Region string
}

// ArtifactConfig defines artifact handling
type ArtifactConfig struct {
	Path string
	S3   *ArtifactS3
	TTL  time.Duration
}

// PodSpecConfig defines pod-level configuration
type PodSpecConfig struct {
	RestartPolicy                 string
	TerminationGracePeriodSeconds int64
	SecurityContext               map[string]interface{}
	ImagePullSecrets              []string
}

// TemplateConfig defines template configuration
type TemplateConfig struct {
	Parallelism    int32
	RetryPolicy    *RetryPolicy
	SchedulePolicy *SchedulePolicy
	ArtifactConfig *ArtifactConfig
	PodSpecConfig  *PodSpecConfig
	Timeout        time.Duration
}

// ExecutionConfig wraps all execution configuration
type ExecutionConfig struct {
	Template TemplateConfig
	Compiler CompilerOptions
}

// ConfigFromPipelineSpec creates execution config from pipeline spec
func ConfigFromPipelineSpec(spec *ir.PipelineSpec) ExecutionConfig {
	cfg := ExecutionConfig{
		Compiler: CompilerOptions{
			ServiceAccount:  "default",
			ImagePullPolicy: "IfNotPresent",
		},
		Template: TemplateConfig{
			Parallelism: 10,
			RetryPolicy: &RetryPolicy{
				Limit: 2,
				Backoff: &Backoff{
					InitialDuration: 1 * time.Second,
					MaxDuration:     60 * time.Second,
					Multiplier:      2.0,
				},
			},
		},
	}

	// Apply spec-level settings if available
	if spec.Metadata["retries"] != nil {
		if retries, ok := spec.Metadata["retries"].(float64); ok {
			cfg.Template.RetryPolicy.Limit = int32(retries)
		}
	}

	return cfg
}

// ApplyConfigToWorkflow applies execution config to workflow
func ApplyConfigToWorkflow(workflow *ArgoWorkflow, cfg ExecutionConfig) {
	workflow.Spec.ServiceAccountName = cfg.Compiler.ServiceAccount

	if cfg.Template.Parallelism > 0 {
		workflow.Spec.Parallelism = cfg.Template.Parallelism
	}

	if cfg.Template.Timeout > 0 {
		workflow.Spec.ActiveDeadlineSeconds = int64(cfg.Template.Timeout.Seconds())
	}
}
