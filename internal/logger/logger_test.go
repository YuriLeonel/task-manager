package logger

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewLogger(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "logger-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test creating logger with debug enabled
	logger, err := NewLogger(tempDir, true)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	if logger == nil {
		t.Fatal("Logger is nil")
	}
	if !logger.debugEnabled {
		t.Error("Debug should be enabled")
	}

	// Test creating logger with debug disabled
	logger, err = NewLogger(tempDir, false)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	if logger == nil {
		t.Fatal("Logger is nil")
	}
	if logger.debugEnabled {
		t.Error("Debug should be disabled")
	}

	// Test creating logger with invalid directory
	_, err = NewLogger("/nonexistent/directory", true)
	if err == nil {
		t.Error("Expected error when creating logger with invalid directory")
	}
}

func TestLoggerMethods(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "logger-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create logger
	logger, err := NewLogger(tempDir, true)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// Test all logging methods
	logger.Debug("debug message %d", 1)
	logger.Info("info message %d", 2)
	logger.Warn("warn message %d", 3)
	logger.Error("error message %d", 4)

	// Verify log file was created
	logFile := filepath.Join(tempDir, "task-manager-"+time.Now().Format("2025-03-30")+".log")
	_, err = os.Stat(logFile)
	if err != nil {
		t.Fatalf("Log file was not created: %v", err)
	}
}

func TestWithFields(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "logger-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create logger
	logger, err := NewLogger(tempDir, true)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// Test WithFields
	fields := map[string]any{
		"user":    "testuser",
		"request": "GET /api/tasks",
	}
	withFields := logger.WithFields(fields)

	// Verify the new logger has the fields in its prefix
	withFields.Info("test message")

	// Verify log file was created
	logFile := filepath.Join(tempDir, "task-manager-"+time.Now().Format("2025-03-30")+".log")
	_, err = os.Stat(logFile)
	if err != nil {
		t.Fatalf("Log file was not created: %v", err)
	}
}

func TestDebugLogging(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "logger-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create logger with debug disabled
	logger, err := NewLogger(tempDir, false)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// Test debug logging when disabled
	logger.Debug("this should not be logged")
	logger.Info("this should be logged")

	// Verify log file was created
	logFile := filepath.Join(tempDir, "task-manager-"+time.Now().Format("2025-03-30")+".log")
	_, err = os.Stat(logFile)
	if err != nil {
		t.Fatalf("Log file was not created: %v", err)
	}
}
