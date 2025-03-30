// Package storage provides interfaces and implementations for task persistence.
package storage

import (
	"context"

	"github.com/YuriLeonel/task-manager/pkg/models"
)

// TaskReader defines the interface for reading tasks.
type TaskReader interface {
	// GetTask retrieves a task by its ID.
	// Returns a copy of the task to prevent external modification.
	GetTask(ctx context.Context, id string) (*models.Task, error)
	// GetAllTasks retrieves all tasks.
	// Returns copies of all tasks to prevent external modification.
	GetAllTasks(ctx context.Context) ([]*models.Task, error)
}

// TaskWriter defines the interface for writing tasks.
type TaskWriter interface {
	// SaveTask saves a task to the storage.
	// It validates the task before saving and ensures atomic operations.
	SaveTask(ctx context.Context, task *models.Task) error
	// UpdateTask updates an existing task.
	// It validates the task and ensures atomic operations.
	UpdateTask(ctx context.Context, task *models.Task) error
	// DeleteTask removes a task by its ID.
	// Ensures atomic operations during deletion.
	DeleteTask(ctx context.Context, id string) error
}

// TaskStorage combines TaskReader and TaskWriter interfaces.
type TaskStorage interface {
	TaskReader
	TaskWriter
}

// BackupOperations defines the interface for managing task backups.
type BackupOperations interface {
	// CreateBackup creates a backup of the current tasks.
	CreateBackup(ctx context.Context) error
	// RestoreBackup restores tasks from a backup.
	RestoreBackup(ctx context.Context, backupID string) error
	// ListBackups lists all available backups.
	ListBackups(ctx context.Context) ([]string, error)
	// DeleteBackup deletes a backup.
	DeleteBackup(ctx context.Context, backupID string) error
}

// Storage combines TaskStorage and BackupOperations interfaces.
type Storage interface {
	TaskStorage
	BackupOperations
}
