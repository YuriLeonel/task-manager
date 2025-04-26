package task

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/YuriLeonel/task-manager/internal/communication"
	"github.com/YuriLeonel/task-manager/internal/errors"
	"github.com/YuriLeonel/task-manager/internal/storage"
	"github.com/YuriLeonel/task-manager/internal/storage/backup"
	"github.com/YuriLeonel/task-manager/pkg/models"
)

// Manager handles concurrent task operations
type Manager struct {
	tasks    map[string]*models.Task
	mu       sync.RWMutex
	events   *EventHandler
	backup   *backup.ConcurrentManager
	storage  storage.Storage
	service  *Service
	stopChan chan struct{}
	wg       sync.WaitGroup
}

// NewManager creates a new task manager
func NewManager(comm *communication.CommunicationManager, backup *backup.ConcurrentManager, storage storage.Storage, service *Service) *Manager {
	m := &Manager{
		tasks:    make(map[string]*models.Task),
		events:   NewEventHandler(comm),
		backup:   backup,
		storage:  storage,
		service:  service,
		stopChan: make(chan struct{}),
	}

	// Start background tasks
	m.wg.Add(1)
	go m.autoBackupLoop()

	return m
}

// CreateTask creates a new task with proper synchronization
func (m *Manager) CreateTask(ctx context.Context, description string, dueDate *time.Time, priority models.Priority) (*models.Task, error) {
	task := models.NewTask(description)
	task.DueDate = dueDate
	task.Priority = priority

	// Save task to storage
	if err := m.storage.SaveTask(ctx, task); err != nil {
		return nil, err
	}

	m.mu.Lock()
	m.tasks[task.ID] = task
	m.mu.Unlock()

	// Notify about task creation
	if err := m.events.NotifyTaskCreated(ctx, task); err != nil {
		// Log error but don't fail the operation
		fmt.Printf("Warning: failed to notify task creation: %v\n", err)
	}

	return task, nil
}

// UpdateTask updates an existing task
func (m *Manager) UpdateTask(ctx context.Context, taskID string, description string, dueDate *time.Time, priority models.Priority) (*models.Task, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	task, exists := m.tasks[taskID]
	if !exists {
		return nil, errors.NewTaskError(errors.TaskNotFound, "task not found", nil)
	}

	task.Description = description
	task.DueDate = dueDate
	task.Priority = priority
	task.UpdatedAt = time.Now()

	// Update task in storage
	if err := m.storage.UpdateTask(ctx, task); err != nil {
		return nil, err
	}

	// Notify about task update
	if err := m.events.NotifyTaskUpdated(ctx, task); err != nil {
		// Log error but don't fail the operation
		fmt.Printf("Warning: failed to notify task update: %v\n", err)
	}

	return task, nil
}

// DeleteTask deletes a task
func (m *Manager) DeleteTask(ctx context.Context, taskID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.tasks[taskID]; !exists {
		return errors.NewTaskError(errors.TaskNotFound, "task not found", nil)
	}

	// Delete task from storage
	if err := m.storage.DeleteTask(ctx, taskID); err != nil {
		return err
	}

	delete(m.tasks, taskID)

	// Notify about task deletion
	if err := m.events.NotifyTaskDeleted(ctx, taskID); err != nil {
		// Log error but don't fail the operation
		fmt.Printf("Warning: failed to notify task deletion: %v\n", err)
	}

	return nil
}

// GetTask retrieves a task by ID
func (m *Manager) GetTask(taskID string) (*models.Task, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	task, exists := m.tasks[taskID]
	if !exists {
		// Try to get from storage
		var err error
		task, err = m.storage.GetTask(context.Background(), taskID)
		if err != nil {
			return nil, err
		}
		if task == nil {
			return nil, errors.NewTaskError(errors.TaskNotFound, "task not found", nil)
		}
		// Cache the task
		m.tasks[taskID] = task
	}

	return task, nil
}

// ListTasks returns all tasks
func (m *Manager) ListTasks() []*models.Task {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Get tasks from storage
	tasks, err := m.storage.GetAllTasks(context.Background())
	if err != nil {
		// Log error and return in-memory tasks as fallback
		fmt.Printf("Warning: failed to get tasks from storage: %v\n", err)
		tasks = make([]*models.Task, 0, len(m.tasks))
		for _, task := range m.tasks {
			tasks = append(tasks, task)
		}
	}

	return tasks
}

// Stop stops the task manager and its background tasks
func (m *Manager) Stop() {
	close(m.stopChan)
	m.wg.Wait()
}

// autoBackupLoop periodically creates backups of tasks
func (m *Manager) autoBackupLoop() {
	defer m.wg.Done()

	ticker := time.NewTicker(15 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-m.stopChan:
			return
		case <-ticker.C:
			m.createBackup()
		}
	}
}

// createBackup creates a backup of all tasks
func (m *Manager) createBackup() {
	m.mu.RLock()
	tasks := m.ListTasks()
	m.mu.RUnlock()

	if err := m.backup.CreateBackup(context.Background(), tasks); err != nil {
		fmt.Printf("Warning: failed to create backup: %v\n", err)
	}
}
