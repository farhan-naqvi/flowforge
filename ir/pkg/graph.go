package ir

// Graph is the interface for graph operations on the pipeline.
type Graph interface {
	// Nodes returns all task IDs in the pipeline.
	Nodes() []string

	// Edges returns all edges in the pipeline.
	Edges() []Edge

	// Successors returns tasks that this task feeds into.
	Successors(taskID string) []string

	// Predecessors returns tasks that feed into this task.
	Predecessors(taskID string) []string

	// TopologicalSort returns tasks in topological order.
	// Returns an error if there are cycles.
	TopologicalSort() ([]string, error)

	// HasCycle checks if the graph has cycles.
	HasCycle() bool

	// GetCycle returns a cycle if one exists (for debugging).
	GetCycle() []string

	// String returns a string representation of the graph.
	String() string
}
