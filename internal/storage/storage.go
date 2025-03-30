// Package storage provides interfaces and implementations for task persistence.
// It includes secure file-based storage with atomic operations, thread safety,
// and comprehensive error handling. The package implements various security
// measures including:
//   - Secure file permissions (0600 for files, 0700 for directories)
//   - Path traversal protection
//   - Resource limits (max 10,000 tasks, 10MB file size)
//   - Atomic file operations
//   - Input validation
//   - Thread-safe operations
//   - Secure temporary file handling
//   - Data integrity checks
package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"slices"

	"github.com/YuriLeonel/task-manager/pkg/models"
)

const (
	// maxFileSize is the maximum allowed file size in bytes (10MB) to prevent
	// disk space exhaustion attacks.
	maxFileSize = 10 * 1024 * 1024
	// defaultFileMode is the default file mode with user read/write only (0600)
	// to ensure file security.
	defaultFileMode = 0600
	// defaultDirMode is the default directory mode with user read/write/execute
	// only (0700) to ensure directory security.
	defaultDirMode = 0700
)

// Storage defines the interface for task persistence.
// All implementations must ensure thread safety and proper error handling.
type Storage interface {
	// SaveTask saves a task to the storage.
	// It validates the task before saving and ensures atomic operations.
	SaveTask(ctx context.Context, task *models.Task) error
	// GetTask retrieves a task by its ID.
	// Returns a copy of the task to prevent external modification.
	GetTask(ctx context.Context, id string) (*models.Task, error)
	// GetAllTasks retrieves all tasks.
	// Returns copies of all tasks to prevent external modification.
	GetAllTasks(ctx context.Context) ([]*models.Task, error)
	// UpdateTask updates an existing task.
	// It validates the task and ensures atomic operations.
	UpdateTask(ctx context.Context, task *models.Task) error
	// DeleteTask removes a task by its ID.
	// Ensures atomic operations during deletion.
	DeleteTask(ctx context.Context, id string) error
}

// JSONFileStorage implements Storage interface using JSON files.
// It provides secure file-based storage with atomic operations and thread safety.
type JSONFileStorage struct {
	// tasksFilePath is the absolute path to the JSON file storing tasks.
	// It is validated to ensure it's within allowed directories.
	tasksFilePath string
	// mu provides thread safety for all file operations.
	mu sync.RWMutex
	// maxTasks is the maximum number of tasks allowed in storage.
	maxTasks int
}

// NewJSONFileStorage creates a new JSON file storage instance.
func NewJSONFileStorage(tasksFilePath string, maxTasks int) (*JSONFileStorage, error) {
	// Validate and sanitize the filepath
	absolutePath, err := filepath.Abs(tasksFilePath)
	if err != nil {
		return nil, fmt.Errorf("invalid filepath: %w", err)
	}

	// Ensure the filepath is within allowed directories
	if err := validateFilePath(absolutePath); err != nil {
		return nil, err
	}

	// Validate maxTasks
	if maxTasks <= 0 {
		return nil, fmt.Errorf("maxTasks must be greater than 0")
	}

	return &JSONFileStorage{
		tasksFilePath: absolutePath,
		maxTasks:      maxTasks,
	}, nil
}

// validateFilePath ensures the filepath is secure and within allowed directories.
func validateFilePath(filePath string) error {
	// Get user's home directory
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	// Ensure the path is within the user's home directory
	if !strings.HasPrefix(filePath, userHomeDir) {
		return fmt.Errorf("filepath must be within user's home directory")
	}

	// Check for path traversal attempts
	if strings.Contains(filePath, "..") {
		return fmt.Errorf("path traversal not allowed")
	}

	return nil
}

// validateTask checks if a task is valid before saving.
func (s *JSONFileStorage) validateTask(task *models.Task) error {
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

// SaveTask saves a task to the JSON file.
func (s *JSONFileStorage) SaveTask(ctx context.Context, newTask *models.Task) error {
	// Check for context cancellation
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context error: %w", err)
	}

	if err := s.validateTask(newTask); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Read existing tasks
	existingTasks, err := s.readTasks(ctx)
	if err != nil {
		return err
	}

	// Check task limit
	if len(existingTasks) >= s.maxTasks {
		return fmt.Errorf("maximum number of tasks (%d) reached", s.maxTasks)
	}

	// Check for duplicate ID
	for _, existingTask := range existingTasks {
		if existingTask.ID == newTask.ID {
			return fmt.Errorf("task with ID %s already exists", newTask.ID)
		}
	}

	// Add new task and save
	existingTasks = append(existingTasks, newTask)
	return s.saveTasks(ctx, existingTasks)
}

// GetTask retrieves a task by its ID.
func (s *JSONFileStorage) GetTask(ctx context.Context, taskID string) (*models.Task, error) {
	// Check for context cancellation
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context error: %w", err)
	}

	if taskID == "" {
		return nil, fmt.Errorf("task ID cannot be empty")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	tasks, err := s.readTasks(ctx)
	if err != nil {
		return nil, err
	}

	for _, task := range tasks {
		if task.ID == taskID {
			// Return a copy to prevent external modification
			return copyTask(task), nil
		}
	}
	return nil, fmt.Errorf("task not found: %s", taskID)
}

