package validator

import (
	"flowforge/ir/pkg"
	"fmt"
)

// DAGValidator checks for cycles and validates pipeline topology.
type DAGValidator struct{}

// NewDAGValidator creates a new DAG validator.
func NewDAGValidator() ir.Validator {
	return &DAGValidator{}
}

// Validate implements Validator.
func (dv *DAGValidator) Validate(spec *ir.PipelineSpec) error {
	// Build adjacency list
	graph := make(map[string][]string)
	for taskID := range spec.Tasks {
		graph[taskID] = []string{}
	}

	for _, edge := range spec.Edges {
		graph[edge.From.Task] = append(graph[edge.From.Task], edge.To.Task)
	}

	// Check for cycles using DFS
	visited := make(map[string]bool)
	recStack := make(map[string]bool)
	var cycle []string

	for taskID := range graph {
		if !visited[taskID] {
			if dv.hasCycleDFS(taskID, graph, visited, recStack, &cycle) {
				return ir.ValidationError{
					Validator: dv.Name(),
					Message:   fmt.Sprintf("cycle detected: %v", cycle),
				}
			}
		}
	}

	return nil
}

// Name implements Validator.
func (dv *DAGValidator) Name() string {
	return "DAGValidator"
}

func (dv *DAGValidator) hasCycleDFS(
	node string,
	graph map[string][]string,
	visited map[string]bool,
	recStack map[string]bool,
	cycle *[]string,
) bool {
	visited[node] = true
	recStack[node] = true
	*cycle = append(*cycle, node)

	for _, neighbor := range graph[node] {
		if !visited[neighbor] {
			if dv.hasCycleDFS(neighbor, graph, visited, recStack, cycle) {
				return true
			}
		} else if recStack[neighbor] {
			// Found cycle
			return true
		}
	}

	recStack[node] = false
	*cycle = (*cycle)[:len(*cycle)-1]
	return false
}
