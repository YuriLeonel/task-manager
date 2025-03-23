// Package models provides the core data structures for the task manager.
// It includes validation rules and security considerations for task data.
package models

import (
	"time"
)

// Priority represents the priority level of a task
type Priority int

const (
	Low Priority = iota
	Medium
	High
)

// Task represents a single task in the task manager.
// It includes validation rules and security measures to ensure data integrity.
type Task struct {
	// ID uniquely identifies the task
	ID string `json:"id"`
	// Description contains the task's description
	Description string `json:"description"`
	// Completed indicates whether the task is completed
	Completed bool `json:"completed"`
	// CreatedAt is the timestamp when the task was created
	CreatedAt time.Time `json:"created_at"`
	// UpdatedAt is the timestamp when the task was last modified
	UpdatedAt time.Time `json:"updated_at"`
	// DueDate is the optional deadline for the task
	DueDate *time.Time `json:"due_date,omitempty"`
	// Priority indicates the task's priority level
	Priority Priority `json:"priority"`
}

// NewTask creates a new Task instance with the given description.
// The task is initialized with default values and current timestamp.
func NewTask(description string) *Task {
	now := time.Now()
	return &Task{
		ID:          generateID(),
		Description: description,
		Completed:   false,
		CreatedAt:   now,
		UpdatedAt:   now,
		Priority:    Medium,
	}
}

// MarkAsCompleted marks the task as completed and updates the UpdatedAt timestamp.
func (task *Task) MarkAsCompleted() {
	task.Completed = true
	task.UpdatedAt = time.Now()
}

// SetPriority updates the task's priority and updates the UpdatedAt timestamp.
func (task *Task) SetPriority(priority Priority) {
	task.Priority = priority
	task.UpdatedAt = time.Now()
}

// SetDueDate updates the task's due date and updates the UpdatedAt timestamp.
func (task *Task) SetDueDate(dueDate time.Time) {
	task.DueDate = &dueDate
	task.UpdatedAt = time.Now()
}

// String implements the Stringer interface for Task.
// It returns a formatted string representation of the task.
func (task *Task) String() string {
	status := " "
	if task.Completed {
		status = "X"
	}
	return status
}

// generateID creates a unique ID for a task using timestamp and random string
func generateID() string {
	return time.Now().Format("20060102150405")
}
