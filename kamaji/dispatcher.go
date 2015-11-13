package kamaji

import (
    log "github.com/Sirupsen/logrus"
    "time"
    "sync"
    "fmt"
)

func init() {
    level, err := log.ParseLevel(Config.Logging.Dispatcher)
    if err == nil {
        log.SetLevel(level)
    }
}

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
        "module":  "dispatcher",
        "action":  "start",
    }).Info("Starting Dispatcher.")
    Dispatch:
        for {
            log.WithField("module", "dispatcher").Debug("Waiting for node")
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
                    log.WithField("module", "dispatcher").Debug("Waiting for command")
                    select {
                    case <-d.done:
                        return
                    case command := <-d.tm.NextCommand:
                        log.WithFields(log.Fields{
                            "module":  "dispatcher",
                            "job":     command.Task.Job.Name,
                            "task":    command.Task.Name,
                            "command": command.Name,
                            "node":  node.Name,
                        }).Debug("Dispatch Task")
                        // Get license
                        fmt.Println("License Requirements: ", command.Task.LicenseRequirements )
                        for _, lic := range command.Task.LicenseRequirements {
                            lic_count, err := d.lm.Borrow(lic)
                            if err != nil {
                                // Return already acquired licenses
                                LicenseReturner <- command.Licenses
                                log.WithFields(log.Fields{
                                    "module":  "dispatcher",
                                    "command": command.Name,
                                    "node":  node.Name,
                                    "action":  "license query",
                                    "lic count": lic_count,
                                }).Warning(err)
                                err := command.FSM.Event("ready")
                                if err != nil {
                                    log.WithField("module", "dispatcher").Error(err)
                                }
                                node.ChangeState("ready")
                                continue Dispatch
                            }else{
                                command.Licenses = append(command.Licenses, lic)
                            }
                        }
                        fmt.Println("Set Licenses: ", command.Licenses)
                        err := command.FSM.Event("start")
                        if err != nil {
                            log.Fatal(err)
                        }
                        err = node.assignCommand(command)
                        if err != nil {
                            log.WithFields(log.Fields{
                                "module":  "dispatcher",
                                "command": command.Name,
                                "node":  node.Name,
                                "action":  "assign",
                            }).Error(err)
                        }
                    }
                }
            }
        }
}

func (d *Dispatcher) Stop() {
    log.WithFields(log.Fields{
        "module":  "dispatcher",
        "action":  "stop",
    }).Info("Stopping Dispatcher.")
    close(d.done)
}
