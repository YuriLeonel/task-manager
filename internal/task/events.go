package task

import (
	"context"

	"github.com/YuriLeonel/task-manager/internal/communication"
	"github.com/YuriLeonel/task-manager/pkg/models"
)

// EventHandler handles task-related events and notifications.
type EventHandler struct {
	comm *communication.CommunicationManager
}

// NewEventHandler creates a new task event handler.
func NewEventHandler(comm *communication.CommunicationManager) *EventHandler {
	return &EventHandler{
		comm: comm,
	}
}

// NotifyTaskCreated notifies about a new task creation.
func (h *EventHandler) NotifyTaskCreated(ctx context.Context, task *models.Task) error {
	return h.comm.Publish(ctx, communication.TaskCreated, task)
}

// NotifyTaskUpdated notifies about a task update.
func (h *EventHandler) NotifyTaskUpdated(ctx context.Context, task *models.Task) error {
	return h.comm.Publish(ctx, communication.TaskUpdated, task)
}

// NotifyTaskDeleted notifies about a task deletion.
func (h *EventHandler) NotifyTaskDeleted(ctx context.Context, taskID string) error {
	return h.comm.Publish(ctx, communication.TaskDeleted, taskID)
}

// SubscribeToTaskEvents subscribes to task-related events.
func (h *EventHandler) SubscribeToTaskEvents(id string) (*communication.Subscriber, error) {
	return h.comm.Subscribe(id, []communication.MessageType{
		communication.TaskCreated,
		communication.TaskUpdated,
		communication.TaskDeleted,
	})
}

// SubscribeToBackupEvents subscribes to backup-related events.
func (h *EventHandler) SubscribeToBackupEvents(id string) (*communication.Subscriber, error) {
	return h.comm.Subscribe(id, []communication.MessageType{
		communication.BackupCreated,
		communication.BackupRestored,
	})
}

// SubscribeToReminderEvents subscribes to reminder-related events.
func (h *EventHandler) SubscribeToReminderEvents(id string) (*communication.Subscriber, error) {
	return h.comm.Subscribe(id, []communication.MessageType{
		communication.ReminderTriggered,
	})
}
