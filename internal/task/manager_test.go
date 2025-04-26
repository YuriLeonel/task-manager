package task

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/YuriLeonel/task-manager/internal/communication"
	"github.com/YuriLeonel/task-manager/internal/storage"
	"github.com/YuriLeonel/task-manager/internal/storage/backup"
	"github.com/YuriLeonel/task-manager/pkg/models"
)

func setupTestManager(t *testing.T) (*Manager, *communication.CommunicationManager, func()) {
	// Create temporary directory for backups
	tmpDir, err := os.MkdirTemp("", "task-manager-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Create communication manager
	comm := communication.NewCommunicationManager()

	// Create backup manager
	backupMgr, err := backup.NewConcurrentManager(tmpDir, 5, comm)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("Failed to create backup manager: %v", err)
	}

	// Create storage
	store, err := storage.NewMemoryStorage(100, tmpDir, 5)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("Failed to create storage: %v", err)
	}

	// Create task service
	service := NewService(store)

	// Create task manager
	manager := NewManager(comm, backupMgr, store, service)

	// Return cleanup function
	cleanup := func() {
		manager.Stop()
		backupMgr.Stop()
		comm.Stop()
		os.RemoveAll(tmpDir)
	}

	return manager, comm, cleanup
}

func TestTaskManager(t *testing.T) {
	manager, _, cleanup := setupTestManager(t)
	defer cleanup()

	ctx := context.Background()

	// Test creating a task
	dueDate := time.Now().Add(24 * time.Hour)
	task, err := manager.CreateTask(ctx, "Test Task", &dueDate, models.High)
	if err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}

	// Test getting the task
	retrieved, err := manager.GetTask(task.ID)
	if err != nil {
		t.Fatalf("Failed to get task: %v", err)
	}
	if retrieved.Description != task.Description {
		t.Errorf("Expected description %s, got %s", task.Description, retrieved.Description)
	}

	// Test updating the task
	newDueDate := time.Now().Add(48 * time.Hour)
	updated, err := manager.UpdateTask(ctx, task.ID, "Updated Task", &newDueDate, models.Medium)
	if err != nil {
		t.Fatalf("Failed to update task: %v", err)
	}
	if updated.Description != "Updated Task" {
		t.Errorf("Expected description 'Updated Task', got %s", updated.Description)
	}

	// Test listing tasks
	tasks := manager.ListTasks()
	if len(tasks) != 1 {
		t.Errorf("Expected 1 task, got %d", len(tasks))
	}

	// Test deleting the task
	err = manager.DeleteTask(ctx, task.ID)
	if err != nil {
		t.Fatalf("Failed to delete task: %v", err)
	}

	// Verify task was deleted
	tasks = manager.ListTasks()
	if len(tasks) != 0 {
		t.Errorf("Expected 0 tasks after deletion, got %d", len(tasks))
	}
}

func TestConcurrentTaskOperations(t *testing.T) {
	manager, _, cleanup := setupTestManager(t)
	defer cleanup()

	ctx := context.Background()
	numTasks := 10
	done := make(chan struct{})

	// Create tasks concurrently
	go func() {
		for i := 0; i < numTasks; i++ {
			dueDate := time.Now().Add(24 * time.Hour)
			_, err := manager.CreateTask(ctx, "Test Task", &dueDate, models.High)
			if err != nil {
				t.Errorf("Failed to create task: %v", err)
			}
		}
		done <- struct{}{}
	}()

	// List tasks concurrently
	go func() {
		for i := 0; i < numTasks; i++ {
			tasks := manager.ListTasks()
			if len(tasks) > numTasks {
				t.Errorf("Expected at most %d tasks, got %d", numTasks, len(tasks))
			}
		}
		done <- struct{}{}
	}()

	// Wait for operations to complete
	<-done
	<-done

	// Verify final state
	tasks := manager.ListTasks()
	if len(tasks) != numTasks {
		t.Errorf("Expected %d tasks, got %d", numTasks, len(tasks))
	}
}

func TestTaskEvents(t *testing.T) {
	manager, comm, cleanup := setupTestManager(t)
	defer cleanup()

	// Subscribe to task events
	subscriber, err := comm.Subscribe("test", []communication.MessageType{
		communication.TaskCreated,
		communication.TaskUpdated,
		communication.TaskDeleted,
	})
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}
	defer comm.Unsubscribe(subscriber.ID)

	ctx := context.Background()
	dueDate := time.Now().Add(24 * time.Hour)

	// Create a task and wait for the event
	task, err := manager.CreateTask(ctx, "Test Task", &dueDate, models.High)
	if err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}

	select {
	case msg := <-subscriber.Channel:
		if msg.Type != communication.TaskCreated {
			t.Errorf("Expected TaskCreated event, got %s", msg.Type)
		}
	case <-time.After(time.Second):
		t.Error("Timeout waiting for TaskCreated event")
	}

	// Update the task and wait for the event
	_, err = manager.UpdateTask(ctx, task.ID, "Updated Task", &dueDate, models.Medium)
	if err != nil {
		t.Fatalf("Failed to update task: %v", err)
	}

	select {
	case msg := <-subscriber.Channel:
		if msg.Type != communication.TaskUpdated {
			t.Errorf("Expected TaskUpdated event, got %s", msg.Type)
		}
	case <-time.After(time.Second):
		t.Error("Timeout waiting for TaskUpdated event")
	}

	// Delete the task and wait for the event
	err = manager.DeleteTask(ctx, task.ID)
	if err != nil {
		t.Fatalf("Failed to delete task: %v", err)
	}

	select {
	case msg := <-subscriber.Channel:
		if msg.Type != communication.TaskDeleted {
			t.Errorf("Expected TaskDeleted event, got %s", msg.Type)
		}
	case <-time.After(time.Second):
		t.Error("Timeout waiting for TaskDeleted event")
	}
}
