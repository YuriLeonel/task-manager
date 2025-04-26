// Package main provides the entry point for the task manager application.
// It initializes the task service and CLI interface with secure storage,
// then starts the interactive command-line interface for managing tasks.
// The application ensures:
//   - Secure file storage in user's home directory
//   - Proper error handling and logging
//   - Graceful shutdown on errors
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/YuriLeonel/task-manager/internal/communication"
	"github.com/YuriLeonel/task-manager/internal/config"
	"github.com/YuriLeonel/task-manager/internal/logger"
	"github.com/YuriLeonel/task-manager/internal/storage"
	"github.com/YuriLeonel/task-manager/internal/storage/backup"
	"github.com/YuriLeonel/task-manager/internal/task"
)

func main() {
	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Load and validate configuration
	cfg, err := config.DefaultConfig()
	if err != nil {
		fmt.Printf("Failed to load default config: %v\n", err)
		os.Exit(1)
	}

	if err := cfg.LoadFromEnv(); err != nil {
		fmt.Printf("Failed to load environment variables: %v\n", err)
		os.Exit(1)
	}

	if err := cfg.Validate(); err != nil {
		fmt.Printf("Invalid configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	log, err := logger.NewLogger(cfg.GetLogDir(), cfg.Debug)
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	log.Info("Logger initialized successfully")

	// Initialize communication manager
	comm := communication.NewCommunicationManager()

	// Initialize backup manager
	backupMgr, err := backup.NewConcurrentManager(cfg.BackupDir, cfg.MaxBackups, comm)
	if err != nil {
		log.Error("Failed to initialize backup manager: %v", err)
		os.Exit(1)
	}
	log.Info("Backup manager initialized successfully")

	// Initialize storage based on type
	var store storage.Storage
	if cfg.StorageType == "memory" {
		store, err = storage.NewMemoryStorage(cfg.MaxTasks, cfg.BackupDir, cfg.MaxBackups)
	} else if cfg.StorageType == "file" {
		store, err = storage.NewJSONFileStorage(cfg.GetTasksFilePath(), cfg.MaxTasks, cfg.BackupDir, cfg.MaxBackups)
	} else {
		log.Error("Invalid storage type: %s", cfg.StorageType)
		os.Exit(1)
	}

	if err != nil {
		log.Error("Failed to initialize storage: %v", err)
		os.Exit(1)
	}
	log.Info("Storage initialized successfully with type: %s", cfg.StorageType)

	// Initialize task service
	taskService := task.NewService(store)
	log.Info("Task service initialized successfully")

	// Initialize task manager with storage and service
	taskManager := task.NewManager(comm, backupMgr, store, taskService)
	log.Info("Task manager initialized successfully")

	// Initialize and start CLI
	cli := task.NewCLI(taskService)
	log.Info("CLI initialized successfully")

	// Start CLI in a goroutine
	cliDone := make(chan error)
	go func() {
		cliDone <- cli.Run()
	}()

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for either CLI exit or system signal
	select {
	case err := <-cliDone:
		if err != nil {
			log.Error("CLI error: %v", err)
		}
		log.Info("CLI exited normally")
	case sig := <-sigChan:
		log.Info("Received signal: %v", sig)
	}

	// Create shutdown context with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 5*time.Second)
	defer shutdownCancel()

	// Stop components in reverse order
	select {
	case <-shutdownCtx.Done():
		log.Error("Shutdown timeout exceeded")
	default:
		taskManager.Stop()
		log.Info("Task manager stopped successfully")

		backupMgr.Stop()
		log.Info("Backup manager stopped successfully")

		comm.Stop()
		log.Info("Communication manager stopped successfully")
	}
}
