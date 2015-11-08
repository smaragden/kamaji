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
	ID    uuid.UUID
	Name  string
	State State
	Task  *Task
	FSM   *fsm.FSM
}

func NewCommand(name string, task *Task) *Command {
	c := new(Command)
	c.ID = uuid.NewRandom()
	c.Name = name
	c.Task = task
	c.State = UNKNOWN
	if task != nil {
		task.Commands = append(task.Commands, c)
	}
	c.FSM = fsm.NewFSM(
		c.State.S(),
		fsm.Events{
			{Name: "ready", Src: []string{UNKNOWN.S(), STOPPED.S()}, Dst: READY.S()},
			{Name: "assign", Src: []string{READY.S()}, Dst: ASSIGNING.S()},
			{Name: "start", Src: []string{UNKNOWN.S(), READY.S(), ASSIGNING.S(), STOPPED.S()}, Dst: WORKING.S()},
			{Name: "restart", Src: []string{DONE.S()}, Dst: WORKING.S()},
			{Name: "finish", Src: []string{WORKING.S()}, Dst: DONE.S()},
			{Name: "stop", Src: []string{WORKING.S()}, Dst: STOPPED.S()},
		},
		fsm.Callbacks{
			"after_event": func(e *fsm.Event) { c.afterEvent(e) },
		},
	)
	return c
}

func (c *Command) afterEvent(e *fsm.Event) {
	c.State = StateFromString(e.Dst)
	log.WithFields(log.Fields{
		"module":  "command",
		"command": c.Name,
		"from":    e.Src,
		"to":      e.Dst,
	}).Debug("Changing Command State")
	c.Task.calculateState()
}
