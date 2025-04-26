package reminder

import (
	"testing"
	"time"

	"github.com/YuriLeonel/task-manager/internal/communication"
	"github.com/YuriLeonel/task-manager/pkg/models"
)

func TestReminderManager(t *testing.T) {
	// Create communication manager
	comm := communication.NewCommunicationManager()
	defer comm.Stop()

	// Create a new reminder manager
	rm := NewReminderManager(comm)
	defer rm.Stop()

	// Create a test task
	dueDate := time.Now().Add(30 * time.Minute)
	task := &models.Task{
		ID:          "test-task",
		Description: "Test task",
		DueDate:     &dueDate,
		Priority:    models.High,
	}

	// Add reminder
	err := rm.AddReminder(task)
	if err != nil {
		t.Fatalf("Failed to add reminder: %v", err)
	}

	// Update reminder
	task.Description = "Updated test task"
	err = rm.UpdateReminder(task)
	if err != nil {
		t.Fatalf("Failed to update reminder: %v", err)
	}

	// Remove reminder
	rm.RemoveReminder(task.ID)
}

func TestReminderChannel(t *testing.T) {
	// Create communication manager
	comm := communication.NewCommunicationManager()
	defer comm.Stop()

	// Create a new reminder manager
	rm := NewReminderManager(comm)
	defer rm.Stop()

	// Create a test task due in 5 seconds
	dueDate := time.Now().Add(5 * time.Second)
	task := &models.Task{
		ID:          "test-task",
		Description: "Test task",
		DueDate:     &dueDate,
		Priority:    models.High,
	}

	// Add reminder
	err := rm.AddReminder(task)
	if err != nil {
		t.Fatalf("Failed to add reminder: %v", err)
	}

	// Subscribe to reminder events
	subscriber, err := comm.Subscribe("test-subscriber", []communication.MessageType{communication.ReminderTriggered})
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}
	defer comm.Unsubscribe(subscriber.ID)

	// Wait for reminder
	select {
	case msg := <-subscriber.Channel:
		if msg.Type != communication.ReminderTriggered {
			t.Errorf("Expected message type %s, got %s", communication.ReminderTriggered, msg.Type)
		}
		reminder := msg.Data.(*Reminder)
		if reminder.TaskID != task.ID {
			t.Errorf("Expected task ID %s, got %s", task.ID, reminder.TaskID)
		}
	case <-time.After(10 * time.Second):
		t.Error("Timeout waiting for reminder")
	}
}

func TestInvalidReminder(t *testing.T) {
	// Create communication manager
	comm := communication.NewCommunicationManager()
	defer comm.Stop()

	// Create a new reminder manager
	rm := NewReminderManager(comm)
	defer rm.Stop()

	// Test nil task
	err := rm.AddReminder(nil)
	if err == nil {
		t.Error("Expected error for nil task")
	}

	// Test task without due date
	task := &models.Task{
		ID:          "test-task",
		Description: "Test task",
		Priority:    models.High,
	}

	err = rm.AddReminder(task)
	if err == nil {
		t.Error("Expected error for task without due date")
	}
}

func TestMultipleReminders(t *testing.T) {
	// Create communication manager
	comm := communication.NewCommunicationManager()
	defer comm.Stop()

	// Create a new reminder manager
	rm := NewReminderManager(comm)
	defer rm.Stop()

	// Create multiple test tasks
	dueDate1 := time.Now().Add(5 * time.Second)
	dueDate2 := time.Now().Add(10 * time.Second)
	dueDate3 := time.Now().Add(15 * time.Second)
	tasks := []*models.Task{
		{
			ID:          "task1",
			Description: "Task 1",
			DueDate:     &dueDate1,
			Priority:    models.High,
		},
		{
			ID:          "task2",
			Description: "Task 2",
			DueDate:     &dueDate2,
			Priority:    models.Medium,
		},
		{
			ID:          "task3",
			Description: "Task 3",
			DueDate:     &dueDate3,
			Priority:    models.Low,
		},
	}

	// Add all reminders
	for _, task := range tasks {
		err := rm.AddReminder(task)
		if err != nil {
			t.Fatalf("Failed to add reminder for task %s: %v", task.ID, err)
		}
	}

	// Subscribe to reminder events
	subscriber, err := comm.Subscribe("test-subscriber", []communication.MessageType{communication.ReminderTriggered})
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}
	defer comm.Unsubscribe(subscriber.ID)

	// Wait for reminders
	received := make(map[string]bool)
	for i := 0; i < len(tasks); i++ {
		select {
		case msg := <-subscriber.Channel:
			if msg.Type != communication.ReminderTriggered {
				t.Errorf("Expected message type %s, got %s", communication.ReminderTriggered, msg.Type)
			}
			reminder := msg.Data.(*Reminder)
			received[reminder.TaskID] = true
		case <-time.After(20 * time.Second):
			t.Error("Timeout waiting for reminders")
			return
		}
	}

	// Verify all reminders were received
	for _, task := range tasks {
		if !received[task.ID] {
			t.Errorf("Did not receive reminder for task %s", task.ID)
		}
	}
}
