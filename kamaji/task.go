package kamaji

import (
	"code.google.com/p/go-uuid/uuid"
	"container/list"
	"fmt"
	"github.com/looplab/fsm"
)

type Task struct {
	ID       uuid.UUID
	Name     string
	Status   Status
	Job      *Job
	Commands *list.List
	FSM      *fsm.FSM
}

// NewTask create a new Task struct, generates a uuid for it and returns the task.
func NewTask(name string, job *Job) *Task {
	t := new(Task)
	t.ID = uuid.NewRandom()
	t.Name = name
	t.Status = UNKNOWN
	t.Job = job
	t.Commands = list.New()
	job.Children.PushBack(t)
	t.FSM = fsm.NewFSM(
		t.Status.String(),
		fsm.Events{
			{Name: "start", Src: []string{UNKNOWN.String(), IDLE.String(), STOPPED.String()}, Dst: RUNNING.String()},
			{Name: "stop", Src: []string{RUNNING.String()}, Dst: STOPPED.String()},
		},
		fsm.Callbacks{
			"enter_state":    func(e *fsm.Event) { t.enterState(e) },
			RUNNING.String(): func(e *fsm.Event) { t.startTask(e) },
			STOPPED.String(): func(e *fsm.Event) { t.stopTask(e) },
		},
	)
	return t
}

func (t *Task) enterState(e *fsm.Event) {
	t.Status = StatusFromString(e.Dst)
	fmt.Printf("The task:%s, %s -> %s\n", t.Name, e.Src, e.Dst)
}

func (t *Task) startTask(e *fsm.Event) {
	fmt.Printf("Starting Task: %s\n", t.Name)
	for it := t.Commands.Front(); it != nil; it = it.Next() {
		command := it.Value.(*Command)
		command.FSM.Event("start")
	}
}

func (t *Task) stopTask(e *fsm.Event) {
	fmt.Printf("Stopping Task: %s\n", t.Name)
	for it := t.Commands.Front(); it != nil; it = it.Next() {
		command := it.Value.(*Command)
		command.FSM.Event("stop")
	}
}

func (t *Task) ParentStatusChanged(status Status) bool {
	return true
}
