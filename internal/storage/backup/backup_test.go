package backup

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/YuriLeonel/task-manager/pkg/models"
)

func TestBackupManager(t *testing.T) {
	// Create temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "task-manager-backup-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create backup manager
	manager, err := NewManager(tmpDir, 5)
	if err != nil {
		t.Fatalf("Failed to create backup manager: %v", err)
	}

	ctx := context.Background()

	// Create some test tasks
	tasks := []*models.Task{
		models.NewTask("Task 1"),
		models.NewTask("Task 2"),
		models.NewTask("Task 3"),
	}

	// Test creating a backup
	err = manager.CreateBackup(ctx, tasks)
	if err != nil {
		t.Fatalf("Failed to create backup: %v", err)
	}

	// Test listing backups
	backupIDs, err := manager.ListBackups(ctx)
	if err != nil {
		t.Fatalf("Failed to list backups: %v", err)
	}
	if len(backupIDs) != 1 {
		t.Errorf("Expected 1 backup, got %d", len(backupIDs))
	}

	// Test restoring backup
	restoredTasks, err := manager.RestoreBackup(ctx, backupIDs[0])
	if err != nil {
		t.Fatalf("Failed to restore backup: %v", err)
	}
	if len(restoredTasks) != len(tasks) {
		t.Errorf("Expected %d tasks, got %d", len(tasks), len(restoredTasks))
	}

	// Test deleting backup
	err = manager.DeleteBackup(ctx, backupIDs[0])
	if err != nil {
		t.Fatalf("Failed to delete backup: %v", err)
	}

	// Verify deletion
	backupIDs, err = manager.ListBackups(ctx)
	if err != nil {
		t.Fatalf("Failed to list backups: %v", err)
	}
	if len(backupIDs) != 0 {
		t.Error("Backup should be deleted")
	}
}

func TestBackupLimit(t *testing.T) {
	// Create temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "task-manager-backup-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create backup manager with limit of 2
	manager, err := NewManager(tmpDir, 2)
	if err != nil {
		t.Fatalf("Failed to create backup manager: %v", err)
	}

	ctx := context.Background()

	// Create some test tasks
	tasks := []*models.Task{
		models.NewTask("Task 1"),
		models.NewTask("Task 2"),
	}

	// Create three backups
	for i := 0; i < 3; i++ {
		err = manager.CreateBackup(ctx, tasks)
		if err != nil {
			t.Fatalf("Failed to create backup %d: %v", i+1, err)
		}
		time.Sleep(time.Millisecond) // Ensure different timestamps
	}

	// Check that only 2 backups exist
	backupIDs, err := manager.ListBackups(ctx)
	if err != nil {
		t.Fatalf("Failed to list backups: %v", err)
	}
	if len(backupIDs) != 2 {
		t.Errorf("Expected 2 backups, got %d", len(backupIDs))
	}
}

func TestInvalidBackupID(t *testing.T) {
	// Create temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "task-manager-backup-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create backup manager
	manager, err := NewManager(tmpDir, 5)
	if err != nil {
		t.Fatalf("Failed to create backup manager: %v", err)
	}

	ctx := context.Background()

	// Test restoring non-existent backup
	_, err = manager.RestoreBackup(ctx, "non-existent")
	if err == nil {
		t.Error("Expected error when restoring non-existent backup")
	}

	// Test deleting non-existent backup
	err = manager.DeleteBackup(ctx, "non-existent")
	if err == nil {
		t.Error("Expected error when deleting non-existent backup")
	}
}

func TestBackupFilePermissions(t *testing.T) {
	// Create temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "task-manager-backup-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create backup manager
	manager, err := NewManager(tmpDir, 5)
	if err != nil {
		t.Fatalf("Failed to create backup manager: %v", err)
	}

	ctx := context.Background()

	// Create test tasks
	tasks := []*models.Task{
		models.NewTask("Test task"),
	}

	// Create backup
	err = manager.CreateBackup(ctx, tasks)
	if err != nil {
		t.Fatalf("Failed to create backup: %v", err)
	}

	// List backups to get the backup ID
	backupIDs, err := manager.ListBackups(ctx)
	if err != nil {
		t.Fatalf("Failed to list backups: %v", err)
	}
	if len(backupIDs) != 1 {
		t.Fatalf("Expected 1 backup, got %d", len(backupIDs))
	}

	// Check backup file permissions
	backupPath := filepath.Join(tmpDir, backupIDs[0]+".json")
	fileInfo, err := os.Stat(backupPath)
	if err != nil {
		t.Fatalf("Failed to get backup file info: %v", err)
	}

	// Check file mode (should be 0600)
	if fileInfo.Mode() != backupFileMode {
		t.Errorf("Expected file mode %v, got %v", backupFileMode, fileInfo.Mode())
	}
}
