// Package logger provides structured logging functionality for the task manager.
// It supports different log levels and includes contextual information in log entries.
package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

// Level represents a log level.
type Level int

const (
	// Debug level for detailed debugging information.
	Debug Level = iota
	// Info level for general operational information.
	Info
	// Warn level for warning messages.
	Warn
	// Error level for error messages.
	Error
)

// Logger provides structured logging functionality.
type Logger struct {
	// debugLogger is used for debug level messages.
	debugLogger *log.Logger
	// infoLogger is used for info level messages.
	infoLogger *log.Logger
	// warnLogger is used for warning level messages.
	warnLogger *log.Logger
	// errorLogger is used for error level messages.
	errorLogger *log.Logger
	// debugEnabled indicates if debug logging is enabled.
	debugEnabled bool
}

// NewLogger creates a new logger instance.
func NewLogger(logDir string, debugEnabled bool) (*Logger, error) {
	// Create log directory if it doesn't exist
	if err := os.MkdirAll(logDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	// Create log file with timestamp
	timestamp := time.Now().Format("2025-03-30")
	logFile := filepath.Join(logDir, fmt.Sprintf("task-manager-%s.log", timestamp))
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	// Create loggers with different prefixes
	logger := &Logger{
		debugLogger:  log.New(file, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile),
		infoLogger:   log.New(file, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile),
		warnLogger:   log.New(file, "WARN: ", log.Ldate|log.Ltime|log.Lshortfile),
		errorLogger:  log.New(file, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile),
		debugEnabled: debugEnabled,
	}

	return logger, nil
}

// Debug logs a debug message.
func (l *Logger) Debug(format string, v ...any) {
	if l.debugEnabled {
		l.debugLogger.Printf(format, v...)
	}
}

// Info logs an info message.
func (l *Logger) Info(format string, v ...any) {
	l.infoLogger.Printf(format, v...)
}

// Warn logs a warning message.
func (l *Logger) Warn(format string, v ...any) {
	l.warnLogger.Printf(format, v...)
}

// Error logs an error message.
func (l *Logger) Error(format string, v ...any) {
	l.errorLogger.Printf(format, v...)
}

// WithFields creates a new logger instance with additional fields.
func (l *Logger) WithFields(fields map[string]any) *Logger {
	// Create a new logger with the same configuration
	newLogger := &Logger{
		debugLogger:  l.debugLogger,
		infoLogger:   l.infoLogger,
		warnLogger:   l.warnLogger,
		errorLogger:  l.errorLogger,
		debugEnabled: l.debugEnabled,
	}

	// Add fields to log messages
	prefix := ""
	for key, value := range fields {
		prefix += fmt.Sprintf("[%s=%v] ", key, value)
	}

	newLogger.debugLogger.SetPrefix("DEBUG: " + prefix)
	newLogger.infoLogger.SetPrefix("INFO: " + prefix)
	newLogger.warnLogger.SetPrefix("WARN: " + prefix)
	newLogger.errorLogger.SetPrefix("ERROR: " + prefix)

	return newLogger
}
