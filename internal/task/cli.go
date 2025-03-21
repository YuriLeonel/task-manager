package task

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// CLI handles the command-line interface for the task manager.
// It provides an interactive interface for users to manage their tasks
// through the terminal. The CLI struct maintains a reference to the
// task service and a scanner for reading user input.
type CLI struct {
	service *Service
	scanner *bufio.Scanner
}

// NewCLI creates a new CLI instance with the provided task service.
// It initializes a scanner for reading user input from standard input.
func NewCLI(service *Service) *CLI {
	return &CLI{
		service: service,
		scanner: bufio.NewScanner(os.Stdin),
	}
}

// Run starts the CLI interface and enters the main program loop.
// It continuously displays the menu and processes user input until
// the user chooses to exit. This method blocks until the program
// is terminated.
func (cli *CLI) Run() {
	for {
		cli.printMenu()
		option := cli.readInput()

		switch strings.TrimSpace(option) {
		case "1":
			cli.handleAddTask()
		case "2":
			cli.handleListTasks()
		case "3":
			cli.handleMarkTaskCompleted()
		case "4":
			fmt.Println("Exiting. Bye!")
			return
		default:
			fmt.Println("Invalid option. Please try again.")
		}
	}
}

// printMenu displays the main menu options to the user.
// It shows all available commands and prompts for user input.
func (cli *CLI) printMenu() {
	fmt.Println("\n=== TASK MANAGER ===")
	fmt.Println("1. Add task")
	fmt.Println("2. List tasks")
	fmt.Println("3. Mark as completed")
	fmt.Println("4. Exit")
	fmt.Print("Choose an option: ")
}

// readInput reads a line of text from standard input.
// It returns the trimmed string containing the user's input.
func (cli *CLI) readInput() string {
	cli.scanner.Scan()
	return cli.scanner.Text()
}

// handleAddTask processes the add task command.
// It prompts the user for a task description and adds it to the task list.
// If the description is empty, an error message is displayed.
func (cli *CLI) handleAddTask() {
	fmt.Print("Enter task description: ")
	description := cli.readInput()

	err := cli.service.AddTask(description)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("Task added successfully!")
}

// handleListTasks displays all tasks in the list.
// If there are no tasks, it shows an appropriate message.
// For each task, it shows the completion status ([X] for completed,
// [ ] for pending) followed by the task number and description.
func (cli *CLI) handleListTasks() {
	tasks := cli.service.ListTasks()

	if len(tasks) == 0 {
		fmt.Println("No tasks in the list.")
		return
	}

	fmt.Println("\nTASK LIST:")
	for i, task := range tasks {
		fmt.Printf("[%s] %d. %s\n", task.String(), i+1, task.Description)
	}
}

// handleMarkTaskCompleted processes the mark task as completed command.
// It first checks if there are any tasks in the list.
// If tasks exist, it shows the current list and prompts for a task number.
// The selected task is then marked as completed if the input is valid.
func (cli *CLI) handleMarkTaskCompleted() {
	if cli.service.GetTaskCount() == 0 {
		fmt.Println("No tasks in the list.")
		return
	}

	cli.handleListTasks()

	fmt.Print("Enter the task number to mark as completed: ")
	indexStr := cli.readInput()

	var index int
	_, err := fmt.Sscanf(indexStr, "%d", &index)
	if err != nil {
		fmt.Println("Invalid task number.")
		return
	}

	err = cli.service.MarkTaskAsCompleted(index - 1)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("Task marked as completed!")
}
