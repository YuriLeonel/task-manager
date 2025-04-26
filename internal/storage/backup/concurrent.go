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
	"sync"
	"time"

	"github.com/YuriLeonel/task-manager/internal/communication"
	"github.com/YuriLeonel/task-manager/internal/errors"
	"github.com/YuriLeonel/task-manager/pkg/models"
)

// ConcurrentManager handles backup operations with concurrent processing.
type ConcurrentManager struct {
	// backupDir is the directory where backups are stored.
	backupDir string
	// maxBackups is the maximum number of backups to keep.
	maxBackups int
	// backupChan is used to send backup requests.
	backupChan chan backupRequest
	// restoreChan is used to send restore requests.
	restoreChan chan restoreRequest
	// listChan is used to send list requests.
	listChan chan listRequest
	// deleteChan is used to send delete requests.
	deleteChan chan deleteRequest
	// stopChan is used to stop the manager.
	stopChan chan struct{}
	// wg is used to wait for goroutines to finish.
	wg sync.WaitGroup
	// comm is used to communicate with other components.
	comm *communication.CommunicationManager
}

// backupRequest represents a backup request.
type backupRequest struct {
	tasks []*models.Task
	resp  chan error
}

// restoreRequest represents a restore request.
type restoreRequest struct {
	backupID string
	resp     chan restoreResponse
}

// restoreResponse represents a restore response.
type restoreResponse struct {
	tasks []*models.Task
	err   error
}

// listRequest represents a list request.
type listRequest struct {
	resp chan listResponse
}

// listResponse represents a list response.
type listResponse struct {
	backupIDs []string
	err       error
}

// deleteRequest represents a delete request.
type deleteRequest struct {
	backupID string
	resp     chan error
}

// NewConcurrentManager creates a new concurrent backup manager.
func NewConcurrentManager(backupDir string, maxBackups int, comm *communication.CommunicationManager) (*ConcurrentManager, error) {
	if maxBackups <= 0 {
		return nil, errors.NewValidationError("maxBackups", "must be greater than 0")
	}

	// Create backup directory if it doesn't exist
	if err := os.MkdirAll(backupDir, backupDirMode); err != nil {
		return nil, errors.NewPersistenceError(errors.PersistenceErrorCode, "failed to create backup directory", err)
	}

	cm := &ConcurrentManager{
		backupDir:   backupDir,
		maxBackups:  maxBackups,
		backupChan:  make(chan backupRequest, 10),
		restoreChan: make(chan restoreRequest, 10),
		listChan:    make(chan listRequest, 10),
		deleteChan:  make(chan deleteRequest, 10),
		stopChan:    make(chan struct{}),
		comm:        comm,
	}

	// Start the manager goroutine
	cm.wg.Add(1)
	go cm.run()

	return cm, nil
}

// run is the main goroutine that processes backup operations.
func (cm *ConcurrentManager) run() {
	defer cm.wg.Done()

	for {
		select {
		case <-cm.stopChan:
			return
		case req := <-cm.backupChan:
			req.resp <- cm.createBackup(req.tasks)
		case req := <-cm.restoreChan:
			tasks, err := cm.restoreBackup(req.backupID)
			req.resp <- restoreResponse{tasks, err}
		case req := <-cm.listChan:
			backupIDs, err := cm.listBackups()
			req.resp <- listResponse{backupIDs, err}
		case req := <-cm.deleteChan:
			req.resp <- cm.deleteBackup(req.backupID)
		}
	}
}

// Stop stops the concurrent backup manager.
func (cm *ConcurrentManager) Stop() {
	close(cm.stopChan)
	cm.wg.Wait()
}

// CreateBackup creates a new backup of the given tasks.
func (cm *ConcurrentManager) CreateBackup(ctx context.Context, tasks []*models.Task) error {
	// Check for context cancellation
	if err := ctx.Err(); err != nil {
		return errors.NewPersistenceError(errors.PersistenceErrorCode, "context error", err)
	}

	resp := make(chan error)
	cm.backupChan <- backupRequest{tasks, resp}
	return <-resp
}

