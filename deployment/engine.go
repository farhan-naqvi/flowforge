// Package deployment provides infrastructure-as-code generation and management
package deployment

import (
	"context"
	"fmt"
	"time"

	"flowforge/ir"
)

// DeploymentEngine orchestrates infrastructure deployment from IR specifications
type DeploymentEngine struct {
	stateManager StateManager
	planner      PlanGenerator
	applier      Applier
	versionMgr   VersionManager
}

// StateManager tracks deployment state and history
type StateManager interface {
	SaveState(ctx context.Context, state *DeploymentState) error
	GetState(ctx context.Context, pipelineID string) (*DeploymentState, error)
	GetHistory(ctx context.Context, pipelineID string, limit int) ([]*DeploymentState, error)
	DeleteState(ctx context.Context, pipelineID string) error
}

// PlanGenerator creates deployment plans
type PlanGenerator interface {
	GeneratePlan(ctx context.Context, spec *ir.PipelineSpec) (*DeploymentPlan, error)
	GenerateTerraform(ctx context.Context, spec *ir.PipelineSpec) (string, error)
	GenerateHelm(ctx context.Context, spec *ir.PipelineSpec) (map[string]interface{}, error)
}

// Applier applies deployment plans
type Applier interface {
	Apply(ctx context.Context, plan *DeploymentPlan) (*DeploymentResult, error)
	Destroy(ctx context.Context, pipelineID string) error
	Rollback(ctx context.Context, pipelineID, toVersion string) error
}

// VersionManager manages deployment versions
type VersionManager interface {
	GetVersion(ctx context.Context, pipelineID string) (string, error)
	ListVersions(ctx context.Context, pipelineID string) ([]string, error)
}

// DeploymentState represents the current state of a deployment
type DeploymentState struct {
	PipelineID     string
	Version        string
	Namespace      string
	Status         string
	Phase          string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	LastApplied    time.Time
	Spec           *ir.PipelineSpec
	Resources      map[string]interface{}
	TerraformState map[string]interface{}
	HelmValues     map[string]interface{}
	Conditions     []Condition
}

// Condition describes the state of a deployment aspect
type Condition struct {
	Type               string
	Status             string
	LastTransitionTime time.Time
	Reason             string
	Message            string
}

// DeploymentPlan describes what will be deployed
type DeploymentPlan struct {
	PipelineID      string
	Version         string
	Action          string // "create", "update", "delete"
	Containers      []ContainerDefinition
	TerraformScript string
	HelmChart       map[string]interface{}
	Estimated       PlanEstimate
	DryRun          bool
}

// ContainerDefinition describes a container image
type ContainerDefinition struct {
	Name     string
	Image    string
	Registry string
	Tag      string
	Digest   string
	Size     int64
	Layers   []string
}

// PlanEstimate provides estimated impact
type PlanEstimate struct {
	ResourcesAdded    int
	ResourcesModified int
	ResourcesDeleted  int
	EstimatedCost     float64
	EstimatedTime     time.Duration
}

// DeploymentResult describes the result of an apply operation
type DeploymentResult struct {
	PipelineID  string
	Version     string
	Status      string
	Success     bool
	Created     int
	Updated     int
	Deleted     int
	Failed      int
	Duration    time.Duration
	StartedAt   time.Time
	CompletedAt time.Time
	Resources   map[string]string // resource ID -> URL
	Warnings    []string
	Errors      []string
}

// New creates a new deployment engine
func New(stateMgr StateManager, planner PlanGenerator, applier Applier, versionMgr VersionManager) *DeploymentEngine {
	return &DeploymentEngine{
		stateManager: stateMgr,
		planner:      planner,
		applier:      applier,
		versionMgr:   versionMgr,
	}
}

// Plan creates a deployment plan
func (e *DeploymentEngine) Plan(ctx context.Context, spec *ir.PipelineSpec) (*DeploymentPlan, error) {
	plan, err := e.planner.GeneratePlan(ctx, spec)
	if err != nil {
		return nil, fmt.Errorf("plan generation failed: %w", err)
	}

	return plan, nil
}

