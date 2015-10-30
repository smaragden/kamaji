package kamaji

import (
	"code.google.com/p/go-uuid/uuid"
	log "github.com/Sirupsen/logrus"
	"github.com/looplab/fsm"
)

func init() {
	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})
}

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
	task.Commands = append(task.Commands, c)
	c.FSM = fsm.NewFSM(
		c.Status.String(),
		fsm.Events{
			{Name: "ready", Src: []string{UNKNOWN.String(), STOPPED.String()}, Dst: READY.String()},
			{Name: "assign", Src: []string{READY.String()}, Dst: ASSIGNING.String()},
			{Name: "start", Src: []string{UNKNOWN.String(), READY.String(), STOPPED.String()}, Dst: RUNNING.String()},
			{Name: "stop", Src: []string{RUNNING.String()}, Dst: STOPPED.String()},
		},
		fsm.Callbacks{
			"enter_state":      func(e *fsm.Event) { c.enterState(e) },
			READY.String():     func(e *fsm.Event) { c.readyCommand(e) },
			ASSIGNING.String(): func(e *fsm.Event) { c.assignCommand(e) },
			RUNNING.String():   func(e *fsm.Event) { c.startCommand(e) },
			STOPPED.String():   func(e *fsm.Event) { c.stopCommand(e) },
		},
	)
	return c
}

func (c *Command) enterState(e *fsm.Event) {
	c.Status = StatusFromString(e.Dst)
	log.WithFields(log.Fields{
		"module":  "command",
		"command": c.Name,
		"from":    e.Src,
		"to":      e.Dst,
	}).Info("Changing Command State")
}

func (c *Command) readyCommand(e *fsm.Event) {

}

func (c *Command) assignCommand(e *fsm.Event) {

}

func (c *Command) startCommand(e *fsm.Event) {

}

func (c *Command) stopCommand(e *fsm.Event) {

}
