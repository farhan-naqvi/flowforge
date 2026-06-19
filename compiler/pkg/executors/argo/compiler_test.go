package argo

import (
	"context"
	"strings"
	"testing"

	"flowforge/ir/pkg"
)

func TestArgoCompileSimple(t *testing.T) {
	spec := &pkg.PipelineSpec{
		Metadata: pkg.Metadata{
			Name:    "simple_etl",
			Version: "1.0.0",
			Owner:   "test",
		},
		Tasks: map[string]*pkg.Task{
			"task1": {
				Handler: &pkg.Handler{Type: "python"},
				Config:  &pkg.TaskConfig{Image: "python:3.11"},
			},
		},
		Edges: []pkg.Edge{},
	}

	compiler := New("default")
	result, err := compiler.Compile(context.Background(), spec)
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	yaml := result.Artifact.(string)
	if !strings.Contains(yaml, "apiVersion") {
		t.Error("Expected YAML to contain apiVersion")
	}
	if !strings.Contains(yaml, "Workflow") {
		t.Error("Expected YAML to contain Workflow kind")
	}
}

func TestArgoCompileWithEdges(t *testing.T) {
	spec := &pkg.PipelineSpec{
		Metadata: pkg.Metadata{Name: "pipeline"},
		Tasks: map[string]*pkg.Task{
			"extract": {
				Handler: &pkg.Handler{Type: "python"},
				Config:  &pkg.TaskConfig{Image: "python:3.11"},
			},
			"load": {
				Handler: &pkg.Handler{Type: "python"},
				Config:  &pkg.TaskConfig{Image: "python:3.11"},
			},
		},
		Edges: []pkg.Edge{
			{From: "extract", To: "load"},
		},
	}

	compiler := New("default")
	result, err := compiler.Compile(context.Background(), spec)
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	yaml := result.Artifact.(string)
	if !strings.Contains(yaml, "extract") {
		t.Error("Expected YAML to contain extract task")
	}
	if !strings.Contains(yaml, "load") {
		t.Error("Expected YAML to contain load task")
	}
}

func TestArgoGetFormat(t *testing.T) {
	compiler := New("default")
	format := compiler.GetFormat()

	if format != "argo" {
		t.Errorf("Expected format 'argo', got %s", format)
	}
}
