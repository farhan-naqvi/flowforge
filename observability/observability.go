// observability/observability.go - Observability system for FlowForge

package observability

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// MetricsCollector collects execution metrics
type MetricsCollector interface {
	RecordTaskStart(ctx context.Context, taskID string) error
	RecordTaskEnd(ctx context.Context, taskID string, duration time.Duration, success bool) error
	RecordMetric(ctx context.Context, name string, value float64, tags map[string]string) error
	GetMetrics(ctx context.Context, filter *MetricFilter) ([]*Metric, error)
}

// LogAggregator aggregates logs from various sources
type LogAggregator interface {
	WriteLog(ctx context.Context, executionID, taskID string, level string, message string) error
	GetLogs(ctx context.Context, executionID string, opts *LogOptions) (string, error)
	StreamLogs(ctx context.Context, executionID string) (<-chan string, error)
}

// LineageTracker tracks data lineage
type LineageTracker interface {
	RecordDataFlow(ctx context.Context, from, to string, dataRef string) error
	GetLineage(ctx context.Context, dataRef string) (*DataLineage, error)
	GetDownstream(ctx context.Context, dataRef string) ([]*DataFlow, error)
	GetUpstream(ctx context.Context, dataRef string) ([]*DataFlow, error)
}

// CostTracker tracks execution costs
type CostTracker interface {
	RecordResourceUsage(ctx context.Context, executionID string, resources *ResourceUsage) error
	GetExecutionCost(ctx context.Context, executionID string) (*CostBreakdown, error)
	EstimateCost(ctx context.Context, spec interface{}) (*CostEstimate, error)
}

// Metric represents a single metric
type Metric struct {
	Name      string
	Value     float64
	Timestamp time.Time
	Tags      map[string]string
	Unit      string
}

// MetricFilter filters metrics
type MetricFilter struct {
	Name      string
	StartTime time.Time
	EndTime   time.Time
	Tags      map[string]string
}

// LogOptions filters logs
type LogOptions struct {
	Level     string
	TaskID    string
	StartTime time.Time
	EndTime   time.Time
	Limit     int
}

// DataLineage represents data lineage
type DataLineage struct {
	DataRef    string
	Upstream   []*DataFlow
	Downstream []*DataFlow
	CreatedAt  time.Time
	CreatedBy  string
}

// DataFlow represents a single data flow
type DataFlow struct {
	From      string
	To        string
	DataRef   string
	Timestamp time.Time
}

// ResourceUsage describes resource consumption
type ResourceUsage struct {
	ExecutionID string
	TaskID      string
	CPUTime     time.Duration
	MemoryPeak  int64 // MB
	MemoryAvg   int64 // MB
	DiskUsed    int64 // MB
	GPUTime     time.Duration
	GPUMemory   int64 // MB
	Timestamp   time.Time
}

// CostBreakdown breaks down execution costs
type CostBreakdown struct {
	ExecutionID string
	ComputeCost float64
	StorageCost float64
	NetworkCost float64
	TotalCost   float64
	Currency    string
}

// CostEstimate estimates future costs
type CostEstimate struct {
	EstimatedCost float64
	MinCost       float64
	MaxCost       float64
	Currency      string
	Confidence    float64
}

// ExecutionTracker tracks pipeline executions
type ExecutionTracker interface {
	StartExecution(ctx context.Context, executionID, pipelineID string) error
	EndExecution(ctx context.Context, executionID string) error
	GetExecutionStatus(ctx context.Context, executionID string) (*ExecutionStatus, error)
	ListExecutions(ctx context.Context, pipelineID string, limit int) ([]*ExecutionStatus, error)
}

// ExecutionStatus describes execution status
type ExecutionStatus struct {
	ExecutionID string
	PipelineID  string
	Status      string
	StartedAt   time.Time
	CompletedAt *time.Time
	Duration    *time.Duration
	TaskStatus  map[string]TaskStatus
}

// TaskStatus describes task status
type TaskStatus struct {
	TaskID      string
	Status      string
	StartedAt   time.Time
	CompletedAt *time.Time
	Duration    *time.Duration
	ExitCode    *int
	Error       string
}

