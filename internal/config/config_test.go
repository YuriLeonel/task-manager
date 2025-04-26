package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	config, err := DefaultConfig()
	if err != nil {
		t.Fatalf("Failed to create default config: %v", err)
	}

	// Verify default values
	if config.BackupEnabled != true {
		t.Errorf("Expected BackupEnabled to be true, got %v", config.BackupEnabled)
	}
	if config.BackupInterval != 24*time.Hour {
		t.Errorf("Expected BackupInterval to be 24h, got %v", config.BackupInterval)
	}
	if config.MaxBackups != 10 {
		t.Errorf("Expected MaxBackups to be 10, got %d", config.MaxBackups)
	}
	if config.MaxTasks != 10000 {
		t.Errorf("Expected MaxTasks to be 10000, got %d", config.MaxTasks)
	}
	if config.StorageType != "file" {
		t.Errorf("Expected StorageType to be 'file', got %s", config.StorageType)
	}
	if config.Debug != false {
		t.Errorf("Expected Debug to be false, got %v", config.Debug)
	}
}

func TestLoadFromEnv(t *testing.T) {
	// Set up test environment variables
	os.Setenv("TASK_MANAGER_DATA_DIR", "/test/data")
	os.Setenv("TASK_MANAGER_BACKUP_DIR", "/test/backups")
	os.Setenv("TASK_MANAGER_BACKUP_ENABLED", "false")
	os.Setenv("TASK_MANAGER_BACKUP_INTERVAL", "1h")
	os.Setenv("TASK_MANAGER_MAX_BACKUPS", "5")
	os.Setenv("TASK_MANAGER_MAX_TASKS", "100")
	os.Setenv("TASK_MANAGER_STORAGE_TYPE", "memory")
	os.Setenv("TASK_MANAGER_DEBUG", "true")
	defer func() {
		os.Unsetenv("TASK_MANAGER_DATA_DIR")
		os.Unsetenv("TASK_MANAGER_BACKUP_DIR")
		os.Unsetenv("TASK_MANAGER_BACKUP_ENABLED")
		os.Unsetenv("TASK_MANAGER_BACKUP_INTERVAL")
		os.Unsetenv("TASK_MANAGER_MAX_BACKUPS")
		os.Unsetenv("TASK_MANAGER_MAX_TASKS")
		os.Unsetenv("TASK_MANAGER_STORAGE_TYPE")
		os.Unsetenv("TASK_MANAGER_DEBUG")
	}()

	config, err := DefaultConfig()
	if err != nil {
		t.Fatalf("Failed to create default config: %v", err)
	}

	err = config.LoadFromEnv()
	if err != nil {
		t.Fatalf("Failed to load from env: %v", err)
	}

	// Verify environment values were loaded correctly
	if config.DataDir != "/test/data" {
		t.Errorf("Expected DataDir to be '/test/data', got %s", config.DataDir)
	}
	if config.BackupDir != "/test/backups" {
		t.Errorf("Expected BackupDir to be '/test/backups', got %s", config.BackupDir)
	}
	if config.BackupEnabled != false {
		t.Errorf("Expected BackupEnabled to be false, got %v", config.BackupEnabled)
	}
	if config.BackupInterval != time.Hour {
		t.Errorf("Expected BackupInterval to be 1h, got %v", config.BackupInterval)
	}
	if config.MaxBackups != 5 {
		t.Errorf("Expected MaxBackups to be 5, got %d", config.MaxBackups)
	}
	if config.MaxTasks != 100 {
		t.Errorf("Expected MaxTasks to be 100, got %d", config.MaxTasks)
	}
	if config.StorageType != "memory" {
		t.Errorf("Expected StorageType to be 'memory', got %s", config.StorageType)
	}
	if config.Debug != true {
		t.Errorf("Expected Debug to be true, got %v", config.Debug)
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &Config{
				DataDir:        "/test/data",
				BackupDir:      "/test/backups",
				BackupEnabled:  true,
				BackupInterval: time.Hour,
				MaxBackups:     10,
				MaxTasks:       100,
				StorageType:    "file",
			},
			wantErr: false,
		},
		{
			name: "empty data dir",
			config: &Config{
				DataDir:        "",
				BackupDir:      "/test/backups",
				BackupEnabled:  true,
				BackupInterval: time.Hour,
				MaxBackups:     10,
				MaxTasks:       100,
				StorageType:    "file",
			},
			wantErr: true,
		},
		{
			name: "invalid storage type",
			config: &Config{
				DataDir:        "/test/data",
				BackupDir:      "/test/backups",
				BackupEnabled:  true,
				BackupInterval: time.Hour,
				MaxBackups:     10,
				MaxTasks:       100,
				StorageType:    "invalid",
			},
			wantErr: true,
		},
		{
			name: "invalid max backups",
			config: &Config{
				DataDir:        "/test/data",
				BackupDir:      "/test/backups",
				BackupEnabled:  true,
				BackupInterval: time.Hour,
				MaxBackups:     0,
				MaxTasks:       100,
				StorageType:    "file",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetTasksFilePath(t *testing.T) {
	config := &Config{
		DataDir: "/test/data",
	}
	expected := filepath.Join("/test/data", "tasks.json")
	if got := config.GetTasksFilePath(); got != expected {
		t.Errorf("GetTasksFilePath() = %v, want %v", got, expected)
	}
}

func TestGetLogDir(t *testing.T) {
	config := &Config{
		DataDir: "/test/data",
	}
	expected := filepath.Join("/test/data", "logs")
	if got := config.GetLogDir(); got != expected {
		t.Errorf("GetLogDir() = %v, want %v", got, expected)
	}
}
