// Package reminder provides functionality for managing task reminders.
// It uses goroutines and channels for parallel processing of reminders.
package reminder

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/YuriLeonel/task-manager/internal/communication"
	"github.com/YuriLeonel/task-manager/pkg/models"
)

// Reminder represents a task reminder.
type Reminder struct {
	TaskID      string
	Description string
	DueDate     time.Time
	Priority    models.Priority
}

// ReminderManager handles task reminders using goroutines and channels.
type ReminderManager struct {
	// reminders is a map of task ID to reminder
	reminders map[string]*Reminder
	// mu provides thread safety
	mu sync.RWMutex
	// reminderChan is used to send reminders
	reminderChan chan *Reminder
	// stopChan is used to stop the reminder goroutine
	stopChan chan struct{}
	// wg is used to wait for goroutines to finish
	wg   sync.WaitGroup
	comm *communication.CommunicationManager
}

// NewReminderManager creates a new reminder manager.
func NewReminderManager(comm *communication.CommunicationManager) *ReminderManager {
	rm := &ReminderManager{
		reminders:    make(map[string]*Reminder),
		reminderChan: make(chan *Reminder, 100),
		stopChan:     make(chan struct{}),
		comm:         comm,
	}

	// Start the reminder goroutine
	rm.wg.Add(1)
	go rm.reminderLoop()

	return rm
}

// AddReminder adds a new reminder for a task.
func (rm *ReminderManager) AddReminder(task *models.Task) error {
	if task == nil {
		return fmt.Errorf("task cannot be nil")
	}

	if task.DueDate == nil {
		return fmt.Errorf("task must have a due date")
	}

	rm.mu.Lock()
	defer rm.mu.Unlock()

	reminder := &Reminder{
		TaskID:      task.ID,
		Description: task.Description,
		DueDate:     *task.DueDate,
		Priority:    task.Priority,
	}

	rm.reminders[task.ID] = reminder
	return nil
}

// RemoveReminder removes a reminder for a task.
func (rm *ReminderManager) RemoveReminder(taskID string) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	delete(rm.reminders, taskID)
}

// UpdateReminder updates an existing reminder.
func (rm *ReminderManager) UpdateReminder(task *models.Task) error {
	if task == nil {
		return fmt.Errorf("task cannot be nil")
	}

	if task.DueDate == nil {
		return fmt.Errorf("task must have a due date")
	}

	rm.mu.Lock()
	defer rm.mu.Unlock()

	if reminder, exists := rm.reminders[task.ID]; exists {
		reminder.Description = task.Description
		reminder.DueDate = *task.DueDate
		reminder.Priority = task.Priority
	}

	return nil
}

// Stop stops the reminder manager and all its goroutines.
func (rm *ReminderManager) Stop() {
	close(rm.stopChan)
	rm.wg.Wait()
}

// reminderLoop is the main goroutine that processes reminders.
func (rm *ReminderManager) reminderLoop() {
	defer rm.wg.Done()

	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-rm.stopChan:
			return
		case <-ticker.C:
			rm.checkReminders()
		}
	}
}

// checkReminders checks all reminders and sends notifications for due tasks.
func (rm *ReminderManager) checkReminders() {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	now := time.Now()
	for _, reminder := range rm.reminders {
		// Check if the task is due within the next hour
		if reminder.DueDate.Sub(now) <= time.Hour && reminder.DueDate.Sub(now) > 0 {
			// Send reminder through the communication system
			if err := rm.comm.Publish(context.Background(), communication.ReminderTriggered, reminder); err != nil {
				// Log error but continue processing other reminders
				fmt.Printf("Warning: failed to notify about reminder: %v\n", err)
			}
		}
	}
}

// GetReminderChan returns the channel for receiving reminders.
func (rm *ReminderManager) GetReminderChan() <-chan *Reminder {
	return rm.reminderChan
}
