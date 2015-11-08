package kamaji

import (
	"code.google.com/p/go-uuid/uuid"
	log "github.com/Sirupsen/logrus"
	"github.com/looplab/fsm"
	"sync"
)

type Task struct {
	sync.RWMutex
	ID       uuid.UUID
	Name     string
	State    State
	Job      *Job
	Commands []*Command
	FSM      *fsm.FSM
}

// NewTask create a new Task struct, generates a uuid for it and returns the task.
func NewTask(name string, job *Job) *Task {
	t := new(Task)
	t.ID = uuid.NewRandom()
	t.Name = name
	t.State = UNKNOWN
	t.Job = job
	t.Commands = []*Command{}
	if job != nil {
		job.Children = append(job.Children, t)
	}
	t.FSM = fsm.NewFSM(
		t.State.S(),
		fsm.Events{
			{Name: "ready", Src: []string{UNKNOWN.S(), STOPPED.S()}, Dst: READY.S()},
			{Name: "work", Src: []string{UNKNOWN.S(), READY.S(), STOPPED.S()}, Dst: WORKING.S()},
			{Name: "stop", Src: []string{WORKING.S()}, Dst: STOPPED.S()},
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

func (t *Task) calculateState() {
	new_state := UNKNOWN
	old_state := t.State
	for _, command := range t.Commands {
		if command.State > new_state {
			new_state = command.State
		}
	}
	if new_state != old_state {
		t.State = new_state
		log.WithFields(log.Fields{
			"module":     "task",
			"task":       t.Name,
			"old_status": old_state,
			"new_status": new_state,
		}).Debug("Calculated new task state")
		t.Job.calculateState()
	}
}
