package main

import (
	"bufio"
	"fmt"
	"os"
)

// Task represents a task in our list
type Task struct {
	Description string
	Completed   bool
}

func main() {
	// Slice to store tasks
	var tasks []Task

	// Scanner to read user input
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Println("\n=== TASK MANAGER ===")
		fmt.Println("1. Add task")
		fmt.Println("2. List tasks")
		fmt.Println("3. Mark as completed")
		fmt.Println("4. Exit")
		fmt.Print("Choose an option: ")

		scanner.Scan()
		option := scanner.Text()

		switch option {
		case "1":
			fmt.Print("Enter task description: ")
			scanner.Scan()
			description := scanner.Text()

			newTask := Task{
				Description: description,
				Completed:   false,
			}
			tasks = append(tasks, newTask)
			fmt.Println("Task added successfully!")

		case "2":
			if len(tasks) == 0 {
				fmt.Println("No tasks in the list.")
			} else {
				fmt.Println("\nTASK LIST:")
				for i, task := range tasks {
					status := " "
					if task.Completed {
						status = "X"
					}
					fmt.Printf("[%s] %d. %s\n", status, i+1, task.Description)
				}
			}

		case "3":
			if len(tasks) == 0 {
				fmt.Println("No tasks in the list.")
				continue
			}

			fmt.Println("\nTASK LIST:")
			for i, task := range tasks {
				status := " "
				if task.Completed {
					status = "X"
				}
				fmt.Printf("[%s] %d. %s\n", status, i+1, task.Description)
			}

			fmt.Print("Enter the task number to mark as completed: ")
			scanner.Scan()
			taskNumberStr := scanner.Text()

			var taskNumber int
			_, err := fmt.Sscanf(taskNumberStr, "%d", &taskNumber)
			if err != nil || taskNumber < 1 || taskNumber > len(tasks) {
				fmt.Println("Invalid task number.")
				continue
			}

			tasks[taskNumber-1].Completed = true
			fmt.Println("Task marked as completed!")

		case "4":
			fmt.Println("Exiting. Bye!")
			return

		default:
			fmt.Println("Invalid option. Please try again.")
		}
	}
}
