package main

import (
    "fmt"
    "github.com/smaragden/kamaji/kamaji"
    "os"
    "os/signal"
//"time"
//"sync"
)

func main() {
    fmt.Println("Starting")
    // Create all managers
    lm := kamaji.NewLicenseManager()
    lm.AddApplication("maya", 4)
    lm.AddApplication("nuke", 4)
    nm := kamaji.NewNodeManager("", 1314)
    tm := kamaji.NewTaskManager()
    d := kamaji.NewDispatcher(lm, nm, tm)
    s := kamaji.NewService("", 8080, lm, nm, tm)
    // Create signal handler
    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt)
    go func() {
        for sig := range c {
            switch sig{
            case os.Interrupt:
                fmt.Printf("\nCatched Interupt Signal, Cleaning up\n")
                d.Stop()
            }
        }
    }()

    // Create Test Jobs
    job_count := 10
    task_count := 2
    command_count := 10
    for i := 1; i < job_count + 1; i++ {
        job := kamaji.NewJob(fmt.Sprintf("Job %d", i))
        for j := 0; j < task_count; j++ {
            lic_name := "maya"
            if i > 6{
                lic_name = "nuke"
            }
            task := kamaji.NewTask(fmt.Sprintf("Task %d [%s]", j, lic_name), job, []string{lic_name})
            for k := 0; k < command_count; k++ {
                _ = kamaji.NewCommand(fmt.Sprintf("Command %d", k), task)
            }
        }
        tm.AddJob(job)
    }
    go lm.Start()
    defer lm.Stop()
    go nm.Start()
    defer nm.Stop()
    go tm.Start()
    defer tm.Stop()
    go s.Start()
    defer s.Stop()
    d.Start()
}
