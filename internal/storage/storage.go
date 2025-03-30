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
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"slices"

	"github.com/YuriLeonel/task-manager/internal/errors"
	"github.com/YuriLeonel/task-manager/internal/storage/backup"
	"github.com/YuriLeonel/task-manager/pkg/models"
	"github.com/YuriLeonel/task-manager/pkg/utils"
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
	// backupManager handles backup operations.
	backupManager *backup.Manager
}

// NewJSONFileStorage creates a new JSON file storage instance.
func NewJSONFileStorage(tasksFilePath string, maxTasks int, backupDir string, maxBackups int) (*JSONFileStorage, error) {
	// Validate and sanitize the filepath
	absolutePath, err := filepath.Abs(tasksFilePath)
	if err != nil {
		return nil, errors.NewPersistenceError(errors.PersistenceErrorCode, "invalid filepath", err)
	}

	// Ensure the filepath is within allowed directories
	if err := validateFilePath(absolutePath); err != nil {
		return nil, err
	}

	// Validate maxTasks
	if maxTasks <= 0 {
		return nil, errors.NewValidationError("maxTasks", "must be greater than 0")
	}

	// Create backup manager
	backupManager, err := backup.NewManager(backupDir, maxBackups)
	if err != nil {
		return nil, err
	}

	return &JSONFileStorage{
		tasksFilePath: absolutePath,
		maxTasks:      maxTasks,
		backupManager: backupManager,
	}, nil
}

// validateFilePath ensures the filepath is secure and within allowed directories.
func validateFilePath(filePath string) error {
	// Get user's home directory
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		return errors.NewPersistenceError(errors.PersistenceErrorCode, "failed to get home directory", err)
	}

	// Ensure the path is within the user's home directory
	if !strings.HasPrefix(filePath, userHomeDir) {
		return errors.NewValidationError("filepath", "must be within user's home directory")
	}

	// Check for path traversal attempts
	if strings.Contains(filePath, "..") {
		return errors.NewValidationError("filepath", "path traversal not allowed")
	}

	return nil
}

// validateTask checks if a task is valid before saving.
func (s *JSONFileStorage) validateTask(task *models.Task) error {
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

// SaveTask saves a task to the JSON file.
func (s *JSONFileStorage) SaveTask(ctx context.Context, newTask *models.Task) error {
	// Check for context cancellation
	if err := ctx.Err(); err != nil {
		return errors.NewPersistenceError(errors.PersistenceErrorCode, "context error", err)
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
		return errors.NewTaskError(errors.TaskLimitExceeded, "maximum number of tasks reached", nil)
	}

	// Check for duplicate ID
	for _, existingTask := range existingTasks {
		if existingTask.ID == newTask.ID {
			return errors.NewTaskError(errors.TaskAlreadyExists, "task with ID already exists", nil)
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
		return nil, errors.NewPersistenceError(errors.PersistenceErrorCode, "context error", err)
	}

	if taskID == "" {
		return nil, errors.NewValidationError("taskID", "cannot be empty")
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
			return utils.CopyTask(task), nil
		}
	}
	return nil, errors.NewTaskError(errors.TaskNotFound, "task not found", nil)
}

// GetAllTasks retrieves all tasks from the JSON file.
func (s *JSONFileStorage) GetAllTasks(ctx context.Context) ([]*models.Task, error) {
	// Check for context cancellation
	if err := ctx.Err(); err != nil {
		return nil, errors.NewPersistenceError(errors.PersistenceErrorCode, "context error", err)
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
		tasksCopy[i] = utils.CopyTask(task)
	}
	return tasksCopy, nil
}

// UpdateTask updates an existing task in the JSON file.
func (s *JSONFileStorage) UpdateTask(ctx context.Context, updatedTask *models.Task) error {
	// Check for context cancellation
	if err := ctx.Err(); err != nil {
		return errors.NewPersistenceError(errors.PersistenceErrorCode, "context error", err)
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
		return errors.NewTaskError(errors.TaskNotFound, "task not found", nil)
	}

	return s.saveTasks(ctx, tasks)
}

// DeleteTask removes a task from the JSON file.
func (s *JSONFileStorage) DeleteTask(ctx context.Context, taskID string) error {
	// Check for context cancellation
	if err := ctx.Err(); err != nil {
		return errors.NewPersistenceError(errors.PersistenceErrorCode, "context error", err)
	}

	if taskID == "" {
		return errors.NewValidationError("taskID", "cannot be empty")
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
		return errors.NewTaskError(errors.TaskNotFound, "task not found", nil)
	}

	// Remove task by index
	tasks = slices.Delete(tasks, index, index+1)
	return s.saveTasks(ctx, tasks)
}

// readTasks reads tasks from the JSON file.
func (s *JSONFileStorage) readTasks(ctx context.Context) ([]*models.Task, error) {
	// Check for context cancellation
	if err := ctx.Err(); err != nil {
		return nil, errors.NewPersistenceError(errors.PersistenceErrorCode, "context error", err)
	}

	// If file doesn't exist, return empty slice
	if _, err := os.Stat(s.tasksFilePath); os.IsNotExist(err) {
		return []*models.Task{}, nil
	}

	// Open the file
	file, err := os.Open(s.tasksFilePath)
	if err != nil {
		return nil, errors.NewPersistenceError(errors.PersistenceErrorCode, "failed to open tasks file", err)
	}
	defer file.Close()

	// Check file size
	fileInfo, err := file.Stat()
	if err != nil {
		return nil, errors.NewPersistenceError(errors.PersistenceErrorCode, "failed to get file info", err)
	}
	if fileInfo.Size() > maxFileSize {
		return nil, errors.NewValidationError("file size", "exceeds maximum allowed size")
	}

	// Decode tasks
	var tasks []*models.Task
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&tasks); err != nil {
		if err.Error() == "EOF" {
			// Empty file, return empty slice
			return []*models.Task{}, nil
		}
		return nil, errors.NewPersistenceError(errors.PersistenceErrorCode, "failed to decode tasks", err)
	}

	return tasks, nil
}

