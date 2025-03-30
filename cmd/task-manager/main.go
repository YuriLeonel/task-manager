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
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/YuriLeonel/task-manager/internal/storage"
	"github.com/YuriLeonel/task-manager/internal/task"
)

const (
	// Version information
	version = "1.1.0"
	// Default maximum number of tasks
	defaultMaxTasks = 10000
	// Default backup interval (1 hour)
	defaultBackupInterval = 1 * time.Hour
	// Default maximum number of backups
	defaultMaxBackups = 5
)

var (
	// Command line flags
	dataDirFlag    = flag.String("data-dir", "", "Custom directory for storing task data")
	backupFlag     = flag.Bool("backup", true, "Enable automatic task backup")
	maxTasksFlag   = flag.Int("max-tasks", defaultMaxTasks, "Maximum number of tasks allowed")
	debugFlag      = flag.Bool("debug", false, "Enable debug mode for detailed logging")
	versionFlag    = flag.Bool("version", false, "Show version information")
	storageType    = flag.String("storage", "json", "Storage type: 'json' or 'memory'")
	backupDir      = flag.String("backup-dir", "", "Directory for storing backups (default: data-dir/backups)")
	backupInterval = flag.Duration("backup-interval", defaultBackupInterval, "Interval between automatic backups")
	maxBackups     = flag.Int("max-backups", defaultMaxBackups, "Maximum number of backups to keep")
)

func init() {
	// Customize flag usage message
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Task Manager - A secure command-line task manager\n\n")
		fmt.Fprintf(os.Stderr, "Usage: task-manager [options]\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  task-manager --data-dir ~/.my-tasks --backup=false\n")
		fmt.Fprintf(os.Stderr, "  task-manager --storage memory --max-tasks 100\n")
		fmt.Fprintf(os.Stderr, "  task-manager --backup-interval 30m --max-backups 10\n")
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

	// Initialize appropriate storage
	var storageImpl storage.Storage
	var backupManager *storage.BackupManager
	var err error
	var taskFilePath string

	switch *storageType {
	case "memory":
		// Memory storage (ephemeral, for testing)
		log.Println("Using in-memory storage")
		storageImpl, err = storage.NewMemoryStorage(*maxTasksFlag)
		if err != nil {
			log.Fatalf("Failed to initialize memory storage: %v", err)
		}

	case "json", "":
		// Default JSON file storage
		log.Println("Using JSON file storage")
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

		// Determine task file path
		taskFilePath = filepath.Join(dataDir, "tasks.json")

		// Initialize file-based storage
		storageImpl, err = storage.NewJSONFileStorage(taskFilePath, *maxTasksFlag)
		if err != nil {
			log.Fatalf("Failed to initialize storage: %v", err)
		}

		// Setup backup system if enabled
		if *backupFlag {
			// Determine backup directory
			var backupDirectory string
			if *backupDir != "" {
				backupDirectory = *backupDir
			} else {
				backupDirectory = filepath.Join(dataDir, "backups")
			}

			// Create backup manager
			backupManager, err = storage.NewBackupManager(
				taskFilePath,
				backupDirectory,
				*backupInterval,
				*maxBackups,
			)
			if err != nil {
				log.Printf("Warning: Failed to initialize backup manager: %v", err)
			} else {
				backupManager.Start()
				log.Printf("Automatic backup enabled with interval %s", backupInterval)
			}
		}

	default:
		log.Fatalf("Unknown storage type: %s. Valid options are 'json' or 'memory'", *storageType)
	}

	// Initialize service with secure storage
	service := task.NewService(storageImpl)

	// Initialize CLI with service
	cli := task.NewCLI(service)

	// Handle SIGINT and SIGTERM for graceful shutdown
	setupGracefulShutdown(backupManager)

	// Run the application with error handling
	if err := cli.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		// Trigger backup before exit if backup manager is enabled
		if backupManager != nil {
			backupManager.TriggerBackup()
			backupManager.Stop()
		}
		os.Exit(1)
	}

	// Clean shutdown
	if backupManager != nil {
		// Trigger final backup
		backupManager.TriggerBackup()
		backupManager.Stop()
	}
}

// setupGracefulShutdown sets up signal handling for graceful shutdown
func setupGracefulShutdown(backupManager *storage.BackupManager) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		fmt.Println("\nShutting down...")

		// Trigger backup before exit if backup manager is enabled
		if backupManager != nil {
			fmt.Println("Creating final backup...")
			backupManager.TriggerBackup()
			backupManager.Stop()
		}

		os.Exit(0)
	}()
}
