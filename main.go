package main

import (
	"fmt"
	"os"
	"encoding/json"
	"time"
	"strconv"
	"errors"
	"io"
	"strings"
)

const version = "0.1.0"

type Task struct {
	Id int				`json:"id"`
	Description string	`json:"description"`
	Status string		`json:"status"`
	CreatedAt string	`json:"createdAt"`
	UpdatedAt string	`json:"updatedAt"`
}

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	if len(args) == 0 {
		rootUsage()
		return 2
	}

	switch args[0] {
	case "add":
		return cmdAdd(args[1:])
	case "update":
		return cmdUpdate(args[1:])
	case "delete":
		return cmdDelete(args[1:])
	case "list":
		return cmdList(args[1:])
	case "mark-in-progress":
		return cmdMarkStatus(args[1:], "in-progress")
	case "mark-done":
		return cmdMarkStatus(args[1:], "done")
	case "-h", "--help", "help":
		rootUsage()
		return 0
	case "-v", "--version":
		fmt.Printf("task-tracker-cli %s\n", version)
		return 0
	default:
		rootUsage()
		return 2
	}
}


func rootUsage() {
	fmt.Fprintf(os.Stderr, `task-tracker-cli %s

Usage:
  task-cli <command> [options]

Commands:
  add     			Add a new task with a description
  update      			Update the description for the specified task
  delete			Delete specified task
  mark-in-progress		Update the status of specified task to in-progress
  mark-done			Update the status of specified task to done
  list				List all tasks or filtered tasks by status

Global:
  -h, --help       Show this help
  -v, --version    Show version
`, version)
}

func cmdAdd(args []string) int {
	if len(args) == 0 {
        fmt.Fprintln(os.Stderr, "Error: no task description provided.")
        fmt.Fprintln(os.Stderr, "Usage: task-cli add <description> [<description> ...]")
        return 2
    }

	tasks, err := loadFileData()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	nextId := 1
	if len(tasks) > 0 {
		nextId = tasks[len(tasks)-1].Id + 1
	}

	// Create new tasks
	for _, task := range args {
		new_task := Task{
			Id: nextId, 
			Description: task, 
			Status: "todo", 
			CreatedAt: time.Now().Format(time.DateTime), 
			UpdatedAt: time.Now().Format(time.DateTime), 
		}
		tasks = append(tasks, new_task)
		nextId += 1
	}

	if err := saveFileData(tasks); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	fmt.Printf("Successfully added %d tasks\n", len(args))
	return 0
}

func cmdUpdate(args []string) int {
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: task-cli update <id> <description>")
		return 2
	}

	id, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error: first argument must be an integer task id.")
		return 2
	}

	tasks, err := loadFileData()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	idx := findIndexByTaskId(tasks, id)
	if idx == -1 {
        fmt.Fprintf(os.Stderr, "Error: id %d not found in task list\n", id)
        return 1
	}

	// Update fields and save back to file
	tasks[idx].Description = args[1]
	tasks[idx].UpdatedAt = time.Now().Format(time.DateTime)

	if err := saveFileData(tasks); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	fmt.Printf("Successfully updated task %d\n", id)
	return 0
}

func cmdDelete(args []string) int {
	if len(args) == 0 {
        fmt.Fprintln(os.Stderr, "Error: no task id provided.")
        fmt.Fprintln(os.Stderr, "Usage: task-cli delete <id>")
        return 2
    }

	id, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error: first argument must be an integer task id.")
		return 2
	}

	tasks, err := loadFileData()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	idx := findIndexByTaskId(tasks, id)
	if idx == -1 {
        fmt.Fprintf(os.Stderr, "Error: id %d not found in task list\n", id)
        return 1
	}

	tasks = append(tasks[:idx], tasks[idx+1:]...)

	if err := saveFileData(tasks); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	fmt.Printf("Successfully deleted task %d\n", id)
	return 0
}

func cmdList(args []string) int {
	// Validate arguments and check for filter
	var filter string
	switch len(args) {
	case 0:
	case 1:
		filter = strings.ToLower(args[0])
		if !isValidStatus(filter) {
			fmt.Fprintln(os.Stderr, "Error: invalid status")
			fmt.Fprintln(os.Stderr, "Usage: task-cli list [todo|in-progress|done]")
			return 2
		}
	default:
		fmt.Fprintln(os.Stderr, "Error: too many arguments")
		fmt.Fprintln(os.Stderr, "Usage: task-cli list [todo|in-progress|done]")
		return 2
	}

	tasks, err := loadFileData()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	filtered_tasks := filterByStatus(filter, tasks)

	if len(filtered_tasks) == 0 {
		if filter == "" {
			fmt.Println("No tasks available")
		} else {
			fmt.Printf("No tasks with status %s.\n", filter)
		}
		return 0
	}

	for _, task := range filtered_tasks {
		fmt.Printf("%-4d %-12s %s\n", task.Id, task.Status, task.Description)
	}
	return 0
}

func cmdMarkStatus(args []string, newStatus string) int {
	if len(args) == 0 {
        fmt.Fprintln(os.Stderr, "Error: no task id provided.")
        fmt.Fprintf(os.Stderr, "Usage: task-cli mark-%s <id>\n", newStatus)
        return 2
    }

	id, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error: first argument must be an integer task id.")
		return 2
	}

	tasks, err := loadFileData()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	idx := findIndexByTaskId(tasks, id)
	if idx == -1 {
        fmt.Fprintf(os.Stderr, "Error: id %d not found in task list\n", id)
        return 1
	}

	if tasks[idx].Status == newStatus {
		fmt.Printf("Task %d is already %s.\n", id, newStatus)
        return 0
	}

	tasks[idx].Status = newStatus
	tasks[idx].UpdatedAt = time.Now().Format(time.DateTime)

	if err := saveFileData(tasks); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	fmt.Printf("Successfully updated task %d status to %s\n", id, newStatus)
	return 0
}

func loadFileData() ([]Task, error) {
	var tasks []Task

	file, err := os.Open("tasks.json")
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []Task{}, nil
		} else {
        	return nil, fmt.Errorf("Error opening tasks.json: %w", err)
		}
	}	
	defer file.Close()
	if err := json.NewDecoder(file).Decode(&tasks) ; err != nil {
		if errors.Is(err, io.EOF) {
			return []Task{}, nil 	// treat empty file as empty tasks list
		}
		return nil, fmt.Errorf("Error decoding json data: %w", err)
	}

	return tasks, nil
}

func saveFileData(tasks []Task) error {
	// Convert Go slice to JSON []byte
	data, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
    	return fmt.Errorf("Error marshaling tasks: %w", err)
	}

	// Write to JSON file
	if err := os.WriteFile("tasks.json", data, 0644); err != nil {
		return fmt.Errorf("Error writing to tasks.json: %w", err)
	}

	return nil
}

func findIndexByTaskId(tasks []Task, id int) int {
	idx := -1
	for i := range tasks {
		if tasks[i].Id == id {
			idx = i
			break
		} 
	}
	return idx
}

func isValidStatus(s string) bool {
	switch s {
	case "todo", "in-progress", "done":
		return true
	default:
		return false
	}
}

func filterByStatus(filter string, tasks []Task) []Task {
	if filter == "" {
		return tasks
	}

	new_tasks := []Task{}
	
	for _, task := range tasks {
		if task.Status == filter {
			new_tasks = append(new_tasks, task)
		}
	}

	return new_tasks
}