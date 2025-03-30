// Package task provides the command-line interface for task management.
// It includes input validation, secure data handling, and user-friendly error messages.
package task

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/YuriLeonel/task-manager/pkg/models"
)

// CLI provides the command-line interface for task management.
// It ensures secure input handling and proper error reporting.
type CLI struct {
	service *Service
	reader  *bufio.Reader
}

// NewCLI creates a new CLI instance with secure service integration.
func NewCLI(service *Service) *CLI {
	return &CLI{
		service: service,
		reader:  bufio.NewReader(os.Stdin),
	}
}

// Run starts the interactive CLI session.
// It continuously displays the menu and processes user input until
// the user chooses to exit. This method blocks until the program
// is terminated.
func (cli *CLI) Run() error {
	for {
		cli.displayMenu()
		choice, err := cli.readInput("Enter your choice: ")
		if err != nil {
			return fmt.Errorf("failed to read input: %w", err)
		}

		switch choice {
		case "1":
			if err := cli.handleAddTask(); err != nil {
				fmt.Printf("Error: %v\n", err)
			}
		case "2":
			if err := cli.handleListTasks(); err != nil {
				fmt.Printf("Error: %v\n", err)
			}
		case "3":
			if err := cli.handleMarkAsCompleted(); err != nil {
				fmt.Printf("Error: %v\n", err)
			}
		case "4":
			if err := cli.handleSetPriority(); err != nil {
				fmt.Printf("Error: %v\n", err)
			}
		case "5":
			if err := cli.handleSetDueDate(); err != nil {
				fmt.Printf("Error: %v\n", err)
			}
		case "6":
			if err := cli.handleDeleteTask(); err != nil {
				fmt.Printf("Error: %v\n", err)
			}
		case "7":
			if err := cli.handleAddTag(); err != nil {
				fmt.Printf("Error: %v\n", err)
			}
		case "8":
			if err := cli.handleRemoveTag(); err != nil {
				fmt.Printf("Error: %v\n", err)
			}
		case "9":
			if err := cli.handleSetProgress(); err != nil {
				fmt.Printf("Error: %v\n", err)
			}
		case "10":
			if err := cli.handleListTasksByTag(); err != nil {
				fmt.Printf("Error: %v\n", err)
			}
		case "11":
			return nil
		default:
			fmt.Println("Invalid choice. Please try again.")
		}
	}
}

// displayMenu shows the available options to the user.
// It shows all available commands and prompts for user input.
func (cli *CLI) displayMenu() {
	fmt.Println("\nTask Manager Menu:")
	fmt.Println("1. Add task")
	fmt.Println("2. List tasks")
	fmt.Println("3. Mark task as completed")
	fmt.Println("4. Set task priority")
	fmt.Println("5. Set task due date")
	fmt.Println("6. Delete task")
	fmt.Println("7. Add tag to task")
	fmt.Println("8. Remove tag from task")
	fmt.Println("9. Set task progress")
	fmt.Println("10. List tasks by tag")
	fmt.Println("11. Exit")
}

// readInput reads a line of input from the user.
// It returns the trimmed string containing the user's input.
func (cli *CLI) readInput(prompt string) (string, error) {
	fmt.Print(prompt)
	input, err := cli.reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(input), nil
}

// handleAddTask handles the task creation flow.
// It prompts the user for a task description and adds it to the task list.
// If the description is empty, an error message is displayed.
func (cli *CLI) handleAddTask() error {
	description, err := cli.readInput("Enter task description: ")
	if err != nil {
		return err
	}

	if err := cli.service.AddTask(description); err != nil {
		return err
	}

	fmt.Println("Task added successfully!")
	return nil
}

// handleListTasks displays all tasks with their status and details.
func (cli *CLI) handleListTasks() error {
	tasks, err := cli.service.ListTasks()
	if err != nil {
		return fmt.Errorf("failed to get tasks: %w", err)
	}

	if len(tasks) == 0 {
		fmt.Println("\nNo tasks found.")
		return nil
	}

	fmt.Println("\nTasks:")
	for i, task := range tasks {
		status := "[ ]"
		if task.Completed {
			status = "[X]"
		}
		priority := getPriorityString(task.Priority)
		dueDate := "No due date"
		if task.DueDate != nil {
			dueDate = task.DueDate.Format("2006-01-02")
		}
		tags := "No tags"
		if len(task.Tags) > 0 {
			tags = strings.Join(task.Tags, ", ")
		}
		fmt.Printf("%d. %s %s (Priority: %s, Due: %s, Progress: %d%%, Tags: %s)\n",
			i+1, status, task.Description, priority, dueDate, task.Progress, tags)
	}
	return nil
}

