// Package errors provides custom error types for the task manager.
package errors

import "fmt"

// TaskError represents a task-related error.
type TaskError struct {
	// Code is the error code
	Code string
	// Message is the error message
	Message string
	// Err is the underlying error
	Err error
}

// Error implements the error interface.
func (e *TaskError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying error.
func (e *TaskError) Unwrap() error {
	return e.Err
}

// PersistenceError represents a storage-related error.
type PersistenceError struct {
	// Code is the error code
	Code string
	// Message is the error message
	Message string
	// Err is the underlying error
	Err error
}

// Error implements the error interface.
func (e *PersistenceError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying error.
func (e *PersistenceError) Unwrap() error {
	return e.Err
}

// ValidationError represents a validation error.
type ValidationError struct {
	// Field is the field that failed validation
	Field string
	// Message is the validation error message
	Message string
}

// Error implements the error interface.
func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error: %s: %s", e.Field, e.Message)
}

// Common error codes
const (
	// TaskNotFound is returned when a task is not found
	TaskNotFound = "TASK_NOT_FOUND"
	// TaskAlreadyExists is returned when a task with the same ID already exists
	TaskAlreadyExists = "TASK_ALREADY_EXISTS"
	// TaskLimitExceeded is returned when the maximum number of tasks is reached
	TaskLimitExceeded = "TASK_LIMIT_EXCEEDED"
	// InvalidTaskData is returned when task data is invalid
	InvalidTaskData = "INVALID_TASK_DATA"
	// PersistenceErrorCode is returned when a storage operation fails
	PersistenceErrorCode = "PERSISTENCE_ERROR"
	// BackupError is returned when a backup operation fails
	BackupError = "BACKUP_ERROR"
	// InvalidInput is returned when user input is invalid
	InvalidInput = "INVALID_INPUT"
)

// NewTaskError creates a new TaskError.
func NewTaskError(code string, message string, err error) *TaskError {
	return &TaskError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// NewPersistenceError creates a new PersistenceError.
func NewPersistenceError(code string, message string, err error) *PersistenceError {
	return &PersistenceError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// NewValidationError creates a new ValidationError.
func NewValidationError(field string, message string) *ValidationError {
	return &ValidationError{
		Field:   field,
		Message: message,
	}
}
