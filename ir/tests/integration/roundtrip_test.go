package integration

import (
	"encoding/json"
	"flowforge/ir/internal/graph"
	"flowforge/ir/internal/validator"
	"flowforge/ir/pkg"
	"io/ioutil"
	"path/filepath"
	"testing"
)

// TestRoundtripJSON tests JSON serialization roundtrip.
func TestRoundtripJSON(t *testing.T) {
	spec, _ := ir.NewBuilder("test").
		SetVersion("1.0.0").
		AddTask("extract", ir.TaskTypeSource, ir.Handler{Type: "python", Source: "test()"}).
		AddOutput("extract", "records", ir.Schema{"type": "array"}).
		AddTask("transform", ir.TaskTypeTransform, ir.Handler{Type: "python", Source: "test()"}).
		AddInput("transform", "data", ir.Schema{"type": "array"}).
		AddOutput("transform", "result", ir.Schema{"type": "array"}).
		AddEdge("extract", "records", "transform", "data").
		Build()

	// Serialize
	data, err := spec.ToJSON()
	if err != nil {
		t.Fatalf("failed to serialize: %v", err)
	}

	// Deserialize
	spec2, err := ir.FromJSON(data)
	if err != nil {
		t.Fatalf("failed to deserialize: %v", err)
	}

	// Verify
	if spec2.Metadata.Name != spec.Metadata.Name {
		t.Errorf("name mismatch: %s != %s", spec2.Metadata.Name, spec.Metadata.Name)
	}
	if len(spec2.Tasks) != len(spec.Tasks) {
		t.Errorf("task count mismatch: %d != %d", len(spec2.Tasks), len(spec.Tasks))
	}
	if len(spec2.Edges) != len(spec.Edges) {
		t.Errorf("edge count mismatch: %d != %d", len(spec2.Edges), len(spec.Edges))
	}
}

// TestLoadExamplePipelines tests loading and validating example pipelines.
func TestLoadExamplePipelines(t *testing.T) {
	examples := []string{
		"simple_etl.json",
		"fan_out_fan_in.json",
		"data_quality.json",
		"scheduled_batch.json",
	}

	validator := ir.NewCompositeValidator(
		validator.NewDAGValidator(),
		validator.NewSchemaValidator(),
	)

	for _, example := range examples {
		t.Run(example, func(t *testing.T) {
			// Find example file
			path := filepath.Join("..", "..", "examples", example)
			data, err := ioutil.ReadFile(path)
			if err != nil {
				t.Skipf("example file not found: %s", path)
			}

			// Parse
			spec, err := ir.FromJSON(data)
			if err != nil {
				t.Fatalf("failed to parse: %v", err)
			}

			// Validate
			if err := validator.Validate(spec); err != nil {
				t.Fatalf("validation failed: %v", err)
			}

			// Check graph
			g := graph.NewDAG(spec)
			if g.HasCycle() {
				t.Fatalf("unexpected cycle in %s", example)
			}

			sorted, err := g.TopologicalSort()
			if err != nil {
				t.Fatalf("topological sort failed: %v", err)
			}
			if len(sorted) != len(spec.Tasks) {
				t.Fatalf("topological sort returned %d tasks, expected %d", len(sorted), len(spec.Tasks))
			}
		})
	}
}