// handleMarkAsCompleted handles marking a task as completed.
// It first checks if there are any tasks in the list.
// If tasks exist, it shows the current list and prompts for a task number.
// The selected task is then marked as completed if the input is valid.
func (cli *CLI) handleMarkAsCompleted() error {
	tasks, err := cli.service.ListTasks()
	if err != nil {
		return err
	}

	if len(tasks) == 0 {
		fmt.Println("No tasks available to mark as completed.")
		return nil
	}

	cli.handleListTasks()
	indexStr, err := cli.readInput("Enter task number to mark as completed: ")
	if err != nil {
		return err
	}

	index, err := strconv.Atoi(indexStr)
	if err != nil {
		return fmt.Errorf("invalid task number")
	}

	if index < 1 || index > len(tasks) {
		return fmt.Errorf("invalid task number")
	}

	task := tasks[index-1]
	if err := cli.service.MarkAsCompleted(task.ID); err != nil {
		return err
	}

	fmt.Println("Task marked as completed!")
	return nil
}

// handleSetPriority handles setting a task's priority.
// It first checks if there are any tasks in the list.
// If tasks exist, it shows the current list and prompts for a task number.
// The selected task's priority is then updated if the input is valid.
func (cli *CLI) handleSetPriority() error {
	tasks, err := cli.service.ListTasks()
	if err != nil {
		return err
	}

	if len(tasks) == 0 {
		fmt.Println("No tasks available to set priority.")
		return nil
	}

	cli.handleListTasks()
	indexStr, err := cli.readInput("Enter task number to set priority: ")
	if err != nil {
		return err
	}

	index, err := strconv.Atoi(indexStr)
	if err != nil {
		return fmt.Errorf("invalid task number")
	}

	if index < 1 || index > len(tasks) {
		return fmt.Errorf("invalid task number")
	}

	fmt.Println("\nPriority levels:")
	fmt.Println("1. Low")
	fmt.Println("2. Medium")
	fmt.Println("3. High")

	priorityStr, err := cli.readInput("Enter priority level (1-3): ")
	if err != nil {
		return err
	}

	priority, err := strconv.Atoi(priorityStr)
	if err != nil {
		return fmt.Errorf("invalid priority level")
	}

	if priority < 1 || priority > 3 {
		return fmt.Errorf("invalid priority level")
	}

	task := tasks[index-1]
	if err := cli.service.SetPriority(task.ID, models.Priority(priority-1)); err != nil {
		return err
	}

	fmt.Println("Task priority updated!")
	return nil
}

// handleSetDueDate handles setting a task's due date.
// It first checks if there are any tasks in the list.
// If tasks exist, it shows the current list and prompts for a task number.
// The selected task's due date is then updated if the input is valid.
func (cli *CLI) handleSetDueDate() error {
	tasks, err := cli.service.ListTasks()
	if err != nil {
		return err
	}

	if len(tasks) == 0 {
		fmt.Println("No tasks available to set due date.")
		return nil
	}

	cli.handleListTasks()
	indexStr, err := cli.readInput("Enter task number to set due date: ")
	if err != nil {
		return err
	}

	index, err := strconv.Atoi(indexStr)
	if err != nil {
		return fmt.Errorf("invalid task number")
	}

	if index < 1 || index > len(tasks) {
		return fmt.Errorf("invalid task number")
	}

	dueDateStr, err := cli.readInput("Enter due date (YYYY-MM-DD): ")
	if err != nil {
		return err
	}

	dueDate, err := time.Parse("2006-01-02", dueDateStr)
	if err != nil {
		return fmt.Errorf("invalid date format. Use YYYY-MM-DD")
	}

	task := tasks[index-1]
	if err := cli.service.SetDueDate(task.ID, dueDate); err != nil {
		return err
	}

	fmt.Println("Task due date updated!")
	return nil
}

// handleDeleteTask handles deleting a task.
// It first checks if there are any tasks in the list.
// If tasks exist, it shows the current list and prompts for a task number.
// The selected task is then deleted if the input is valid.
func (cli *CLI) handleDeleteTask() error {
	tasks, err := cli.service.ListTasks()
	if err != nil {
		return err
	}

	if len(tasks) == 0 {
		fmt.Println("No tasks available to delete.")
		return nil
	}

	cli.handleListTasks()
	indexStr, err := cli.readInput("Enter task number to delete: ")
	if err != nil {
		return err
	}

	index, err := strconv.Atoi(indexStr)
	if err != nil {
		return fmt.Errorf("invalid task number")
	}

	if index < 1 || index > len(tasks) {
		return fmt.Errorf("invalid task number")
	}

	task := tasks[index-1]
	if err := cli.service.DeleteTask(task.ID); err != nil {
		return err
	}

	fmt.Println("Task deleted successfully!")
	return nil
}

// getPriorityString returns a string representation of the task priority.
func getPriorityString(priority models.Priority) string {
	switch priority {
	case models.Low:
		return "Low"
	case models.Medium:
		return "Medium"
	case models.High:
		return "High"
	default:
		return "Unknown"
	}
}

