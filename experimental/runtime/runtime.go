// Package runtime provides transformation runtime for FlowForge
package runtime

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// TransformationRuntime manages containerization, versioning, and execution
type TransformationRuntime struct {
	containerMgr ContainerManager
	registry     RegistryClient
	executor     Executor
	versionMgr   VersionManager
	logCollector LogCollector
}

// ContainerManager handles Docker container operations
type ContainerManager interface {
	Build(ctx context.Context, config *ContainerConfig) (*ContainerImage, error)
	Push(ctx context.Context, image *ContainerImage) error
	Run(ctx context.Context, image *ContainerImage, opts *ExecutionOptions) (*ExecutionResult, error)
	GetLogs(ctx context.Context, containerID string) (string, error)
	Stop(ctx context.Context, containerID string) error
}

// RegistryClient manages container registry operations
type RegistryClient interface {
	Push(ctx context.Context, image *ContainerImage) (string, error)
	Pull(ctx context.Context, imageRef string) (*ContainerImage, error)
	List(ctx context.Context, filter string) ([]*ContainerImage, error)
	Delete(ctx context.Context, imageRef string) error
}

// Executor runs containers
type Executor interface {
	Execute(ctx context.Context, image *ContainerImage, opts *ExecutionOptions) (*ExecutionResult, error)
	GetStatus(ctx context.Context, executionID string) (*ExecutionStatus, error)
	Stop(ctx context.Context, executionID string) error
}

// VersionManager tracks container image versions
type VersionManager interface {
	SaveVersion(ctx context.Context, version *ImageVersion) error
	GetVersion(ctx context.Context, functionID, tag string) (*ImageVersion, error)
	ListVersions(ctx context.Context, functionID string) ([]*ImageVersion, error)
	DeleteVersion(ctx context.Context, functionID, tag string) error
}

// LogCollector collects execution logs
type LogCollector interface {
	CollectLogs(ctx context.Context, executionID string) (string, error)
	StreamLogs(ctx context.Context, executionID string) (<-chan string, error)
}

// ContainerConfig describes a container to build
type ContainerConfig struct {
	FunctionID   string
	FunctionName string
	SourceCode   string
	Requirements []string
	BaseImage    string
	BuildArgs    map[string]string
	Environment  map[string]string
	Entrypoint   []string
	WorkDir      string
	ExposePorts  []string
}

// ContainerImage represents a built container image
type ContainerImage struct {
	ID        string
	Name      string
	Tag       string
	Registry  string
	ImageRef  string
	Digest    string
	Size      int64
	BuildTime time.Duration
	CreatedAt time.Time
	Metadata  map[string]string
	Layers    []string
}

// ImageVersion tracks a version of a container image
type ImageVersion struct {
	FunctionID  string
	Tag         string
	Image       *ContainerImage
	SourceCode  string
	CreatedAt   time.Time
	CreatedBy   string
	Description string
	Rollback    bool
}

// ExecutionOptions configures execution behavior
type ExecutionOptions struct {
	Timeout     time.Duration
	Memory      int64 // MB
	CPU         float64
	GPUs        int
	Environment map[string]string
	Volumes     map[string]string
	Stdin       string
	Wait        bool
	Detach      bool
}

// ExecutionResult describes the result of execution
type ExecutionResult struct {
	ExecutionID string
	FunctionID  string
	ContainerID string
	Status      string
	ExitCode    int
	StartedAt   time.Time
	CompletedAt time.Time
	Duration    time.Duration
	StdOut      string
	StdErr      string
	Artifacts   map[string]string
	Logs        []string
}

// ExecutionStatus describes the current status of an execution
type ExecutionStatus struct {
	ExecutionID string
	Status      string
	Progress    int
	Message     string
	UpdatedAt   time.Time
}

// New creates a new transformation runtime
func New(containerMgr ContainerManager, registry RegistryClient, executor Executor, versionMgr VersionManager, logCollector LogCollector) *TransformationRuntime {
	return &TransformationRuntime{
		containerMgr: containerMgr,
		registry:     registry,
		executor:     executor,
		versionMgr:   versionMgr,
		logCollector: logCollector,
	}
}

