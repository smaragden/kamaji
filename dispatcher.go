package main

import (
	"fmt"
	"github.com/smaragden/kamaji/kamaji"
	"github.com/docopt/docopt-go"
	"strconv"
)

func main() {
	usage := `Kamaji Dispatcher.
Usage:
	client_spawner [-a=0.0.0.0] [-p=1314]

Options:
	-a --address=N  	Bind to.  [default: 0.0.0.0]
	-p --port=N  		Port.     [default: 1314]
	-h --help           Show this screen.`
	arguments, err := docopt.Parse(usage, nil, true, "Kamaji Client Spawner", false)
    if err != nil {
        fmt.Println(err)
    }

	address, _ := arguments["--address"].(string)
	port, _ := strconv.Atoi(arguments["--port"].(string))
	fmt.Printf("Starting on %s:%d\n", address, port)
	lm := kamaji.NewLicenseManager()
	nm := kamaji.NewNodeManager(address, port)
	go nm.Start()
	tm := kamaji.NewTaskManager()
	d := kamaji.NewDispatcher(lm, nm, tm)
	job_count := 10
	task_count := 2
	command_count := 100
	for i := 1; i < job_count+1; i++ {
		job := kamaji.NewJob(fmt.Sprintf("Job %d", i))
		for j := 0; j < task_count; j++ {
			task := kamaji.NewTask(fmt.Sprintf("Task %d", j), job, []string{})
			for k := 0; k < command_count; k++ {
				_ = kamaji.NewCommand(fmt.Sprintf("Command %d", k), task)
			}
		}
		tm.AddJob(job)
	}
	d.Start()
	fmt.Println("Exiting")
}
