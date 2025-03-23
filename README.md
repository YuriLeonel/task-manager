# Task Manager

A secure and efficient command-line task manager written in Go that helps you manage your daily tasks through a terminal interface. The project follows Go best practices for documentation, security, and code organization.

## Project Structure

```
.
├── cmd/
│   └── task-manager/     # Application entry point
│       └── main.go       # Main application with CLI flags
├── internal/
│   ├── storage/         # Storage implementations
│   │   └── storage.go   # JSON file storage with security features
│   └── task/            # Internal package for task management
│       ├── service.go   # Business logic
│       └── cli.go       # CLI interface
├── pkg/
│   └── models/          # Shared models
│       └── task.go      # Task entity definition
├── LICENSE              # MIT License
└── go.mod              # Go module definition
```

## Features

- Add new tasks with descriptions
- List all tasks with their completion status
- Mark tasks as completed
- Set task priority (Low, Medium, High)
- Set task due dates
- Delete tasks
- Custom data directory support
- Automatic task backup
- Configurable maximum tasks limit
- Debug mode for detailed logging
- Version information display
- Simple and intuitive command-line interface
- Secure file storage with atomic operations
- Thread-safe operations
- Resource limits and validation
- Comprehensive error handling
- Data integrity protection

## Security Features

- Secure file permissions (0600 for files, 0700 for directories)
- Path traversal protection
- Resource limits (configurable max tasks, 10MB file size)
- Atomic file operations
- Input validation
- Thread-safe operations
- Secure temporary file handling
- Data integrity checks

## Requirements

- Go 1.21 or higher

## Installation

1. Clone the repository:

   ```bash
   git clone https://github.com/YuriLeonel/task-manager.git
   ```

2. Navigate to the project directory:

   ```bash
   cd task-manager
   ```

3. Build the application:
   ```bash
   go build ./cmd/task-manager
   ```

## Running the Application

You can run the application with various command-line options:

```bash
# Show help and available options
task-manager --help

# Show version information
task-manager --version

# Use custom data directory
task-manager --data-dir ~/.my-tasks

# Disable automatic backup
task-manager --backup=false

# Set maximum number of tasks
task-manager --max-tasks 5000

# Enable debug mode for detailed logging
task-manager --debug
```

## Interactive Menu

The application provides an interactive menu with the following options:

1. **Add task**: Add a new task to your list

   - Enter a description for your task
   - Task is validated and stored securely

2. **List tasks**: Display all tasks

   - Shows tasks with their completion status ([ ] for pending, [X] for completed)
   - Displays priority level and due date
   - Tasks are displayed with their creation and update times

3. **Mark as completed**: Mark a task as completed

   - Select a task by its number
   - Updates are performed atomically

4. **Set task priority**: Update a task's priority level

   - Choose between Low, Medium, or High
   - Changes are saved immediately

5. **Set task due date**: Add or update a task's due date

   - Enter date in YYYY-MM-DD format
   - Due dates are displayed in the task list

6. **Delete task**: Remove a task from the list

   - Select a task by its number
   - Deletion is performed atomically

7. **Exit**: Close the application
   - Ensures all data is properly saved

## Data Storage

Tasks are stored in a JSON file located in the user's home directory by default:

- Default path: `~/.task-manager/tasks.json`
- Custom path: Configurable via `--data-dir` option
- Secure file permissions
- Atomic write operations
- Automatic backup handling (can be disabled)

## Documentation

The codebase follows Go's documentation conventions:

- Package documentation is available in each package's main file
- Functions and types are documented using godoc format
- Security considerations are documented
- Error handling is clearly documented
- To view the documentation locally, run:
  ```bash
  go doc ./...
  ```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

### MIT License Summary

- ✔️ Commercial use
- ✔️ Modification
- ✔️ Distribution
- ✔️ Private use
- ❗ Must include copy of license and copyright notice
- ❌ No liability or warranty

For more information about the MIT License, visit [Choose a License](https://choosealicense.com/licenses/mit/).
