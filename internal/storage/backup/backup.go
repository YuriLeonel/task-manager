// Package backup provides functionality for managing task backups.
// It includes secure backup creation, restoration, and management with
// proper error handling and validation.
package backup

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/YuriLeonel/task-manager/internal/errors"
	"github.com/YuriLeonel/task-manager/pkg/models"
)

const (
	// backupFileMode is the file mode for backup files (0600)
	backupFileMode = 0600
	// backupDirMode is the directory mode for backup directories (0700)
	backupDirMode = 0700
	// maxBackupAge is the maximum age of backups (30 days)
	maxBackupAge = 30 * 24 * time.Hour
)

// Backup represents a task backup with metadata.
type Backup struct {
	// ID is the unique identifier for the backup.
	ID string
	// CreatedAt is the timestamp when the backup was created.
	CreatedAt time.Time
	// Tasks is the list of tasks in the backup.
	Tasks []*models.Task
}

// Manager handles backup operations for tasks.
type Manager struct {
	// backupDir is the directory where backups are stored.
	backupDir string
	// maxBackups is the maximum number of backups to keep.
	maxBackups int
}

// NewManager creates a new backup manager.
func NewManager(backupDir string, maxBackups int) (*Manager, error) {
	if maxBackups <= 0 {
		return nil, errors.NewValidationError("maxBackups", "must be greater than 0")
	}

	// Create backup directory if it doesn't exist
	if err := os.MkdirAll(backupDir, backupDirMode); err != nil {
		return nil, errors.NewPersistenceError(errors.PersistenceErrorCode, "failed to create backup directory", err)
	}

	return &Manager{
		backupDir:  backupDir,
		maxBackups: maxBackups,
	}, nil
}

// CreateBackup creates a new backup of the given tasks.
func (m *Manager) CreateBackup(ctx context.Context, tasks []*models.Task) error {
	// Check for context cancellation
	if err := ctx.Err(); err != nil {
		return errors.NewPersistenceError(errors.PersistenceErrorCode, "context error", err)
	}

	// Create backup
	backup := &Backup{
		ID:        generateBackupID(),
		CreatedAt: time.Now(),
		Tasks:     tasks,
	}

	// Create backup file
	backupPath := filepath.Join(m.backupDir, fmt.Sprintf("%s.json", backup.ID))
	file, err := os.Create(backupPath)
	if err != nil {
		return errors.NewPersistenceError(errors.PersistenceErrorCode, "failed to create backup file", err)
	}
	defer file.Close()

	// Set secure file permissions
	if err := os.Chmod(backupPath, backupFileMode); err != nil {
		return errors.NewPersistenceError(errors.PersistenceErrorCode, "failed to set backup file permissions", err)
	}

	// Encode backup to file
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(backup); err != nil {
		return errors.NewPersistenceError(errors.PersistenceErrorCode, "failed to encode backup", err)
	}

	// Clean up old backups
	if err := m.cleanupOldBackups(ctx); err != nil {
		// Log error but don't fail the backup creation
		fmt.Printf("Warning: failed to clean up old backups: %v\n", err)
	}

	return nil
}

// RestoreBackup restores tasks from a backup.
func (m *Manager) RestoreBackup(ctx context.Context, backupID string) ([]*models.Task, error) {
	// Check for context cancellation
	if err := ctx.Err(); err != nil {
		return nil, errors.NewPersistenceError(errors.PersistenceErrorCode, "context error", err)
	}

	// Read backup file
	backupPath := filepath.Join(m.backupDir, fmt.Sprintf("%s.json", backupID))
	file, err := os.Open(backupPath)
	if err != nil {
		return nil, errors.NewPersistenceError(errors.PersistenceErrorCode, "failed to open backup file", err)
	}
	defer file.Close()

	// Decode backup
	var backup Backup
	if err := json.NewDecoder(file).Decode(&backup); err != nil {
		return nil, errors.NewPersistenceError(errors.PersistenceErrorCode, "failed to decode backup", err)
	}

	return backup.Tasks, nil
}

// ListBackups lists all available backups.
func (m *Manager) ListBackups(ctx context.Context) ([]string, error) {
	// Check for context cancellation
	if err := ctx.Err(); err != nil {
		return nil, errors.NewPersistenceError(errors.PersistenceErrorCode, "context error", err)
	}

	// Read backup directory
	entries, err := os.ReadDir(m.backupDir)
	if err != nil {
		return nil, errors.NewPersistenceError(errors.PersistenceErrorCode, "failed to read backup directory", err)
	}

	// Filter and sort backup files
	var backupIDs []string
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" {
			backupID := entry.Name()[:len(entry.Name())-5] // Remove .json extension
			backupIDs = append(backupIDs, backupID)
		}
	}

	return backupIDs, nil
}

// DeleteBackup deletes a backup.
func (m *Manager) DeleteBackup(ctx context.Context, backupID string) error {
	// Check for context cancellation
	if err := ctx.Err(); err != nil {
		return errors.NewPersistenceError(errors.PersistenceErrorCode, "context error", err)
	}

	// Delete backup file
	backupPath := filepath.Join(m.backupDir, fmt.Sprintf("%s.json", backupID))
	if err := os.Remove(backupPath); err != nil {
		if os.IsNotExist(err) {
			return errors.NewTaskError(errors.TaskNotFound, "backup not found", nil)
		}
		return errors.NewPersistenceError(errors.PersistenceErrorCode, "failed to delete backup file", err)
	}

	return nil
}

// cleanupOldBackups removes old backups to maintain the maximum number of backups.
func (m *Manager) cleanupOldBackups(ctx context.Context) error {
	// List all backups
	backupIDs, err := m.ListBackups(ctx)
	if err != nil {
		return err
	}

	// If we have more backups than the maximum, remove the oldest ones
	if len(backupIDs) > m.maxBackups {
		// Sort backups by creation time
		type backupInfo struct {
			ID        string
			CreatedAt time.Time
		}
		backups := make([]backupInfo, len(backupIDs))
		for i, id := range backupIDs {
			backupPath := filepath.Join(m.backupDir, fmt.Sprintf("%s.json", id))
			file, err := os.Open(backupPath)
			if err != nil {
				continue // Skip if we can't read the backup
			}

			var backup Backup
			if err := json.NewDecoder(file).Decode(&backup); err != nil {
				file.Close()
				continue // Skip if we can't decode the backup
			}
			file.Close()

			backups[i] = backupInfo{
				ID:        id,
				CreatedAt: backup.CreatedAt,
			}
		}

		// Sort by creation time (oldest first)
		sort.Slice(backups, func(i, j int) bool {
			return backups[i].CreatedAt.Before(backups[j].CreatedAt)
		})

		// Remove oldest backups
		for i := 0; i < len(backups)-m.maxBackups; i++ {
			if err := m.DeleteBackup(ctx, backups[i].ID); err != nil {
				// Log error but continue with cleanup
				fmt.Printf("Warning: failed to delete old backup %s: %v\n", backups[i].ID, err)
			}
		}
	}

	return nil
}

// generateBackupID creates a unique backup ID based on timestamp.
func generateBackupID() string {
	return fmt.Sprintf("backup-%d", time.Now().UnixNano())
}