// ObservabilityEngine coordinates observability features
type ObservabilityEngine struct {
	metrics    MetricsCollector
	logs       LogAggregator
	lineage    LineageTracker
	costs      CostTracker
	executions ExecutionTracker
}

// New creates a new observability engine
func New(metrics MetricsCollector, logs LogAggregator, lineage LineageTracker, costs CostTracker, executions ExecutionTracker) *ObservabilityEngine {
	return &ObservabilityEngine{
		metrics:    metrics,
		logs:       logs,
		lineage:    lineage,
		costs:      costs,
		executions: executions,
	}
}

// RecordExecution records a full execution with all observability data
func (oe *ObservabilityEngine) RecordExecution(ctx context.Context, executionID, pipelineID string, tasks map[string]*TaskRecord) error {
	// Start execution
	if err := oe.executions.StartExecution(ctx, executionID, pipelineID); err != nil {
		return fmt.Errorf("start execution failed: %w", err)
	}

	// Record task data
	for taskID, task := range tasks {
		// Record start
		if err := oe.metrics.RecordTaskStart(ctx, taskID); err != nil {
			return fmt.Errorf("record task start failed: %w", err)
		}

		// Record metrics
		if task.ResourceUsage != nil {
			if err := oe.costs.RecordResourceUsage(ctx, executionID, task.ResourceUsage); err != nil {
				return fmt.Errorf("record resource usage failed: %w", err)
			}
		}

		// Record logs
		for _, logLine := range task.Logs {
			if err := oe.logs.WriteLog(ctx, executionID, taskID, logLine.Level, logLine.Message); err != nil {
				return fmt.Errorf("write log failed: %w", err)
			}
		}

		// Record lineage
		for _, flow := range task.DataFlows {
			if err := oe.lineage.RecordDataFlow(ctx, flow.From, flow.To, flow.DataRef); err != nil {
				return fmt.Errorf("record data flow failed: %w", err)
			}
		}

		// Record end
		if err := oe.metrics.RecordTaskEnd(ctx, taskID, task.Duration, task.Success); err != nil {
			return fmt.Errorf("record task end failed: %w", err)
		}
	}

	// End execution
	if err := oe.executions.EndExecution(ctx, executionID); err != nil {
		return fmt.Errorf("end execution failed: %w", err)
	}

	return nil
}

// GetExecutionReport gets a complete execution report
func (oe *ObservabilityEngine) GetExecutionReport(ctx context.Context, executionID string) (*ExecutionReport, error) {
	status, err := oe.executions.GetExecutionStatus(ctx, executionID)
	if err != nil {
		return nil, fmt.Errorf("get status failed: %w", err)
	}

	logs, err := oe.logs.GetLogs(ctx, executionID, nil)
	if err != nil {
		return nil, fmt.Errorf("get logs failed: %w", err)
	}

	costs, err := oe.costs.GetExecutionCost(ctx, executionID)
	if err != nil {
		return nil, fmt.Errorf("get costs failed: %w", err)
	}

	return &ExecutionReport{
		ExecutionStatus: status,
		Logs:            logs,
		CostBreakdown:   costs,
		GeneratedAt:     time.Now(),
	}, nil
}

// TaskRecord records task execution details
type TaskRecord struct {
	TaskID        string
	Duration      time.Duration
	Success       bool
	Logs          []*LogLine
	ResourceUsage *ResourceUsage
	DataFlows     []*DataFlow
}

// LogLine represents a log line
type LogLine struct {
	Level   string
	Message string
	Time    time.Time
}

// ExecutionReport is a comprehensive execution report
type ExecutionReport struct {
	ExecutionStatus *ExecutionStatus
	Logs            string
	CostBreakdown   *CostBreakdown
	GeneratedAt     time.Time
}

// Mock implementations

// MockMetricsCollector is a mock metrics collector
type MockMetricsCollector struct {
	mu      sync.RWMutex
	metrics []*Metric
}

// NewMockMetricsCollector creates a mock metrics collector
func NewMockMetricsCollector() *MockMetricsCollector {
	return &MockMetricsCollector{
		metrics: make([]*Metric, 0),
	}
}

