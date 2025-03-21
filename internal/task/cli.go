package task

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// CLI handles the command-line interface
type CLI struct {
	service *Service
	scanner *bufio.Scanner
}

// NewCLI creates a new CLI instance
func NewCLI(service *Service) *CLI {
	return &CLI{
		service: service,
		scanner: bufio.NewScanner(os.Stdin),
	}
}

// Run starts the CLI interface
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

// printMenu displays the main menu
func (cli *CLI) printMenu() {
	fmt.Println("\n=== TASK MANAGER ===")
	fmt.Println("1. Add task")
	fmt.Println("2. List tasks")
	fmt.Println("3. Mark as completed")
	fmt.Println("4. Exit")
	fmt.Print("Choose an option: ")
}

// readInput reads user input
func (cli *CLI) readInput() string {
	cli.scanner.Scan()
	return cli.scanner.Text()
}

// handleAddTask handles the add task operation
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

// handleListTasks handles the list tasks operation
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

// handleMarkTaskCompleted handles marking a task as completed
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
