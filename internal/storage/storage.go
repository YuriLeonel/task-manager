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
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

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
	SaveTask(task *models.Task) error
	// GetTask retrieves a task by its ID.
	// Returns a copy of the task to prevent external modification.
	GetTask(id string) (*models.Task, error)
	// GetAllTasks retrieves all tasks.
	// Returns copies of all tasks to prevent external modification.
	GetAllTasks() ([]*models.Task, error)
	// UpdateTask updates an existing task.
	// It validates the task and ensures atomic operations.
	UpdateTask(task *models.Task) error
	// DeleteTask removes a task by its ID.
	// Ensures atomic operations during deletion.
	DeleteTask(id string) error
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
func (s *JSONFileStorage) SaveTask(newTask *models.Task) error {
	if err := s.validateTask(newTask); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Read existing tasks
	existingTasks, err := s.readTasks()
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
	return s.saveTasks(existingTasks)
}

// GetTask retrieves a task by its ID.
func (s *JSONFileStorage) GetTask(taskID string) (*models.Task, error) {
	if taskID == "" {
		return nil, fmt.Errorf("task ID cannot be empty")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	tasks, err := s.readTasks()
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
func (s *JSONFileStorage) GetAllTasks() ([]*models.Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tasks, err := s.readTasks()
	if err != nil {
		return nil, err
	}

	// Return copies to prevent external modification
	taskCopies := make([]*models.Task, len(tasks))
	for i, task := range tasks {
		taskCopies[i] = copyTask(task)
	}
	return taskCopies, nil
}

// UpdateTask updates an existing task in the JSON file.
func (s *JSONFileStorage) UpdateTask(updatedTask *models.Task) error {
	if err := s.validateTask(updatedTask); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	tasks, err := s.readTasks()
	if err != nil {
		return err
	}

	for i, existingTask := range tasks {
		if existingTask.ID == updatedTask.ID {
			// Preserve original creation time
			updatedTask.CreatedAt = existingTask.CreatedAt
			// Update the modification time
			updatedTask.UpdatedAt = time.Now()
			tasks[i] = updatedTask
			return s.saveTasks(tasks)
		}
	}
	return fmt.Errorf("task not found: %s", updatedTask.ID)
}

// DeleteTask removes a task by its ID from the JSON file.
func (s *JSONFileStorage) DeleteTask(taskID string) error {
	if taskID == "" {
		return fmt.Errorf("task ID cannot be empty")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	tasks, err := s.readTasks()
	if err != nil {
		return err
	}

	for i, task := range tasks {
		if task.ID == taskID {
			tasks = append(tasks[:i], tasks[i+1:]...)
			return s.saveTasks(tasks)
		}
	}
	return fmt.Errorf("task not found: %s", taskID)
}

// readTasks reads all tasks from the JSON file.
func (s *JSONFileStorage) readTasks() ([]*models.Task, error) {
	// Check if file exists and get its size
	fileInfo, err := os.Stat(s.tasksFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []*models.Task{}, nil
		}
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	// Check file size limit
	if fileInfo.Size() > maxFileSize {
		return nil, fmt.Errorf("file size exceeds maximum allowed size of %d bytes", maxFileSize)
	}

	fileContent, err := os.ReadFile(s.tasksFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var tasks []*models.Task
	if err := json.Unmarshal(fileContent, &tasks); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tasks: %w", err)
	}

	return tasks, nil
}

// saveTasks saves all tasks to the JSON file.
func (s *JSONFileStorage) saveTasks(tasks []*models.Task) error {
	jsonData, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal tasks: %w", err)
	}

	// Ensure the directory exists with secure permissions
	if err := os.MkdirAll(filepath.Dir(s.tasksFilePath), defaultDirMode); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Create a temporary file in the same directory
	tempFilePath := s.tasksFilePath + ".tmp"
	tempFile, err := os.OpenFile(tempFilePath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, defaultFileMode)
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}

	// Write data to temporary file
	if _, err := tempFile.Write(jsonData); err != nil {
		tempFile.Close()
		os.Remove(tempFilePath)
		return fmt.Errorf("failed to write to temporary file: %w", err)
	}

	// Close the temporary file
	if err := tempFile.Close(); err != nil {
		os.Remove(tempFilePath)
		return fmt.Errorf("failed to close temporary file: %w", err)
	}

	// Rename the temporary file to the actual file
	if err := os.Rename(tempFilePath, s.tasksFilePath); err != nil {
		os.Remove(tempFilePath)
		return fmt.Errorf("failed to rename temporary file: %w", err)
	}

	return nil
}

// copyTask creates a deep copy of a task to prevent external modification.
func copyTask(sourceTask *models.Task) *models.Task {
	if sourceTask == nil {
		return nil
	}
	return &models.Task{
		ID:          sourceTask.ID,
		Description: sourceTask.Description,
		Completed:   sourceTask.Completed,
		CreatedAt:   sourceTask.CreatedAt,
		UpdatedAt:   sourceTask.UpdatedAt,
		DueDate:     sourceTask.DueDate,
		Priority:    sourceTask.Priority,
	}
}
