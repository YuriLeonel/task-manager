// Package config provides configuration management for the task manager.
// It includes validation, default values, and environment variable support.
package config

import (
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/YuriLeonel/task-manager/internal/errors"
)

// Config holds all configuration options for the task manager.
type Config struct {
	// DataDir is the directory where task data is stored.
	DataDir string
	// BackupDir is the directory where backups are stored.
	BackupDir string
	// BackupEnabled indicates if automatic backups are enabled.
	BackupEnabled bool
	// BackupInterval is the interval between automatic backups.
	BackupInterval time.Duration
	// MaxBackups is the maximum number of backups to keep.
	MaxBackups int
	// MaxTasks is the maximum number of tasks allowed.
	MaxTasks int
	// StorageType is the type of storage to use ("memory" or "file").
	StorageType string
	// Debug indicates if debug logging is enabled.
	Debug bool
}

// DefaultConfig returns a Config instance with default values.
func DefaultConfig() (*Config, error) {
	// Get user's home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	// Create default paths
	dataDir := filepath.Join(homeDir, ".task-manager", "data")
	backupDir := filepath.Join(homeDir, ".task-manager", "backups")

	return &Config{
		DataDir:        dataDir,
		BackupDir:      backupDir,
		BackupEnabled:  true,
		BackupInterval: 24 * time.Hour,
		MaxBackups:     10,
		MaxTasks:       10000,
		StorageType:    "file",
		Debug:          false,
	}, nil
}

// LoadFromEnv loads configuration from environment variables.
// It updates the config with any environment variables that are set.
func (c *Config) LoadFromEnv() error {
	// Data directory
	if dataDir := os.Getenv("TASK_MANAGER_DATA_DIR"); dataDir != "" {
		c.DataDir = dataDir
	}

	// Backup directory
	if backupDir := os.Getenv("TASK_MANAGER_BACKUP_DIR"); backupDir != "" {
		c.BackupDir = backupDir
	}

	// Backup enabled
	if backupEnabled := os.Getenv("TASK_MANAGER_BACKUP_ENABLED"); backupEnabled != "" {
		enabled, err := strconv.ParseBool(backupEnabled)
		if err != nil {
			return err
		}
		c.BackupEnabled = enabled
	}

	// Backup interval
	if backupInterval := os.Getenv("TASK_MANAGER_BACKUP_INTERVAL"); backupInterval != "" {
		interval, err := time.ParseDuration(backupInterval)
		if err != nil {
			return err
		}
		c.BackupInterval = interval
	}

	// Max backups
	if maxBackups := os.Getenv("TASK_MANAGER_MAX_BACKUPS"); maxBackups != "" {
		max, err := strconv.Atoi(maxBackups)
		if err != nil {
			return err
		}
		c.MaxBackups = max
	}

	// Max tasks
	if maxTasks := os.Getenv("TASK_MANAGER_MAX_TASKS"); maxTasks != "" {
		max, err := strconv.Atoi(maxTasks)
		if err != nil {
			return err
		}
		c.MaxTasks = max
	}

	// Storage type
	if storageType := os.Getenv("TASK_MANAGER_STORAGE_TYPE"); storageType != "" {
		c.StorageType = storageType
	}

	// Debug mode
	if debug := os.Getenv("TASK_MANAGER_DEBUG"); debug != "" {
		enabled, err := strconv.ParseBool(debug)
		if err != nil {
			return err
		}
		c.Debug = enabled
	}

	return nil
}

// Validate checks if the configuration is valid.
func (c *Config) Validate() error {
	// Check required fields
	if c.DataDir == "" {
		return errors.NewValidationError("dataDir", "cannot be empty")
	}
	if c.BackupDir == "" {
		return errors.NewValidationError("backupDir", "cannot be empty")
	}
	if c.StorageType == "" {
		return errors.NewValidationError("storageType", "cannot be empty")
	}

	// Validate storage type
	if c.StorageType != "memory" && c.StorageType != "file" {
		return errors.NewValidationError("storageType", "must be either 'memory' or 'file'")
	}

	// Validate numeric fields
	if c.MaxBackups <= 0 {
		return errors.NewValidationError("maxBackups", "must be greater than 0")
	}
	if c.MaxTasks <= 0 {
		return errors.NewValidationError("maxTasks", "must be greater than 0")
	}
	if c.BackupInterval <= 0 {
		return errors.NewValidationError("backupInterval", "must be greater than 0")
	}

	return nil
}

// GetTasksFilePath returns the path to the tasks file.
func (c *Config) GetTasksFilePath() string {
	return filepath.Join(c.DataDir, "tasks.json")
}

// GetLogDir returns the path to the log directory.
func (c *Config) GetLogDir() string {
	return filepath.Join(c.DataDir, "logs")
}
