package task

import (
	"fmt"

	"github.com/YuriLeonel/task-manager/pkg/models"
)

// Service handles the business logic for task management
type Service struct {
	tasks []*models.Task
}

// NewService creates a new task service
func NewService() *Service {
	return &Service{
		tasks: make([]*models.Task, 0),
	}
}

// AddTask adds a new task to the list
func (service *Service) AddTask(description string) error {
	if description == "" {
		return fmt.Errorf("task description cannot be empty")
	}

	task := models.NewTask(description)
	service.tasks = append(service.tasks, task)
	return nil
}

// ListTasks returns all tasks
func (service *Service) ListTasks() []*models.Task {
	return service.tasks
}

// MarkTaskAsCompleted marks a task as completed by its index
func (service *Service) MarkTaskAsCompleted(index int) error {
	if index < 0 || index >= len(service.tasks) {
		return fmt.Errorf("invalid task index: %d", index)
	}

	service.tasks[index].MarkAsCompleted()
	return nil
}

// GetTaskCount returns the total number of tasks
func (service *Service) GetTaskCount() int {
	return len(service.tasks)
}
