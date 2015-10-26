package kamaji_test

import (
	"github.com/smaragden/kamaji/kamaji"
	"testing"
)

func TestTaskStates(t *testing.T) {
	job := kamaji.NewJob("Test Job 01")
	task := kamaji.NewTask("Test Task 01", job)
	t.Logf("Task: %s is %s", task.Name, task.Status)
	err := task.FSM.Event("start")
	if err != nil {
		t.Log(err)
	}
	t.Logf("Task: %s is %s", task.Name, task.Status)
	err = task.FSM.Event("start")
	if err != nil {
		t.Log(err)
	}
	t.Logf("Task: %s is %s", task.Name, task.Status)
	err = task.FSM.Event("stop")
	if err != nil {
		t.Log(err)
	}
	t.Logf("Task: %s is %s", task.Name, task.Status)
}
