package metrics

import (
	"testing"
	"time"
)

func TestMetrics(t *testing.T) {
	metrics := NewMetrics()

	// Test task count
	metrics.IncrementTaskCount()
	if metrics.GetTaskCount() != 1 {
		t.Errorf("Expected task count 1, got %d", metrics.GetTaskCount())
	}

	metrics.DecrementTaskCount()
	if metrics.GetTaskCount() != 0 {
		t.Errorf("Expected task count 0, got %d", metrics.GetTaskCount())
	}

	// Test completed task count
	metrics.IncrementCompletedTaskCount()
	if metrics.GetCompletedTaskCount() != 1 {
		t.Errorf("Expected completed task count 1, got %d", metrics.GetCompletedTaskCount())
	}

	metrics.DecrementCompletedTaskCount()
	if metrics.GetCompletedTaskCount() != 0 {
		t.Errorf("Expected completed task count 0, got %d", metrics.GetCompletedTaskCount())
	}

	// Test backup count
	metrics.IncrementBackupCount()
	if metrics.GetBackupCount() != 1 {
		t.Errorf("Expected backup count 1, got %d", metrics.GetBackupCount())
	}

	metrics.DecrementBackupCount()
	if metrics.GetBackupCount() != 0 {
		t.Errorf("Expected backup count 0, got %d", metrics.GetBackupCount())
	}

	// Test operation recording
	operationType := "test_operation"
	latency := 100 * time.Millisecond
	metrics.RecordOperation(operationType, latency)

	if metrics.GetOperationCount(operationType) != 1 {
		t.Errorf("Expected operation count 1, got %d", metrics.GetOperationCount(operationType))
	}

	latencies := metrics.GetOperationLatencies(operationType)
	if len(latencies) != 1 {
		t.Errorf("Expected 1 latency, got %d", len(latencies))
	}
	if latencies[0] != latency {
		t.Errorf("Expected latency %v, got %v", latency, latencies[0])
	}

	// Test average operation latency
	avgLatency := metrics.GetAverageOperationLatency(operationType)
	if avgLatency != latency {
		t.Errorf("Expected average latency %v, got %v", latency, avgLatency)
	}

	// Test last backup time
	now := time.Now()
	metrics.SetLastBackupTime(now)
	if !metrics.GetLastBackupTime().Equal(now) {
		t.Errorf("Expected last backup time %v, got %v", now, metrics.GetLastBackupTime())
	}
}

func TestMetricsSnapshot(t *testing.T) {
	metrics := NewMetrics()

	// Set up some test data
	metrics.IncrementTaskCount()
	metrics.IncrementCompletedTaskCount()
	metrics.IncrementBackupCount()
	metrics.RecordOperation("test_op", 100*time.Millisecond)
	metrics.SetLastBackupTime(time.Now())

	// Get snapshot
	snapshot := metrics.GetMetricsSnapshot()

	// Verify snapshot contents
	if snapshot["taskCount"] != uint64(1) {
		t.Errorf("Expected task count 1, got %v", snapshot["taskCount"])
	}
	if snapshot["completedTaskCount"] != uint64(1) {
		t.Errorf("Expected completed task count 1, got %v", snapshot["completedTaskCount"])
	}
	if snapshot["backupCount"] != uint64(1) {
		t.Errorf("Expected backup count 1, got %v", snapshot["backupCount"])
	}

	// Verify operations
	operations, ok := snapshot["operations"].(map[string]map[string]interface{})
	if !ok {
		t.Error("Expected operations map in snapshot")
	}

	testOp, ok := operations["test_op"]
	if !ok {
		t.Error("Expected test_op in operations")
	}

	if testOp["count"] != uint64(1) {
		t.Errorf("Expected operation count 1, got %v", testOp["count"])
	}
}

func TestMultipleOperations(t *testing.T) {
	metrics := NewMetrics()
	operationType := "test_operation"

	// Record two operations with fixed latencies
	metrics.RecordOperation(operationType, 500*time.Millisecond)
	metrics.RecordOperation(operationType, 500*time.Millisecond)

	// Verify operation count
	if metrics.GetOperationCount(operationType) != 2 {
		t.Errorf("Expected operation count 2, got %d", metrics.GetOperationCount(operationType))
	}

	// Verify latencies are recorded
	latencies := metrics.GetOperationLatencies(operationType)
	if len(latencies) != 2 {
		t.Errorf("Expected 2 latencies, got %d", len(latencies))
	}

	// Verify average latency
	avgLatency := metrics.GetAverageOperationLatency(operationType)
	expectedAvg := 500 * time.Millisecond
	if avgLatency != expectedAvg {
		t.Errorf("Expected average latency %v, got %v", expectedAvg, avgLatency)
	}
}

func TestConcurrentOperations(t *testing.T) {
	metrics := NewMetrics()
	operationType := "test_operation"

	// Run concurrent operations
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				metrics.RecordOperation(operationType, time.Duration(j)*time.Millisecond)
			}
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify total operation count
	if metrics.GetOperationCount(operationType) != 1000 {
		t.Errorf("Expected operation count 1000, got %d", metrics.GetOperationCount(operationType))
	}
}
