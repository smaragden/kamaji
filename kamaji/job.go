package kamaji

import (
	"code.google.com/p/go-uuid/uuid"
	"container/list"
	"fmt"
	"github.com/looplab/fsm"
	"time"
)

// Job is the structure that holds tasks.
type Job struct {
	ID       uuid.UUID
	Name     string
	Status   Status
	Children *list.List
	created  time.Time
	FSM      *fsm.FSM
}

// NewJob create a new Job struct, generates a uuid for it and returns the job.
func NewJob(name string) *Job {
	j := new(Job)
	j.ID = uuid.NewRandom()
	j.Name = name
	j.Status = UNKNOWN
	j.Children = list.New()
	j.created = time.Now()
	j.FSM = fsm.NewFSM(
		j.Status.String(),
		fsm.Events{
			{Name: "start", Src: []string{UNKNOWN.String(), IDLE.String(), STOPPED.String()}, Dst: RUNNING.String()},
			{Name: "stop", Src: []string{RUNNING.String()}, Dst: STOPPED.String()},
		},
		fsm.Callbacks{
			"enter_state":    func(e *fsm.Event) { j.enterState(e) },
			RUNNING.String(): func(e *fsm.Event) { j.startJob(e) },
			STOPPED.String(): func(e *fsm.Event) { j.stopJob(e) },
		},
	)
	return j
}

func (j *Job) enterState(e *fsm.Event) {
	j.Status = StatusFromString(e.Dst)
	fmt.Printf("The job:%s, %s -> %s\n", j.Name, e.Src, e.Dst)
}

func (j *Job) startJob(e *fsm.Event) {
	fmt.Printf("Starting Job: %s\n", j.Name)
	for it := j.Children.Front(); it != nil; it = it.Next() {
		task := it.Value.(*Task)
		task.FSM.Event("start")
	}
}

func (j *Job) stopJob(e *fsm.Event) {
	fmt.Printf("Stopping Job: %s\n", j.Name)
	for it := j.Children.Front(); it != nil; it = it.Next() {
		task := it.Value.(*Task)
		task.FSM.Event("stop")
	}
}

func (j Job) ChangeStatus(status Status) bool {
	j.Status = status
	for it := j.Children.Front(); it != nil; it = it.Next() {
		task := it.Value.(*Task)
		return task.ParentStatusChanged(status)
	}
	return false
}

func (j Job) GetCreated() time.Time {
	return j.created
}

func (j Job) Store() bool {
	db := NewDatabase()
	if _, err := db.Client.Do("HMSET", fmt.Sprintf("job:%s", j.ID),
		"Name", j.Name,
		"Status", j.Status.String(),
		"created", j.created.String()); err != nil {
		panic(err)
	}
	if _, err := db.Client.Do("LPUSH", "jobs", j.ID); err != nil {
		panic(err)
	}
	return true
}
