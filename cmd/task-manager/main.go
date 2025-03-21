// Package main is the entry point for the task manager application.
// It initializes the task service and CLI interface, then starts
// the interactive command-line interface for managing tasks.
package main

import (
	"github.com/YuriLeonel/task-manager/internal/task"
)

func main() {
	// Initialize the task service
	service := task.NewService()

	// Initialize the CLI with the task service
	cli := task.NewCLI(service)

	// Start the application's main loop
	cli.Run()
}
