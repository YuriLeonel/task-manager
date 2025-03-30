package storage

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/YuriLeonel/task-manager/pkg/models"
)

// MemoryStorage implements Storage interface using in-memory structures.
// It provides thread-safe operations for testing and development purposes.
type MemoryStorage struct {
	// tasks is a map of task ID to task
	tasks map[string]*models.Task
	// mu provides thread safety for all operations
	mu sync.RWMutex
	// maxTasks is the maximum number of tasks allowed
	maxTasks int
}

// NewMemoryStorage creates a new in-memory storage instance.
func NewMemoryStorage(maxTasks int) (*MemoryStorage, error) {
	if maxTasks <= 0 {
		return nil, fmt.Errorf("maxTasks must be greater than 0")
	}

	return &MemoryStorage{
		tasks:    make(map[string]*models.Task),
		maxTasks: maxTasks,
	}, nil
}

// validateTask checks if a task is valid before saving.
func (s *MemoryStorage) validateTask(task *models.Task) error {
	if task == nil {
		return fmt.Errorf("task cannot be nil")
	}
	if task.ID == "" {
		return fmt.Errorf("task ID cannot be empty")
	}
	if task.Description == "" {
		return fmt.Errorf("task description cannot be empty")
	}
	if task.CreatedAt.IsZero() {
		task.CreatedAt = time.Now()
	}
	if task.UpdatedAt.IsZero() {
		task.UpdatedAt = time.Now()
	}
	return nil
}

// SaveTask saves a task to memory.
func (s *MemoryStorage) SaveTask(ctx context.Context, task *models.Task) error {
	// Check for context cancellation
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context error: %w", err)
	}

	if err := s.validateTask(task); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Check task limit
	if len(s.tasks) >= s.maxTasks {
		return fmt.Errorf("maximum number of tasks (%d) reached", s.maxTasks)
	}

	// Check for duplicate ID
	if _, exists := s.tasks[task.ID]; exists {
		return fmt.Errorf("task with ID %s already exists", task.ID)
	}

	// Add new task (create a deep copy)
	s.tasks[task.ID] = copyTask(task)
	return nil
}

// GetTask retrieves a task by its ID.
func (s *MemoryStorage) GetTask(ctx context.Context, id string) (*models.Task, error) {
	// Check for context cancellation
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context error: %w", err)
	}

	if id == "" {
		return nil, fmt.Errorf("task ID cannot be empty")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	task, exists := s.tasks[id]
	if !exists {
		return nil, fmt.Errorf("task not found: %s", id)
	}

	// Return a copy to prevent external modification
	return copyTask(task), nil
}

// GetAllTasks retrieves all tasks.
func (s *MemoryStorage) GetAllTasks(ctx context.Context) ([]*models.Task, error) {
	// Check for context cancellation
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context error: %w", err)
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	tasks := make([]*models.Task, 0, len(s.tasks))
	for _, task := range s.tasks {
		tasks = append(tasks, copyTask(task))
	}

	return tasks, nil
}

// UpdateTask updates an existing task.
func (s *MemoryStorage) UpdateTask(ctx context.Context, updatedTask *models.Task) error {
	// Check for context cancellation
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context error: %w", err)
	}

	if err := s.validateTask(updatedTask); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.tasks[updatedTask.ID]; !exists {
		return fmt.Errorf("task not found: %s", updatedTask.ID)
	}

	// Update the task (create a deep copy)
	s.tasks[updatedTask.ID] = copyTask(updatedTask)
	return nil
}

// DeleteTask removes a task by its ID.
func (s *MemoryStorage) DeleteTask(ctx context.Context, id string) error {
	// Check for context cancellation
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context error: %w", err)
	}

	if id == "" {
		return fmt.Errorf("task ID cannot be empty")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.tasks[id]; !exists {
		return fmt.Errorf("task not found: %s", id)
	}

	delete(s.tasks, id)
	return nil
}
