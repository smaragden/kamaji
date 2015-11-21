package kamaji

import (
//"code.google.com/p/go-uuid/uuid"
    log "github.com/Sirupsen/logrus"
    "sync"
    "sort"
)

type TaskManager struct {
    sync.RWMutex
    Jobs        Jobs
    NextCommand chan *Command
    reset chan bool
}

func NewTaskManager() *TaskManager {
    log.Debug("Creating Taskmanager")
    tm := new(TaskManager)
    tm.NextCommand = make(chan *Command)
    tm.reset = make(chan bool)
    return tm
}

func (tm *TaskManager) Start() {
    log.WithFields(log.Fields{
        "module":  "taskmanager",
        "action":  "start",
    }).Info("Starting Task Manager.")

    go tm.commandProvider()
}

func (tm *TaskManager) Stop() {
    log.WithFields(log.Fields{
        "module":  "taskmanager",
        "action":  "stop",
    }).Info("Stopping Task Manager")
}

func (tm *TaskManager) ResetProvider() {
    tm.reset <- true
}

func (tm *TaskManager) GetNumJobs() int {
    return len(tm.Jobs)
}

func (tm *TaskManager) GetNumTasks() int {
    num_tasks := 0
    for _, job := range tm.Jobs {
        num_tasks += len(job.Children)
    }
    return num_tasks
}

func (tm *TaskManager) GetNumCommands() int {
    num_tasks := 0
    for _, job := range tm.Jobs {
        for _, task := range job.GetTasks() {
            num_tasks += len(task.Commands)
        }
    }
    return num_tasks
}

func (tm *TaskManager) taskSorter() {

}

func (tm *TaskManager) GetJobFromId(id string) *Job {
    for _, job := range tm.Jobs {
        if job.ID.String() == id {
            return job
        }
    }
    return nil
}

func (tm *TaskManager) getCommandsFromId(id string) *Command {
    for _, job := range tm.Jobs {
        if job.State != STOPPED {
            for _, task := range job.GetTasks() {
                if task.State != STOPPED {
                    for _, command := range task.getCommands() {
                        if command.ID.String() == id {
                            return command
                        }
                    }
                }
            }
        }
    }
    return nil
}

func (tm *TaskManager) getReadyCommands() []*Command {
    tm.Lock()
    defer tm.Unlock()
    var readyCommands Commands
    sort.Sort(tm.Jobs)
    for _, job := range tm.Jobs {
        if job.State != STOPPED {
            tasks := job.GetTasks()
            sort.Sort(tasks)
            for _, task := range  tasks{
                if task.State != STOPPED {
                    commands := task.getCommands()
                    sort.Sort(commands)
                    for _, command := range commands{
                        if command.State == READY {
                            readyCommands = append(readyCommands, command)
                        }
                    }
                }
            }
        }
    }
    return readyCommands
}

func (tm *TaskManager) commandProvider() {
    Provider:
    for {
        readyCommands := tm.getReadyCommands()
        for _, command := range readyCommands {
            select {
            case tm.NextCommand <- command:
            case <-tm.reset:
                continue Provider
            }
        }
    }
}

func (tm *TaskManager) AddJob(job *Job) {
    tm.Lock()
    defer tm.Unlock()
    tm.Jobs = append(tm.Jobs, job)
    err := job.FSM.Event("ready")
    if err != nil {
        log.WithField("module", "taskmanager").Error(err)
    }
}