// TestComplexPipeline tests a complex multi-branch pipeline.
func TestComplexPipeline(t *testing.T) {
	// Create a pipeline with multiple branches
	spec, err := ir.NewBuilder("complex").
		SetVersion("1.0.0").
		AddTask("source", ir.TaskTypeSource, ir.Handler{Type: "python", Source: "src()"}).
		AddOutput("source", "data", ir.Schema{"type": "array"}).
		// Branch A
		AddTask("branch_a", ir.TaskTypeTransform, ir.Handler{Type: "python", Source: "a()"}).
		AddInput("branch_a", "data", ir.Schema{"type": "array"}).
		AddOutput("branch_a", "result", ir.Schema{"type": "array"}).
		// Branch B
		AddTask("branch_b", ir.TaskTypeTransform, ir.Handler{Type: "python", Source: "b()"}).
		AddInput("branch_b", "data", ir.Schema{"type": "array"}).
		AddOutput("branch_b", "result", ir.Schema{"type": "array"}).
		// Join
		AddTask("join", ir.TaskTypeTransform, ir.Handler{Type: "python", Source: "join()"}).
		AddInput("join", "left", ir.Schema{"type": "array"}).
		AddInput("join", "right", ir.Schema{"type": "array"}).
		AddOutput("join", "merged", ir.Schema{"type": "array"}).
		// Sink
		AddTask("sink", ir.TaskTypeSink, ir.Handler{Type: "python", Source: "sink()"}).
		AddInput("sink", "data", ir.Schema{"type": "array"}).
		// Edges
		AddEdge("source", "data", "branch_a", "data").
		AddEdge("source", "data", "branch_b", "data").
		AddEdge("branch_a", "result", "join", "left").
		AddEdge("branch_b", "result", "join", "right").
		AddEdge("join", "merged", "sink", "data").
		Build()

	if err != nil {
		t.Fatalf("failed to build: %v", err)
	}

	// Validate
	validator := ir.NewCompositeValidator(
		validator.NewDAGValidator(),
		validator.NewSchemaValidator(),
	)
	if err := validator.Validate(spec); err != nil {
		t.Fatalf("validation failed: %v", err)
	}

	// Check execution order
	g := graph.NewDAG(spec)
	sorted, err := g.TopologicalSort()
	if err != nil {
		t.Fatalf("topological sort failed: %v", err)
	}

	// Source must be first, sink must be last
	if sorted[0] != "source" {
		t.Errorf("source should be first, got %s", sorted[0])
	}
	if sorted[len(sorted)-1] != "sink" {
		t.Errorf("sink should be last, got %s", sorted[len(sorted)-1])
	}

	// Both branches should come before join
	sourceIdx := 0
	joinIdx := 0
	branchAIdx := 0
	branchBIdx := 0
	for i, task := range sorted {
		if task == "source" {
			sourceIdx = i
		} else if task == "join" {
			joinIdx = i
		} else if task == "branch_a" {
			branchAIdx = i
		} else if task == "branch_b" {
			branchBIdx = i
		}
	}

	if sourceIdx >= branchAIdx || sourceIdx >= branchBIdx {
		t.Error("source should come before branches")
	}
	if branchAIdx >= joinIdx || branchBIdx >= joinIdx {
		t.Error("branches should come before join")
	}
}

// TestEdgeCases tests edge cases in spec construction.
func TestEdgeCases(t *testing.T) {
	// Single task pipeline
	spec, err := ir.NewBuilder("single").
		AddTask("only", ir.TaskTypeSource, ir.Handler{Type: "python", Source: "test()"}).
		AddOutput("only", "out", ir.Schema{"type": "string"}).
		Build()

	if err != nil {
		t.Fatalf("single task build failed: %v", err)
	}

	validator := ir.NewCompositeValidator(
		validator.NewDAGValidator(),
		validator.NewSchemaValidator(),
	)
	if err := validator.Validate(spec); err != nil {
		t.Fatalf("validation failed: %v", err)
	}

	// Verify serialization preserves structure
	json, _ := spec.ToJSON()
	spec2, _ := ir.FromJSON(json)

	if len(spec2.Tasks) != 1 {
		t.Errorf("expected 1 task, got %d", len(spec2.Tasks))
	}
}

// TestJSONUnmarshal tests JSON unmarshaling handles edge cases.
func TestJSONUnmarshal(t *testing.T) {
	const sampleJSON = `{
		"apiVersion": "flowforge.io/v1",
		"kind": "Pipeline",
		"metadata": {
			"name": "test",
			"version": "1.0.0"
		},
		"tasks": {
			"task1": {
				"type": "Source",
				"handler": {"type": "python", "source": "test()"},
				"outputs": {"out": {"type": "array"}}
			}
		},
		"edges": []
	}`

	spec, err := ir.FromJSON([]byte(sampleJSON))
	if err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if spec.Metadata.Name != "test" {
		t.Errorf("name mismatch: %s", spec.Metadata.Name)
	}
	if len(spec.Tasks) != 1 {
		t.Errorf("expected 1 task, got %d", len(spec.Tasks))
	}
}