// RecordTaskStart records task start
func (m *MockMetricsCollector) RecordTaskStart(ctx context.Context, taskID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.metrics = append(m.metrics, &Metric{
		Name:      "task_start",
		Timestamp: time.Now(),
		Tags:      map[string]string{"task_id": taskID},
	})
	return nil
}

// RecordTaskEnd records task end
func (m *MockMetricsCollector) RecordTaskEnd(ctx context.Context, taskID string, duration time.Duration, success bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	status := "failure"
	if success {
		status = "success"
	}
	m.metrics = append(m.metrics, &Metric{
		Name:      "task_end",
		Value:     duration.Seconds(),
		Timestamp: time.Now(),
		Unit:      "seconds",
		Tags: map[string]string{
			"task_id": taskID,
			"status":  status,
		},
	})
	return nil
}

// RecordMetric records a metric
func (m *MockMetricsCollector) RecordMetric(ctx context.Context, name string, value float64, tags map[string]string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.metrics = append(m.metrics, &Metric{
		Name:      name,
		Value:     value,
		Timestamp: time.Now(),
		Tags:      tags,
	})
	return nil
}

// GetMetrics gets metrics
func (m *MockMetricsCollector) GetMetrics(ctx context.Context, filter *MetricFilter) ([]*Metric, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.metrics, nil
}

// MockLogAggregator is a mock log aggregator
type MockLogAggregator struct {
	mu   sync.RWMutex
	logs map[string][]string
}

// NewMockLogAggregator creates a mock log aggregator
func NewMockLogAggregator() *MockLogAggregator {
	return &MockLogAggregator{
		logs: make(map[string][]string),
	}
}

// WriteLog writes a log
func (m *MockLogAggregator) WriteLog(ctx context.Context, executionID, taskID, level, message string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := fmt.Sprintf("%s/%s", executionID, taskID)
	m.logs[key] = append(m.logs[key], fmt.Sprintf("[%s] %s", level, message))
	return nil
}

// GetLogs gets logs
func (m *MockLogAggregator) GetLogs(ctx context.Context, executionID string, opts *LogOptions) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var logs string
	for key, lines := range m.logs {
		if len(key) > len(executionID) && key[:len(executionID)] == executionID {
			logs += fmt.Sprintf("%s:\n%s\n", key, fmt.Sprint(lines))
		}
	}
	return logs, nil
}

// StreamLogs streams logs
func (m *MockLogAggregator) StreamLogs(ctx context.Context, executionID string) (<-chan string, error) {
	ch := make(chan string, 10)
	go func() {
		defer close(ch)
		m.mu.RLock()
		defer m.mu.RUnlock()
		for key, lines := range m.logs {
			if len(key) > len(executionID) && key[:len(executionID)] == executionID {
				for _, line := range lines {
					ch <- line
				}
			}
		}
	}()
	return ch, nil
}

// MockLineageTracker is a mock lineage tracker
type MockLineageTracker struct {
	mu       sync.RWMutex
	lineages map[string]*DataLineage
}

// NewMockLineageTracker creates a mock lineage tracker
func NewMockLineageTracker() *MockLineageTracker {
	return &MockLineageTracker{
		lineages: make(map[string]*DataLineage),
	}
}

// RecordDataFlow records a data flow
func (m *MockLineageTracker) RecordDataFlow(ctx context.Context, from, to, dataRef string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.lineages[dataRef]; !ok {
		m.lineages[dataRef] = &DataLineage{
			DataRef:    dataRef,
			Upstream:   make([]*DataFlow, 0),
			Downstream: make([]*DataFlow, 0),
			CreatedAt:  time.Now(),
		}
	}
	m.lineages[dataRef].Upstream = append(m.lineages[dataRef].Upstream, &DataFlow{
		From:      from,
		To:        to,
		DataRef:   dataRef,
		Timestamp: time.Now(),
	})
	return nil
}

// GetLineage gets lineage
func (m *MockLineageTracker) GetLineage(ctx context.Context, dataRef string) (*DataLineage, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if lineage, ok := m.lineages[dataRef]; ok {
		return lineage, nil
	}
	return nil, fmt.Errorf("lineage not found")
}

