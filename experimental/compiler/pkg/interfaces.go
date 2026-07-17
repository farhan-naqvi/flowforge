package compiler

import (
	"context"

	"flowforge/ir/pkg"
)

// ExecutorCompiler is the interface for executor-specific compilers
type ExecutorCompiler interface {
	// Compile converts an optimized PipelineSpec to executor-specific artifact
	Compile(ctx context.Context, spec *pkg.PipelineSpec) (CompileResult, error)

	// GetFormat returns the executor format this compiler targets
	GetFormat() ExecutorFormat
}

// OptimizationEngine performs IR optimization passes
type OptimizationEngine interface {
	// Optimize applies optimization passes to the IR
	Optimize(ctx context.Context, spec *pkg.PipelineSpec) *pkg.PipelineSpec

	// GetOptimizations returns the list of applied optimizations
	GetOptimizations() []string
}

// OutputValidator validates executor-specific output
type OutputValidator interface {
	// Validate checks if the compiled output is valid
	Validate(ctx context.Context, result CompileResult) error
}

// ArgoCompiler compiles IR to Argo Workflows YAML
type ArgoCompiler struct{}

// NewArgoCompiler creates a new Argo compiler
func NewArgoCompiler() ExecutorCompiler {
	return &ArgoCompiler{}
}

// Compile implements ExecutorCompiler
func (c *ArgoCompiler) Compile(
	ctx context.Context,
	spec *pkg.PipelineSpec,
) (CompileResult, error) {
	// TODO: Implement Argo compilation
	return CompileResult{}, nil
}

// GetFormat implements ExecutorCompiler
func (c *ArgoCompiler) GetFormat() ExecutorFormat {
	return ExecutorFormatArgo
}

// AirflowCompiler compiles IR to Apache Airflow Python DAG
type AirflowCompiler struct{}

// NewAirflowCompiler creates a new Airflow compiler
func NewAirflowCompiler() ExecutorCompiler {
	return &AirflowCompiler{}
}

// Compile implements ExecutorCompiler
func (c *AirflowCompiler) Compile(
	ctx context.Context,
	spec *pkg.PipelineSpec,
) (CompileResult, error) {
	// TODO: Implement Airflow compilation
	return CompileResult{}, nil
}

// GetFormat implements ExecutorCompiler
func (c *AirflowCompiler) GetFormat() ExecutorFormat {
	return ExecutorFormatAirflow
}

// OptimizationEngineImpl is the default optimization engine
type OptimizationEngineImpl struct {
	optimizations []string
}

// NewOptimizationEngine creates a new optimization engine
func NewOptimizationEngine() OptimizationEngine {
	return &OptimizationEngineImpl{}
}

// Optimize implements OptimizationEngine
func (e *OptimizationEngineImpl) Optimize(
	ctx context.Context,
	spec *pkg.PipelineSpec,
) *pkg.PipelineSpec {
	e.optimizations = []string{}

	// Apply optimization passes
	result := spec

	// 1. Detect parallelizable tasks
	if canParallelize(result) {
		e.optimizations = append(e.optimizations, "parallelization")
	}

	// 2. Merge sequential tasks (future)
	// result = mergeSequentialTasks(result)
	// e.optimizations = append(e.optimizations, "sequential_merge")

	return result
}

// GetOptimizations implements OptimizationEngine
func (e *OptimizationEngineImpl) GetOptimizations() []string {
	return e.optimizations
}

// canParallelize checks if pipeline contains parallelizable tasks (fan-out/fan-in)
func canParallelize(spec *pkg.PipelineSpec) bool {
	// Check for tasks with multiple outgoing edges
	outDegree := make(map[string]int)
	for _, edge := range spec.Edges {
		outDegree[edge.From]++
	}

	for _, degree := range outDegree {
		if degree > 1 {
			return true
		}
	}

	return false
}
