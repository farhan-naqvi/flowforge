package compiler

import (
	"context"
	"fmt"

	"flowforge/ir/pkg"
)

// ValidationResult contains validation errors and warnings
type ValidationResult struct {
	Valid    bool
	Errors   []string
	Warnings []string
}

// Validator validates executor-specific output
type Validator struct {
	format ExecutorFormat
}

// NewValidator creates a new validator
func NewValidator(format ExecutorFormat) *Validator {
	return &Validator{format: format}
}

// Validate validates the compiled output
func (v *Validator) Validate(ctx context.Context, result CompileResult) error {
	if result.Format != v.format {
		return fmt.Errorf("format mismatch: expected %s, got %s", v.format, result.Format)
	}

	switch v.format {
	case ExecutorFormatArgo:
		return v.validateArgo(ctx, result)
	case ExecutorFormatAirflow:
		return v.validateAirflow(ctx, result)
	default:
		return fmt.Errorf("unknown format: %s", v.format)
	}
}

// validateArgo validates Argo Workflow YAML
func (v *Validator) validateArgo(ctx context.Context, result CompileResult) error {
	yaml, ok := result.Artifact.(string)
	if !ok {
		return fmt.Errorf("Argo artifact must be a string")
	}

	if len(yaml) == 0 {
		return fmt.Errorf("Argo YAML is empty")
	}

	// Basic YAML structure validation
	if !contains(yaml, "apiVersion") {
		return fmt.Errorf("Argo YAML missing apiVersion")
	}

	if !contains(yaml, "kind") {
		return fmt.Errorf("Argo YAML missing kind")
	}

	if !contains(yaml, "metadata") {
		return fmt.Errorf("Argo YAML missing metadata")
	}

	if !contains(yaml, "spec") {
		return fmt.Errorf("Argo YAML missing spec")
	}

	return nil
}

// validateAirflow validates Airflow DAG Python code
func (v *Validator) validateAirflow(ctx context.Context, result CompileResult) error {
	code, ok := result.Artifact.(string)
	if !ok {
		return fmt.Errorf("Airflow artifact must be a string")
	}

	if len(code) == 0 {
		return fmt.Errorf("Airflow DAG code is empty")
	}

	// Basic Python code validation
	if !contains(code, "from airflow") {
		return fmt.Errorf("Airflow DAG missing airflow imports")
	}

	if !contains(code, "dag =") && !contains(code, "DAG(") {
		return fmt.Errorf("Airflow DAG missing DAG definition")
	}

	return nil
}

// contains checks if a string contains a substring (case-insensitive)
func contains(haystack, needle string) bool {
	for i := 0; i <= len(haystack)-len(needle); i++ {
		if matchCaseInsensitive(haystack[i:i+len(needle)], needle) {
			return true
		}
	}
	return false
}

// matchCaseInsensitive checks if two strings match (case-insensitive)
func matchCaseInsensitive(s1, s2 string) bool {
	if len(s1) != len(s2) {
		return false
	}
	for i := 0; i < len(s1); i++ {
		c1 := s1[i]
		c2 := s2[i]
		if c1 >= 'A' && c1 <= 'Z' {
			c1 = c1 - 'A' + 'a'
		}
		if c2 >= 'A' && c2 <= 'Z' {
			c2 = c2 - 'A' + 'a'
		}
		if c1 != c2 {
			return false
		}
	}
	return true
}

// IRValidator validates IR specifications
type IRValidator struct{}

// NewIRValidator creates a new IR validator
func NewIRValidator() *IRValidator {
	return &IRValidator{}
}

// ValidateSpec validates a PipelineSpec
func (v *IRValidator) ValidateSpec(ctx context.Context, spec *pkg.PipelineSpec) ValidationResult {
	result := ValidationResult{
		Valid:    true,
		Errors:   []string{},
		Warnings: []string{},
	}

	// Check required fields
	if spec.Metadata.Name == "" {
		result.Valid = false
		result.Errors = append(result.Errors, "pipeline name is required")
	}

	if len(spec.Tasks) == 0 {
		result.Valid = false
		result.Errors = append(result.Errors, "pipeline must have at least one task")
	}

	// Check for cycles
	if hasCycle(spec) {
		result.Valid = false
		result.Errors = append(result.Errors, "pipeline contains a cycle")
	}

	// Check edges
	for _, edge := range spec.Edges {
		if _, exists := spec.Tasks[edge.From]; !exists {
			result.Valid = false
			result.Errors = append(result.Errors, fmt.Sprintf("edge from non-existent task: %s", edge.From))
		}
		if _, exists := spec.Tasks[edge.To]; !exists {
			result.Valid = false
			result.Errors = append(result.Errors, fmt.Sprintf("edge to non-existent task: %s", edge.To))
		}
	}

	// Check for unreachable tasks
	unreachable := findUnreachableTasks(spec)
	for _, task := range unreachable {
		result.Warnings = append(result.Warnings, fmt.Sprintf("task %s is unreachable", task))
	}

	// Check tasks
	for taskID, task := range spec.Tasks {
		if task.Handler == nil || task.Handler.Type == "" {
			result.Valid = false
			result.Errors = append(result.Errors, fmt.Sprintf("task %s has no handler type", taskID))
		}

		if len(task.Inputs) > 0 && task.Config.Image == "" {
			result.Warnings = append(result.Warnings, fmt.Sprintf("task %s has no image specified", taskID))
		}
	}

	return result
}
