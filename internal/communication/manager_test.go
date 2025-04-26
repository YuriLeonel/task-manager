package communication

import (
	"context"
	"testing"
	"time"

	"github.com/YuriLeonel/task-manager/pkg/models"
)

func TestCommunicationManager(t *testing.T) {
	// Create a new communication manager
	cm := NewCommunicationManager()
	defer cm.Stop()

	// Create a subscriber for task-related messages
	subscriber, err := cm.Subscribe("task-subscriber", []MessageType{TaskCreated, TaskUpdated, TaskDeleted})
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}
	defer cm.Unsubscribe(subscriber.ID)

	// Create a test task
	task := models.NewTask("Test Task")

	// Test publishing a task created message
	ctx := context.Background()
	err = cm.Publish(ctx, TaskCreated, task)
	if err != nil {
		t.Fatalf("Failed to publish message: %v", err)
	}

	// Wait for the message
	select {
	case msg := <-subscriber.Channel:
		if msg.Type != TaskCreated {
			t.Errorf("Expected message type %s, got %s", TaskCreated, msg.Type)
		}
		if msg.Data.(*models.Task).Description != task.Description {
			t.Errorf("Expected task description %s, got %s", task.Description, msg.Data.(*models.Task).Description)
		}
	case <-time.After(1 * time.Second):
		t.Error("Timeout waiting for message")
	}
}

func TestMultipleSubscribers(t *testing.T) {
	// Create a new communication manager
	cm := NewCommunicationManager()
	defer cm.Stop()

	// Create multiple subscribers
	subscriber1, err := cm.Subscribe("subscriber1", []MessageType{TaskCreated})
	if err != nil {
		t.Fatalf("Failed to subscribe subscriber1: %v", err)
	}
	defer cm.Unsubscribe(subscriber1.ID)

	subscriber2, err := cm.Subscribe("subscriber2", []MessageType{TaskCreated})
	if err != nil {
		t.Fatalf("Failed to subscribe subscriber2: %v", err)
	}
	defer cm.Unsubscribe(subscriber2.ID)

	// Create a test task
	task := models.NewTask("Test Task")

	// Publish a message
	ctx := context.Background()
	err = cm.Publish(ctx, TaskCreated, task)
	if err != nil {
		t.Fatalf("Failed to publish message: %v", err)
	}

	// Wait for messages on both subscribers
	received1 := false
	received2 := false

	for i := 0; i < 2; i++ {
		select {
		case msg := <-subscriber1.Channel:
			if msg.Type != TaskCreated {
				t.Errorf("Expected message type %s, got %s", TaskCreated, msg.Type)
			}
			received1 = true
		case msg := <-subscriber2.Channel:
			if msg.Type != TaskCreated {
				t.Errorf("Expected message type %s, got %s", TaskCreated, msg.Type)
			}
			received2 = true
		case <-time.After(1 * time.Second):
			t.Error("Timeout waiting for messages")
		}
	}

	if !received1 || !received2 {
		t.Error("Not all subscribers received the message")
	}
}

func TestMessageFiltering(t *testing.T) {
	// Create a new communication manager
	cm := NewCommunicationManager()
	defer cm.Stop()

	// Create a subscriber that only listens for TaskCreated messages
	subscriber, err := cm.Subscribe("task-subscriber", []MessageType{TaskCreated})
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}
	defer cm.Unsubscribe(subscriber.ID)

	// Create a test task
	task := models.NewTask("Test Task")

	// Publish a TaskUpdated message (which the subscriber should not receive)
	ctx := context.Background()
	err = cm.Publish(ctx, TaskUpdated, task)
	if err != nil {
		t.Fatalf("Failed to publish message: %v", err)
	}

	// The subscriber should not receive the message
	select {
	case <-subscriber.Channel:
		t.Error("Subscriber received message it should not have")
	case <-time.After(100 * time.Millisecond):
		// Expected behavior
	}
}

func TestConcurrentPublishing(t *testing.T) {
	// Create a new communication manager
	cm := NewCommunicationManager()
	defer cm.Stop()

	// Create a subscriber
	subscriber, err := cm.Subscribe("subscriber", []MessageType{TaskCreated})
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}
	defer cm.Unsubscribe(subscriber.ID)

	// Create a test task
	task := models.NewTask("Test Task")

	// Run multiple publish operations concurrently
	ctx := context.Background()
	done := make(chan struct{})
	go func() {
		for i := 0; i < 10; i++ {
			err := cm.Publish(ctx, TaskCreated, task)
			if err != nil {
				t.Errorf("Failed to publish message: %v", err)
			}
			// Add a small delay between publishes
			time.Sleep(10 * time.Millisecond)
		}
		done <- struct{}{}
	}()

	// Wait for all messages
	received := 0
	timeout := time.After(5 * time.Second)
	for {
		select {
		case msg := <-subscriber.Channel:
			if msg.Type != TaskCreated {
				t.Errorf("Expected message type %s, got %s", TaskCreated, msg.Type)
			}
			received++
			if received == 10 {
				return
			}
		case <-done:
			if received != 10 {
				t.Errorf("Expected 10 messages, received %d", received)
			}
			return
		case <-timeout:
			t.Error("Timeout waiting for messages")
			return
		}
	}
}
