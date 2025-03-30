package task

import (
	"context"
	"testing"
	"time"

	"github.com/YuriLeonel/task-manager/pkg/models"
)

// MockStorage implements storage.Storage for testing
type MockStorage struct {
	tasks map[string]*models.Task
}

func NewMockStorage() *MockStorage {
	return &MockStorage{
		tasks: make(map[string]*models.Task),
	}
}

func (m *MockStorage) SaveTask(ctx context.Context, task *models.Task) error {
	m.tasks[task.ID] = task
	return nil
}

func (m *MockStorage) GetTask(ctx context.Context, id string) (*models.Task, error) {
	if task, ok := m.tasks[id]; ok {
		return task, nil
	}
	return nil, nil
}

func (m *MockStorage) GetAllTasks(ctx context.Context) ([]*models.Task, error) {
	tasks := make([]*models.Task, 0, len(m.tasks))
	for _, task := range m.tasks {
		tasks = append(tasks, task)
	}
	return tasks, nil
}

func (m *MockStorage) UpdateTask(ctx context.Context, task *models.Task) error {
	if _, ok := m.tasks[task.ID]; ok {
		m.tasks[task.ID] = task
		return nil
	}
	return nil
}

func (m *MockStorage) DeleteTask(ctx context.Context, id string) error {
	delete(m.tasks, id)
	return nil
}

func (m *MockStorage) CreateBackup(ctx context.Context) error {
	return nil
}

func (m *MockStorage) RestoreBackup(ctx context.Context, backupID string) error {
	return nil
}

func (m *MockStorage) ListBackups(ctx context.Context) ([]string, error) {
	return nil, nil
}

func (m *MockStorage) DeleteBackup(ctx context.Context, backupID string) error {
	return nil
}

func TestService(t *testing.T) {
	mockStorage := NewMockStorage()
	service := NewService(mockStorage)

	// Test adding a task
	err := service.AddTask("Test task")
	if err != nil {
		t.Fatalf("Failed to add task: %v", err)
	}

	// Test listing tasks
	tasks, err := service.ListTasks()
	if err != nil {
		t.Fatalf("Failed to list tasks: %v", err)
	}
	if len(tasks) != 1 {
		t.Errorf("Expected 1 task, got %d", len(tasks))
	}

	// Test marking task as completed
	err = service.MarkAsCompleted(tasks[0].ID)
	if err != nil {
		t.Fatalf("Failed to mark task as completed: %v", err)
	}

	// Verify task is completed
	task, err := service.GetTask(tasks[0].ID)
	if err != nil {
		t.Fatalf("Failed to get task: %v", err)
	}
	if !task.Completed {
		t.Error("Task should be marked as completed")
	}

	// Test setting priority
	err = service.SetPriority(tasks[0].ID, models.High)
	if err != nil {
		t.Fatalf("Failed to set priority: %v", err)
	}

	// Verify priority
	task, err = service.GetTask(tasks[0].ID)
	if err != nil {
		t.Fatalf("Failed to get task: %v", err)
	}
	if task.Priority != models.High {
		t.Errorf("Expected priority High, got %v", task.Priority)
	}

	// Test setting due date
	dueDate := time.Now().Add(24 * time.Hour)
	err = service.SetDueDate(tasks[0].ID, dueDate)
	if err != nil {
		t.Fatalf("Failed to set due date: %v", err)
	}

	// Verify due date
	task, err = service.GetTask(tasks[0].ID)
	if err != nil {
		t.Fatalf("Failed to get task: %v", err)
	}
	if !task.DueDate.Equal(dueDate) {
		t.Errorf("Expected due date %v, got %v", dueDate, task.DueDate)
	}

	// Test adding tag
	err = service.AddTag(tasks[0].ID, "test-tag")
	if err != nil {
		t.Fatalf("Failed to add tag: %v", err)
	}

	// Verify tag
	task, err = service.GetTask(tasks[0].ID)
	if err != nil {
		t.Fatalf("Failed to get task: %v", err)
	}
	if len(task.Tags) != 1 || task.Tags[0] != "test-tag" {
		t.Errorf("Expected tag 'test-tag', got %v", task.Tags)
	}

	// Test removing tag
	err = service.RemoveTag(tasks[0].ID, "test-tag")
	if err != nil {
		t.Fatalf("Failed to remove tag: %v", err)
	}

	// Verify tag removal
	task, err = service.GetTask(tasks[0].ID)
	if err != nil {
		t.Fatalf("Failed to get task: %v", err)
	}
	if len(task.Tags) != 0 {
		t.Error("Tag should be removed")
	}

	// Test setting progress
	err = service.SetProgress(tasks[0].ID, 50)
	if err != nil {
		t.Fatalf("Failed to set progress: %v", err)
	}

	// Verify progress
	task, err = service.GetTask(tasks[0].ID)
	if err != nil {
		t.Fatalf("Failed to get task: %v", err)
	}
	if task.Progress != 50 {
		t.Errorf("Expected progress 50, got %d", task.Progress)
	}

	// Test deleting task
	err = service.DeleteTask(tasks[0].ID)
	if err != nil {
		t.Fatalf("Failed to delete task: %v", err)
	}

	// Verify deletion
	task, err = service.GetTask(tasks[0].ID)
	if err != nil {
		t.Fatalf("Failed to get task: %v", err)
	}
	if task != nil {
		t.Error("Task should be deleted")
	}
}

