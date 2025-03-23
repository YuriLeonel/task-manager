// Package main provides the entry point for the task manager application.
// It initializes the task service and CLI interface with secure storage,
// then starts the interactive command-line interface for managing tasks.
// The application ensures:
//   - Secure file storage in user's home directory
//   - Proper error handling and logging
//   - Graceful shutdown on errors
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/YuriLeonel/task-manager/internal/storage"
	"github.com/YuriLeonel/task-manager/internal/task"
)

const (
	// Version information
	version = "1.0.0"
	// Default maximum number of tasks
	defaultMaxTasks = 10000
)

var (
	// Command line flags
	dataDirFlag  = flag.String("data-dir", "", "Custom directory for storing task data")
	backupFlag   = flag.Bool("backup", true, "Enable automatic task backup")
	maxTasksFlag = flag.Int("max-tasks", defaultMaxTasks, "Maximum number of tasks allowed")
	debugFlag    = flag.Bool("debug", false, "Enable debug mode for detailed logging")
	versionFlag  = flag.Bool("version", false, "Show version information")
)

func init() {
	// Customize flag usage message
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Task Manager - A secure command-line task manager\n\n")
		fmt.Fprintf(os.Stderr, "Usage: task-manager [options]\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExample:\n")
		fmt.Fprintf(os.Stderr, "  task-manager --data-dir ~/.my-tasks --backup=false\n")
	}
}

func main() {
	// Parse command line flags
	flag.Parse()

	// Show version if requested
	if *versionFlag {
		fmt.Printf("Task Manager version %s\n", version)
		return
	}

	// Configure logging based on debug flag
	if *debugFlag {
		log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	} else {
		log.SetFlags(log.Ldate | log.Ltime)
	}

	// Get user's home directory for secure storage location
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Failed to get home directory: %v", err)
	}

	// Determine data directory
	var dataDir string
	if *dataDirFlag != "" {
		dataDir = *dataDirFlag
	} else {
		dataDir = filepath.Join(homeDir, ".task-manager")
	}

	// Create data directory with secure permissions
	if err := os.MkdirAll(dataDir, 0700); err != nil {
		log.Fatalf("Failed to create data directory: %v", err)
	}

	// Initialize secure storage with proper error handling
	storage, err := storage.NewJSONFileStorage(filepath.Join(dataDir, "tasks.json"), *maxTasksFlag)
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
