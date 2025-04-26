package communication

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// MessageType represents the type of message being sent
type MessageType string

const (
	// TaskCreated is sent when a new task is created
	TaskCreated MessageType = "task_created"
	// TaskUpdated is sent when a task is updated
	TaskUpdated MessageType = "task_updated"
	// TaskDeleted is sent when a task is deleted
	TaskDeleted MessageType = "task_deleted"
	// BackupCreated is sent when a backup is created
	BackupCreated MessageType = "backup_created"
	// BackupRestored is sent when a backup is restored
	BackupRestored MessageType = "backup_restored"
	// ReminderTriggered is sent when a reminder is triggered
	ReminderTriggered MessageType = "reminder_triggered"
)

// Message represents a message sent through the communication system
type Message struct {
	Type      MessageType
	Timestamp time.Time
	Data      any
}

// Subscriber represents a component that subscribes to messages
type Subscriber struct {
	ID      string
	Channel chan Message
	Types   []MessageType
}

// CommunicationManager handles message passing between components
type CommunicationManager struct {
	subscribers map[string]*Subscriber
	mu          sync.RWMutex
	stopChan    chan struct{}
	wg          sync.WaitGroup
}

// NewCommunicationManager creates a new communication manager
func NewCommunicationManager() *CommunicationManager {
	cm := &CommunicationManager{
		subscribers: make(map[string]*Subscriber),
		stopChan:    make(chan struct{}),
	}

	// Start the message processing goroutine
	cm.wg.Add(1)
	go cm.processMessages()

	return cm
}

// Subscribe adds a new subscriber to the communication manager
func (cm *CommunicationManager) Subscribe(id string, types []MessageType) (*Subscriber, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if _, exists := cm.subscribers[id]; exists {
		return nil, fmt.Errorf("subscriber with ID %s already exists", id)
	}

	subscriber := &Subscriber{
		ID:      id,
		Channel: make(chan Message, 100),
		Types:   types,
	}

	cm.subscribers[id] = subscriber
	return subscriber, nil
}

// Unsubscribe removes a subscriber from the communication manager
func (cm *CommunicationManager) Unsubscribe(id string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if subscriber, exists := cm.subscribers[id]; exists {
		close(subscriber.Channel)
		delete(cm.subscribers, id)
	}
}

// Publish sends a message to all relevant subscribers
func (cm *CommunicationManager) Publish(ctx context.Context, msgType MessageType, data any) error {
	// Create the message before acquiring the lock
	message := Message{
		Type:      msgType,
		Timestamp: time.Now(),
		Data:      data,
	}

	cm.mu.RLock()
	defer cm.mu.RUnlock()

	// Create a slice of interested subscribers to avoid holding the lock while sending
	interestedSubscribers := make([]*Subscriber, 0)
	for _, subscriber := range cm.subscribers {
		for _, t := range subscriber.Types {
			if t == msgType {
				interestedSubscribers = append(interestedSubscribers, subscriber)
				break
			}
		}
	}

	// Send the message to all interested subscribers
	for _, subscriber := range interestedSubscribers {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case subscriber.Channel <- message:
			// Message sent successfully
		default:
			// If the channel is full, try one more time after a short delay
			select {
			case <-ctx.Done():
				return ctx.Err()
			case subscriber.Channel <- message:
				// Message sent successfully on retry
			case <-time.After(100 * time.Millisecond):
				// Skip this subscriber if still full after retry
				fmt.Printf("Warning: subscriber %s channel is full, skipping message\n", subscriber.ID)
			}
		}
	}

	return nil
}

// Stop stops the communication manager and all its goroutines
func (cm *CommunicationManager) Stop() {
	close(cm.stopChan)
	cm.wg.Wait()

	// Close all subscriber channels
	cm.mu.Lock()
	defer cm.mu.Unlock()
	for _, subscriber := range cm.subscribers {
		close(subscriber.Channel)
	}
	cm.subscribers = make(map[string]*Subscriber)
}

// processMessages handles message processing in a background goroutine
func (cm *CommunicationManager) processMessages() {
	defer cm.wg.Done()

	for {
		select {
		case <-cm.stopChan:
			return
		}
	}
}
