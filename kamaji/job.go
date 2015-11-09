package kamaji

import (
    "code.google.com/p/go-uuid/uuid"
    log "github.com/Sirupsen/logrus"
    "github.com/looplab/fsm"
    "sync"
    "time"
)

func init() {
    level, err := log.ParseLevel(Config.Logging.Task)
    if err == nil {
        log.SetLevel(level)
    }
}

// Job is the structure that holds tasks.
type Job struct {
    sync.RWMutex
    ID          uuid.UUID
    Name        string
    State       State
    Children    []*Task
    created     time.Time
    FSM         *fsm.FSM
    priority    int
    index       int
    ChangeState chan string
}

// NewJob create a new Job struct, generates a uuid for it and returns the job.
func NewJob(name string) *Job {
    j := new(Job)
    j.ID = uuid.NewRandom()
    j.Name = name
    j.State = UNKNOWN
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
    j.ChangeState = make(chan string)
    go j.stateChanger()
    return j
}


func (j *Job) stateChanger() {
    for {
        state := <-j.ChangeState
        //if j.FSM.Cannot(state) {
        //	continue
        //}
        err := j.FSM.Event(state)
        if err != nil {
            log.WithFields(log.Fields{"module": "nodemanager", "fuction": "stateChanger", "job": j.ID}).Error(err)
        }
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
    for _, task := range j.Children {
        if task.State > new_state {
            new_state = task.State
        }
    }
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

func (j *Job) GetCreated() time.Time {
    return j.created
}

func (j *Job) GetTasks() []*Task {
    j.Lock()
    defer j.Unlock()
    return append([]*Task(nil), j.Children...)
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