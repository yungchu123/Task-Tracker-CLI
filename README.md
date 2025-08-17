# Task-Tracker-CLI
Sample solution for the [task-tracker](https://roadmap.sh/projects/task-tracker) challenge from [roadmap.sh](https://roadmap.sh/).

## How to run
Clone the repository first.
Run the following command to build and run the project:
```
go build -o task-cli

# To see the list of all available commands
./task-cli --help

# To add a task or tasks
./task-cli add "Buy groceries"
./task-cli add "Sweep the floor" "Homework"

# To update a task
./task-cli update 1 "Buy groceries and shopping"

# To mark a task as in progress / done
./task-cli mark-in-progress 1
./task-cli mark-done 1

# To list tasks
./task-cli list
./task-cli list todo
./task-cli list in-progress
./task-cli list done
```
