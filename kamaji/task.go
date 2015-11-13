package kamaji

import (
    "code.google.com/p/go-uuid/uuid"
    log "github.com/Sirupsen/logrus"
    "github.com/looplab/fsm"
    "sync"
    "time"
    "fmt"
)

type Task struct {
    sync.RWMutex
    ID       uuid.UUID
    Name     string
    State    State
    Completion float32
    Job      *Job
    Commands []*Command
    created  time.Time
    FSM      *fsm.FSM
    priority int
    LicenseRequirements []string
}

// NewTask create a new Task struct, generates a uuid for it and returns the task.
func NewTask(name string, job *Job, licenses []string) *Task {
    t := new(Task)
    t.ID = uuid.NewRandom()
    t.Name = name
    t.State = UNKNOWN
    t.Completion = 0.0
    t.Job = job
    t.Commands = []*Command{}
    t.created = time.Now()
    t.priority = 0
    if job != nil {
        job.Children = append(job.Children, t)
    }
    t.LicenseRequirements = licenses
    fmt.Println("Lic Req: ", licenses)
    t.FSM = fsm.NewFSM(
        t.State.S(),
        fsm.Events{
            {Name: "ready", Src: StateList(UNKNOWN, STOPPED), Dst: READY.S()},
            {Name: "work", Src: StateList(UNKNOWN, READY, STOPPED), Dst: WORKING.S()},
            {Name: "stop", Src: StateList(WORKING), Dst: STOPPED.S()},
        },
        fsm.Callbacks{
            "after_event": func(e *fsm.Event) { t.afterEvent(e) },
        },
    )
    return t
}

func (t *Task) afterEvent(e *fsm.Event) {
    t.State = StateFromString(e.Dst)
    log.WithFields(log.Fields{
        "module": "task",
        "task":   t.Name,
        "from":   e.Src,
        "to":     e.Dst,
    }).Debug("Changing Task State")
    for _, command := range t.Commands {
        command.FSM.Event(e.Event)
    }
    t.Job.calculateState()
}

func (t *Task) getCommands() []*Command {
    t.Lock()
    defer t.Unlock()
    return append([]*Command(nil), t.Commands...)
}

func (t *Task) GetCommandFromId(id string) *Command {
    t.Lock()
    defer t.Unlock()
    for _, command := range t.Commands {
        if command.ID.String() == id {
            return command
        }
    }
    return nil
}

func (t *Task) calculateState() {
    new_state := UNKNOWN
    old_state := t.State
    var completion float32
    for _, command := range t.Commands {
        completion+=command.Completion
        if command.State > new_state {
            new_state = command.State
        }
    }
    t.Completion = completion/float32(len(t.Commands))
    //if new_state != old_state {
        t.State = new_state
        log.WithFields(log.Fields{
            "module":     "task",
            "task":       t.Name,
            "old_status": old_state,
            "new_status": new_state,
        }).Debug("Calculated new task state")
        t.Job.calculateState()
    //}
}

func (t *Task) GetPrio() int {
    return t.priority
}

func (t *Task) GetCreated() time.Time {
    return t.created
}
