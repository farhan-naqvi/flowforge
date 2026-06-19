package ir

import (
	"encoding/json"
	"fmt"
)

// PipelineSpec represents the complete pipeline specification.
// It is immutable and versioned for audit and replay purposes.
type PipelineSpec struct {
	APIVersion string           `json:"apiVersion"`
	Kind       string           `json:"kind"`
	Metadata   PipelineMetadata `json:"metadata"`
	Tasks      map[string]*Task `json:"tasks"`
	Edges      []Edge           `json:"edges"`
}

// PipelineMetadata contains pipeline metadata.
type PipelineMetadata struct {
	Name        string            `json:"name"`
	Version     string            `json:"version,omitempty"`
	Namespace   string            `json:"namespace,omitempty"`
	Owner       string            `json:"owner,omitempty"`
	Description string            `json:"description,omitempty"`
	Tags        map[string]string `json:"tags,omitempty"`
}

// TaskType represents the type of task in the pipeline.
type TaskType string

const (
	TaskTypeSource      TaskType = "Source"
	TaskTypeTransform   TaskType = "Transform"
	TaskTypeSink        TaskType = "Sink"
	TaskTypeConditional TaskType = "Conditional"
	TaskTypeRetry       TaskType = "Retry"
	TaskTypeSchedule    TaskType = "Schedule"
)

// Task represents a node in the pipeline DAG.
type Task struct {
	Type           TaskType               `json:"type"`
	Handler        Handler                `json:"handler"`
	Description    string                 `json:"description,omitempty"`
	Inputs         map[string]Schema      `json:"inputs,omitempty"`
	Outputs        map[string]Schema      `json:"outputs,omitempty"`
	ExecutorConfig map[string]interface{} `json:"executorConfig,omitempty"`
	CostEstimate   *CostEstimate          `json:"costEstimate,omitempty"`
	Retry          *RetryPolicy           `json:"retry,omitempty"`
	Timeout        string                 `json:"timeout,omitempty"`
	Metadata       map[string]string      `json:"metadata,omitempty"`
}

// Handler specifies how a task is executed.
type Handler struct {
	Type   string            `json:"type"` // "python", "sql", "spark", "docker", "http"
	Source string            `json:"source"`
	Env    map[string]string `json:"env,omitempty"`
}

// Schema represents a JSON Schema for task ports (inputs/outputs).
// It can be any valid JSON Schema.
type Schema map[string]interface{}

// Edge represents a data flow connection between task outputs and inputs.
type Edge struct {
	From TaskPort `json:"from"`
	To   TaskPort `json:"to"`
}

// TaskPort represents a specific port on a task (input or output).
type TaskPort struct {
	Task string `json:"task"`
	Port string `json:"port"`
}

// RetryPolicy defines retry behavior.
type RetryPolicy struct {
	MaxAttempts         int     `json:"maxAttempts"`
	Backoff             string  `json:"backoff"` // "linear", "exponential", "fixed"
	BackoffMultiplier   float64 `json:"backoffMultiplier,omitempty"`
	InitialDelaySeconds int     `json:"initialDelaySeconds,omitempty"`
}

// CostEstimate represents the estimated cost of running a task.
type CostEstimate struct {
	Compute *CostDimension `json:"compute,omitempty"`
	Storage *CostDimension `json:"storage,omitempty"`
	Network *CostDimension `json:"network,omitempty"`
}

// CostDimension represents a single cost dimension.
type CostDimension struct {
	Unit     string  `json:"unit"`
	Quantity float64 `json:"quantity"`
}

// NewPipelineSpec creates a new pipeline specification.
func NewPipelineSpec(name string) *PipelineSpec {
	return &PipelineSpec{
		APIVersion: "flowforge.io/v1",
		Kind:       "Pipeline",
		Metadata: PipelineMetadata{
			Name: name,
		},
		Tasks: make(map[string]*Task),
		Edges: []Edge{},
	}
}

// Validate checks if the PipelineSpec is valid.
// It does NOT perform executor-specific validation.
func (ps *PipelineSpec) Validate() error {
	if ps.APIVersion != "flowforge.io/v1" {
		return fmt.Errorf("invalid apiVersion: %s", ps.APIVersion)
	}
	if ps.Kind != "Pipeline" {
		return fmt.Errorf("invalid kind: %s", ps.Kind)
	}
	if ps.Metadata.Name == "" {
		return fmt.Errorf("metadata.name is required")
	}
	if len(ps.Tasks) == 0 {
		return fmt.Errorf("at least one task is required")
	}

	// Validate each task exists and edges reference valid tasks/ports
	for edgeIdx, edge := range ps.Edges {
		fromTask, ok := ps.Tasks[edge.From.Task]
		if !ok {
			return fmt.Errorf("edge %d: source task '%s' not found", edgeIdx, edge.From.Task)
		}
		toTask, ok := ps.Tasks[edge.To.Task]
		if !ok {
			return fmt.Errorf("edge %d: target task '%s' not found", edgeIdx, edge.To.Task)
		}

		// Check ports exist
		if _, ok := fromTask.Outputs[edge.From.Port]; !ok {
			return fmt.Errorf("edge %d: output port '%s' not found on task '%s'", edgeIdx, edge.From.Port, edge.From.Task)
		}
		if _, ok := toTask.Inputs[edge.To.Port]; !ok {
			return fmt.Errorf("edge %d: input port '%s' not found on task '%s'", edgeIdx, edge.To.Port, edge.To.Task)
		}
	}

	return nil
}

// ToJSON serializes the PipelineSpec to JSON.
func (ps *PipelineSpec) ToJSON() ([]byte, error) {
	return json.MarshalIndent(ps, "", "  ")
}

// FromJSON deserializes a PipelineSpec from JSON.
func FromJSON(data []byte) (*PipelineSpec, error) {
	var spec PipelineSpec
	if err := json.Unmarshal(data, &spec); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}
	return &spec, nil
}