// Apply applies a deployment plan
func (e *DeploymentEngine) Apply(ctx context.Context, spec *ir.PipelineSpec) (*DeploymentResult, error) {
	// Validate spec
	if spec == nil || spec.Metadata == nil {
		return nil, fmt.Errorf("invalid specification")
	}

	pipelineID, ok := spec.Metadata["name"].(string)
	if !ok || pipelineID == "" {
		return nil, fmt.Errorf("pipeline name required")
	}

	// Generate plan
	plan, err := e.Plan(ctx, spec)
	if err != nil {
		return nil, err
	}

	// Apply plan
	result, err := e.applier.Apply(ctx, plan)
	if err != nil {
		return nil, fmt.Errorf("apply failed: %w", err)
	}

	// Save state
	state := &DeploymentState{
		PipelineID:  pipelineID,
		Version:     result.Version,
		Status:      result.Status,
		Phase:       "Applied",
		CreatedAt:   result.StartedAt,
		UpdatedAt:   time.Now(),
		LastApplied: time.Now(),
		Spec:        spec,
		Resources:   result.Resources,
		Conditions: []Condition{
			{
				Type:   "Applied",
				Status: "True",
				Reason: "DeploymentSucceeded",
			},
		},
	}

	if err := e.stateManager.SaveState(ctx, state); err != nil {
		return nil, fmt.Errorf("failed to save state: %w", err)
	}

	return result, nil
}

// Destroy removes deployed resources
func (e *DeploymentEngine) Destroy(ctx context.Context, pipelineID string) error {
	// Get current state
	state, err := e.stateManager.GetState(ctx, pipelineID)
	if err != nil {
		return fmt.Errorf("get state failed: %w", err)
	}

	if state == nil {
		return fmt.Errorf("no state found for pipeline: %s", pipelineID)
	}

	// Destroy resources
	if err := e.applier.Destroy(ctx, pipelineID); err != nil {
		return fmt.Errorf("destroy failed: %w", err)
	}

	// Update state
	state.Status = "Destroyed"
	state.Phase = "Destroyed"
	state.UpdatedAt = time.Now()
	state.Conditions = append(state.Conditions, Condition{
		Type:   "Destroyed",
		Status: "True",
		Reason: "DestroyOperationCompleted",
	})

	if err := e.stateManager.SaveState(ctx, state); err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}

	return nil
}

// Rollback reverts to a previous version
func (e *DeploymentEngine) Rollback(ctx context.Context, pipelineID, toVersion string) error {
	// Get target state
	state, err := e.stateManager.GetState(ctx, pipelineID)
	if err != nil {
		return fmt.Errorf("get state failed: %w", err)
	}

	if state == nil {
		return fmt.Errorf("no state found for pipeline: %s", pipelineID)
	}

	// Perform rollback
	if err := e.applier.Rollback(ctx, pipelineID, toVersion); err != nil {
		return fmt.Errorf("rollback failed: %w", err)
	}

	// Update state
	state.Version = toVersion
	state.Status = "Rolled Back"
	state.UpdatedAt = time.Now()
	state.Conditions = append(state.Conditions, Condition{
		Type:    "RolledBack",
		Status:  "True",
		Reason:  "RollbackOperationCompleted",
		Message: fmt.Sprintf("Rolled back to version %s", toVersion),
	})

	if err := e.stateManager.SaveState(ctx, state); err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}

	return nil
}

// GetStatus retrieves deployment status
func (e *DeploymentEngine) GetStatus(ctx context.Context, pipelineID string) (*DeploymentState, error) {
	return e.stateManager.GetState(ctx, pipelineID)
}

// GetHistory retrieves deployment history
func (e *DeploymentEngine) GetHistory(ctx context.Context, pipelineID string, limit int) ([]*DeploymentState, error) {
	return e.stateManager.GetHistory(ctx, pipelineID, limit)
}

// DryRun performs a dry-run deployment
func (e *DeploymentEngine) DryRun(ctx context.Context, spec *ir.PipelineSpec) (*DeploymentPlan, error) {
	plan, err := e.Plan(ctx, spec)
	if err != nil {
		return nil, err
	}

	plan.DryRun = true
	return plan, nil
}

// GetVersion returns the current version of a pipeline deployment
func (e *DeploymentEngine) GetVersion(ctx context.Context, pipelineID string) (string, error) {
	return e.versionMgr.GetVersion(ctx, pipelineID)
}

// ListVersions lists all versions of a pipeline deployment
func (e *DeploymentEngine) ListVersions(ctx context.Context, pipelineID string) ([]string, error) {
	return e.versionMgr.ListVersions(ctx, pipelineID)
}