// handleAddTag handles adding a tag to a task.
func (cli *CLI) handleAddTag() error {
	tasks, err := cli.service.ListTasks()
	if err != nil {
		return err
	}

	if len(tasks) == 0 {
		fmt.Println("No tasks available to add tags.")
		return nil
	}

	cli.handleListTasks()
	indexStr, err := cli.readInput("Enter task number to add tag: ")
	if err != nil {
		return err
	}

	index, err := strconv.Atoi(indexStr)
	if err != nil {
		return fmt.Errorf("invalid task number")
	}

	if index < 1 || index > len(tasks) {
		return fmt.Errorf("invalid task number")
	}

	tag, err := cli.readInput("Enter tag name: ")
	if err != nil {
		return err
	}

	if tag == "" {
		return fmt.Errorf("tag cannot be empty")
	}

	task := tasks[index-1]
	if err := cli.service.AddTag(task.ID, tag); err != nil {
		return err
	}

	fmt.Println("Tag added successfully!")
	return nil
}

// handleRemoveTag handles removing a tag from a task.
func (cli *CLI) handleRemoveTag() error {
	tasks, err := cli.service.ListTasks()
	if err != nil {
		return err
	}

	if len(tasks) == 0 {
		fmt.Println("No tasks available to remove tags.")
		return nil
	}

	cli.handleListTasks()
	indexStr, err := cli.readInput("Enter task number to remove tag: ")
	if err != nil {
		return err
	}

	index, err := strconv.Atoi(indexStr)
	if err != nil {
		return fmt.Errorf("invalid task number")
	}

	if index < 1 || index > len(tasks) {
		return fmt.Errorf("invalid task number")
	}

	task := tasks[index-1]
	if len(task.Tags) == 0 {
		fmt.Println("This task has no tags.")
		return nil
	}

	// Display existing tags
	fmt.Println("Existing tags:")
	for i, tag := range task.Tags {
		fmt.Printf("%d. %s\n", i+1, tag)
	}

	tagIndexStr, err := cli.readInput("Enter tag number to remove: ")
	if err != nil {
		return err
	}

	tagIndex, err := strconv.Atoi(tagIndexStr)
	if err != nil {
		return fmt.Errorf("invalid tag number")
	}

	if tagIndex < 1 || tagIndex > len(task.Tags) {
		return fmt.Errorf("invalid tag number")
	}

	tag := task.Tags[tagIndex-1]
	if err := cli.service.RemoveTag(task.ID, tag); err != nil {
		return err
	}

	fmt.Println("Tag removed successfully!")
	return nil
}

// handleSetProgress handles setting the progress of a task.
func (cli *CLI) handleSetProgress() error {
	tasks, err := cli.service.ListTasks()
	if err != nil {
		return err
	}

	if len(tasks) == 0 {
		fmt.Println("No tasks available to set progress.")
		return nil
	}

	cli.handleListTasks()
	indexStr, err := cli.readInput("Enter task number to set progress: ")
	if err != nil {
		return err
	}

	index, err := strconv.Atoi(indexStr)
	if err != nil {
		return fmt.Errorf("invalid task number")
	}

	if index < 1 || index > len(tasks) {
		return fmt.Errorf("invalid task number")
	}

	progressStr, err := cli.readInput("Enter progress percentage (0-100): ")
	if err != nil {
		return err
	}

	progress, err := strconv.Atoi(progressStr)
	if err != nil {
		return fmt.Errorf("invalid progress value")
	}

	if progress < 0 || progress > 100 {
		return fmt.Errorf("progress must be between 0 and 100")
	}

	task := tasks[index-1]
	if err := cli.service.SetProgress(task.ID, progress); err != nil {
		return err
	}

	fmt.Printf("Progress set to %d%%\n", progress)
	if progress == 100 {
		fmt.Println("Task marked as completed!")
	} else if task.Completed {
		fmt.Println("Task marked as incomplete.")
	}

	return nil
}

// handleListTasksByTag displays all tasks with a specific tag.
func (cli *CLI) handleListTasksByTag() error {
	tag, err := cli.readInput("Enter tag to filter by: ")
	if err != nil {
		return err
	}

	if tag == "" {
		return fmt.Errorf("tag cannot be empty")
	}

	tasks, err := cli.service.GetTasksByTag(tag)
	if err != nil {
		return err
	}

	if len(tasks) == 0 {
		fmt.Printf("\nNo tasks found with tag '%s'.\n", tag)
		return nil
	}

	fmt.Printf("\nTasks with tag '%s':\n", tag)
	for i, task := range tasks {
		status := "[ ]"
		if task.Completed {
			status = "[X]"
		}
		priority := getPriorityString(task.Priority)
		dueDate := "No due date"
		if task.DueDate != nil {
			dueDate = task.DueDate.Format("2006-01-02")
		}

		fmt.Printf("%d. %s %s (Priority: %s, Due: %s, Progress: %d%%)\n",
			i+1, status, task.Description, priority, dueDate, task.Progress)
	}
	return nil
}