// ContainerizeFunction builds a container from Python function code
func (tr *TransformationRuntime) ContainerizeFunction(ctx context.Context, functionID, functionName, sourceCode string, requirements []string) (*ContainerImage, error) {
	config := &ContainerConfig{
		FunctionID:   functionID,
		FunctionName: functionName,
		SourceCode:   sourceCode,
		Requirements: requirements,
		BaseImage:    "python:3.11",
		WorkDir:      "/app",
		Entrypoint:   []string{"python", "-m", "flowforge.runtime"},
	}

	return tr.containerMgr.Build(ctx, config)
}

// PushToRegistry pushes a container image to registry
func (tr *TransformationRuntime) PushToRegistry(ctx context.Context, image *ContainerImage) (string, error) {
	imageRef, err := tr.registry.Push(ctx, image)
	if err != nil {
		return "", fmt.Errorf("push failed: %w", err)
	}

	return imageRef, nil
}

// TagVersion creates a versioned tag for an image
func (tr *TransformationRuntime) TagVersion(ctx context.Context, functionID string, image *ContainerImage, sourceCode, description string) (*ImageVersion, error) {
	tag := fmt.Sprintf("v%d-%s", time.Now().Unix(), image.Digest[:12])

	version := &ImageVersion{
		FunctionID:  functionID,
		Tag:         tag,
		Image:       image,
		SourceCode:  sourceCode,
		CreatedAt:   time.Now(),
		Description: description,
	}

	if err := tr.versionMgr.SaveVersion(ctx, version); err != nil {
		return nil, fmt.Errorf("save version failed: %w", err)
	}

	return version, nil
}

// ExecuteFunction executes a transformation function
func (tr *TransformationRuntime) ExecuteFunction(ctx context.Context, functionID, tag string, opts *ExecutionOptions) (*ExecutionResult, error) {
	// Get version
	version, err := tr.versionMgr.GetVersion(ctx, functionID, tag)
	if err != nil {
		return nil, fmt.Errorf("get version failed: %w", err)
	}

	if version == nil {
		return nil, fmt.Errorf("version not found: %s:%s", functionID, tag)
	}

	// Execute
	result, err := tr.executor.Execute(ctx, version.Image, opts)
	if err != nil {
		return nil, fmt.Errorf("execution failed: %w", err)
	}

	// Collect logs
	logs, err := tr.logCollector.CollectLogs(ctx, result.ExecutionID)
	if err == nil {
		result.Logs = []string{logs}
	}

	return result, nil
}

// GetExecutionLogs retrieves logs from an execution
func (tr *TransformationRuntime) GetExecutionLogs(ctx context.Context, executionID string) (string, error) {
	return tr.logCollector.CollectLogs(ctx, executionID)
}

// StreamExecutionLogs streams logs from an ongoing execution
func (tr *TransformationRuntime) StreamExecutionLogs(ctx context.Context, executionID string) (<-chan string, error) {
	return tr.logCollector.StreamLogs(ctx, executionID)
}

// RollbackVersion reverts to a previous container version
func (tr *TransformationRuntime) RollbackVersion(ctx context.Context, functionID, toTag string) (*ImageVersion, error) {
	version, err := tr.versionMgr.GetVersion(ctx, functionID, toTag)
	if err != nil {
		return nil, fmt.Errorf("get version failed: %w", err)
	}

	if version == nil {
		return nil, fmt.Errorf("version not found")
	}

	version.Rollback = true
	if err := tr.versionMgr.SaveVersion(ctx, version); err != nil {
		return nil, fmt.Errorf("save version failed: %w", err)
	}

	return version, nil
}

// ListVersions lists all versions of a function
func (tr *TransformationRuntime) ListVersions(ctx context.Context, functionID string) ([]*ImageVersion, error) {
	return tr.versionMgr.ListVersions(ctx, functionID)
}

// GetVersion retrieves a specific version
func (tr *TransformationRuntime) GetVersion(ctx context.Context, functionID, tag string) (*ImageVersion, error) {
	return tr.versionMgr.GetVersion(ctx, functionID, tag)
}

