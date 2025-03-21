# Task Manager

A command-line task manager written in Go that helps you manage your daily tasks through a terminal interface.

## Project Structure

```
.
├── cmd/
│   └── task-manager/     # Application entry point
│       └── main.go
├── internal/
│   └── task/            # Internal package for task management
│       ├── service.go   # Business logic
│       └── cli.go       # CLI interface
├── pkg/
│   └── models/          # Shared models
│       └── task.go      # Task entity definition
└── go.mod              # Go module definition
```

## Features

- Add new tasks with descriptions
- List all tasks with their completion status
- Mark tasks as completed
- Simple and intuitive command-line interface
- Error handling for invalid inputs

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

You can run the application in two ways:

1. Using `go run`:

   ```bash
   go run ./cmd/task-manager
   ```

2. Using the built binary (after building):
   ```bash
   ./task-manager
   ```

## Usage

The application provides an interactive menu with the following options:

1. **Add task**: Add a new task to your list

   - Enter a description for your task

2. **List tasks**: Display all tasks

   - Shows tasks with their completion status ([ ] for pending, [X] for completed)

3. **Mark as completed**: Mark a task as completed

   - Select a task by its number

4. **Exit**: Close the application

## Documentation

The codebase follows Go's documentation conventions:

- Package documentation is available in each package's main file
- Functions and types are documented using godoc format
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