// RestoreBackup restores tasks from a backup.
func (cm *ConcurrentManager) RestoreBackup(ctx context.Context, backupID string) ([]*models.Task, error) {
	// Check for context cancellation
	if err := ctx.Err(); err != nil {
		return nil, errors.NewPersistenceError(errors.PersistenceErrorCode, "context error", err)
	}

	resp := make(chan restoreResponse)
	cm.restoreChan <- restoreRequest{backupID, resp}
	result := <-resp
	return result.tasks, result.err
}

// ListBackups lists all available backups.
func (cm *ConcurrentManager) ListBackups(ctx context.Context) ([]string, error) {
	// Check for context cancellation
	if err := ctx.Err(); err != nil {
		return nil, errors.NewPersistenceError(errors.PersistenceErrorCode, "context error", err)
	}

	resp := make(chan listResponse)
	cm.listChan <- listRequest{resp}
	result := <-resp
	return result.backupIDs, result.err
}

// DeleteBackup deletes a backup.
func (cm *ConcurrentManager) DeleteBackup(ctx context.Context, backupID string) error {
	// Check for context cancellation
	if err := ctx.Err(); err != nil {
		return errors.NewPersistenceError(errors.PersistenceErrorCode, "context error", err)
	}

	resp := make(chan error)
	cm.deleteChan <- deleteRequest{backupID, resp}
	return <-resp
}

// createBackup creates a new backup of the given tasks.
func (cm *ConcurrentManager) createBackup(tasks []*models.Task) error {
	// Create backup
	backup := &Backup{
		ID:        generateBackupID(),
		CreatedAt: time.Now(),
		Tasks:     tasks,
	}

	// Create backup file
	backupPath := filepath.Join(cm.backupDir, fmt.Sprintf("%s.json", backup.ID))
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

	// Notify about backup creation
	if err := cm.comm.Publish(context.Background(), communication.BackupCreated, backup); err != nil {
		// Log error but don't fail the backup creation
		fmt.Printf("Warning: failed to notify about backup creation: %v\n", err)
	}

	// Clean up old backups
	if err := cm.cleanupOldBackups(); err != nil {
		// Log error but don't fail the backup creation
		fmt.Printf("Warning: failed to clean up old backups: %v\n", err)
	}

	return nil
}

// restoreBackup restores tasks from a backup.
func (cm *ConcurrentManager) restoreBackup(backupID string) ([]*models.Task, error) {
	// Read backup file
	backupPath := filepath.Join(cm.backupDir, fmt.Sprintf("%s.json", backupID))
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

	// Notify about backup restoration
	if err := cm.comm.Publish(context.Background(), communication.BackupRestored, backup); err != nil {
		// Log error but don't fail the backup restoration
		fmt.Printf("Warning: failed to notify about backup restoration: %v\n", err)
	}

	return backup.Tasks, nil
}

// listBackups lists all available backups.
func (cm *ConcurrentManager) listBackups() ([]string, error) {
	// Read backup directory
	entries, err := os.ReadDir(cm.backupDir)
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

// deleteBackup deletes a backup.
func (cm *ConcurrentManager) deleteBackup(backupID string) error {
	// Delete backup file
	backupPath := filepath.Join(cm.backupDir, fmt.Sprintf("%s.json", backupID))
	if err := os.Remove(backupPath); err != nil {
		if os.IsNotExist(err) {
			return errors.NewTaskError(errors.TaskNotFound, "backup not found", nil)
		}
		return errors.NewPersistenceError(errors.PersistenceErrorCode, "failed to delete backup file", err)
	}

	return nil
}

// cleanupOldBackups removes old backups to maintain the maximum number of backups.
func (cm *ConcurrentManager) cleanupOldBackups() error {
	// List all backups
	backupIDs, err := cm.listBackups()
	if err != nil {
		return err
	}

	// If we have more backups than the maximum, remove the oldest ones
	if len(backupIDs) > cm.maxBackups {
		// Sort backups by creation time
		type backupInfo struct {
			ID        string
			CreatedAt time.Time
		}
		backups := make([]backupInfo, len(backupIDs))
		for i, id := range backupIDs {
			backupPath := filepath.Join(cm.backupDir, fmt.Sprintf("%s.json", id))
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
		for i := 0; i < len(backups)-cm.maxBackups; i++ {
			if err := cm.deleteBackup(backups[i].ID); err != nil {
				// Log error but continue with cleanup
				fmt.Printf("Warning: failed to delete old backup %s: %v\n", backups[i].ID, err)
			}
		}
	}

	return nil
}
