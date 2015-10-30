package kamaji

import (
	"code.google.com/p/go-uuid/uuid"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/looplab/fsm"
	"sync"
)

type Task struct {
	sync.RWMutex
	ID       uuid.UUID
	Name     string
	Status   Status
	Job      *Job
	Commands []*Command
	FSM      *fsm.FSM
}

// NewTask create a new Task struct, generates a uuid for it and returns the task.
func NewTask(name string, job *Job) *Task {
	t := new(Task)
	t.ID = uuid.NewRandom()
	t.Name = name
	t.Status = UNKNOWN
	t.Job = job
	t.Commands = []*Command{}
	job.Children = append(job.Children, t)
	t.FSM = fsm.NewFSM(
		t.Status.String(),
		fsm.Events{
			{Name: "ready", Src: []string{UNKNOWN.String(), STOPPED.String()}, Dst: READY.String()},
			{Name: "start", Src: []string{UNKNOWN.String(), READY.String(), STOPPED.String()}, Dst: RUNNING.String()},
			{Name: "stop", Src: []string{RUNNING.String()}, Dst: STOPPED.String()},
		},
		fsm.Callbacks{
			"enter_state":    func(e *fsm.Event) { t.enterState(e) },
			READY.String():   func(e *fsm.Event) { t.readyTask(e) },
			RUNNING.String(): func(e *fsm.Event) { t.startTask(e) },
			STOPPED.String(): func(e *fsm.Event) { t.stopTask(e) },
		},
	)
	return t
}

func (t *Task) enterState(e *fsm.Event) {
	t.Status = StatusFromString(e.Dst)
	log.WithFields(log.Fields{
		"module": "task",
		"task":   t.Name,
		"from":   e.Src,
		"to":     e.Dst,
	}).Info("Changing Task State")
}

func (t *Task) readyTask(e *fsm.Event) {
	fmt.Printf("Ready Task: %s\n", t.Name)
	for _, command := range t.Commands {
		command.FSM.Event("ready")
	}
}

func (t *Task) startTask(e *fsm.Event) {
	fmt.Printf("Starting Task: %s\n", t.Name)
	for _, command := range t.Commands {
		command.FSM.Event("start")
	}
}

func (t *Task) stopTask(e *fsm.Event) {
	fmt.Printf("Stopping Task: %s\n", t.Name)
	for _, command := range t.Commands {
		command.FSM.Event("stop")
	}
}

func (t *Task) getCommands() []*Command {
	t.Lock()
	defer t.Unlock()
	return append([]*Command(nil), t.Commands...)
}

func (t *Task) ParentStatusChanged(status Status) bool {
	return true
}
