package deployment

import (
	"context"
	"fmt"
	"sync"
	"time"

	"flowforge/ir"
)

// InMemoryStateManager is an in-memory implementation of StateManager
type InMemoryStateManager struct {
	mu      sync.RWMutex
	states  map[string]*DeploymentState
	history map[string][]*DeploymentState
}

// NewInMemoryStateManager creates a new in-memory state manager
func NewInMemoryStateManager() *InMemoryStateManager {
	return &InMemoryStateManager{
		states:  make(map[string]*DeploymentState),
		history: make(map[string][]*DeploymentState),
	}
}

// SaveState saves a deployment state
func (m *InMemoryStateManager) SaveState(ctx context.Context, state *DeploymentState) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.states[state.PipelineID] = state

	// Keep history
	if _, ok := m.history[state.PipelineID]; !ok {
		m.history[state.PipelineID] = make([]*DeploymentState, 0)
	}

	// Store copy in history
	historyCopy := *state
	m.history[state.PipelineID] = append(m.history[state.PipelineID], &historyCopy)

	// Keep only last 100 entries
	if len(m.history[state.PipelineID]) > 100 {
		m.history[state.PipelineID] = m.history[state.PipelineID][1:]
	}

	return nil
}

// GetState retrieves a deployment state
func (m *InMemoryStateManager) GetState(ctx context.Context, pipelineID string) (*DeploymentState, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	state, ok := m.states[pipelineID]
	if !ok {
		return nil, fmt.Errorf("state not found: %s", pipelineID)
	}

	return state, nil
}

// GetHistory retrieves deployment history
func (m *InMemoryStateManager) GetHistory(ctx context.Context, pipelineID string, limit int) ([]*DeploymentState, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	history, ok := m.history[pipelineID]
	if !ok {
		return nil, fmt.Errorf("history not found: %s", pipelineID)
	}

	// Return most recent entries
	start := 0
	if len(history) > limit {
		start = len(history) - limit
	}

	result := make([]*DeploymentState, len(history[start:]))
	copy(result, history[start:])

	return result, nil
}

// DeleteState deletes a deployment state
func (m *InMemoryStateManager) DeleteState(ctx context.Context, pipelineID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.states, pipelineID)
	delete(m.history, pipelineID)

	return nil
}

// MockPlanGenerator is a mock plan generator
type MockPlanGenerator struct{}

// GeneratePlan generates a mock deployment plan
func (m *MockPlanGenerator) GeneratePlan(ctx context.Context, spec *ir.PipelineSpec) (*DeploymentPlan, error) {
	pipelineID := spec.Metadata["name"].(string)

	plan := &DeploymentPlan{
		PipelineID:      pipelineID,
		Version:         time.Now().Format("20060102150405"),
		Action:          "create",
		Containers:      m.generateContainers(spec),
		TerraformScript: m.generateTerraform(spec),
		HelmChart:       m.generateHelm(spec),
		Estimated: PlanEstimate{
			ResourcesAdded: len(spec.Tasks) + 1,
			EstimatedCost:  float64(len(spec.Tasks)) * 0.05,
			EstimatedTime:  5 * time.Minute,
		},
	}

	return plan, nil
}

// GenerateTerraform generates mock Terraform
func (m *MockPlanGenerator) GenerateTerraform(ctx context.Context, spec *ir.PipelineSpec) (string, error) {
	return `
resource "kubernetes_namespace" "pipeline" {
  metadata {
    name = "flowforge"
  }
}
`, nil
}

// GenerateHelm generates mock Helm values
func (m *MockPlanGenerator) GenerateHelm(ctx context.Context, spec *ir.PipelineSpec) (map[string]interface{}, error) {
	return map[string]interface{}{
		"namespace": "flowforge",
		"replicas":  1,
	}, nil
}

// generateContainers generates container definitions
func (m *MockPlanGenerator) generateContainers(spec *ir.PipelineSpec) []ContainerDefinition {
	containers := make([]ContainerDefinition, 0)

	for taskID, task := range spec.Tasks {
		containers = append(containers, ContainerDefinition{
			Name:     taskID,
			Image:    task.Config.Image,
			Registry: "docker.io",
			Tag:      "latest",
			Size:     1024 * 1024 * 512, // 512MB estimate
		})
	}

	return containers
}

// generateTerraform generates Terraform script
func (m *MockPlanGenerator) generateTerraform(spec *ir.PipelineSpec) string {
	return `resource "kubernetes_namespace" "pipeline" {
  metadata {
    name = "flowforge"
  }
}
`
}

// generateHelm generates Helm configuration
func (m *MockPlanGenerator) generateHelm(spec *ir.PipelineSpec) map[string]interface{} {
	return map[string]interface{}{
		"namespace": "flowforge",
		"replicas":  1,
	}
}

// MockApplier is a mock applier for testing
type MockApplier struct{}

// Apply applies a deployment plan (mock)
func (ma *MockApplier) Apply(ctx context.Context, plan *DeploymentPlan) (*DeploymentResult, error) {
	startTime := time.Now()

	result := &DeploymentResult{
		PipelineID:  plan.PipelineID,
		Version:     plan.Version,
		Status:      "Succeeded",
		Success:     true,
		Created:     plan.Estimated.ResourcesAdded,
		Updated:     0,
		Deleted:     0,
		Failed:      0,
		Duration:    time.Since(startTime),
		StartedAt:   startTime,
		CompletedAt: time.Now(),
		Resources:   make(map[string]string),
		Warnings:    make([]string, 0),
		Errors:      make([]string, 0),
	}

	// Mock resource creation
	for i, container := range plan.Containers {
		result.Resources[container.Name] = fmt.Sprintf("https://k8s-cluster/api/v1/namespaces/flowforge/pods/%s-%d", container.Name, i)
	}

	return result, nil
}

// Destroy destroys deployed resources (mock)
func (ma *MockApplier) Destroy(ctx context.Context, pipelineID string) error {
	return nil
}

// Rollback rolls back to a previous version (mock)
func (ma *MockApplier) Rollback(ctx context.Context, pipelineID, toVersion string) error {
	return nil
}

// MockVersionManager is a mock version manager
type MockVersionManager struct {
	mu       sync.RWMutex
	versions map[string][]string
}

// NewMockVersionManager creates a new mock version manager
func NewMockVersionManager() *MockVersionManager {
	return &MockVersionManager{
		versions: make(map[string][]string),
	}
}

// GetVersion returns the current version
func (m *MockVersionManager) GetVersion(ctx context.Context, pipelineID string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	versions, ok := m.versions[pipelineID]
	if !ok || len(versions) == 0 {
		return "1.0.0", nil
	}

	return versions[len(versions)-1], nil
}

// ListVersions lists all versions
func (m *MockVersionManager) ListVersions(ctx context.Context, pipelineID string) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	versions, ok := m.versions[pipelineID]
	if !ok {
		return []string{}, nil
	}

	return versions, nil
}

// AddVersion adds a new version
func (m *MockVersionManager) AddVersion(pipelineID, version string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.versions[pipelineID]; !ok {
		m.versions[pipelineID] = make([]string, 0)
	}

	m.versions[pipelineID] = append(m.versions[pipelineID], version)
}
