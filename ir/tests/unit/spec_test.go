package ir

import (
	"testing"
)

func TestNewPipelineSpec(t *testing.T) {
	spec := NewPipelineSpec("test-pipeline")
	if spec.Metadata.Name != "test-pipeline" {
		t.Errorf("expected name 'test-pipeline', got %s", spec.Metadata.Name)
	}
	if spec.APIVersion != "flowforge.io/v1" {
		t.Errorf("expected apiVersion 'flowforge.io/v1', got %s", spec.APIVersion)
	}
	if spec.Kind != "Pipeline" {
		t.Errorf("expected kind 'Pipeline', got %s", spec.Kind)
	}
}

func TestBuilderSimplePipeline(t *testing.T) {
	spec, err := NewBuilder("simple-etl").
		SetVersion("1.0.0").
		SetOwner("data_team").
		SetDescription("Simple ETL pipeline").
		AddTag("env", "dev").
		AddTaskWithDescription(
			"extract",
			TaskTypeSource,
			Handler{Type: "python", Source: "extract_data()"},
			"Extract data from S3",
		).
		AddOutput("extract", "records", Schema{"type": "array"}).
		AddTaskWithDescription(
			"transform",
			TaskTypeTransform,
			Handler{Type: "python", Source: "transform_data()"},
			"Transform data",
		).
		AddInput("transform", "data", Schema{"type": "array"}).
		AddOutput("transform", "result", Schema{"type": "array"}).
		AddTaskWithDescription(
			"load",
			TaskTypeSink,
			Handler{Type: "python", Source: "load_data()"},
			"Load to Postgres",
		).
		AddInput("load", "data", Schema{"type": "array"}).
		AddEdge("extract", "records", "transform", "data").
		AddEdge("transform", "result", "load", "data").
		Build()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(spec.Tasks) != 3 {
		t.Errorf("expected 3 tasks, got %d", len(spec.Tasks))
	}
	if len(spec.Edges) != 2 {
		t.Errorf("expected 2 edges, got %d", len(spec.Edges))
	}
	if spec.Metadata.Version != "1.0.0" {
		t.Errorf("expected version '1.0.0', got %s", spec.Metadata.Version)
	}
	if spec.Metadata.Owner != "data_team" {
		t.Errorf("expected owner 'data_team', got %s", spec.Metadata.Owner)
	}
}

func TestBuilderValidation(t *testing.T) {
	tests := []struct {
		name    string
		builder func() Builder
		wantErr bool
		errMsg  string
	}{
		{
			name: "empty task ID",
			builder: func() Builder {
				return NewBuilder("test").
					AddTask("", TaskTypeSource, Handler{Type: "python", Source: "test()"})
			},
			wantErr: true,
			errMsg:  "task ID cannot be empty",
		},
		{
			name: "duplicate task ID",
			builder: func() Builder {
				return NewBuilder("test").
					AddTask("extract", TaskTypeSource, Handler{Type: "python", Source: "test()"}).
					AddTask("extract", TaskTypeSource, Handler{Type: "python", Source: "test()"})
			},
			wantErr: true,
			errMsg:  "already exists",
		},
		{
			name: "edge to non-existent task",
			builder: func() Builder {
				return NewBuilder("test").
					AddTask("extract", TaskTypeSource, Handler{Type: "python", Source: "test()"}).
					AddOutput("extract", "records", Schema{"type": "array"}).
					AddEdge("extract", "records", "non_existent", "data")
			},
			wantErr: true,
			errMsg:  "target task",
		},
		{
			name: "no tasks",
			builder: func() Builder {
				return NewBuilder("empty")
			},
			wantErr: true,
			errMsg:  "at least one task",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.builder().Build()
			if (err != nil) != tt.wantErr {
				t.Errorf("got error %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && err != nil && tt.errMsg != "" {
				if !contains(err.Error(), tt.errMsg) {
					t.Errorf("error message %q doesn't contain %q", err.Error(), tt.errMsg)
				}
			}
		})
	}
}

func TestJSONSerialization(t *testing.T) {
	spec, _ := NewBuilder("test").
		SetVersion("1.0.0").
		AddTask("task1", TaskTypeSource, Handler{Type: "python", Source: "test()"}).
		AddOutput("task1", "out", Schema{"type": "string"}).
		Build()

	// Serialize to JSON
	data, err := spec.ToJSON()
	if err != nil {
		t.Fatalf("failed to serialize: %v", err)
	}

	// Deserialize from JSON
	spec2, err := FromJSON(data)
	if err != nil {
		t.Fatalf("failed to deserialize: %v", err)
	}

	// Verify
	if spec2.Metadata.Name != "test" {
		t.Errorf("name mismatch after roundtrip")
	}
	if spec2.Metadata.Version != "1.0.0" {
		t.Errorf("version mismatch after roundtrip")
	}
	if len(spec2.Tasks) != 1 {
		t.Errorf("tasks count mismatch after roundtrip")
	}
}