// StopExecution stops a running execution
func (tr *TransformationRuntime) StopExecution(ctx context.Context, executionID string) error {
	return tr.executor.Stop(ctx, executionID)
}

// GetExecutionStatus retrieves the status of an execution
func (tr *TransformationRuntime) GetExecutionStatus(ctx context.Context, executionID string) (*ExecutionStatus, error) {
	return tr.executor.GetStatus(ctx, executionID)
}

// MockContainerManager is a mock container manager for testing
type MockContainerManager struct {
	mu         sync.RWMutex
	images     map[string]*ContainerImage
	containers map[string]*ExecutionResult
}

// NewMockContainerManager creates a mock container manager
func NewMockContainerManager() *MockContainerManager {
	return &MockContainerManager{
		images:     make(map[string]*ContainerImage),
		containers: make(map[string]*ExecutionResult),
	}
}

// Build builds a container image
func (m *MockContainerManager) Build(ctx context.Context, config *ContainerConfig) (*ContainerImage, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	image := &ContainerImage{
		ID:        fmt.Sprintf("img-%d", time.Now().Unix()),
		Name:      config.FunctionName,
		Tag:       "latest",
		Registry:  "docker.io",
		Digest:    "sha256:abc123def456",
		Size:      1024 * 1024 * 512, // 512MB
		BuildTime: 5 * time.Second,
		CreatedAt: time.Now(),
		Metadata: map[string]string{
			"base_image": config.BaseImage,
		},
	}

	m.images[image.ID] = image
	return image, nil
}

// Push pushes to registry
func (m *MockContainerManager) Push(ctx context.Context, image *ContainerImage) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	image.ImageRef = fmt.Sprintf("%s/%s:%s", image.Registry, image.Name, image.Tag)
	return nil
}

// Run runs a container
func (m *MockContainerManager) Run(ctx context.Context, image *ContainerImage, opts *ExecutionOptions) (*ExecutionResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	result := &ExecutionResult{
		ExecutionID: fmt.Sprintf("exec-%d", time.Now().Unix()),
		ContainerID: fmt.Sprintf("container-%d", time.Now().Unix()),
		Status:      "Succeeded",
		ExitCode:    0,
		StartedAt:   time.Now(),
		CompletedAt: time.Now().Add(5 * time.Second),
		Duration:    5 * time.Second,
		StdOut:      "Function executed successfully",
		StdErr:      "",
	}

	m.containers[result.ExecutionID] = result
	return result, nil
}

// GetLogs retrieves logs
func (m *MockContainerManager) GetLogs(ctx context.Context, containerID string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return fmt.Sprintf("Logs from container %s", containerID), nil
}

// Stop stops a container
func (m *MockContainerManager) Stop(ctx context.Context, containerID string) error {
	return nil
}

// MockRegistryClient is a mock registry client
type MockRegistryClient struct {
	mu     sync.RWMutex
	images map[string]*ContainerImage
}

// NewMockRegistryClient creates a mock registry client
func NewMockRegistryClient() *MockRegistryClient {
	return &MockRegistryClient{
		images: make(map[string]*ContainerImage),
	}
}

// Push pushes to registry
func (m *MockRegistryClient) Push(ctx context.Context, image *ContainerImage) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	ref := fmt.Sprintf("registry.example.com/%s:%s", image.Name, image.Tag)
	m.images[ref] = image
	return ref, nil
}

// Pull pulls from registry
func (m *MockRegistryClient) Pull(ctx context.Context, imageRef string) (*ContainerImage, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if image, ok := m.images[imageRef]; ok {
		return image, nil
	}

	return nil, fmt.Errorf("image not found: %s", imageRef)
}

// List lists images
func (m *MockRegistryClient) List(ctx context.Context, filter string) ([]*ContainerImage, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	images := make([]*ContainerImage, 0)
	for _, img := range m.images {
		images = append(images, img)
	}

	return images, nil
}

// Delete deletes an image
func (m *MockRegistryClient) Delete(ctx context.Context, imageRef string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.images, imageRef)
	return nil
}

