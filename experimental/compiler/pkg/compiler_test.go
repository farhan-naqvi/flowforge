package compiler

import (
	"context"
	"testing"

	"flowforge/ir/pkg"
)

func TestCompilerBasic(t *testing.T) {
	spec := &pkg.PipelineSpec{
		Metadata: pkg.Metadata{
			Name:        "test_pipeline",
			Version:     "1.0.0",
			Owner:       "test_user",
			Description: "Test pipeline",
		},
		Tasks: map[string]*pkg.Task{
			"task1": {
				Handler: &pkg.Handler{Type: "python", Command: "python script.py"},
				Config: &pkg.TaskConfig{
					Image: "python:3.11",
				},
			},
		},
		Edges: []pkg.Edge{},
	}

	compiler := New()
	opts := CompileOptions{
		Format:    ExecutorFormatArgo,
		Namespace: "default",
	}

	result, err := compiler.Compile(context.Background(), spec, opts)
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	if result.Format != ExecutorFormatArgo {
		t.Errorf("Expected format Argo, got %s", result.Format)
	}

	if result.Artifact == nil {
		t.Error("Expected artifact, got nil")
	}
}

func TestCompilerWithEdges(t *testing.T) {
	spec := &pkg.PipelineSpec{
		Metadata: pkg.Metadata{
			Name:    "etl_pipeline",
			Version: "1.0.0",
			Owner:   "data_team",
		},
		Tasks: map[string]*pkg.Task{
			"extract": {
				Handler: &pkg.Handler{Type: "python"},
				Config:  &pkg.TaskConfig{Image: "python:3.11"},
				Outputs: map[string]*pkg.Port{"result": {Schema: map[string]interface{}{"type": "array"}}},
			},
			"transform": {
				Handler: &pkg.Handler{Type: "python"},
				Config:  &pkg.TaskConfig{Image: "python:3.11"},
				Inputs:  map[string]*pkg.Port{"data": {}},
				Outputs: map[string]*pkg.Port{"result": {}},
			},
			"load": {
				Handler: &pkg.Handler{Type: "python"},
				Config:  &pkg.TaskConfig{Image: "python:3.11"},
				Inputs:  map[string]*pkg.Port{"data": {}},
			},
		},
		Edges: []pkg.Edge{
			{From: "extract", FromPort: "result", To: "transform", ToPort: "data"},
			{From: "transform", FromPort: "result", To: "load", ToPort: "data"},
		},
	}

	compiler := New()
	opts := CompileOptions{Format: ExecutorFormatArgo}

	result, err := compiler.Compile(context.Background(), spec, opts)
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	if result.Metadata["task_count"] != 3 {
		t.Errorf("Expected 3 tasks, got %v", result.Metadata["task_count"])
	}

	if result.Metadata["edge_count"] != 2 {
		t.Errorf("Expected 2 edges, got %v", result.Metadata["edge_count"])
	}
}

func TestCompilerAirflow(t *testing.T) {
	spec := &pkg.PipelineSpec{
		Metadata: pkg.Metadata{
			Name:    "airflow_test",
			Version: "1.0.0",
			Owner:   "airflow_user",
		},
		Tasks: map[string]*pkg.Task{
			"task1": {
				Handler: &pkg.Handler{Type: "python"},
				Config:  &pkg.TaskConfig{Image: "python:3.11"},
			},
		},
		Edges: []pkg.Edge{},
	}

	compiler := New()
	opts := CompileOptions{Format: ExecutorFormatAirflow}

	result, err := compiler.Compile(context.Background(), spec, opts)
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	if result.Format != ExecutorFormatAirflow {
		t.Errorf("Expected format Airflow, got %s", result.Format)
	}

	code := result.Artifact.(string)
	if len(code) == 0 {
		t.Error("Expected generated code, got empty")
	}
}

func TestCompilerValidation(t *testing.T) {
	// Empty pipeline
	emptySpec := &pkg.PipelineSpec{
		Metadata: pkg.Metadata{Name: "empty"},
		Tasks:    map[string]*pkg.Task{},
		Edges:    []pkg.Edge{},
	}

	compiler := New()
	opts := CompileOptions{Format: ExecutorFormatArgo}

	_, err := compiler.Compile(context.Background(), emptySpec, opts)
	if err == nil {
		t.Error("Expected validation error for empty pipeline")
	}
}

