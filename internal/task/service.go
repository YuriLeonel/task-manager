// Package task provides the core business logic and CLI interface for task management.
// It implements the task service layer that handles all task operations and the
// command-line interface that interacts with users.
package task

import (
	"fmt"

	"github.com/YuriLeonel/task-manager/pkg/models"
)

// Service handles the business logic for task management.
// It maintains a slice of tasks and provides methods to
// add, list, and modify tasks in the collection.
type Service struct {
	tasks []*models.Task
}

// NewService creates a new task service instance.
// It initializes an empty task collection and returns
// a pointer to the new Service.
func NewService() *Service {
	return &Service{
		tasks: make([]*models.Task, 0),
	}
}

// AddTask creates and adds a new task with the given description to the service.
// It returns an error if the description is empty.
// The task is automatically initialized as not completed.
func (service *Service) AddTask(description string) error {
	if description == "" {
		return fmt.Errorf("task description cannot be empty")
	}

	task := models.NewTask(description)
	service.tasks = append(service.tasks, task)
	return nil
}

// ListTasks returns all tasks currently managed by the service.
// The returned slice contains pointers to Task objects in the order
// they were added.
func (service *Service) ListTasks() []*models.Task {
	return service.tasks
}

// MarkTaskAsCompleted marks the task at the specified index as completed.
// It returns an error if the index is out of bounds.
// Index is zero-based, meaning the first task is at index 0.
func (service *Service) MarkTaskAsCompleted(index int) error {
	if index < 0 || index >= len(service.tasks) {
		return fmt.Errorf("invalid task index: %d", index)
	}

	service.tasks[index].MarkAsCompleted()
	return nil
}

// GetTaskCount returns the total number of tasks in the service.
// This count includes both completed and incomplete tasks.
func (service *Service) GetTaskCount() int {
	return len(service.tasks)
}
