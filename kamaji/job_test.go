package kamaji_test

import (
	"fmt"
	"github.com/smaragden/kamaji/kamaji"
	"sync"
	"testing"
	"time"
)

func changeState(job *kamaji.Job, state string, t *testing.T, wg *sync.WaitGroup) {
	job.ChangeState <- state
	wg.Done()
}
func TestJobState(t *testing.T) {
	job_count := 10
	task_count := 2
	command_count := 10
	var jobs []*kamaji.Job
	for i := 1; i < job_count+1; i++ {
		job := kamaji.NewJob(fmt.Sprintf("Job %d", i))
		for j := 0; j < task_count; j++ {
			task := kamaji.NewTask(fmt.Sprintf("Task %d", j), job)
			for k := 0; k < command_count; k++ {
				_ = kamaji.NewCommand(fmt.Sprintf("Command %d", k), task)
			}
		}
		jobs = append(jobs, job)
	}
	stateSequence := []string{"ready", "start", "finish"}
	var wg sync.WaitGroup
	for _, state := range stateSequence {
		for _, job := range jobs {
			wg.Add(1)
			go changeState(job, state, t, &wg)
		}
		time.Sleep(time.Millisecond * 10)
	}
	wg.Wait()
	time.Sleep(time.Millisecond * 1000)
}