func TestCompilerInvalidExecutorFormat(t *testing.T) {
	spec := &pkg.PipelineSpec{
		Metadata: pkg.Metadata{Name: "test"},
		Tasks: map[string]*pkg.Task{
			"task1": {
				Handler: &pkg.Handler{Type: "python"},
				Config:  &pkg.TaskConfig{Image: "python:3.11"},
			},
		},
	}

	compiler := New()
	opts := CompileOptions{Format: "unknown_executor"}

	_, err := compiler.Compile(context.Background(), spec, opts)
	if err == nil {
		t.Error("Expected error for unknown executor format")
	}
}

func TestValidateIR(t *testing.T) {
	validator := NewIRValidator()

	// Valid spec
	validSpec := &pkg.PipelineSpec{
		Metadata: pkg.Metadata{Name: "valid"},
		Tasks: map[string]*pkg.Task{
			"task1": {
				Handler: &pkg.Handler{Type: "python"},
				Config:  &pkg.TaskConfig{Image: "python:3.11"},
			},
		},
	}

	result := validator.ValidateSpec(context.Background(), validSpec)
	if !result.Valid {
		t.Errorf("Expected valid spec, got errors: %v", result.Errors)
	}

	// Invalid: missing name
	invalidSpec := &pkg.PipelineSpec{
		Metadata: pkg.Metadata{Name: ""},
		Tasks: map[string]*pkg.Task{
			"task1": {Handler: &pkg.Handler{Type: "python"}, Config: &pkg.TaskConfig{Image: "python:3.11"}},
		},
	}

	result = validator.ValidateSpec(context.Background(), invalidSpec)
	if result.Valid {
		t.Error("Expected invalid spec with missing name")
	}
}

func TestHasCycle(t *testing.T) {
	tests := []struct {
		name string
		spec *pkg.PipelineSpec
		want bool
	}{
		{
			name: "no cycle",
			spec: &pkg.PipelineSpec{
				Tasks: map[string]*pkg.Task{
					"a": {Handler: &pkg.Handler{Type: "python"}, Config: &pkg.TaskConfig{}},
					"b": {Handler: &pkg.Handler{Type: "python"}, Config: &pkg.TaskConfig{}},
				},
				Edges: []pkg.Edge{
					{From: "a", To: "b"},
				},
			},
			want: false,
		},
		{
			name: "has cycle",
			spec: &pkg.PipelineSpec{
				Tasks: map[string]*pkg.Task{
					"a": {Handler: &pkg.Handler{Type: "python"}, Config: &pkg.TaskConfig{}},
					"b": {Handler: &pkg.Handler{Type: "python"}, Config: &pkg.TaskConfig{}},
				},
				Edges: []pkg.Edge{
					{From: "a", To: "b"},
					{From: "b", To: "a"},
				},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hasCycle(tt.spec); got != tt.want {
				t.Errorf("hasCycle() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOptimizer(t *testing.T) {
	spec := &pkg.PipelineSpec{
		Metadata: pkg.Metadata{Name: "opt_test"},
		Tasks: map[string]*pkg.Task{
			"source": {Handler: &pkg.Handler{Type: "python"}, Config: &pkg.TaskConfig{Image: "python:3.11"}},
			"proc_a": {Handler: &pkg.Handler{Type: "python"}, Config: &pkg.TaskConfig{Image: "python:3.11"}},
			"proc_b": {Handler: &pkg.Handler{Type: "python"}, Config: &pkg.TaskConfig{Image: "python:3.11"}},
		},
		Edges: []pkg.Edge{
			{From: "source", To: "proc_a"},
			{From: "source", To: "proc_b"},
		},
	}

	optimizer := NewOptimizer()
	optimized := optimizer.Optimize(context.Background(), spec)

	if optimized == nil {
		t.Error("Expected optimized spec, got nil")
	}

	passes := optimizer.GetPasses()
	if len(passes) == 0 {
		t.Error("Expected optimization passes, got none")
	}

	// Check that parallelization was detected
	parallelFound := false
	for _, pass := range passes {
		if pass.Name == "Parallelization Detection" && pass.Applied {
			parallelFound = true
			break
		}
	}
	if !parallelFound {
		t.Error("Expected parallelization to be detected")
	}
}
