// Package task provides the business logic for task management.
// It includes security measures, validation, and thread-safe operations.
package task

import (
	"context"
	"fmt"
	"time"

	"slices"

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
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	task := models.NewTask(description)
	return s.storage.SaveTask(ctx, task)
}

// ListTasks returns all tasks from storage.
func (s *Service) ListTasks() ([]*models.Task, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return s.storage.GetAllTasks(ctx)
}

// MarkAsCompleted marks a task as completed by its ID.
func (s *Service) MarkAsCompleted(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	task, err := s.storage.GetTask(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}
	if task == nil {
		return fmt.Errorf("task not found: %s", id)
	}

	task.MarkAsCompleted()
	return s.storage.UpdateTask(ctx, task)
}

// SetPriority updates a task's priority.
func (s *Service) SetPriority(id string, priority models.Priority) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	task, err := s.storage.GetTask(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}
	if task == nil {
		return fmt.Errorf("task not found: %s", id)
	}

	task.SetPriority(priority)
	return s.storage.UpdateTask(ctx, task)
}

// SetDueDate updates a task's due date.
func (s *Service) SetDueDate(id string, dueDate time.Time) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	task, err := s.storage.GetTask(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}
	if task == nil {
		return fmt.Errorf("task not found: %s", id)
	}

	task.SetDueDate(dueDate)
	return s.storage.UpdateTask(ctx, task)
}

// AddTag adds a tag to a task
func (s *Service) AddTag(id string, tag string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if tag == "" {
		return fmt.Errorf("tag cannot be empty")
	}

	task, err := s.storage.GetTask(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}
	if task == nil {
		return fmt.Errorf("task not found: %s", id)
	}

	task.AddTag(tag)
	return s.storage.UpdateTask(ctx, task)
}

// RemoveTag removes a tag from a task
func (s *Service) RemoveTag(id string, tag string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if tag == "" {
		return fmt.Errorf("tag cannot be empty")
	}

	task, err := s.storage.GetTask(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}
	if task == nil {
		return fmt.Errorf("task not found: %s", id)
	}

	task.RemoveTag(tag)
	return s.storage.UpdateTask(ctx, task)
}

// SetProgress updates a task's progress percentage
func (s *Service) SetProgress(id string, progress int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	task, err := s.storage.GetTask(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}
	if task == nil {
		return fmt.Errorf("task not found: %s", id)
	}

	task.SetProgress(progress)
	return s.storage.UpdateTask(ctx, task)
}

// DeleteTask removes a task by its ID.
func (s *Service) DeleteTask(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	task, err := s.storage.GetTask(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}
	if task == nil {
		return fmt.Errorf("task not found: %s", id)
	}

	return s.storage.DeleteTask(ctx, id)
}

// GetTask retrieves a task by its ID.
func (s *Service) GetTask(id string) (*models.Task, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return s.storage.GetTask(ctx, id)
}

// GetTasksByTag returns all tasks that have the specified tag
func (s *Service) GetTasksByTag(tag string) ([]*models.Task, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if tag == "" {
		return nil, fmt.Errorf("tag cannot be empty")
	}

	tasks, err := s.storage.GetAllTasks(ctx)
	if err != nil {
		return nil, err
	}

	var filteredTasks []*models.Task
	for _, task := range tasks {
		if slices.Contains(task.Tags, tag) {
			filteredTasks = append(filteredTasks, task)
		}
	}

	return filteredTasks, nil
}
