package storage

import (
	"context"
	"sync"
	"time"

	"github.com/YuriLeonel/task-manager/internal/errors"
	"github.com/YuriLeonel/task-manager/internal/storage/backup"
	"github.com/YuriLeonel/task-manager/pkg/models"
	"github.com/YuriLeonel/task-manager/pkg/utils"
)

// MemoryStorage implements Storage interface using in-memory storage.
// It provides thread-safe operations for task management.
type MemoryStorage struct {
	// tasks is a map of task ID to task pointer.
	tasks map[string]*models.Task
	// mu provides thread safety for all operations.
	mu sync.RWMutex
	// maxTasks is the maximum number of tasks allowed in storage.
	maxTasks int
	// backupManager handles backup operations.
	backupManager *backup.Manager
}

// NewMemoryStorage creates a new memory storage instance.
func NewMemoryStorage(maxTasks int, backupDir string, maxBackups int) (*MemoryStorage, error) {
	if maxTasks <= 0 {
		return nil, errors.NewValidationError("maxTasks", "must be greater than 0")
	}

	// Create backup manager
	backupManager, err := backup.NewManager(backupDir, maxBackups)
	if err != nil {
		return nil, err
	}

	return &MemoryStorage{
		tasks:         make(map[string]*models.Task),
		maxTasks:      maxTasks,
		backupManager: backupManager,
	}, nil
}

// validateTask checks if a task is valid before saving.
func (s *MemoryStorage) validateTask(task *models.Task) error {
	if task == nil {
		return errors.NewValidationError("task", "cannot be nil")
	}
	if task.ID == "" {
		return errors.NewValidationError("task.id", "cannot be empty")
	}
	if task.Description == "" {
		return errors.NewValidationError("task.description", "cannot be empty")
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
		return errors.NewPersistenceError(errors.PersistenceErrorCode, "context error", err)
	}

	if err := s.validateTask(task); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Check task limit
	if len(s.tasks) >= s.maxTasks {
		return errors.NewTaskError(errors.TaskLimitExceeded, "maximum number of tasks reached", nil)
	}

	// Check for duplicate ID
	if _, exists := s.tasks[task.ID]; exists {
		return errors.NewTaskError(errors.TaskAlreadyExists, "task with ID already exists", nil)
	}

	// Save task
	s.tasks[task.ID] = task
	return nil
}

// GetTask retrieves a task by its ID.
func (s *MemoryStorage) GetTask(ctx context.Context, taskID string) (*models.Task, error) {
	// Check for context cancellation
	if err := ctx.Err(); err != nil {
		return nil, errors.NewPersistenceError(errors.PersistenceErrorCode, "context error", err)
	}

	if taskID == "" {
		return nil, errors.NewValidationError("taskID", "cannot be empty")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	task, exists := s.tasks[taskID]
	if !exists {
		return nil, errors.NewTaskError(errors.TaskNotFound, "task not found", nil)
	}

	// Return a copy to prevent external modification
	return copyTask(task), nil
}

// GetAllTasks retrieves all tasks from memory.
func (s *MemoryStorage) GetAllTasks(ctx context.Context) ([]*models.Task, error) {
	// Check for context cancellation
	if err := ctx.Err(); err != nil {
		return nil, errors.NewPersistenceError(errors.PersistenceErrorCode, "context error", err)
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	// Create a slice to store all tasks
	tasks := make([]*models.Task, 0, len(s.tasks))
	for _, task := range s.tasks {
		// Add a copy of each task to prevent external modification
		tasks = append(tasks, copyTask(task))
	}

	return tasks, nil
}

// UpdateTask updates an existing task in memory.
func (s *MemoryStorage) UpdateTask(ctx context.Context, task *models.Task) error {
	// Check for context cancellation
	if err := ctx.Err(); err != nil {
		return errors.NewPersistenceError(errors.PersistenceErrorCode, "context error", err)
	}

	if err := s.validateTask(task); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if task exists
	if _, exists := s.tasks[task.ID]; !exists {
		return errors.NewTaskError(errors.TaskNotFound, "task not found", nil)
	}

	// Update task
	s.tasks[task.ID] = task
	return nil
}

// DeleteTask removes a task from memory.
func (s *MemoryStorage) DeleteTask(ctx context.Context, taskID string) error {
	// Check for context cancellation
	if err := ctx.Err(); err != nil {
		return errors.NewPersistenceError(errors.PersistenceErrorCode, "context error", err)
	}

	if taskID == "" {
		return errors.NewValidationError("taskID", "cannot be empty")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if task exists
	if _, exists := s.tasks[taskID]; !exists {
		return errors.NewTaskError(errors.TaskNotFound, "task not found", nil)
	}

	// Delete task
	delete(s.tasks, taskID)
	return nil
}

// CreateBackup creates a backup of the current tasks.
func (s *MemoryStorage) CreateBackup(ctx context.Context) error {
	// Check for context cancellation
	if err := ctx.Err(); err != nil {
		return errors.NewPersistenceError(errors.PersistenceErrorCode, "context error", err)
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	// Convert tasks map to slice
	tasks := make([]*models.Task, 0, len(s.tasks))
	for _, task := range s.tasks {
		tasks = append(tasks, utils.CopyTask(task))
	}

	// Create backup
	return s.backupManager.CreateBackup(ctx, tasks)
}

// RestoreBackup restores tasks from a backup.
func (s *MemoryStorage) RestoreBackup(ctx context.Context, backupID string) error {
	// Check for context cancellation
	if err := ctx.Err(); err != nil {
		return errors.NewPersistenceError(errors.PersistenceErrorCode, "context error", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Restore tasks from backup
	tasks, err := s.backupManager.RestoreBackup(ctx, backupID)
	if err != nil {
		return err
	}

	// Validate restored tasks
	for _, task := range tasks {
		if err := s.validateTask(task); err != nil {
			return err
		}
	}

	// Clear existing tasks
	s.tasks = make(map[string]*models.Task)

	// Add restored tasks
	for _, task := range tasks {
		s.tasks[task.ID] = task
	}

	return nil
}

// ListBackups lists all available backups.
func (s *MemoryStorage) ListBackups(ctx context.Context) ([]string, error) {
	// Check for context cancellation
	if err := ctx.Err(); err != nil {
		return nil, errors.NewPersistenceError(errors.PersistenceErrorCode, "context error", err)
	}

	return s.backupManager.ListBackups(ctx)
}

// DeleteBackup deletes a backup.
func (s *MemoryStorage) DeleteBackup(ctx context.Context, backupID string) error {
	// Check for context cancellation
	if err := ctx.Err(); err != nil {
		return errors.NewPersistenceError(errors.PersistenceErrorCode, "context error", err)
	}

	return s.backupManager.DeleteBackup(ctx, backupID)
}