// GetDownstream gets downstream
func (m *MockLineageTracker) GetDownstream(ctx context.Context, dataRef string) ([]*DataFlow, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if lineage, ok := m.lineages[dataRef]; ok {
		return lineage.Downstream, nil
	}
	return nil, fmt.Errorf("lineage not found")
}

// GetUpstream gets upstream
func (m *MockLineageTracker) GetUpstream(ctx context.Context, dataRef string) ([]*DataFlow, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if lineage, ok := m.lineages[dataRef]; ok {
		return lineage.Upstream, nil
	}
	return nil, fmt.Errorf("lineage not found")
}

// MockCostTracker is a mock cost tracker
type MockCostTracker struct {
	mu   sync.RWMutex
	data map[string]*CostBreakdown
}

// NewMockCostTracker creates a mock cost tracker
func NewMockCostTracker() *MockCostTracker {
	return &MockCostTracker{
		data: make(map[string]*CostBreakdown),
	}
}

// RecordResourceUsage records resource usage
func (m *MockCostTracker) RecordResourceUsage(ctx context.Context, executionID string, resources *ResourceUsage) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	// Calculate costs (simplified)
	computeCost := float64(resources.MemoryPeak) * 0.001 // $0.001 per MB
	storageCost := float64(resources.DiskUsed) * 0.0001  // $0.0001 per MB
	m.data[executionID] = &CostBreakdown{
		ExecutionID: executionID,
		ComputeCost: computeCost,
		StorageCost: storageCost,
		TotalCost:   computeCost + storageCost,
		Currency:    "USD",
	}
	return nil
}

// GetExecutionCost gets execution cost
func (m *MockCostTracker) GetExecutionCost(ctx context.Context, executionID string) (*CostBreakdown, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if cost, ok := m.data[executionID]; ok {
		return cost, nil
	}
	return &CostBreakdown{Currency: "USD"}, nil
}

// EstimateCost estimates cost
func (m *MockCostTracker) EstimateCost(ctx context.Context, spec interface{}) (*CostEstimate, error) {
	return &CostEstimate{
		EstimatedCost: 1.5,
		MinCost:       1.0,
		MaxCost:       2.0,
		Currency:      "USD",
		Confidence:    0.8,
	}, nil
}

// MockExecutionTracker is a mock execution tracker
type MockExecutionTracker struct {
	mu         sync.RWMutex
	executions map[string]*ExecutionStatus
}

// NewMockExecutionTracker creates a mock execution tracker
func NewMockExecutionTracker() *MockExecutionTracker {
	return &MockExecutionTracker{
		executions: make(map[string]*ExecutionStatus),
	}
}

// StartExecution starts execution
func (m *MockExecutionTracker) StartExecution(ctx context.Context, executionID, pipelineID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.executions[executionID] = &ExecutionStatus{
		ExecutionID: executionID,
		PipelineID:  pipelineID,
		Status:      "Running",
		StartedAt:   time.Now(),
		TaskStatus:  make(map[string]TaskStatus),
	}
	return nil
}

// EndExecution ends execution
func (m *MockExecutionTracker) EndExecution(ctx context.Context, executionID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if exec, ok := m.executions[executionID]; ok {
		now := time.Now()
		exec.CompletedAt = &now
		duration := now.Sub(exec.StartedAt)
		exec.Duration = &duration
		exec.Status = "Completed"
	}
	return nil
}

// GetExecutionStatus gets execution status
func (m *MockExecutionTracker) GetExecutionStatus(ctx context.Context, executionID string) (*ExecutionStatus, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if exec, ok := m.executions[executionID]; ok {
		return exec, nil
	}
	return nil, fmt.Errorf("execution not found")
}

// ListExecutions lists executions
func (m *MockExecutionTracker) ListExecutions(ctx context.Context, pipelineID string, limit int) ([]*ExecutionStatus, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	executions := make([]*ExecutionStatus, 0)
	count := 0
	for _, exec := range m.executions {
		if exec.PipelineID == pipelineID && count < limit {
			executions = append(executions, exec)
			count++
		}
	}
	return executions, nil
}
