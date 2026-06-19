package graph

import (
	ir "flowforge/ir/pkg"
	"fmt"
	"sort"
)

// DAG implements the Graph interface.
type DAG struct {
	spec *ir.PipelineSpec
	adj  map[string][]string // adjacency list
	radj map[string][]string // reverse adjacency list
}

// NewDAG creates a new DAG from a PipelineSpec.
func NewDAG(spec *ir.PipelineSpec) ir.Graph {
	g := &DAG{
		spec: spec,
		adj:  make(map[string][]string),
		radj: make(map[string][]string),
	}

	// Initialize vertices
	for taskID := range spec.Tasks {
		g.adj[taskID] = []string{}
		g.radj[taskID] = []string{}
	}

	// Add edges
	for _, edge := range spec.Edges {
		g.adj[edge.From.Task] = append(g.adj[edge.From.Task], edge.To.Task)
		g.radj[edge.To.Task] = append(g.radj[edge.To.Task], edge.From.Task)
	}

	// Sort for consistency
	for _, neighbors := range g.adj {
		sort.Strings(neighbors)
	}
	for _, neighbors := range g.radj {
		sort.Strings(neighbors)
	}

	return g
}

// Nodes implements Graph.
func (g *DAG) Nodes() []string {
	nodes := make([]string, 0, len(g.spec.Tasks))
	for id := range g.spec.Tasks {
		nodes = append(nodes, id)
	}
	sort.Strings(nodes)
	return nodes
}

// Edges implements Graph.
func (g *DAG) Edges() []ir.Edge {
	return g.spec.Edges
}

// Successors implements Graph.
func (g *DAG) Successors(taskID string) []string {
	return g.adj[taskID]
}

// Predecessors implements Graph.
func (g *DAG) Predecessors(taskID string) []string {
	return g.radj[taskID]
}

// TopologicalSort implements Graph.
func (g *DAG) TopologicalSort() ([]string, error) {
	// Kahn's algorithm
	inDegree := make(map[string]int)
	for taskID := range g.spec.Tasks {
		inDegree[taskID] = 0
	}

	for _, successors := range g.adj {
		for _, succ := range successors {
			inDegree[succ]++
		}
	}

	queue := make([]string, 0)
	for taskID, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, taskID)
		}
	}
	sort.Strings(queue)

	result := make([]string, 0)
	for len(queue) > 0 {
		taskID := queue[0]
		queue = queue[1:]
		result = append(result, taskID)

		for _, succ := range g.adj[taskID] {
			inDegree[succ]--
			if inDegree[succ] == 0 {
				queue = append(queue, succ)
				sort.Strings(queue)
			}
		}
	}

	if len(result) != len(g.spec.Tasks) {
		cycle := g.GetCycle()
		return nil, fmt.Errorf("cycle detected: %v", cycle)
	}

	return result, nil
}

// HasCycle implements Graph.
func (g *DAG) HasCycle() bool {
	_, err := g.TopologicalSort()
	return err != nil
}

// GetCycle implements Graph.
func (g *DAG) GetCycle() []string {
	visited := make(map[string]bool)
	recStack := make(map[string]bool)
	var path []string

	for taskID := range g.spec.Tasks {
		if !visited[taskID] {
			if g.dfsCycle(taskID, visited, recStack, &path) {
				return path
			}
		}
	}

	return nil
}

// String implements Graph.
func (g *DAG) String() string {
	s := fmt.Sprintf("DAG with %d nodes and %d edges:\n", len(g.spec.Tasks), len(g.spec.Edges))
	for _, edge := range g.spec.Edges {
		s += fmt.Sprintf("  %s.%s -> %s.%s\n", edge.From.Task, edge.From.Port, edge.To.Task, edge.To.Port)
	}
	return s
}

func (g *DAG) dfsCycle(
	node string,
	visited map[string]bool,
	recStack map[string]bool,
	path *[]string,
) bool {
	visited[node] = true
	recStack[node] = true
	*path = append(*path, node)

	for _, neighbor := range g.adj[node] {
		if !visited[neighbor] {
			if g.dfsCycle(neighbor, visited, recStack, path) {
				return true
			}
		} else if recStack[neighbor] {
			// Found cycle, trim path
			for i, n := range *path {
				if n == neighbor {
					*path = (*path)[i:]
					return true
				}
			}
		}
	}

	recStack[node] = false
	*path = (*path)[:len(*path)-1]
	return false
}
