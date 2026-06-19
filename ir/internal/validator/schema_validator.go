package validator

import (
	"flowforge/ir/pkg"
	"fmt"
)

// SchemaValidator checks that task input/output schemas are properly connected.
type SchemaValidator struct{}

// NewSchemaValidator creates a new schema validator.
func NewSchemaValidator() ir.Validator {
	return &SchemaValidator{}
}

// Validate implements Validator.
func (sv *SchemaValidator) Validate(spec *ir.PipelineSpec) error {
	// Validate edges have matching schemas
	for idx, edge := range spec.Edges {
		fromTask, ok := spec.Tasks[edge.From.Task]
		if !ok {
			return ir.ValidationError{
				Validator: sv.Name(),
				Message:   fmt.Sprintf("edge %d references non-existent source task: %s", idx, edge.From.Task),
			}
		}

		toTask, ok := spec.Tasks[edge.To.Task]
		if !ok {
			return ir.ValidationError{
				Validator: sv.Name(),
				Message:   fmt.Sprintf("edge %d references non-existent target task: %s", idx, edge.To.Task),
			}
		}

		// Check output exists
		fromSchema, ok := fromTask.Outputs[edge.From.Port]
		if !ok {
			return ir.ValidationError{
				Validator: sv.Name(),
				Message:   fmt.Sprintf("edge %d: output port '%s' not found on task '%s'", idx, edge.From.Port, edge.From.Task),
			}
		}

		// Check input exists
		toSchema, ok := toTask.Inputs[edge.To.Port]
		if !ok {
			return ir.ValidationError{
				Validator: sv.Name(),
				Message:   fmt.Sprintf("edge %d: input port '%s' not found on task '%s'", idx, edge.To.Port, edge.To.Task),
			}
		}

		// TODO: Validate schema compatibility (for now just check they exist)
		_ = fromSchema
		_ = toSchema
	}

	return nil
}

// Name implements Validator.
func (sv *SchemaValidator) Name() string {
	return "SchemaValidator"
}
