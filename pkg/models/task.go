// Package models provides the data structures and methods for task management.
// It defines the core Task type and its associated operations that are used
// throughout the application to represent and manipulate tasks.
package models

// Task represents a single task in the task manager.
// It contains a description of the task and its completion status.
// The Description field holds the task's text content while
// Completed indicates whether the task has been finished.
type Task struct {
	Description string
	Completed   bool
}

// NewTask creates a new Task instance with the given description.
// The task is initialized as not completed.
// The description parameter specifies the task's content.
func NewTask(description string) *Task {
	return &Task{
		Description: description,
		Completed:   false,
	}
}

// MarkAsCompleted marks the task as completed.
// This is an irreversible operation - once a task is marked as completed,
// there is no method to mark it as incomplete.
func (task *Task) MarkAsCompleted() {
	task.Completed = true
}

// String implements the Stringer interface for Task.
// It returns "X" for completed tasks and " " for incomplete tasks,
// which is used for displaying the task's status in the CLI.
func (task *Task) String() string {
	status := " "
	if task.Completed {
		status = "X"
	}
	return status
}