// MockExecutor is a mock executor
type MockExecutor struct {
	mu         sync.RWMutex
	executions map[string]*ExecutionStatus
}

// NewMockExecutor creates a mock executor
func NewMockExecutor() *MockExecutor {
	return &MockExecutor{
		executions: make(map[string]*ExecutionStatus),
	}
}

// Execute runs a container
func (m *MockExecutor) Execute(ctx context.Context, image *ContainerImage, opts *ExecutionOptions) (*ExecutionResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	execID := fmt.Sprintf("exec-%d", time.Now().Unix())

	m.executions[execID] = &ExecutionStatus{
		ExecutionID: execID,
		Status:      "Succeeded",
		Progress:    100,
		UpdatedAt:   time.Now(),
	}

	return &ExecutionResult{
		ExecutionID: execID,
		Status:      "Succeeded",
		ExitCode:    0,
		StartedAt:   time.Now(),
		CompletedAt: time.Now().Add(time.Second),
		Duration:    time.Second,
	}, nil
}

// GetStatus gets execution status
func (m *MockExecutor) GetStatus(ctx context.Context, executionID string) (*ExecutionStatus, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if status, ok := m.executions[executionID]; ok {
		return status, nil
	}

	return nil, fmt.Errorf("execution not found: %s", executionID)
}

// Stop stops an execution
func (m *MockExecutor) Stop(ctx context.Context, executionID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if status, ok := m.executions[executionID]; ok {
		status.Status = "Stopped"
	}

	return nil
}

// MockVersionManager is a mock version manager
type MockVersionManager struct {
	mu       sync.RWMutex
	versions map[string][]*ImageVersion
}

// NewMockVersionManager creates a mock version manager
func NewMockVersionManager() *MockVersionManager {
	return &MockVersionManager{
		versions: make(map[string][]*ImageVersion),
	}
}

// SaveVersion saves a version
func (m *MockVersionManager) SaveVersion(ctx context.Context, version *ImageVersion) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.versions[version.FunctionID]; !ok {
		m.versions[version.FunctionID] = make([]*ImageVersion, 0)
	}

	m.versions[version.FunctionID] = append(m.versions[version.FunctionID], version)
	return nil
}

// GetVersion gets a specific version
func (m *MockVersionManager) GetVersion(ctx context.Context, functionID, tag string) (*ImageVersion, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	versions, ok := m.versions[functionID]
	if !ok {
		return nil, fmt.Errorf("function not found: %s", functionID)
	}

	for _, v := range versions {
		if v.Tag == tag {
			return v, nil
		}
	}

	return nil, fmt.Errorf("version not found: %s:%s", functionID, tag)
}

// ListVersions lists all versions
func (m *MockVersionManager) ListVersions(ctx context.Context, functionID string) ([]*ImageVersion, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	versions, ok := m.versions[functionID]
	if !ok {
		return nil, fmt.Errorf("function not found: %s", functionID)
	}

	return versions, nil
}

// DeleteVersion deletes a version
func (m *MockVersionManager) DeleteVersion(ctx context.Context, functionID, tag string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	versions, ok := m.versions[functionID]
	if !ok {
		return fmt.Errorf("function not found: %s", functionID)
	}

	for i, v := range versions {
		if v.Tag == tag {
			m.versions[functionID] = append(versions[:i], versions[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("version not found: %s:%s", functionID, tag)
}

// MockLogCollector is a mock log collector
type MockLogCollector struct{}

// CollectLogs collects logs
func (m *MockLogCollector) CollectLogs(ctx context.Context, executionID string) (string, error) {
	return fmt.Sprintf("Logs from execution: %s\n...", executionID), nil
}

// StreamLogs streams logs
func (m *MockLogCollector) StreamLogs(ctx context.Context, executionID string) (<-chan string, error) {
	ch := make(chan string, 10)
	go func() {
		defer close(ch)
		ch <- fmt.Sprintf("Log line 1 from %s", executionID)
		ch <- "Log line 2"
		ch <- "Log line 3"
	}()
	return ch, nil
}
