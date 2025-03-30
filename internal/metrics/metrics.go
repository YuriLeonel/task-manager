// Package metrics provides functionality for tracking and reporting
// important metrics in the task manager application.
package metrics

import (
	"sync"
	"sync/atomic"
	"time"
)

// Metrics tracks various metrics for the task manager.
type Metrics struct {
	// mu provides thread safety for all operations.
	mu sync.RWMutex
	// taskCount is the current number of tasks.
	taskCount uint64
	// completedTaskCount is the number of completed tasks.
	completedTaskCount uint64
	// backupCount is the number of backups.
	backupCount uint64
	// operationCounts tracks the number of operations by type.
	operationCounts map[string]uint64
	// operationLatencies tracks the latency of operations by type.
	operationLatencies map[string][]time.Duration
	// lastBackupTime is the timestamp of the last backup.
	lastBackupTime time.Time
}

// NewMetrics creates a new metrics instance.
func NewMetrics() *Metrics {
	return &Metrics{
		operationCounts:    make(map[string]uint64),
		operationLatencies: make(map[string][]time.Duration),
	}
}

// IncrementTaskCount increments the task count.
func (m *Metrics) IncrementTaskCount() {
	atomic.AddUint64(&m.taskCount, 1)
}

// DecrementTaskCount decrements the task count.
func (m *Metrics) DecrementTaskCount() {
	atomic.AddUint64(&m.taskCount, ^uint64(0))
}

// GetTaskCount returns the current task count.
func (m *Metrics) GetTaskCount() uint64 {
	return atomic.LoadUint64(&m.taskCount)
}

// IncrementCompletedTaskCount increments the completed task count.
func (m *Metrics) IncrementCompletedTaskCount() {
	atomic.AddUint64(&m.completedTaskCount, 1)
}

// DecrementCompletedTaskCount decrements the completed task count.
func (m *Metrics) DecrementCompletedTaskCount() {
	atomic.AddUint64(&m.completedTaskCount, ^uint64(0))
}

// GetCompletedTaskCount returns the current completed task count.
func (m *Metrics) GetCompletedTaskCount() uint64 {
	return atomic.LoadUint64(&m.completedTaskCount)
}

// IncrementBackupCount increments the backup count.
func (m *Metrics) IncrementBackupCount() {
	atomic.AddUint64(&m.backupCount, 1)
}

// DecrementBackupCount decrements the backup count.
func (m *Metrics) DecrementBackupCount() {
	atomic.AddUint64(&m.backupCount, ^uint64(0))
}

// GetBackupCount returns the current backup count.
func (m *Metrics) GetBackupCount() uint64 {
	return atomic.LoadUint64(&m.backupCount)
}

// RecordOperation records an operation with its latency.
func (m *Metrics) RecordOperation(operationType string, latency time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Increment operation count
	m.operationCounts[operationType]++

	// Record operation latency
	m.operationLatencies[operationType] = append(m.operationLatencies[operationType], latency)

	// Keep only the last 1000 latencies per operation type
	if len(m.operationLatencies[operationType]) > 1000 {
		m.operationLatencies[operationType] = m.operationLatencies[operationType][1:]
	}
}

// GetOperationCount returns the count of operations of a specific type.
func (m *Metrics) GetOperationCount(operationType string) uint64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.operationCounts[operationType]
}

// GetOperationLatencies returns the latencies of operations of a specific type.
func (m *Metrics) GetOperationLatencies(operationType string) []time.Duration {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.operationLatencies[operationType]
}

// GetAverageOperationLatency returns the average latency of operations of a specific type.
func (m *Metrics) GetAverageOperationLatency(operationType string) time.Duration {
	m.mu.RLock()
	defer m.mu.RUnlock()

	latencies := m.operationLatencies[operationType]
	if len(latencies) == 0 {
		return 0
	}

	var total time.Duration
	for _, latency := range latencies {
		total += latency
	}
	return total / time.Duration(len(latencies))
}

// SetLastBackupTime sets the timestamp of the last backup.
func (m *Metrics) SetLastBackupTime(t time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.lastBackupTime = t
}

// GetLastBackupTime returns the timestamp of the last backup.
func (m *Metrics) GetLastBackupTime() time.Time {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.lastBackupTime
}

// GetMetricsSnapshot returns a snapshot of all metrics.
func (m *Metrics) GetMetricsSnapshot() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	snapshot := make(map[string]interface{})
	snapshot["taskCount"] = m.GetTaskCount()
	snapshot["completedTaskCount"] = m.GetCompletedTaskCount()
	snapshot["backupCount"] = m.GetBackupCount()
	snapshot["lastBackupTime"] = m.lastBackupTime

	// Add operation counts and average latencies
	operationMetrics := make(map[string]map[string]interface{})
	for opType := range m.operationCounts {
		operationMetrics[opType] = map[string]interface{}{
			"count":      m.operationCounts[opType],
			"avgLatency": m.GetAverageOperationLatency(opType),
		}
	}
	snapshot["operations"] = operationMetrics

	return snapshot
}
