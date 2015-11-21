package kamaji

import (
    log "github.com/Sirupsen/logrus"
    "time"
    "sync"
    //"fmt"
)

var contextLogger *log.Entry

// Dispatcher is the orchestrator of kamaji. The dispatcher is using the other managers to combine resources to
// create a task to assign to a Node.
type Dispatcher struct {
    lm        *LicenseManager
    nm        *NodeManager
    tm        *TaskManager
    running   bool
    done      chan bool
    waitGroup *sync.WaitGroup
}

func NewDispatcher(lm *LicenseManager, nm *NodeManager, tm *TaskManager) *Dispatcher {
    log.Debug("Create Dispatcher")
    d := new(Dispatcher)
    d.lm = lm
    d.nm = nm
    d.tm = tm
    d.running = false
    d.done = make(chan bool)
    d.waitGroup = &sync.WaitGroup{}
    return d
}

func (d *Dispatcher) Start() {
    d.waitGroup.Add(1)
    defer d.waitGroup.Done()
    log.WithFields(log.Fields{
        "action":  "start",
    }).Info("Starting Dispatcher.")
    DispatchNode:
    for {
        log.Debug("Waiting for node")
        select {
        case <-d.done:
            return
        case node := <-d.nm.NextNode:
            {
                if node == nil {
                    time.Sleep(time.Millisecond * 1000)
                    continue
                }
                node.ChangeState("assign")
                log.Debug("Waiting for command")
                DispatchCommand:
                for {
                    select {
                    case <-d.done:
                    node.ChangeState("ready")
                        return
                    case command := <-d.tm.NextCommand:
                        // Get license
                        licenses, err := d.lm.matchRequirements(command.Task.LicenseRequirements)
                        if err != nil{
                            continue DispatchCommand
                        }
                        command.Licenses = licenses
                        command.ChangeState("assign")
                        command.ChangeState("start")
                        err = node.assignCommand(command)
                        if err != nil {
                            log.WithFields(log.Fields{
                                "command": command.Name,
                                "node":  node.Name,
                                "action":  "assign",
                            }).Error(err)
                        }
                        log.WithFields(log.Fields{
                            "job":     command.Task.Job.Name,
                            "task":    command.Task.Name,
                            "command": command.Name,
                            "node":  node.Name,
                        }).Debug("Task Dispatched")
                        d.tm.ResetProvider()
                        continue DispatchNode
                    }
                }
            }
        }
    }
}

func (d *Dispatcher) Stop() {
    log.WithFields(log.Fields{
        "module":  "taskmanager",
        "action":  "stop",
    }).Info("Stopping Dispatcher")
    close(d.done)
}
