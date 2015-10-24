package main

import (
	"fmt"
	"github.com/smaragden/kamaji/kamaji"
)

func main() {
	db := kamaji.NewDatabase("Dataaaabaaas", "localhost:6379")
	fmt.Println("Database Client: ", db.Client)
	err := db.Client.Set("fkey", 15.0, 0).Err()
	if err != nil {
		panic(err)
	}
	job := kamaji.Job{Name: "Jobbet"}
	task := kamaji.Task{Name: "Tasken"}
	command := kamaji.Command{Name: "Commandet"}
	license := kamaji.License{Name: "Licensen"}
	fmt.Println("Database: ", db.Name)
	fmt.Println("Job: ", job.Name)
	fmt.Println("Task: ", task.Name)
	fmt.Println("Command: ", command.Name)
	fmt.Println("License: ", license.Name)
}
