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
	job.Children = append(job.Children, t)
	t.FSM = fsm.NewFSM(
		t.State.S(),
		fsm.Events{
			{Name: "ready", Src: []string{UNKNOWN.S(), STOPPED.S()}, Dst: READY.S()},
			{Name: "work", Src: []string{UNKNOWN.S(), READY.S(), STOPPED.S()}, Dst: WORKING.S()},
			{Name: "stop", Src: []string{WORKING.S()}, Dst: STOPPED.S()},
		},
		fsm.Callbacks{
			"enter_state": func(e *fsm.Event) { t.enterState(e) },
			READY.S():     func(e *fsm.Event) { t.readyTask(e) },
			WORKING.S():   func(e *fsm.Event) { t.workTask(e) },
			STOPPED.S():   func(e *fsm.Event) { t.stopTask(e) },
		},
	)
	return t
}

func (t *Task) enterState(e *fsm.Event) {
	t.State = StateFromString(e.Dst)
	log.WithFields(log.Fields{
		"module": "task",
		"task":   t.Name,
		"from":   e.Src,
		"to":     e.Dst,
	}).Debug("Changing Task State")
}

func (t *Task) readyTask(e *fsm.Event) {
	//fmt.Printf("Ready Task: %s\n", t.Name)
	for _, command := range t.Commands {
		command.FSM.Event("ready")
	}
}

func (t *Task) workTask(e *fsm.Event) {
	//fmt.Printf("Starting Task: %s\n", t.Name)
	for _, command := range t.Commands {
		command.FSM.Event("start")
	}
}

func (t *Task) stopTask(e *fsm.Event) {
	//fmt.Printf("Stopping Task: %s\n", t.Name)
	for _, command := range t.Commands {
		command.FSM.Event("stop")
	}
}

func (t *Task) getCommands() []*Command {
	t.Lock()
	defer t.Unlock()
	return append([]*Command(nil), t.Commands...)
}
