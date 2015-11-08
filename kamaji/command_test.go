package kamaji_test

import (
	"github.com/smaragden/kamaji/kamaji"
	"testing"
)

func TestCommandStates(t *testing.T) {
	command := kamaji.NewCommand("Test Command 01", nil)
	stateSequence := []string{"ready", "assign", "start", "finish", "restart", "stop"}
	for _, state := range stateSequence {
		err := command.FSM.Event(state)
		if err != nil {
			t.Error(err)
		}
	}
}
