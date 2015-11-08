package main

import (
	"fmt"
	"github.com/smaragden/kamaji/kamaji"
)

func main() {
	fmt.Println("Starting")
	lm := kamaji.NewLicenseManager()
	nm := kamaji.NewNodeManager("ClientManager", "", 1314)
	go nm.Start()
	tm := kamaji.NewTaskManager()
	d := kamaji.NewDispatcher(lm, nm, tm)
	job_count := 10
	task_count := 2
	command_count := 100
	for i := 1; i < job_count+1; i++ {
		job := kamaji.NewJob(fmt.Sprintf("Job %d", i))
		for j := 0; j < task_count; j++ {
			task := kamaji.NewTask(fmt.Sprintf("Task %d", j), job)
			for k := 0; k < command_count; k++ {
				_ = kamaji.NewCommand(fmt.Sprintf("Command %d", k), task)
			}
		}
		tm.AddJob(job)
	}
	d.Start()
	fmt.Println("Exiting")
}
