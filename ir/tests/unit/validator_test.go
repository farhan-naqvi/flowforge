package validator

import (
	"flowforge/ir/internal/validator"
	"flowforge/ir/pkg"
	"testing"
)

func TestDAGValidatorNoCycle(t *testing.T) {
	spec := &ir.PipelineSpec{
		APIVersion: "flowforge.io/v1",
		Kind:       "Pipeline",
		Metadata:   ir.PipelineMetadata{Name: "test"},
		Tasks: map[string]*ir.Task{
			"task1": {Type: ir.TaskTypeSource, Handler: ir.Handler{}},
			"task2": {Type: ir.TaskTypeTransform, Handler: ir.Handler{}},
			"task3": {Type: ir.TaskTypeSink, Handler: ir.Handler{}},
		},
		Edges: []ir.Edge{
			{From: ir.TaskPort{Task: "task1", Port: "out"}, To: ir.TaskPort{Task: "task2", Port: "in"}},
			{From: ir.TaskPort{Task: "task2", Port: "out"}, To: ir.TaskPort{Task: "task3", Port: "in"}},
		},
	}

	v := validator.NewDAGValidator()
	if err := v.Validate(spec); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestDAGValidatorWithCycle(t *testing.T) {
	spec := &ir.PipelineSpec{
		APIVersion: "flowforge.io/v1",
		Kind:       "Pipeline",
		Metadata:   ir.PipelineMetadata{Name: "test"},
		Tasks: map[string]*ir.Task{
			"task1": {Type: ir.TaskTypeSource, Handler: ir.Handler{}},
			"task2": {Type: ir.TaskTypeTransform, Handler: ir.Handler{}},
		},
		Edges: []ir.Edge{
			{From: ir.TaskPort{Task: "task1", Port: "out"}, To: ir.TaskPort{Task: "task2", Port: "in"}},
			{From: ir.TaskPort{Task: "task2", Port: "out"}, To: ir.TaskPort{Task: "task1", Port: "in"}},
		},
	}

	v := validator.NewDAGValidator()
	err := v.Validate(spec)
	if err == nil {
		t.Error("expected cycle error, got nil")
	}
	if !stringContains(err.Error(), "cycle") {
		t.Errorf("error should mention cycle: %v", err)
	}
}

func TestSchemaValidatorValid(t *testing.T) {
	spec := &ir.PipelineSpec{
		APIVersion: "flowforge.io/v1",
		Kind:       "Pipeline",
		Metadata:   ir.PipelineMetadata{Name: "test"},
		Tasks: map[string]*ir.Task{
			"task1": {
				Type:    ir.TaskTypeSource,
				Handler: ir.Handler{},
				Outputs: map[string]ir.Schema{"out": {"type": "array"}},
			},
			"task2": {
				Type:    ir.TaskTypeTransform,
				Handler: ir.Handler{},
				Inputs:  map[string]ir.Schema{"in": {"type": "array"}},
				Outputs: map[string]ir.Schema{"out": {"type": "object"}},
			},
		},
		Edges: []ir.Edge{
			{From: ir.TaskPort{Task: "task1", Port: "out"}, To: ir.TaskPort{Task: "task2", Port: "in"}},
		},
	}

	v := validator.NewSchemaValidator()
	if err := v.Validate(spec); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSchemaValidatorMissingPort(t *testing.T) {
	spec := &ir.PipelineSpec{
		APIVersion: "flowforge.io/v1",
		Kind:       "Pipeline",
		Metadata:   ir.PipelineMetadata{Name: "test"},
		Tasks: map[string]*ir.Task{
			"task1": {
				Type:    ir.TaskTypeSource,
				Handler: ir.Handler{},
				Outputs: map[string]ir.Schema{"out": {"type": "array"}},
			},
		},
		Edges: []ir.Edge{
			{From: ir.TaskPort{Task: "task1", Port: "nonexistent"}, To: ir.TaskPort{Task: "task1", Port: "out"}},
		},
	}

	v := validator.NewSchemaValidator()
	err := v.Validate(spec)
	if err == nil {
		t.Error("expected error, got nil")
	}
	if !stringContains(err.Error(), "not found") {
		t.Errorf("error should mention missing port: %v", err)
	}
}

func TestCompositeValidator(t *testing.T) {
	spec := &ir.PipelineSpec{
		APIVersion: "flowforge.io/v1",
		Kind:       "Pipeline",
		Metadata:   ir.PipelineMetadata{Name: "test"},
		Tasks: map[string]*ir.Task{
			"task1": {
				Type:    ir.TaskTypeSource,
				Handler: ir.Handler{},
				Outputs: map[string]ir.Schema{"out": {"type": "array"}},
			},
			"task2": {
				Type:    ir.TaskTypeTransform,
				Handler: ir.Handler{},
				Inputs:  map[string]ir.Schema{"in": {"type": "array"}},
			},
		},
		Edges: []ir.Edge{
			{From: ir.TaskPort{Task: "task1", Port: "out"}, To: ir.TaskPort{Task: "task2", Port: "in"}},
		},
	}

	composite := ir.NewCompositeValidator(
		validator.NewDAGValidator(),
		validator.NewSchemaValidator(),
	)

	if err := composite.Validate(spec); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func stringContains(s, substr string) bool {
	for i := 0; i < len(s)-len(substr)+1; i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