func TestSetExecutorConfig(t *testing.T) {
	spec, _ := NewBuilder("test").
		AddTask("task1", TaskTypeSource, Handler{Type: "python", Source: "test()"}).
		SetExecutorConfig("task1", "argo", map[string]interface{}{
			"image":     "python:3.11",
			"resources": map[string]interface{}{"cpu": "2", "memory": "4Gi"},
		}).
		SetExecutorConfig("task1", "airflow", map[string]interface{}{
			"pool": "default",
		}).
		Build()

	task := spec.Tasks["task1"]
	argoConfig, ok := task.ExecutorConfig["argo"]
	if !ok {
		t.Fatalf("argo config not set")
	}

	argoCfg := argoConfig.(map[string]interface{})
	if argoCfg["image"] != "python:3.11" {
		t.Errorf("argo image mismatch")
	}
}

func TestRetryPolicy(t *testing.T) {
	spec, _ := NewBuilder("test").
		AddTask("task1", TaskTypeSource, Handler{Type: "python", Source: "test()"}).
		SetRetryPolicy("task1", &RetryPolicy{
			MaxAttempts:         3,
			Backoff:             "exponential",
			BackoffMultiplier:   2.0,
			InitialDelaySeconds: 5,
		}).
		Build()

	task := spec.Tasks["task1"]
	if task.Retry == nil {
		t.Fatalf("retry policy not set")
	}
	if task.Retry.MaxAttempts != 3 {
		t.Errorf("expected 3 attempts, got %d", task.Retry.MaxAttempts)
	}
}

func TestCostEstimate(t *testing.T) {
	spec, _ := NewBuilder("test").
		AddTask("task1", TaskTypeSource, Handler{Type: "python", Source: "test()"}).
		SetCostEstimate("task1", &CostEstimate{
			Compute: &CostDimension{Unit: "cpu-second", Quantity: 60},
			Storage: &CostDimension{Unit: "GB-month", Quantity: 10},
		}).
		Build()

	task := spec.Tasks["task1"]
	if task.CostEstimate == nil {
		t.Fatalf("cost estimate not set")
	}
	if task.CostEstimate.Compute.Quantity != 60 {
		t.Errorf("expected 60, got %v", task.CostEstimate.Compute.Quantity)
	}
}

func TestSpecValidate(t *testing.T) {
	spec := &PipelineSpec{
		APIVersion: "flowforge.io/v1",
		Kind:       "Pipeline",
		Metadata:   PipelineMetadata{Name: "test"},
		Tasks: map[string]*Task{
			"task1": {Type: TaskTypeSource, Handler: Handler{}},
		},
		Edges: []Edge{},
	}

	if err := spec.Validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSpecValidateFailures(t *testing.T) {
	tests := []struct {
		name    string
		spec    *PipelineSpec
		wantErr string
	}{
		{
			name: "invalid apiVersion",
			spec: &PipelineSpec{
				APIVersion: "wrong",
				Kind:       "Pipeline",
				Metadata:   PipelineMetadata{Name: "test"},
			},
			wantErr: "invalid apiVersion",
		},
		{
			name: "no name",
			spec: &PipelineSpec{
				APIVersion: "flowforge.io/v1",
				Kind:       "Pipeline",
				Metadata:   PipelineMetadata{},
			},
			wantErr: "name is required",
		},
		{
			name: "no tasks",
			spec: &PipelineSpec{
				APIVersion: "flowforge.io/v1",
				Kind:       "Pipeline",
				Metadata:   PipelineMetadata{Name: "test"},
				Tasks:      map[string]*Task{},
			},
			wantErr: "at least one task",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.spec.Validate()
			if err == nil {
				t.Errorf("expected error, got nil")
			}
			if !contains(err.Error(), tt.wantErr) {
				t.Errorf("error %q doesn't contain %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && contains(s, substr) || s == "" || s == substr || (len(s) > len(substr) && contains(s[1:], substr))
}

// Better contains function
func stringContains(s, substr string) bool {
	for i := 0; i < len(s)-len(substr)+1; i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
