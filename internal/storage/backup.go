package storage

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// BackupManager handles automatic backups of task data.
// It runs in a background goroutine and performs backups
// on a schedule or when triggered manually.
type BackupManager struct {
	// sourceFile is the path to the file to be backed up
	sourceFile string
	// backupDir is the directory where backups will be stored
	backupDir string
	// interval is the time between automatic backups
	interval time.Duration
	// maxBackups is the maximum number of backups to keep
	maxBackups int
	// stopChan is used to signal the backup goroutine to stop
	stopChan chan struct{}
	// wg is used to wait for the backup goroutine to finish
	wg sync.WaitGroup
	// backupChan is used to trigger manual backups
	backupChan chan struct{}
	// mu provides thread safety
	mu sync.Mutex
}

// NewBackupManager creates a new backup manager.
func NewBackupManager(sourceFile, backupDir string, interval time.Duration, maxBackups int) (*BackupManager, error) {
	// Verify that source file exists
	if _, err := os.Stat(sourceFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("source file does not exist: %w", err)
	}

	// Create backup directory if it doesn't exist
	if err := os.MkdirAll(backupDir, defaultDirMode); err != nil {
		return nil, fmt.Errorf("failed to create backup directory: %w", err)
	}

	return &BackupManager{
		sourceFile: sourceFile,
		backupDir:  backupDir,
		interval:   interval,
		maxBackups: maxBackups,
		stopChan:   make(chan struct{}),
		backupChan: make(chan struct{}, 1),
	}, nil
}

// Start begins the backup process in a background goroutine.
func (bm *BackupManager) Start() {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	// Don't start if already running
	select {
	case <-bm.stopChan:
		// Channel is closed, need to recreate it
		bm.stopChan = make(chan struct{})
	default:
		// Already running
		return
	}

	bm.wg.Add(1)
	go bm.backupLoop()

	log.Println("Backup manager started")
}

// Stop stops the backup process.
func (bm *BackupManager) Stop() {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	select {
	case <-bm.stopChan:
		// Already stopped
		return
	default:
		close(bm.stopChan)
	}

	bm.wg.Wait()
	log.Println("Backup manager stopped")
}

// TriggerBackup triggers an immediate backup.
func (bm *BackupManager) TriggerBackup() {
	select {
	case bm.backupChan <- struct{}{}:
		// Signal sent
	default:
		// Channel full, backup already pending
	}
}

// backupLoop is the main backup goroutine.
func (bm *BackupManager) backupLoop() {
	defer bm.wg.Done()

	ticker := time.NewTicker(bm.interval)
	defer ticker.Stop()

	for {
		select {
		case <-bm.stopChan:
			return
		case <-ticker.C:
			if err := bm.performBackup(); err != nil {
				log.Printf("Scheduled backup failed: %v", err)
			}
		case <-bm.backupChan:
			if err := bm.performBackup(); err != nil {
				log.Printf("Manual backup failed: %v", err)
			}
		}
	}
}

// performBackup creates a backup of the source file.
func (bm *BackupManager) performBackup() error {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Use the current time for the backup filename
	backupTime := time.Now().Format("20060102-150405")
	backupPath := filepath.Join(bm.backupDir, fmt.Sprintf("tasks-%s.json.bak", backupTime))

	// Copy the source file to the backup file
	if err := bm.copyFile(ctx, bm.sourceFile, backupPath); err != nil {
		return err
	}

	// Rotate old backups
	if err := bm.rotateBackups(ctx); err != nil {
		log.Printf("Failed to rotate backups: %v", err)
	}

	log.Printf("Backup created at %s", backupPath)
	return nil
}

// copyFile copies a file from src to dst.
func (bm *BackupManager) copyFile(ctx context.Context, src, dst string) error {
	// Check for context cancellation
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context error: %w", err)
	}

	// Open source file
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	// Create destination file
	dstFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_EXCL, defaultFileMode)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dstFile.Close()

	// Copy the file
	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	return nil
}

// rotateBackups removes old backups when there are too many.
func (bm *BackupManager) rotateBackups(ctx context.Context) error {
	// Check for context cancellation
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context error: %w", err)
	}

	// List all backup files
	pattern := filepath.Join(bm.backupDir, "tasks-*.json.bak")
	backupFiles, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("failed to list backup files: %w", err)
	}

	// If we have more backups than allowed, delete the oldest ones
	if len(backupFiles) > bm.maxBackups {
		// Sort backup files by name (which includes the timestamp)
		// This will sort from oldest to newest
		for i := 0; i < len(backupFiles)-bm.maxBackups; i++ {
			if err := os.Remove(backupFiles[i]); err != nil {
				log.Printf("Failed to remove old backup %s: %v", backupFiles[i], err)
			} else {
				log.Printf("Removed old backup %s", backupFiles[i])
			}
		}
	}

	return nil
}
