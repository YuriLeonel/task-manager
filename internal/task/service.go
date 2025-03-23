// Package task provides the business logic for task management.
// It includes security measures, validation, and thread-safe operations.
package task

import (
	"fmt"
	"time"

	"github.com/YuriLeonel/task-manager/internal/storage"
	"github.com/YuriLeonel/task-manager/pkg/models"
)

// Service handles the business logic for task management.
// It ensures thread-safe operations and proper validation.
type Service struct {
	storage storage.Storage
}

// NewService creates a new task service instance with secure storage.
func NewService(storage storage.Storage) *Service {
	return &Service{
		storage: storage,
	}
}

// AddTask creates a new task with the given description.
func (s *Service) AddTask(description string) error {
	task := models.NewTask(description)
	return s.storage.SaveTask(task)
}

// ListTasks returns all tasks from storage.
func (s *Service) ListTasks() ([]*models.Task, error) {
	return s.storage.GetAllTasks()
}

// MarkAsCompleted marks a task as completed by its ID.
func (s *Service) MarkAsCompleted(id string) error {
	task, err := s.storage.GetTask(id)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}

	task.MarkAsCompleted()
	return s.storage.UpdateTask(task)
}

// SetPriority updates a task's priority.
func (s *Service) SetPriority(id string, priority models.Priority) error {
	task, err := s.storage.GetTask(id)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}

	task.SetPriority(priority)
	return s.storage.UpdateTask(task)
}

// SetDueDate updates a task's due date.
func (s *Service) SetDueDate(id string, dueDate time.Time) error {
	task, err := s.storage.GetTask(id)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}

	task.SetDueDate(dueDate)
	return s.storage.UpdateTask(task)
}

// DeleteTask removes a task by its ID.
func (s *Service) DeleteTask(id string) error {
	return s.storage.DeleteTask(id)
}

// GetTask retrieves a task by its ID.
func (s *Service) GetTask(id string) (*models.Task, error) {
	return s.storage.GetTask(id)
}
