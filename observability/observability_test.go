package observability

import (
	"context"
	"testing"
	"time"
)

// TestMetricsCollection tests metric collection
func TestMetricsCollection(t *testing.T) {
	mc := NewMockMetricsCollector()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := mc.RecordTaskStart(ctx, "task1"); err != nil {
		t.Fatalf("record start failed: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	if err := mc.RecordTaskEnd(ctx, "task1", 100*time.Millisecond, true); err != nil {
		t.Fatalf("record end failed: %v", err)
	}

	metrics, err := mc.GetMetrics(ctx, nil)
	if err != nil {
		t.Fatalf("get metrics failed: %v", err)
	}

	if len(metrics) != 2 {
		t.Fatalf("expected 2 metrics, got %d", len(metrics))
	}

	t.Logf("Collected %d metrics", len(metrics))
}

// TestLogAggregation tests log aggregation
func TestLogAggregation(t *testing.T) {
	la := NewMockLogAggregator()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := la.WriteLog(ctx, "exec1", "task1", "INFO", "Starting task"); err != nil {
		t.Fatalf("write log failed: %v", err)
	}

	if err := la.WriteLog(ctx, "exec1", "task1", "INFO", "Task completed"); err != nil {
		t.Fatalf("write log failed: %v", err)
	}

	logs, err := la.GetLogs(ctx, "exec1", nil)
	if err != nil {
		t.Fatalf("get logs failed: %v", err)
	}

	if logs == "" {
		t.Fatal("no logs retrieved")
	}

	t.Logf("Aggregated logs: %d bytes", len(logs))
}

// TestLogStreaming tests log streaming
func TestLogStreaming(t *testing.T) {
	la := NewMockLogAggregator()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	la.WriteLog(ctx, "exec1", "task1", "INFO", "Line 1")
	la.WriteLog(ctx, "exec1", "task1", "INFO", "Line 2")
	la.WriteLog(ctx, "exec1", "task1", "INFO", "Line 3")

	ch, err := la.StreamLogs(ctx, "exec1")
	if err != nil {
		t.Fatalf("stream logs failed: %v", err)
	}

	count := 0
	for range ch {
		count++
	}

	if count == 0 {
		t.Fatal("no logs streamed")
	}

	t.Logf("Streamed %d log lines", count)
}

// TestLineageTracking tests data lineage tracking
func TestLineageTracking(t *testing.T) {
	lt := NewMockLineageTracker()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := lt.RecordDataFlow(ctx, "extract", "transform", "data1"); err != nil {
		t.Fatalf("record flow failed: %v", err)
	}

	if err := lt.RecordDataFlow(ctx, "transform", "load", "data1"); err != nil {
		t.Fatalf("record flow failed: %v", err)
	}

	lineage, err := lt.GetLineage(ctx, "data1")
	if err != nil {
		t.Fatalf("get lineage failed: %v", err)
	}

	if lineage == nil || len(lineage.Upstream) == 0 {
		t.Fatal("no lineage data")
	}

	t.Logf("Tracked lineage with %d flows", len(lineage.Upstream))
}

// TestCostTracking tests cost tracking
func TestCostTracking(t *testing.T) {
	ct := NewMockCostTracker()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resources := &ResourceUsage{
		ExecutionID: "exec1",
		TaskID:      "task1",
		MemoryPeak:  512,  // MB
		MemoryAvg:   256,  // MB
		DiskUsed:    1024, // MB
	}

	if err := ct.RecordResourceUsage(ctx, "exec1", resources); err != nil {
		t.Fatalf("record usage failed: %v", err)
	}

	cost, err := ct.GetExecutionCost(ctx, "exec1")
	if err != nil {
		t.Fatalf("get cost failed: %v", err)
	}

	if cost.TotalCost <= 0 {
		t.Fatal("no cost calculated")
	}

	t.Logf("Tracked costs: $%.4f", cost.TotalCost)
}

// TestCostEstimation tests cost estimation
func TestCostEstimation(t *testing.T) {
	ct := NewMockCostTracker()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	estimate, err := ct.EstimateCost(ctx, nil)
	if err != nil {
		t.Fatalf("estimate failed: %v", err)
	}

	if estimate.EstimatedCost <= 0 {
		t.Fatal("no estimate provided")
	}

	t.Logf("Cost estimate: $%.2f (±$%.2f)", estimate.EstimatedCost, estimate.MaxCost-estimate.MinCost)
}

// TestExecutionTracking tests execution tracking
func TestExecutionTracking(t *testing.T) {
	et := NewMockExecutionTracker()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := et.StartExecution(ctx, "exec1", "pipeline1"); err != nil {
		t.Fatalf("start failed: %v", err)
	}

	time.Sleep(50 * time.Millisecond)

	if err := et.EndExecution(ctx, "exec1"); err != nil {
		t.Fatalf("end failed: %v", err)
	}

	status, err := et.GetExecutionStatus(ctx, "exec1")
	if err != nil {
		t.Fatalf("get status failed: %v", err)
	}

	if status.Status != "Completed" {
		t.Fatalf("expected Completed, got %s", status.Status)
	}

	if status.Duration == nil || *status.Duration < 50*time.Millisecond {
		t.Fatal("duration not recorded correctly")
	}

	t.Logf("Execution tracked: %v duration", *status.Duration)
}

// TestExecutionListing tests listing executions
func TestExecutionListing(t *testing.T) {
	et := NewMockExecutionTracker()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create multiple executions
	for i := 1; i <= 3; i++ {
		et.StartExecution(ctx, fmt.Sprintf("exec%d", i), "pipeline1")
		et.EndExecution(ctx, fmt.Sprintf("exec%d", i))
	}

	executions, err := et.ListExecutions(ctx, "pipeline1", 10)
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}

	if len(executions) != 3 {
		t.Fatalf("expected 3 executions, got %d", len(executions))
	}

	t.Logf("Listed %d executions", len(executions))
}

// TestComprehensiveObservability tests full observability engine
func TestComprehensiveObservability(t *testing.T) {
	engine := New(
		NewMockMetricsCollector(),
		NewMockLogAggregator(),
		NewMockLineageTracker(),
		NewMockCostTracker(),
		NewMockExecutionTracker(),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Record a complete execution
	tasks := map[string]*TaskRecord{
		"task1": {
			TaskID:   "task1",
			Duration: 100 * time.Millisecond,
			Success:  true,
			Logs: []*LogLine{
				{Level: "INFO", Message: "Task started"},
				{Level: "INFO", Message: "Task completed"},
			},
			ResourceUsage: &ResourceUsage{
				MemoryPeak: 512,
				DiskUsed:   1024,
			},
		},
	}

	if err := engine.RecordExecution(ctx, "exec1", "pipeline1", tasks); err != nil {
		t.Fatalf("record execution failed: %v", err)
	}

	// Get report
	report, err := engine.GetExecutionReport(ctx, "exec1")
	if err != nil {
		t.Fatalf("get report failed: %v", err)
	}

	if report == nil || report.ExecutionStatus == nil {
		t.Fatal("no execution report")
	}

	t.Logf("Recorded comprehensive execution: %s status", report.ExecutionStatus.Status)
}

import "fmt"
