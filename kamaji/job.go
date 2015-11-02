package kamaji

import (
	"code.google.com/p/go-uuid/uuid"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/looplab/fsm"
	"sync"
	"time"
)

func init() {
	log.SetLevel(log.DebugLevel)
}

// Job is the structure that holds tasks.
type Job struct {
	sync.RWMutex
	ID       uuid.UUID
	Name     string
	State    State
	Children []*Task
	created  time.Time
	FSM      *fsm.FSM
}

// NewJob create a new Job struct, generates a uuid for it and returns the job.
func NewJob(name string) *Job {
	j := new(Job)
	j.ID = uuid.NewRandom()
	j.Name = name
	j.State = UNKNOWN
	j.Children = []*Task{}
	j.created = time.Now()
	j.FSM = fsm.NewFSM(
		j.State.String(),
		fsm.Events{
			{Name: "ready", Src: []string{UNKNOWN.S(), STOPPED.S()}, Dst: READY.S()},
			{Name: "work", Src: []string{UNKNOWN.S(), READY.S(), STOPPED.S()}, Dst: WORKING.S()},
			{Name: "stop", Src: []string{WORKING.S()}, Dst: STOPPED.S()},
		},
		fsm.Callbacks{
			"enter_state":    func(e *fsm.Event) { j.enterState(e) },
			READY.String():   func(e *fsm.Event) { j.readyJob(e) },
			WORKING.String(): func(e *fsm.Event) { j.workJob(e) },
			STOPPED.String(): func(e *fsm.Event) { j.stopJob(e) },
		},
	)
	return j
}

func (j *Job) enterState(e *fsm.Event) {
	j.State = StateFromString(e.Dst)
	log.WithFields(log.Fields{
		"module": "job",
		"job":    j.Name,
		"from":   e.Src,
		"to":     e.Dst,
	}).Debug("Changing Job State")
}

func (j *Job) readyJob(e *fsm.Event) {
	//fmt.Printf("Ready Job: %s\n", j.Name)
	for _, task := range j.Children {
		task.FSM.Event("ready")
	}
}

func (j *Job) workJob(e *fsm.Event) {
	//fmt.Printf("Starting Job: %s\n", j.Name)
	for _, task := range j.Children {
		task.FSM.Event("work")
	}
}

func (j *Job) stopJob(e *fsm.Event) {
	//fmt.Printf("Stopping Job: %s\n", j.Name)
	for _, task := range j.Children {
		task.FSM.Event("stop")
	}
}

func (j *Job) GetCreated() time.Time {
	return j.created
}

func (j *Job) getTasks() []*Task {
	j.Lock()
	defer j.Unlock()
	return append([]*Task(nil), j.Children...)
}

func (j Job) Store() bool {
	db := NewDatabase()
	if _, err := db.Client.Do("HMSET", fmt.Sprintf("job:%s", j.ID),
		"Name", j.Name,
		"Status", j.State.S(),
		"created", j.created.String()); err != nil {
		panic(err)
	}
	if _, err := db.Client.Do("LPUSH", "jobs", j.ID); err != nil {
		panic(err)
	}
	return true
}
