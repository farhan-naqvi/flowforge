package compiler

import (
	"context"
	"fmt"

	"flowforge/ir/pkg"
)

// OptimizationPass represents a single optimization
type OptimizationPass struct {
	Name        string
	Description string
	Applied     bool
	Changes     []string
}

// OptimizerOptions configures optimization behavior
type OptimizerOptions struct {
	EnableParallelization bool
	EnableMerging         bool
	EnableCaching         bool
	ResourcePlanning      bool
}

// Optimizer implements optimization passes
type Optimizer struct {
	options OptimizerOptions
	passes  []OptimizationPass
}

// NewOptimizer creates a new optimizer with default options
func NewOptimizer() *Optimizer {
	return &Optimizer{
		options: OptimizerOptions{
			EnableParallelization: true,
			EnableMerging:         false, // Disabled by default - may change semantics
			EnableCaching:         false, // Future feature
			ResourcePlanning:      true,
		},
		passes: []OptimizationPass{},
	}
}

// Optimize applies all enabled optimizations
func (o *Optimizer) Optimize(ctx context.Context, spec *pkg.PipelineSpec) *pkg.PipelineSpec {
	result := spec

	// Pass 1: Parallelization detection
	if o.options.EnableParallelization {
		result = o.parallelizationPass(ctx, result)
	}

	// Pass 2: Sequential task merging (disabled by default)
	if o.options.EnableMerging {
		result = o.mergingPass(ctx, result)
	}

	// Pass 3: Resource planning
	if o.options.ResourcePlanning {
		result = o.resourcePlanningPass(ctx, result)
	}

	return result
}

// parallelizationPass detects fan-out/fan-in patterns
func (o *Optimizer) parallelizationPass(
	ctx context.Context,
	spec *pkg.PipelineSpec,
) *pkg.PipelineSpec {
	pass := OptimizationPass{
		Name:        "Parallelization Detection",
		Description: "Detect fan-out/fan-in task patterns",
		Applied:     false,
		Changes:     []string{},
	}

	// Analyze out-degree
	outDegree := make(map[string]int)
	inDegree := make(map[string]int)

	for _, edge := range spec.Edges {
		outDegree[edge.From]++
		inDegree[edge.To]++
	}

	// Find fan-out tasks (out-degree > 1)
	for taskID, degree := range outDegree {
		if degree > 1 {
			pass.Applied = true
			pass.Changes = append(pass.Changes,
				fmt.Sprintf("Task %s can execute in parallel (fan-out, out-degree=%d)", taskID, degree))
		}
	}

	// Find fan-in tasks (in-degree > 1)
	for taskID, degree := range inDegree {
		if degree > 1 {
			pass.Applied = true
			pass.Changes = append(pass.Changes,
				fmt.Sprintf("Task %s waits for multiple tasks (fan-in, in-degree=%d)", taskID, degree))
		}
	}

	o.passes = append(o.passes, pass)
	return spec
}

// mergingPass merges sequential tasks (disabled by default)
func (o *Optimizer) mergingPass(
	ctx context.Context,
	spec *pkg.PipelineSpec,
) *pkg.PipelineSpec {
	pass := OptimizationPass{
		Name:        "Sequential Task Merging",
		Description: "Merge sequential single-input single-output tasks",
		Applied:     false,
		Changes:     []string{},
	}

	// Build graph
	inDegree := make(map[string]int)
	outDegree := make(map[string]int)
	outgoing := make(map[string][]string)
	incoming := make(map[string][]string)

	for taskID := range spec.Tasks {
		inDegree[taskID] = 0
		outDegree[taskID] = 0
	}

	for _, edge := range spec.Edges {
		outDegree[edge.From]++
		inDegree[edge.To]++
		outgoing[edge.From] = append(outgoing[edge.From], edge.To)
		incoming[edge.To] = append(incoming[edge.To], edge.From)
	}

	// Find mergeable sequences (linear chains)
	visited := make(map[string]bool)
	for taskID := range spec.Tasks {
		if !visited[taskID] && outDegree[taskID] == 1 && inDegree[taskID] <= 1 {
			// Potential start of a sequence
			next := outgoing[taskID][0]
			if inDegree[next] == 1 && outDegree[next] == 1 {
				pass.Applied = true
				pass.Changes = append(pass.Changes,
					fmt.Sprintf("Could merge sequential tasks: %s → %s", taskID, next))
				visited[taskID] = true
				visited[next] = true
			}
		}
	}

	o.passes = append(o.passes, pass)
	return spec // Return unchanged (don't actually merge without more analysis)
}

// resourcePlanningPass suggests resource configurations
func (o *Optimizer) resourcePlanningPass(
	ctx context.Context,
	spec *pkg.PipelineSpec,
) *pkg.PipelineSpec {
	pass := OptimizationPass{
		Name:        "Resource Planning",
		Description: "Suggest resource configurations",
		Applied:     true,
		Changes:     []string{},
	}

	// Analyze tasks for resource needs
	for taskID, task := range spec.Tasks {
		if task.Config.Resources == nil {
			pass.Changes = append(pass.Changes,
				fmt.Sprintf("Task %s has no resource limits specified", taskID))
		}

		// Suggest based on handler type
		if task.Handler.Type == "python" {
			pass.Changes = append(pass.Changes,
				fmt.Sprintf("Task %s (Python) suggested resources: 1-2 CPU, 512Mi-1Gi memory", taskID))
		}
	}

	o.passes = append(o.passes, pass)
	return spec
}

// GetPasses returns all applied optimization passes
func (o *Optimizer) GetPasses() []OptimizationPass {
	return o.passes
}

// Summary returns a human-readable summary of optimizations
func (o *Optimizer) Summary() string {
	summary := "Optimization Summary:\n"
	for _, pass := range o.passes {
		status := "SKIPPED"
		if pass.Applied {
			status = "APPLIED"
		}
		summary += fmt.Sprintf("  %s [%s]\n", pass.Name, status)
		for _, change := range pass.Changes {
			summary += fmt.Sprintf("    - %s\n", change)
		}
	}
	return summary
}
