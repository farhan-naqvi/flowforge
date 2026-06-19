package airflow

import (
	"context"
	"strings"
	"testing"

	"flowforge/ir/pkg"
)

func TestAirflowCompileSimple(t *testing.T) {
	spec := &pkg.PipelineSpec{
		Metadata: pkg.Metadata{
			Name:    "test_dag",
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

	compiler := New("test_dag_1")
	result, err := compiler.Compile(context.Background(), spec)
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	code := result.Artifact.(string)
	if !strings.Contains(code, "from airflow") {
		t.Error("Expected code to contain airflow imports")
	}
	if !strings.Contains(code, "dag =") {
		t.Error("Expected code to contain dag definition")
	}
	if !strings.Contains(code, "task1") {
		t.Error("Expected code to contain task1")
	}
}

func TestAirflowCompileWithDependencies(t *testing.T) {
	spec := &pkg.PipelineSpec{
		Metadata: pkg.Metadata{Name: "etl_dag"},
		Tasks: map[string]*pkg.Task{
			"extract": {
				Handler: &pkg.Handler{Type: "python"},
				Config:  &pkg.TaskConfig{Image: "python:3.11"},
			},
			"transform": {
				Handler: &pkg.Handler{Type: "python"},
				Config:  &pkg.TaskConfig{Image: "python:3.11"},
			},
			"load": {
				Handler: &pkg.Handler{Type: "python"},
				Config:  &pkg.TaskConfig{Image: "python:3.11"},
			},
		},
		Edges: []pkg.Edge{
			{From: "extract", To: "transform"},
			{From: "transform", To: "load"},
		},
	}

	compiler := New("etl_dag")
	result, err := compiler.Compile(context.Background(), spec)
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	code := result.Artifact.(string)
	if !strings.Contains(code, "extract >> transform") {
		t.Error("Expected code to contain extract >> transform")
	}
	if !strings.Contains(code, "transform >> load") {
		t.Error("Expected code to contain transform >> load")
	}
}

func TestAirflowGetFormat(t *testing.T) {
	compiler := New("test_dag")
	format := compiler.GetFormat()

	if format != "airflow" {
		t.Errorf("Expected format 'airflow', got %s", format)
	}
}

func TestSanitizeTaskName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"task1", "task1"},
		{"task-1", "task_1"},
		{"task.1", "task_1"},
		{"123task", "t_123task"},
		{"task@2024", "task_2024"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := sanitizeTaskName(tt.input); got != tt.expected {
				t.Errorf("sanitizeTaskName(%s) = %s, want %s", tt.input, got, tt.expected)
			}
		})
	}
}
