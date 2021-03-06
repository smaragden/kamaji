package kamaji

import (
    "github.com/pborman/uuid"
    log "github.com/Sirupsen/logrus"
    "github.com/looplab/fsm"
    "sync"
    "time"
)

type Jobs []*Job
// Job is the structure that holds tasks.
type Job struct {
    sync.RWMutex
    ID          uuid.UUID
    Name        string
    State       State
    Completion float32
    Children    Tasks
    created     time.Time
    FSM         *fsm.FSM
    priority    int
    index       int
}

// NewJob create a new Job struct, generates a uuid for it and returns the job.
func NewJob(name string) *Job {
    j := new(Job)
    j.ID = uuid.NewRandom()
    j.Name = name
    j.State = UNKNOWN
    j.Completion = 0.0
    j.Children = []*Task{}
    j.created = time.Now()
    j.priority = 0
    j.FSM = fsm.NewFSM(
        j.State.String(),
        fsm.Events{
            {Name: "ready", Src: StateList(UNKNOWN, STOPPED), Dst: READY.S()},
            {Name: "start", Src: StateList(READY), Dst: WORKING.S()},
            {Name: "finish", Src: StateList(WORKING), Dst: DONE.S()},
            {Name: "restart", Src: StateList(DONE), Dst: WORKING.S()},
            {Name: "stop", Src: StateList(WORKING), Dst: STOPPED.S()},
        },
        fsm.Callbacks{
            "after_event": func(e *fsm.Event) { j.afterEvent(e) },
        },
    )
    return j
}

// Synchronous state changer. This method should almost always be called when you want to change state.
func (j *Job) ChangeState(state string) {
    j.Lock()
    defer j.Unlock()
    err := j.FSM.Event(state)
    if err != nil {
        log.WithFields(log.Fields{"module": "job", "fuction": "ChangeState", "node": j.Name}).Fatal(err)
    }
}


func (j *Job) afterEvent(e *fsm.Event) {
    j.State = StateFromString(e.Dst)
    for _, task := range j.Children {
        task.FSM.Event(e.Event)
    }
}

func (j *Job) calculateState() {
    new_state := UNKNOWN
    old_state := j.State
    var completion float32
    for _, task := range j.Children {
        completion+=task.Completion
        if task.State > new_state {
            new_state = task.State
        }
    }
    j.Completion = completion/float32(len(j.Children))
    if new_state != old_state {
        j.State = new_state
        log.WithFields(log.Fields{
            "module":     "job",
            "job":        j.Name,
            "old_status": old_state,
            "new_status": new_state,
        }).Debug("Calculated new job state")
    }
}

func (j *Job) GetPrio() int {
    return j.priority
}

func (j *Job) SetPrio(prio int) {
    j.priority = prio
}

func (j *Job) GetCreated() time.Time {
    return j.created
}

func (j *Job) GetTasks() Tasks {
    j.Lock()
    defer j.Unlock()
    return j.Children
}

func (j *Job) GetTaskFromId(id string) *Task {
    j.Lock()
    defer j.Unlock()
    for _, task := range j.Children {
        if task.ID.String() == id {
            return task
        }
    }
    return nil
}

// Sort Interface
func (slice Jobs) Len() int {
    return len(slice)
}

func (slice Jobs) Less(i, j int) bool {
    if slice[i].priority==slice[j].priority{
        return slice[i].created.UnixNano() < slice[j].created.UnixNano();
    }
    return slice[i].priority > slice[j].priority;
}

func (slice Jobs) Swap(i, j int) {
    slice[i], slice[j] = slice[j], slice[i]
}