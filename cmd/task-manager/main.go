// Package main provides the entry point for the task manager application.
// It initializes the task service and CLI interface with secure storage,
// then starts the interactive command-line interface for managing tasks.
// The application ensures:
//   - Secure file storage in user's home directory
//   - Proper error handling and logging
//   - Graceful shutdown on errors
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/YuriLeonel/task-manager/internal/storage"
	"github.com/YuriLeonel/task-manager/internal/task"
)

func main() {
	// Get user's home directory for secure storage location
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Failed to get home directory: %v", err)
	}

	// Create data directory in user's home with secure permissions
	dataDir := filepath.Join(homeDir, ".task-manager")
	if err := os.MkdirAll(dataDir, 0700); err != nil {
		log.Fatalf("Failed to create data directory: %v", err)
	}

	// Initialize secure storage with proper error handling
	storage, err := storage.NewJSONFileStorage(filepath.Join(dataDir, "tasks.json"))
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}

	// Initialize service with secure storage
	service := task.NewService(storage)

	// Initialize CLI with service
	cli := task.NewCLI(service)

	// Run the application with error handling
	if err := cli.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
