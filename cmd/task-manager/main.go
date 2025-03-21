package main

import (
	"github.com/YuriLeonel/task-manager/internal/task"
)

func main() {
	// Initialize the task service
	service := task.NewService()

	// Initialize the CLI
	cli := task.NewCLI(service)

	// Start the application
	cli.Run()
}