// GetAllTasks retrieves all tasks from the JSON file.
func (s *JSONFileStorage) GetAllTasks(ctx context.Context) ([]*models.Task, error) {
	// Check for context cancellation
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context error: %w", err)
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	tasks, err := s.readTasks(ctx)
	if err != nil {
		return nil, err
	}

	// Return copies to prevent external modification
	tasksCopy := make([]*models.Task, len(tasks))
	for i, task := range tasks {
		tasksCopy[i] = copyTask(task)
	}
	return tasksCopy, nil
}

// UpdateTask updates an existing task in the JSON file.
func (s *JSONFileStorage) UpdateTask(ctx context.Context, updatedTask *models.Task) error {
	// Check for context cancellation
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context error: %w", err)
	}

	if err := s.validateTask(updatedTask); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	tasks, err := s.readTasks(ctx)
	if err != nil {
		return err
	}

	found := false
	for i, task := range tasks {
		if task.ID == updatedTask.ID {
			tasks[i] = updatedTask
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("task not found: %s", updatedTask.ID)
	}

	return s.saveTasks(ctx, tasks)
}

// DeleteTask removes a task from the JSON file.
func (s *JSONFileStorage) DeleteTask(ctx context.Context, taskID string) error {
	// Check for context cancellation
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context error: %w", err)
	}

	if taskID == "" {
		return fmt.Errorf("task ID cannot be empty")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	tasks, err := s.readTasks(ctx)
	if err != nil {
		return err
	}

	index := -1
	for i, task := range tasks {
		if task.ID == taskID {
			index = i
			break
		}
	}

	if index == -1 {
		return fmt.Errorf("task not found: %s", taskID)
	}

	// Remove task by index
	tasks = slices.Delete(tasks, index, index+1)
	return s.saveTasks(ctx, tasks)
}

// readTasks reads tasks from the JSON file.
func (s *JSONFileStorage) readTasks(ctx context.Context) ([]*models.Task, error) {
	// Check for context cancellation
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context error: %w", err)
	}

	// If file doesn't exist, return empty slice
	if _, err := os.Stat(s.tasksFilePath); os.IsNotExist(err) {
		return []*models.Task{}, nil
	}

	// Open the file
	file, err := os.Open(s.tasksFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open tasks file: %w", err)
	}
	defer file.Close()

	// Check file size
	fileInfo, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}
	if fileInfo.Size() > maxFileSize {
		return nil, fmt.Errorf("file size exceeds maximum allowed size")
	}

	// Decode tasks
	var tasks []*models.Task
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&tasks); err != nil {
		if err.Error() == "EOF" {
			// Empty file, return empty slice
			return []*models.Task{}, nil
		}
		return nil, fmt.Errorf("failed to decode tasks: %w", err)
	}

	return tasks, nil
}

// saveTasks writes tasks to the JSON file atomically.
func (s *JSONFileStorage) saveTasks(ctx context.Context, tasks []*models.Task) error {
	// Check for context cancellation
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context error: %w", err)
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(s.tasksFilePath)
	if err := os.MkdirAll(dir, defaultDirMode); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Create a temporary file in the same directory
	tempFile, err := os.CreateTemp(dir, "tasks-*.json.tmp")
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	tempFilePath := tempFile.Name()

	// Clean up temporary file on error
	defer func() {
		tempFile.Close()
		os.Remove(tempFilePath) // Best effort, ignore error
	}()

	// Set secure file permissions
	if err := os.Chmod(tempFilePath, defaultFileMode); err != nil {
		return fmt.Errorf("failed to set file permissions: %w", err)
	}

	// Encode tasks to the temporary file
	encoder := json.NewEncoder(tempFile)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(tasks); err != nil {
		return fmt.Errorf("failed to encode tasks: %w", err)
	}

	// Flush to ensure data is written to disk
	if err := tempFile.Sync(); err != nil {
		return fmt.Errorf("failed to sync file: %w", err)
	}

	// Close the file before renaming
	if err := tempFile.Close(); err != nil {
		return fmt.Errorf("failed to close temporary file: %w", err)
	}

	// Atomically replace the old file with the new one
	if err := os.Rename(tempFilePath, s.tasksFilePath); err != nil {
		return fmt.Errorf("failed to save tasks file: %w", err)
	}

	return nil
}

// copyTask creates a deep copy of a task.
func copyTask(sourceTask *models.Task) *models.Task {
	task := &models.Task{
		ID:          sourceTask.ID,
		Description: sourceTask.Description,
		Completed:   sourceTask.Completed,
		CreatedAt:   sourceTask.CreatedAt,
		UpdatedAt:   sourceTask.UpdatedAt,
		Priority:    sourceTask.Priority,
		Progress:    sourceTask.Progress,
		Tags:        make([]string, len(sourceTask.Tags)),
	}

	// Copy due date if exists
	if sourceTask.DueDate != nil {
		dueDate := *sourceTask.DueDate
		task.DueDate = &dueDate
	}

	// Copy tags
	copy(task.Tags, sourceTask.Tags)

	return task
}
