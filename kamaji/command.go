package kamaji

import (
    "github.com/pborman/uuid"
    log "github.com/Sirupsen/logrus"
    "github.com/looplab/fsm"
    "sync"
    "time"
)

type Commands []*Command
// Command represents the command that is going to be executed on the remote Node.
type Command struct {
    sync.RWMutex
    ID    uuid.UUID
    Name  string
    State State
    Completion float32
    created  time.Time
    priority int
    Task  *Task
    FSM   *fsm.FSM
    Licenses []*License
}

// Create a new Command instance and return it.
func NewCommand(name string, task *Task) *Command {
    c := new(Command)
    c.ID = uuid.NewRandom()
    c.Name = name
    c.State = UNKNOWN
    c.Completion = 0.0
    c.created = time.Now()
    c.priority = 0
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

// Synchronous state changer. This method should almost always be called when you want to change state.
func (c *Command) ChangeState(state string) {
    c.Lock()
    defer c.Unlock()
    err := c.FSM.Event(state)
    if err != nil {
        log.WithFields(log.Fields{"module": "command", "fuction": "stateChanger", "node": c.Name}).Fatal(err)
    }
}

func (c *Command) finishCommand(e *fsm.Event) {
    log.Info("Command Finish.")
    // Return licenses
    for _, lic := range c.Licenses{
        lic.Return()
    }
    c.Licenses = c.Licenses[:0]
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


func (c *Command) SetPrio(prio int) {
    c.priority = prio
}

func (c *Command) GetPrio() int {
    return c.priority
}

func (c *Command) GetCreated() time.Time {
    return c.created
}

// Sort Interface
func (slice Commands) Len() int {
    return len(slice)
}

func (slice Commands) Less(i, j int) bool {
    if slice[i].priority == slice[j].priority{
        return slice[i].created.UnixNano() < slice[j].created.UnixNano();
    }
    return slice[i].priority > slice[j].priority;
}

func (slice Commands) Swap(i, j int) {
    slice[i], slice[j] = slice[j], slice[i]
}
