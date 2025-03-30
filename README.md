# Task Manager CLI

A secure command-line task manager written in Go that helps you manage your tasks efficiently.

[![MIT License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.18+-00ADD8.svg)](https://golang.org/)

![Task Manager](https://images.unsplash.com/photo-1540350394557-8d14678e7f91?ixlib=rb-1.2.1&ixid=eyJhcHBfaWQiOjEyMDd9&auto=format&fit=crop&w=1489&q=80)

## Features

- 📝 Create, read, update, and delete tasks
- ✅ Mark tasks as completed
- ⭐ Set task priorities (Low, Medium, High)
- 📅 Set due dates for tasks
- 🏷️ Add and remove tags for better task organization
- 📊 Track task progress (0-100%)
- 🔍 Filter tasks by tags
- 💾 Multiple storage backends (JSON file, in-memory)
- 🔄 Automatic backups with configurable interval
- 🛡️ Secure storage with proper error handling

## Getting Started

### Prerequisites

- Go 1.18 or higher

### Installation

1. Clone the repository:

```bash
git clone https://github.com/yourusername/task-manager.git
cd task-manager
```

2. Build the application:

```bash
go build -o task-manager ./cmd/task-manager
```

3. Run the application:

```bash
./task-manager
```

## Usage

### Command Line Options

The task manager supports various command line options:

```
Usage: task-manager [options]

Options:
  -backup
        Enable automatic task backup (default true)
  -backup-dir string
        Directory for storing backups (default: data-dir/backups)
  -backup-interval duration
        Interval between automatic backups (default 1h0m0s)
  -data-dir string
        Custom directory for storing task data
  -debug
        Enable debug mode for detailed logging
  -max-backups int
        Maximum number of backups to keep (default 5)
  -max-tasks int
        Maximum number of tasks allowed (default 10000)
  -storage string
        Storage type: 'json' or 'memory' (default "json")
  -version
        Show version information

Examples:
  task-manager --data-dir ~/.my-tasks --backup=false
  task-manager --storage memory --max-tasks 100
  task-manager --backup-interval 30m --max-backups 10
```

### Interactive Interface

Once running, the task manager provides an interactive menu:

```
Task Manager Menu:
1. Add task
2. List tasks
3. Mark task as completed
4. Set task priority
5. Set task due date
6. Delete task
7. Add tag to task
8. Remove tag from task
9. Set task progress
10. List tasks by tag
11. Exit
```

## Project Structure

```
task-manager/
├── cmd/                   # Command-line applications
│   └── task-manager/      # Main application entry point
├── internal/              # Private application and library code
│   ├── storage/           # Storage implementations
│   └── task/              # Task management business logic
├── pkg/                   # Library code that can be used by external applications
│   └── models/            # Data models
├── go.mod                 # Go module definition
├── LICENSE                # MIT License
└── README.md              # This file
```

## Architecture

The Task Manager follows a clean architecture with clear separation of concerns:

1. **Models** (`pkg/models`): Core data structures
2. **Storage** (`internal/storage`): Data persistence layer with multiple implementations
3. **Service** (`internal/task`): Business logic layer
4. **CLI** (`internal/task`): Presentation layer
5. **Main** (`cmd/task-manager`): Application entry point

## Security Features

- Secure file permissions (0600 for files, 0700 for directories)
- Path traversal protection
- Resource limits (configurable maximum tasks)
- Atomic file operations
- Input validation
- Thread-safe operations
- Secure temporary file handling
- Data integrity checks

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Task Manager was created as a learning project for Go programming
- Inspired by various task management systems and productivity tools
