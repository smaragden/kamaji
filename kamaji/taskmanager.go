package kamaji

import (
//"code.google.com/p/go-uuid/uuid"
    log "github.com/Sirupsen/logrus"
    "sync"
    "time"
    "github.com/smaragden/kamaji/kamaji/proto"
    "fmt"
)

var CommandEvent chan *proto_msg.KamajiMessage

func init() {
    level, err := log.ParseLevel(Config.Logging.Taskmanager)
    if err == nil {
        log.SetLevel(level)
    }
    CommandEvent = make(chan *proto_msg.KamajiMessage, 5)
}

type JobList []*Job

type TaskManager struct {
    sync.RWMutex
    Jobs        JobList
    NextCommand chan *Command
}

func NewTaskManager() *TaskManager {
    log.Debug("Creating Taskmanager")
    tm := new(TaskManager)
    tm.NextCommand = make(chan *Command)
    return tm
}

func (tm *TaskManager) Start() {
    log.WithFields(log.Fields{
        "module":  "taskmanager",
        "action":  "start",
    }).Info("Starting Task Manager.")

    go tm.commandProvider()
    go tm.taskEventReciever()
}

func (tm *TaskManager) Stop() {
    log.WithFields(log.Fields{
        "module":  "taskmanager",
        "action":  "stop",
    }).Info("Stopping Task Manager.")
}

func (tm *TaskManager) taskEventReciever() {
    for {
        message := <-CommandEvent
        command := tm.getCommandsFromId(message.GetId())
        if command != nil {
            err := command.FSM.Event("finish")
            if err != nil {
                log.Fatal(err)
            }
        }
    }

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
    readyCommands := []*Command{}
    OrderedBy(prio, created).Sort(tm.Jobs)
    for _, job := range tm.Jobs {
        if job.State != STOPPED {
            for _, task := range job.GetTasks() {
                if task.State != STOPPED {
                    for _, command := range task.getCommands() {
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
    for {
        readyCommands := tm.getReadyCommands()
        // Sort the commands
        if len(readyCommands) == 0 {
            time.Sleep(time.Millisecond * 100)
            continue
        }
        for _, command := range readyCommands {
            fmt.Println("Assigning Next Command: ", command.Task.Name, ", ", command.Name)
            err := command.FSM.Event("assign")
            if err != nil {
                log.WithField("module", "taskmanager").Error(err)
            }
            tm.NextCommand <- command
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
