package kamaji_test

import (
	"github.com/smaragden/kamaji/kamaji"
	//"math/rand"
	"fmt"
	//"sync"
	"testing"
)

func TestJobState(t *testing.T) {
	js := kamaji.UNKNOWN
	if js != 0 {
		t.Errorf("Status %q != %q", js, 0)
	}
	t.Logf("Status: %d, %s", kamaji.UNKNOWN, kamaji.UNKNOWN)
	t.Logf("Status: %d, %s", kamaji.CREATING, kamaji.RUNNING)
	t.Logf("Status: %d, %s", kamaji.IDLE, kamaji.IDLE)
	t.Logf("Status: %d, %s", kamaji.RUNNING, kamaji.RUNNING)
	t.Logf("Status: %d, %s", kamaji.STOPPING, kamaji.STOPPING)
	t.Logf("Status: %d, %s", kamaji.STOPPED, kamaji.STOPPED)
	t.Logf("Status: %d, %s", kamaji.PAUSED, kamaji.PAUSED)
	t.Logf("Status: %d, %s", kamaji.DONE, kamaji.DONE)
	t.Logf("Status: %d, %s", kamaji.ERROR, kamaji.ERROR)
	t.Logf("Status: %d, %s", kamaji.ARCHIVING, kamaji.ARCHIVING)
}

func TestJobCreation(t *testing.T) {
	count := 10
	for i := 0; i < count; i++ {
		job := kamaji.NewJob(fmt.Sprintf("Job %d", i))
		t.Logf("Job: %q, %q, %s, %d, %s", job.Name, job.ID, job.Status, job.Status, job.GetCreated())
	}
}

func TestJobTaskCommandCreation(t *testing.T) {
	job_count := 1
	task_count := 1
	command_count := 1
	var jobs []*kamaji.Job
	for i := 0; i < job_count; i++ {
		job := kamaji.NewJob(fmt.Sprintf("Job %d", i))
		for j := 0; j < task_count; j++ {
			task := kamaji.NewTask(fmt.Sprintf("Task %d", j), job)
			for k := 0; k < command_count; k++ {
				_ = kamaji.NewCommand(fmt.Sprintf("Command %d", k), task)
			}
		}
		jobs = append(jobs, job)
	}
	for _, job := range jobs {
		err := job.FSM.Event("start")
		if err != nil {
			t.Log(err)
		}
	}
	for _, job := range jobs {
		err := job.FSM.Event("stop")
		if err != nil {
			t.Log(err)
		}
	}
}