func TestGetTasksByTag(t *testing.T) {
	mockStorage := NewMockStorage()
	service := NewService(mockStorage)

	// Add tasks with tags
	task1 := models.NewTask("Task 1")
	task1.ID = "task1"
	task1.AddTag("tag1")
	task1.AddTag("tag2")
	mockStorage.SaveTask(context.Background(), task1)

	task2 := models.NewTask("Task 2")
	task2.ID = "task2"
	task2.AddTag("tag2")
	task2.AddTag("tag3")
	mockStorage.SaveTask(context.Background(), task2)

	task3 := models.NewTask("Task 3")
	task3.ID = "task3"
	task3.AddTag("tag1")
	task3.AddTag("tag3")
	mockStorage.SaveTask(context.Background(), task3)

	// Test getting tasks by tag
	taggedTasks, err := service.GetTasksByTag("tag2")
	if err != nil {
		t.Fatalf("Failed to get tasks by tag: %v", err)
	}
	if len(taggedTasks) != 2 {
		t.Errorf("Expected 2 tasks with tag 'tag2', got %d", len(taggedTasks))
	}

	// Test getting tasks by non-existent tag
	taggedTasks, err = service.GetTasksByTag("non-existent")
	if err != nil {
		t.Fatalf("Failed to get tasks by tag: %v", err)
	}
	if len(taggedTasks) != 0 {
		t.Error("Expected no tasks with non-existent tag")
	}
}

func TestInvalidOperations(t *testing.T) {
	mockStorage := NewMockStorage()
	service := NewService(mockStorage)

	// Test operations on non-existent task
	nonExistentID := "non-existent"

	// Test marking non-existent task as completed
	err := service.MarkAsCompleted(nonExistentID)
	if err == nil {
		t.Error("Expected error when marking non-existent task as completed")
	}

	// Test setting priority of non-existent task
	err = service.SetPriority(nonExistentID, models.High)
	if err == nil {
		t.Error("Expected error when setting priority of non-existent task")
	}

	// Test setting due date of non-existent task
	err = service.SetDueDate(nonExistentID, time.Now())
	if err == nil {
		t.Error("Expected error when setting due date of non-existent task")
	}

	// Test adding tag to non-existent task
	err = service.AddTag(nonExistentID, "tag")
	if err == nil {
		t.Error("Expected error when adding tag to non-existent task")
	}

	// Test removing tag from non-existent task
	err = service.RemoveTag(nonExistentID, "tag")
	if err == nil {
		t.Error("Expected error when removing tag from non-existent task")
	}

	// Test setting progress of non-existent task
	err = service.SetProgress(nonExistentID, 50)
	if err == nil {
		t.Error("Expected error when setting progress of non-existent task")
	}

	// Test deleting non-existent task
	err = service.DeleteTask(nonExistentID)
	if err == nil {
		t.Error("Expected error when deleting non-existent task")
	}
}