// saveTasks writes tasks to the JSON file atomically.
func (s *JSONFileStorage) saveTasks(ctx context.Context, tasks []*models.Task) error {
	// Check for context cancellation
	if err := ctx.Err(); err != nil {
		return errors.NewPersistenceError(errors.PersistenceErrorCode, "context error", err)
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(s.tasksFilePath)
	if err := os.MkdirAll(dir, defaultDirMode); err != nil {
		return errors.NewPersistenceError(errors.PersistenceErrorCode, "failed to create directory", err)
	}

	// Create a temporary file in the same directory
	tempFile, err := os.CreateTemp(dir, "tasks-*.json.tmp")
	if err != nil {
		return errors.NewPersistenceError(errors.PersistenceErrorCode, "failed to create temporary file", err)
	}
	tempFilePath := tempFile.Name()

	// Clean up temporary file on error
	defer func() {
		tempFile.Close()
		os.Remove(tempFilePath) // Best effort, ignore error
	}()

	// Set secure file permissions
	if err := os.Chmod(tempFilePath, defaultFileMode); err != nil {
		return errors.NewPersistenceError(errors.PersistenceErrorCode, "failed to set file permissions", err)
	}

	// Encode tasks to the temporary file
	encoder := json.NewEncoder(tempFile)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(tasks); err != nil {
		return errors.NewPersistenceError(errors.PersistenceErrorCode, "failed to encode tasks", err)
	}

	// Flush to ensure data is written to disk
	if err := tempFile.Sync(); err != nil {
		return errors.NewPersistenceError(errors.PersistenceErrorCode, "failed to sync file", err)
	}

	// Close the file before renaming
	if err := tempFile.Close(); err != nil {
		return errors.NewPersistenceError(errors.PersistenceErrorCode, "failed to close temporary file", err)
	}

	// Atomically replace the old file with the new one
	if err := os.Rename(tempFilePath, s.tasksFilePath); err != nil {
		return errors.NewPersistenceError(errors.PersistenceErrorCode, "failed to save tasks file", err)
	}

	return nil
}

// CreateBackup creates a backup of the current tasks.
func (s *JSONFileStorage) CreateBackup(ctx context.Context) error {
	// Check for context cancellation
	if err := ctx.Err(); err != nil {
		return errors.NewPersistenceError(errors.PersistenceErrorCode, "context error", err)
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	// Read current tasks
	tasks, err := s.readTasks(ctx)
	if err != nil {
		return err
	}

	// Create backup
	return s.backupManager.CreateBackup(ctx, tasks)
}

// RestoreBackup restores tasks from a backup.
func (s *JSONFileStorage) RestoreBackup(ctx context.Context, backupID string) error {
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

	// Save restored tasks
	return s.saveTasks(ctx, tasks)
}

// ListBackups lists all available backups.
func (s *JSONFileStorage) ListBackups(ctx context.Context) ([]string, error) {
	// Check for context cancellation
	if err := ctx.Err(); err != nil {
		return nil, errors.NewPersistenceError(errors.PersistenceErrorCode, "context error", err)
	}

	return s.backupManager.ListBackups(ctx)
}

// DeleteBackup deletes a backup.
func (s *JSONFileStorage) DeleteBackup(ctx context.Context, backupID string) error {
	// Check for context cancellation
	if err := ctx.Err(); err != nil {
		return errors.NewPersistenceError(errors.PersistenceErrorCode, "context error", err)
	}

	return s.backupManager.DeleteBackup(ctx, backupID)
}
