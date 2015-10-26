package kamaji

import (
	"code.google.com/p/go-uuid/uuid"
	"fmt"
	"github.com/looplab/fsm"
)

type Command struct {
	ID     uuid.UUID
	Name   string
	Status Status
	Task   *Task
	FSM    *fsm.FSM
}

func NewCommand(name string, task *Task) *Command {
	c := new(Command)
	c.ID = uuid.NewRandom()
	c.Name = name
	c.Task = task
	c.Status = UNKNOWN
	task.Commands.PushBack(c)
	c.FSM = fsm.NewFSM(
		c.Status.String(),
		fsm.Events{
			{Name: "start", Src: []string{UNKNOWN.String(), IDLE.String(), STOPPED.String()}, Dst: RUNNING.String()},
			{Name: "stop", Src: []string{RUNNING.String()}, Dst: STOPPED.String()},
		},
		fsm.Callbacks{
			"enter_state":    func(e *fsm.Event) { c.enterState(e) },
			RUNNING.String(): func(e *fsm.Event) { c.startTask(e) },
			STOPPED.String(): func(e *fsm.Event) { c.stopTask(e) },
		},
	)
	return c
}

func (c *Command) enterState(e *fsm.Event) {
	c.Status = StatusFromString(e.Dst)
	fmt.Printf("The command:%s, %s -> %s\n", c.Name, e.Src, e.Dst)
}

func (c *Command) startTask(e *fsm.Event) {
	fmt.Printf("Starting Command: %s\n", c.Name)
}

func (c *Command) stopTask(e *fsm.Event) {
	fmt.Printf("Stopping Command: %s\n", c.Name)
}
