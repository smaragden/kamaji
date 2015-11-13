package kamaji

import (
    "code.google.com/p/go-uuid/uuid"
    log "github.com/Sirupsen/logrus"
    "github.com/looplab/fsm"
)

func init() {
    log.SetFormatter(&log.TextFormatter{FullTimestamp: true})
}


// Command represents the command that is going to be executed on the remote Node.
type Command struct {
    ID    uuid.UUID
    Name  string
    State State
    Completion float32
    Task  *Task
    FSM   *fsm.FSM
    Licenses []string
}

// Create a new Command instance and return it.
func NewCommand(name string, task *Task) *Command {
    c := new(Command)
    c.ID = uuid.NewRandom()
    c.Name = name
    c.State = UNKNOWN
    c.Completion = 0.0
    c.Task = task
    if task != nil {
        task.Commands = append(task.Commands, c)
    }
    c.FSM = fsm.NewFSM(
        c.State.S(),
        fsm.Events{
            {Name: "ready", Src: StateList(UNKNOWN, STOPPED, ASSIGNING), Dst: READY.S()},
            {Name: "assign", Src: StateList(READY), Dst: ASSIGNING.S()},
            {Name: "start", Src: StateList(UNKNOWN, READY, ASSIGNING, STOPPED), Dst: WORKING.S()},
            {Name: "restart", Src: StateList(DONE), Dst: WORKING.S()},
            {Name: "finish", Src: StateList(WORKING), Dst: DONE.S()},
            {Name: "stop", Src: StateList(WORKING), Dst: STOPPED.S()},
        },
        fsm.Callbacks{
            "after_event": func(e *fsm.Event) { c.afterEvent(e) },
            DONE.S():   func(e *fsm.Event) { c.finishCommand(e) },
        },
    )
    return c
}

func (c *Command) finishCommand(e *fsm.Event) {
    // Return licenses
    LicenseReturner <- c.Licenses
    c.Licenses = []string{}
    c.Completion = 1.0
}


// Set the state of the Command after a successful state transition.
// If the command have a Task, tell the task to recalculate it's State
func (c *Command) afterEvent(e *fsm.Event) {
    c.State = StateFromString(e.Dst)
    log.WithFields(log.Fields{
        "module":  "command",
        "command": c.Name,
        "from":    e.Src,
        "to":      e.Dst,
    }).Debug("Changing Command State")
    if c.Task != nil {
        c.Task.calculateState()
    }
}

