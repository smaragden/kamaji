package kamaji_test

import (
	"github.com/smaragden/kamaji/kamaji"
	"testing"
)

func TestCommandStates(t *testing.T) {
	job := kamaji.NewJob("Test Job 01")
	task := kamaji.NewTask("Test Task 01", job)
	command := kamaji.NewCommand("Test Command 01", task)
	t.Logf("Command: %s is %s", command.Name, command.Status)
	err := command.FSM.Event("start")
	if err != nil {
		t.Log(err)
	}
	t.Logf("Command: %s is %s", command.Name, command.Status)
	err = command.FSM.Event("start")
	if err != nil {
		t.Log(err)
	}
	t.Logf("Command: %s is %s", command.Name, command.Status)
	err = command.FSM.Event("stop")
	if err != nil {
		t.Log(err)
	}
	t.Logf("Command: %s is %s", command.Name, command.Status)
}
