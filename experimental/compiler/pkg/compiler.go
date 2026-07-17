package compiler

import (
	"context"
	"fmt"

	"flowforge/ir/pkg"
)

// ExecutorFormat represents the target executor format
type ExecutorFormat string

const (
	ExecutorFormatArgo    ExecutorFormat = "argo"
	ExecutorFormatAirflow ExecutorFormat = "airflow"
)

// CompileResult contains the compiled artifact
type CompileResult struct {
	Format   ExecutorFormat
	Artifact interface{} // YAML string for Argo, Python string for Airflow
	Metadata map[string]interface{}
}

// CompileOptions contains compilation configuration
type CompileOptions struct {
	Format        ExecutorFormat
	Namespace     string                 // For Argo
	ImageRegistry string                 // Override image registry
	Variables     map[string]interface{} // Runtime variables
}

// Compiler manages the compilation pipeline
type Compiler struct {
	optimizer  OptimizationEngine
	validators map[ExecutorFormat]OutputValidator
}

// New creates a new compiler
func New() *Compiler {
	return &Compiler{
		optimizer:  NewOptimizationEngine(),
		validators: make(map[ExecutorFormat]OutputValidator),
	}
}

// Compile orchestrates the compilation pipeline:
// Parse → Validate → Optimize → Compile
func (c *Compiler) Compile(
	ctx context.Context,
	spec *pkg.PipelineSpec,
	opts CompileOptions,
) (CompileResult, error) {
	// Stage 1: Parse (deserialize IR if needed)
	if spec == nil {
		return CompileResult{}, fmt.Errorf("pipeline spec is nil")
	}

	// Stage 2: Validate IR
	if errs := c.validateIR(ctx, spec); len(errs) > 0 {
		return CompileResult{}, fmt.Errorf("validation errors: %v", errs)
	}

	// Stage 3: Optimize
	optimized := c.optimizer.Optimize(ctx, spec)

	// Stage 4: Compile to executor format
	executor, err := c.getExecutor(opts.Format)
	if err != nil {
		return CompileResult{}, err
	}

	result, err := executor.Compile(ctx, optimized)
	if err != nil {
		return CompileResult{}, fmt.Errorf("compilation failed: %w", err)
	}

	// Stage 5: Validate output
	if validator, ok := c.validators[opts.Format]; ok {
		if err := validator.Validate(ctx, result); err != nil {
			return CompileResult{}, fmt.Errorf("output validation failed: %w", err)
		}
	}

	return result, nil
}

// validateIR performs semantic validation on the IR
func (c *Compiler) validateIR(ctx context.Context, spec *pkg.PipelineSpec) []error {
	var errs []error

	// Check basic properties
	if spec.Metadata.Name == "" {
		errs = append(errs, fmt.Errorf("pipeline name is required"))
	}

	if len(spec.Tasks) == 0 {
		errs = append(errs, fmt.Errorf("pipeline must contain at least one task"))
	}

	// Check for cycles (should be done in SDK, but double-check)
	if hasCycle(spec) {
		errs = append(errs, fmt.Errorf("pipeline contains a cycle"))
	}

	// Check for unreachable tasks
	unreachable := findUnreachableTasks(spec)
	for _, task := range unreachable {
		errs = append(errs, fmt.Errorf("task %s is unreachable", task))
	}

	// Check edge validity
	for _, edge := range spec.Edges {
		if _, exists := spec.Tasks[edge.From.Task]; !exists {
			errs = append(errs, fmt.Errorf("edge references non-existent task: %s", edge.From.Task))
		}
		if _, exists := spec.Tasks[edge.To.Task]; !exists {
			errs = append(errs, fmt.Errorf("edge references non-existent task: %s", edge.To.Task))
		}
	}

	return errs
}

// getExecutor returns the appropriate ExecutorCompiler
func (c *Compiler) getExecutor(format ExecutorFormat) (ExecutorCompiler, error) {
	switch format {
	case ExecutorFormatArgo:
		return NewArgoCompiler(), nil
	case ExecutorFormatAirflow:
		return NewAirflowCompiler(), nil
	default:
		return nil, fmt.Errorf("unsupported executor format: %s", format)
	}
}

// hasCycle checks if the DAG contains a cycle using DFS
func hasCycle(spec *pkg.PipelineSpec) bool {
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	var dfs func(taskID string) bool
	dfs = func(taskID string) bool {
		visited[taskID] = true
		recStack[taskID] = true

		// Find all outgoing edges
		for _, edge := range spec.Edges {
			if edge.From.Task == taskID {
				if !visited[edge.To.Task] {
					if dfs(edge.To.Task) {
						return true
					}
				} else if recStack[edge.To.Task] {
					return true
				}
			}
		}

		recStack[taskID] = false
		return false
	}

	for taskID := range spec.Tasks {
		if !visited[taskID] {
			if dfs(taskID) {
				return true
			}
		}
	}

	return false
}

// findUnreachableTasks identifies tasks with no incoming edges (sources) and
// tasks not reachable from any source
func findUnreachableTasks(spec *pkg.PipelineSpec) []string {
	inDegree := make(map[string]int)
	outgoing := make(map[string][]string)

	// Initialize
	for taskID := range spec.Tasks {
		inDegree[taskID] = 0
		outgoing[taskID] = []string{}
	}

	// Build graph
	for _, edge := range spec.Edges {
		inDegree[edge.To.Task]++
		outgoing[edge.From.Task] = append(outgoing[edge.From.Task], edge.To.Task)
	}

	// BFS from all sources
	queue := []string{}
	for taskID, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, taskID)
		}
	}

	visited := make(map[string]bool)
	for len(queue) > 0 {
		taskID := queue[0]
		queue = queue[1:]

		if visited[taskID] {
			continue
		}
		visited[taskID] = true

		for _, next := range outgoing[taskID] {
			if !visited[next] {
				queue = append(queue, next)
			}
		}
	}

	// Find unreachable
	var unreachable []string
	for taskID := range spec.Tasks {
		if !visited[taskID] && inDegree[taskID] > 0 {
			unreachable = append(unreachable, taskID)
		}
	}

	return unreachable
}
