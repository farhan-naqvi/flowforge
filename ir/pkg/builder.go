package ir

import "fmt"

// Builder is the interface for constructing pipelines programmatically.
// Implementations should validate at each step and provide a fluent API.
type Builder interface {
	// SetVersion sets the pipeline version.
	SetVersion(version string) Builder

	// SetOwner sets the pipeline owner.
	SetOwner(owner string) Builder

	// SetDescription sets the pipeline description.
	SetDescription(description string) Builder

	// AddTag adds a tag to the pipeline.
	AddTag(key, value string) Builder

	// AddTask adds a task to the pipeline.
	AddTask(taskID string, taskType TaskType, handler Handler) Builder

	// AddTaskWithDescription adds a task with a description.
	AddTaskWithDescription(taskID string, taskType TaskType, handler Handler, description string) Builder

	// AddInput adds an input port to a task with schema.
	AddInput(taskID string, portName string, schema Schema) Builder

	// AddOutput adds an output port to a task with schema.
	AddOutput(taskID string, portName string, schema Schema) Builder

	// AddEdge connects an output port to an input port.
	AddEdge(fromTask string, fromPort string, toTask string, toPort string) Builder

	// SetExecutorConfig sets executor-specific configuration for a task.
	SetExecutorConfig(taskID string, executor string, config map[string]interface{}) Builder

	// SetRetryPolicy sets the retry policy for a task.
	SetRetryPolicy(taskID string, policy *RetryPolicy) Builder

	// SetTimeout sets the timeout for a task.
	SetTimeout(taskID string, timeout string) Builder

	// SetCostEstimate sets the cost estimate for a task.
	SetCostEstimate(taskID string, estimate *CostEstimate) Builder

	// Build validates and returns the PipelineSpec.
	Build() (*PipelineSpec, error)
}

// BuilderImpl is the default implementation of Builder.
type BuilderImpl struct {
	spec   *PipelineSpec
	errors []error
}

// NewBuilder creates a new pipeline builder.
func NewBuilder(pipelineName string) Builder {
	return &BuilderImpl{
		spec:   NewPipelineSpec(pipelineName),
		errors: []error{},
	}
}

// SetVersion implements Builder.
func (b *BuilderImpl) SetVersion(version string) Builder {
	b.spec.Metadata.Version = version
	return b
}

// SetOwner implements Builder.
func (b *BuilderImpl) SetOwner(owner string) Builder {
	b.spec.Metadata.Owner = owner
	return b
}

// SetDescription implements Builder.
func (b *BuilderImpl) SetDescription(description string) Builder {
	b.spec.Metadata.Description = description
	return b
}

// AddTag implements Builder.
func (b *BuilderImpl) AddTag(key, value string) Builder {
	if b.spec.Metadata.Tags == nil {
		b.spec.Metadata.Tags = make(map[string]string)
	}
	b.spec.Metadata.Tags[key] = value
	return b
}

// AddTask implements Builder.
func (b *BuilderImpl) AddTask(taskID string, taskType TaskType, handler Handler) Builder {
	return b.AddTaskWithDescription(taskID, taskType, handler, "")
}

// AddTaskWithDescription implements Builder.
func (b *BuilderImpl) AddTaskWithDescription(taskID string, taskType TaskType, handler Handler, description string) Builder {
	if taskID == "" {
		b.errors = append(b.errors, fmt.Errorf("task ID cannot be empty"))
		return b
	}
	if _, exists := b.spec.Tasks[taskID]; exists {
		b.errors = append(b.errors, fmt.Errorf("task '%s' already exists", taskID))
		return b
	}

	b.spec.Tasks[taskID] = &Task{
		Type:        taskType,
		Handler:     handler,
		Description: description,
		Inputs:      make(map[string]Schema),
		Outputs:     make(map[string]Schema),
	}
	return b
}

// AddInput implements Builder.
func (b *BuilderImpl) AddInput(taskID string, portName string, schema Schema) Builder {
	task, exists := b.spec.Tasks[taskID]
	if !exists {
		b.errors = append(b.errors, fmt.Errorf("task '%s' does not exist", taskID))
		return b
	}
	task.Inputs[portName] = schema
	return b
}

// AddOutput implements Builder.
func (b *BuilderImpl) AddOutput(taskID string, portName string, schema Schema) Builder {
	task, exists := b.spec.Tasks[taskID]
	if !exists {
		b.errors = append(b.errors, fmt.Errorf("task '%s' does not exist", taskID))
		return b
	}
	task.Outputs[portName] = schema
	return b
}

// AddEdge implements Builder.
func (b *BuilderImpl) AddEdge(fromTask string, fromPort string, toTask string, toPort string) Builder {
	// Validation done at Build() time
	b.spec.Edges = append(b.spec.Edges, Edge{
		From: TaskPort{Task: fromTask, Port: fromPort},
		To:   TaskPort{Task: toTask, Port: toPort},
	})
	return b
}

// SetExecutorConfig implements Builder.
func (b *BuilderImpl) SetExecutorConfig(taskID string, executor string, config map[string]interface{}) Builder {
	task, exists := b.spec.Tasks[taskID]
	if !exists {
		b.errors = append(b.errors, fmt.Errorf("task '%s' does not exist", taskID))
		return b
	}
	if task.ExecutorConfig == nil {
		task.ExecutorConfig = make(map[string]interface{})
	}
	task.ExecutorConfig[executor] = config
	return b
}

// SetRetryPolicy implements Builder.
func (b *BuilderImpl) SetRetryPolicy(taskID string, policy *RetryPolicy) Builder {
	task, exists := b.spec.Tasks[taskID]
	if !exists {
		b.errors = append(b.errors, fmt.Errorf("task '%s' does not exist", taskID))
		return b
	}
	task.Retry = policy
	return b
}

// SetTimeout implements Builder.
func (b *BuilderImpl) SetTimeout(taskID string, timeout string) Builder {
	task, exists := b.spec.Tasks[taskID]
	if !exists {
		b.errors = append(b.errors, fmt.Errorf("task '%s' does not exist", taskID))
		return b
	}
	task.Timeout = timeout
	return b
}

// SetCostEstimate implements Builder.
func (b *BuilderImpl) SetCostEstimate(taskID string, estimate *CostEstimate) Builder {
	task, exists := b.spec.Tasks[taskID]
	if !exists {
		b.errors = append(b.errors, fmt.Errorf("task '%s' does not exist", taskID))
		return b
	}
	task.CostEstimate = estimate
	return b
}

// Build implements Builder.
func (b *BuilderImpl) Build() (*PipelineSpec, error) {
	// Check for builder errors
	if len(b.errors) > 0 {
		return nil, fmt.Errorf("builder errors: %v", b.errors)
	}

	// Validate the spec
	if err := b.spec.Validate(); err != nil {
		return nil, err
	}

	return b.spec, nil
}
