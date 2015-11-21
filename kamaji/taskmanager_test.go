package kamaji_test

import (
    "github.com/smaragden/kamaji/kamaji"
    "testing"
	"fmt"
	"math/rand"
	"sort"
)

func testEq(a, b []string) bool {
    if a == nil && b == nil {
        return true;
    }
    if a == nil || b == nil {
        return false;
    }
    if len(a) != len(b) {
        return false
    }
    for i := range a {
        if a[i] != b[i] {
            return false
        }
    }
    return true
}

func fillTaskManager(tm *kamaji.TaskManager, job_count int, task_count int, command_count int) {
	// Create Test Jobs
	for i := 1; i < job_count + 1; i++ {
		lic_name := "maya"
		if i > 6 {
			lic_name = "nuke"
		}
		job := kamaji.NewJob(fmt.Sprintf("Job %d | %s", i, lic_name))
		for j := 0; j < task_count; j++ {
			task := kamaji.NewTask(fmt.Sprintf("Task %d [%s]", j, lic_name), job, []string{lic_name})
			for k := 0; k < command_count; k++ {
				_ = kamaji.NewCommand(fmt.Sprintf("Command %d", k), task)
			}
		}
		tm.AddJob(job)
	}
}

func TestTaskManager(t *testing.T) {
	job_count := 1000
	task_count := 2
	command_count := 100
	tm := kamaji.NewTaskManager()
	fillTaskManager(tm, job_count, task_count, command_count)
	expected := job_count
	got := tm.GetNumJobs()
	if  got != expected{
		t.Errorf("Expected taskmanager to contain %d jobs but it reported %d jobs.", expected, got)
	}
	expected = task_count*job_count
	got = tm.GetNumTasks()
	if  got != expected{
		t.Errorf("Expected taskmanager to contain %d tasks but it reported %d tasks.", expected, got)
	}
	expected = command_count*task_count*job_count
	got = tm.GetNumCommands()
	if  got != expected{
		t.Errorf("Expected taskmanager to contain %d commands but it reported %d commands.", expected, got)
	}
}

func TestTaskManagerProviderOrder(t *testing.T) {
	tm := kamaji.NewTaskManager()
	job := kamaji.NewJob("JOB")
	task := kamaji.NewTask("TASK", job, []string{})
	_ = kamaji.NewCommand("0", task)
	_ = kamaji.NewCommand("1", task)
	_ = kamaji.NewCommand("2", task)
	_ = kamaji.NewCommand("3", task)
	_ = kamaji.NewCommand("4", task)
	_ = kamaji.NewCommand("5", task)
	tm.AddJob(job)
	tm.Start()
	defer tm.Stop()
	expected := []string{"0", "1", "2", "3", "4", "5"}
	var got []string
	for _ = range task.Commands{
		command := <-tm.NextCommand
		got = append(got, command.Name)
	}
	if !testEq(expected, got){
		t.Errorf("Expected %+v got %+v.", expected, got)
	}
	got = got[:0]
	tm.ResetProvider()
	for _ = range task.Commands{
		command := <-tm.NextCommand
		got = append(got, command.Name)
	}
	if !testEq(expected, got){
		t.Errorf("After reset. Expected %+v got %+v.", expected, got)
	}
}


func TestTaskManagerProviderPriority(t *testing.T) {
	tm := kamaji.NewTaskManager()
	job0 := kamaji.NewJob("JOB0")
	job0.SetPrio(0)
	job1 := kamaji.NewJob("JOB0")
	job1.SetPrio(1)
	task0 := kamaji.NewTask("TASK0", job0, []string{})
	_ = kamaji.NewCommand("0", task0)
	tm.AddJob(job0)
	task1 := kamaji.NewTask("TASK1", job1, []string{})
	_ = kamaji.NewCommand("1", task1)
	tm.AddJob(job1)
	tm.Start()
	defer tm.Stop()
	expected := []string{"1", "0"}
	var got []string
	for _ = range tm.Jobs{
		command := <-tm.NextCommand
		got = append(got, command.Name)
	}
	if !testEq(expected, got){
		t.Errorf("Expected %+v got %+v.", expected, got)
	}
	got = got[:0]
	tm.ResetProvider()
	for _ = range tm.Jobs{
		command := <-tm.NextCommand
		got = append(got, command.Name)
	}
	if !testEq(expected, got){
		t.Errorf("After reset. Expected %+v got %+v.", expected, got)
	}
}

func TestTaskManagerSortInterface(t *testing.T) {
	job_count := 10
	var jobs kamaji.Jobs
	for i:=0;i<job_count;i++{
		prio := rand.Int31n(100)
		job := kamaji.NewJob(fmt.Sprintf("Job: %d with prio: %d", i, prio))
		job.SetPrio(int(prio))
		jobs = append(jobs, job)
		t.Logf(job.Name)
	}
	t.Logf("Sort")
	sort.Sort(jobs)
	for _,job:=range jobs{
		t.Logf(job.Name)
	}
	t.Logf("Setting all prios to zero.")
	for _,job:=range jobs{
		job.SetPrio(0)
	}
	t.Logf("Sort")
	sort.Sort(jobs)
	for _,job:=range jobs{
		t.Logf(job.Name)
	}
}

func TestTaskManagerGetById(t *testing.T) {
	tm := kamaji.NewTaskManager()
	job := kamaji.NewJob("JOB")
	task := kamaji.NewTask("TASK", job, []string{})
	command := kamaji.NewCommand("1", task)
	tm.AddJob(job)
	q_job := tm.GetJobFromId(job.ID.String())
	if q_job.ID.String() != job.ID.String(){
		t.Logf("Expected %s, got %s", job.ID, q_job.ID)
	}

	q_cmd := tm.GetCommandsFromId(command.ID.String())
	if q_cmd.ID.String() != command.ID.String(){
		t.Logf("Expected %s, got %s", command.ID, q_cmd.ID)
	}
}