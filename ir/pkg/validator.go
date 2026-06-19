package ir

import "fmt"

// Validator is the interface for validating PipelineSpec.
// Validators can be composed and extended.
type Validator interface {
	// Validate checks the PipelineSpec and returns an error if invalid.
	Validate(spec *PipelineSpec) error

	// Name returns the validator name for debugging.
	Name() string
}

// ValidationError wraps validation errors with context.
type ValidationError struct {
	Validator string
	Message   string
	Cause     error
}

// Error implements error interface.
func (ve ValidationError) Error() string {
	if ve.Cause != nil {
		return fmt.Sprintf("[%s] %s: %v", ve.Validator, ve.Message, ve.Cause)
	}
	return fmt.Sprintf("[%s] %s", ve.Validator, ve.Message)
}

// CompositeValidator combines multiple validators.
type CompositeValidator struct {
	validators []Validator
}

// NewCompositeValidator creates a new composite validator.
func NewCompositeValidator(validators ...Validator) Validator {
	return &CompositeValidator{
		validators: validators,
	}
}

// Validate implements Validator.
func (cv *CompositeValidator) Validate(spec *PipelineSpec) error {
	for _, v := range cv.validators {
		if err := v.Validate(spec); err != nil {
			return err
		}
	}
	return nil
}

// Name implements Validator.
func (cv *CompositeValidator) Name() string {
	return "CompositeValidator"
}
