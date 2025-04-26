package errors

import (
	"errors"
	"testing"
)

func TestTaskError(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		message  string
		err      error
		expected string
	}{
		{
			name:     "without underlying error",
			code:     TaskNotFound,
			message:  "task not found",
			err:      nil,
			expected: "TASK_NOT_FOUND: task not found",
		},
		{
			name:     "with underlying error",
			code:     TaskNotFound,
			message:  "task not found",
			err:      errors.New("file not found"),
			expected: "TASK_NOT_FOUND: task not found: file not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewTaskError(tt.code, tt.message, tt.err)
			if err.Error() != tt.expected {
				t.Errorf("Error() = %v, want %v", err.Error(), tt.expected)
			}
			if err.Unwrap() != tt.err {
				t.Errorf("Unwrap() = %v, want %v", err.Unwrap(), tt.err)
			}
		})
	}
}

func TestPersistenceError(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		message  string
		err      error
		expected string
	}{
		{
			name:     "without underlying error",
			code:     PersistenceErrorCode,
			message:  "failed to save",
			err:      nil,
			expected: "PERSISTENCE_ERROR: failed to save",
		},
		{
			name:     "with underlying error",
			code:     PersistenceErrorCode,
			message:  "failed to save",
			err:      errors.New("disk full"),
			expected: "PERSISTENCE_ERROR: failed to save: disk full",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewPersistenceError(tt.code, tt.message, tt.err)
			if err.Error() != tt.expected {
				t.Errorf("Error() = %v, want %v", err.Error(), tt.expected)
			}
			if err.Unwrap() != tt.err {
				t.Errorf("Unwrap() = %v, want %v", err.Unwrap(), tt.err)
			}
		})
	}
}

func TestValidationError(t *testing.T) {
	tests := []struct {
		name     string
		field    string
		message  string
		expected string
	}{
		{
			name:     "simple validation error",
			field:    "name",
			message:  "cannot be empty",
			expected: "validation error: name: cannot be empty",
		},
		{
			name:     "complex validation error",
			field:    "email",
			message:  "must be a valid email address",
			expected: "validation error: email: must be a valid email address",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewValidationError(tt.field, tt.message)
			if err.Error() != tt.expected {
				t.Errorf("Error() = %v, want %v", err.Error(), tt.expected)
			}
		})
	}
}
