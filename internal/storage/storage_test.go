package storage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/YuriLeonel/task-manager/pkg/models"
)

func TestJSONFileStorage(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get user home directory: %v", err)
	}

	// Create test directories in user's home directory
	testDir := filepath.Join(homeDir, ".task-manager-test")
	err = os.MkdirAll(testDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	defer os.RemoveAll(testDir)

	tasksFilePath := filepath.Join(testDir, "tasks.json")
	backupDir := filepath.Join(testDir, "backups")
	err = os.MkdirAll(backupDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create backup directory: %v", err)
	}

	// Create storage instance
	storage, err := NewJSONFileStorage(tasksFilePath, 100, backupDir, 5)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	ctx := context.Background()

	// Test adding a task
	task := models.NewTask("Test task")
	err = storage.SaveTask(ctx, task)
	if err != nil {
		t.Fatalf("Failed to save task: %v", err)
	}

	// Test retrieving the task
	retrievedTask, err := storage.GetTask(ctx, task.ID)
	if err != nil {
		t.Fatalf("Failed to get task: %v", err)
	}
	if retrievedTask.ID != task.ID {
		t.Errorf("Expected task ID %s, got %s", task.ID, retrievedTask.ID)
	}

	// Test updating the task
	task.Description = "Updated task"
	err = storage.UpdateTask(ctx, task)
	if err != nil {
		t.Fatalf("Failed to update task: %v", err)
	}

	// Verify update
	retrievedTask, err = storage.GetTask(ctx, task.ID)
	if err != nil {
		t.Fatalf("Failed to get updated task: %v", err)
	}
	if retrievedTask.Description != "Updated task" {
		t.Errorf("Expected description 'Updated task', got '%s'", retrievedTask.Description)
	}

	// Test deleting the task
	err = storage.DeleteTask(ctx, task.ID)
	if err != nil {
		t.Fatalf("Failed to delete task: %v", err)
	}

	// Verify deletion
	_, err = storage.GetTask(ctx, task.ID)
	if err == nil {
		t.Error("Expected error when getting deleted task")
	}
}

func TestMemoryStorage(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get user home directory: %v", err)
	}

	// Create test backup directory in user's home directory
	backupDir := filepath.Join(homeDir, ".task-manager-test", "memory-backups")
	err = os.MkdirAll(backupDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create backup directory: %v", err)
	}
	defer os.RemoveAll(filepath.Join(homeDir, ".task-manager-test"))

	// Create storage instance
	storage, err := NewMemoryStorage(100, backupDir, 5)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	ctx := context.Background()

	// Test adding a task
	task := models.NewTask("Test task")
	err = storage.SaveTask(ctx, task)
	if err != nil {
		t.Fatalf("Failed to save task: %v", err)
	}

	// Test retrieving the task
	retrievedTask, err := storage.GetTask(ctx, task.ID)
	if err != nil {
		t.Fatalf("Failed to get task: %v", err)
	}
	if retrievedTask.ID != task.ID {
		t.Errorf("Expected task ID %s, got %s", task.ID, retrievedTask.ID)
	}

	// Test updating the task
	task.Description = "Updated task"
	err = storage.UpdateTask(ctx, task)
	if err != nil {
		t.Fatalf("Failed to update task: %v", err)
	}

	// Verify update
	retrievedTask, err = storage.GetTask(ctx, task.ID)
	if err != nil {
		t.Fatalf("Failed to get updated task: %v", err)
	}
	if retrievedTask.Description != "Updated task" {
		t.Errorf("Expected description 'Updated task', got '%s'", retrievedTask.Description)
	}

	// Test deleting the task
	err = storage.DeleteTask(ctx, task.ID)
	if err != nil {
		t.Fatalf("Failed to delete task: %v", err)
	}

	// Verify deletion
	_, err = storage.GetTask(ctx, task.ID)
	if err == nil {
		t.Error("Expected error when getting deleted task")
	}
}

func TestTaskLimit(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get user home directory: %v", err)
	}

	backupDir := filepath.Join(homeDir, ".task-manager-test", "limit-backups")
	err = os.MkdirAll(backupDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create backup directory: %v", err)
	}
	defer os.RemoveAll(filepath.Join(homeDir, ".task-manager-test"))

	// Create storage instance with low limit
	storage, err := NewMemoryStorage(2, backupDir, 5)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	ctx := context.Background()

	// Add two tasks
	task1 := models.NewTask("Task 1")
	task1.ID = "task1"
	task2 := models.NewTask("Task 2")
	task2.ID = "task2"
	task3 := models.NewTask("Task 3")
	task3.ID = "task3"

	err = storage.SaveTask(ctx, task1)
	if err != nil {
		t.Fatalf("Failed to save task1: %v", err)
	}
	err = storage.SaveTask(ctx, task2)
	if err != nil {
		t.Fatalf("Failed to save task2: %v", err)
	}

	// Try to add a third task
	err = storage.SaveTask(ctx, task3)
	if err == nil {
		t.Error("Expected error when exceeding task limit")
	}
}

func TestDuplicateTaskID(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get user home directory: %v", err)
	}

	backupDir := filepath.Join(homeDir, ".task-manager-test", "duplicate-backups")
	err = os.MkdirAll(backupDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create backup directory: %v", err)
	}
	defer os.RemoveAll(filepath.Join(homeDir, ".task-manager-test"))

	storage, err := NewMemoryStorage(100, backupDir, 5)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	ctx := context.Background()

	// Add a task
	task := models.NewTask("Test task")
	err = storage.SaveTask(ctx, task)
	if err != nil {
		t.Fatalf("Failed to save task: %v", err)
	}

	// Try to add a task with the same ID
	err = storage.SaveTask(ctx, task)
	if err == nil {
		t.Error("Expected error when adding task with duplicate ID")
	}
}

func TestGetAllTasks(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get user home directory: %v", err)
	}

	backupDir := filepath.Join(homeDir, ".task-manager-test", "getall-backups")
	err = os.MkdirAll(backupDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create backup directory: %v", err)
	}
	defer os.RemoveAll(filepath.Join(homeDir, ".task-manager-test"))

	storage, err := NewMemoryStorage(100, backupDir, 5)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	ctx := context.Background()

	// Add multiple tasks
	tasks := []*models.Task{
		models.NewTask("Task 1"),
		models.NewTask("Task 2"),
		models.NewTask("Task 3"),
	}

	// Set unique IDs
	for i, task := range tasks {
		task.ID = fmt.Sprintf("task%d", i+1)
		err = storage.SaveTask(ctx, task)
		if err != nil {
			t.Fatalf("Failed to save task: %v", err)
		}
	}

	// Get all tasks
	retrievedTasks, err := storage.GetAllTasks(ctx)
	if err != nil {
		t.Fatalf("Failed to get all tasks: %v", err)
	}

	// Verify task count
	if len(retrievedTasks) != len(tasks) {
		t.Errorf("Expected %d tasks, got %d", len(tasks), len(retrievedTasks))
	}

	// Verify task contents
	for i, task := range tasks {
		if retrievedTasks[i].Description != task.Description {
			t.Errorf("Expected task description '%s', got '%s'", task.Description, retrievedTasks[i].Description)
		}
	}
}
