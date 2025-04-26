package backup

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/YuriLeonel/task-manager/internal/communication"
	"github.com/YuriLeonel/task-manager/pkg/models"
)

func TestConcurrentManager(t *testing.T) {
	// Create temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "task-manager-concurrent-backup-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create communication manager
	comm := communication.NewCommunicationManager()
	defer comm.Stop()

	// Create concurrent backup manager
	manager, err := NewConcurrentManager(tmpDir, 5, comm)
	if err != nil {
		t.Fatalf("Failed to create backup manager: %v", err)
	}
	defer manager.Stop()

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

func TestConcurrentManagerMultipleBackups(t *testing.T) {
	// Create temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "task-manager-concurrent-backup-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create communication manager
	comm := communication.NewCommunicationManager()
	defer comm.Stop()

	// Create concurrent backup manager with limit of 2
	manager, err := NewConcurrentManager(tmpDir, 2, comm)
	if err != nil {
		t.Fatalf("Failed to create backup manager: %v", err)
	}
	defer manager.Stop()

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

func TestConcurrentManagerInvalidBackupID(t *testing.T) {
	// Create temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "task-manager-concurrent-backup-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create communication manager
	comm := communication.NewCommunicationManager()
	defer comm.Stop()

	// Create concurrent backup manager
	manager, err := NewConcurrentManager(tmpDir, 5, comm)
	if err != nil {
		t.Fatalf("Failed to create backup manager: %v", err)
	}
	defer manager.Stop()

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

func TestConcurrentManagerConcurrentOperations(t *testing.T) {
	// Create temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "task-manager-concurrent-backup-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create communication manager
	comm := communication.NewCommunicationManager()
	defer comm.Stop()

	// Create concurrent backup manager
	manager, err := NewConcurrentManager(tmpDir, 5, comm)
	if err != nil {
		t.Fatalf("Failed to create backup manager: %v", err)
	}
	defer manager.Stop()

	ctx := context.Background()

	// Create some test tasks
	tasks := []*models.Task{
		models.NewTask("Task 1"),
		models.NewTask("Task 2"),
	}

	// Run multiple operations concurrently
	done := make(chan struct{})
	go func() {
		for i := 0; i < 10; i++ {
			err := manager.CreateBackup(ctx, tasks)
			if err != nil {
				t.Errorf("Failed to create backup: %v", err)
			}
		}
		done <- struct{}{}
	}()

	go func() {
		for i := 0; i < 10; i++ {
			_, err := manager.ListBackups(ctx)
			if err != nil {
				t.Errorf("Failed to list backups: %v", err)
			}
		}
		done <- struct{}{}
	}()

	// Wait for goroutines to finish
	<-done
	<-done
}
