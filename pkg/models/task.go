package models

// Task represents a single task in the task manager
type Task struct {
	Description string
	Completed   bool
}

// NewTask creates a new task with the given description
func NewTask(description string) *Task {
	return &Task{
		Description: description,
		Completed:   false,
	}
}

// MarkAsCompleted marks the task as completed
func (task *Task) MarkAsCompleted() {
	task.Completed = true
}

// String returns a string representation of the task
func (task *Task) String() string {
	status := " "
	if task.Completed {
		status = "X"
	}
	return status
}
