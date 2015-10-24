package kamaji_test

import (
	"github.com/smaragden/kamaji/kamaji"
	//"math/rand"
	"fmt"
	//"sync"
	"testing"
	//"time"
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
		t.Logf("Job: %q, %q, %s, %d", job.Name, job.ID, job.Status, job.Status)
	}
}
