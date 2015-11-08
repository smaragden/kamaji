package kamaji

import (
	//"code.google.com/p/go-uuid/uuid"
	log "github.com/Sirupsen/logrus"
	"sort"
	"sync"
	"time"
)

var CommandEvent chan *KamajiMessage
var AllJobs chan []*Job
var AllTasks chan []*Task
var AllCommands chan []*Command

func init() {
	log.SetLevel(Config.LOG_LEVEL_TASKMANAGER)
	CommandEvent = make(chan *KamajiMessage, 5)
	AllJobs = make(chan []*Job)
	AllTasks = make(chan []*Task)
	AllCommands = make(chan []*Command)
}

type JobList []*Job
type JobListPrio []*Job
type JobListCreated []*Job

type TaskManager struct {
	sync.RWMutex
	Jobs        JobList
	NextCommand chan *Command
}

func prio(c1, c2 *Job) bool {
	return c1.priority < c2.priority
}

func created(c1, c2 *Job) bool {
	return c1.created.UnixNano() < c2.created.UnixNano()
}

type lessFunc func(p1 *Job, p2 *Job) bool

// multiSorter implements the Sort interface, sorting the changes within.
type multiSorter struct {
	jobs JobList
	less []lessFunc
}

// Sort sorts the argument slice according to the less functions passed to OrderedBy.
func (ms *multiSorter) Sort(jobs JobList) {
	ms.jobs = jobs
	sort.Sort(ms)
}

// OrderedBy returns a Sorter that sorts using the less functions, in order.
// Call its Sort method to sort the data.
func OrderedBy(less ...lessFunc) *multiSorter {
	return &multiSorter{
		less: less,
	}
}

// Len is part of sort.Interface.
func (ms *multiSorter) Len() int {
	return len(ms.jobs)
}

// Swap is part of sort.Interface.
func (ms *multiSorter) Swap(i, j int) {
	ms.jobs[i], ms.jobs[j] = ms.jobs[j], ms.jobs[i]
}

// Less is part of sort.Interface. It is implemented by looping along the
// less functions until it finds a comparison that is either Less or
// !Less. Note that it can call the less functions twice per call. We
// could change the functions to return -1, 0, 1 and reduce the
// number of calls for greater efficiency: an exercise for the reader.
func (ms *multiSorter) Less(i, j int) bool {
	p, q := ms.jobs[i], ms.jobs[j]
	// Try all but the last comparison.
	var k int
	for k = 0; k < len(ms.less)-1; k++ {
		less := ms.less[k]
		switch {
		case less(p, q):
			// p < q, so we have a decision.
			return true
		case less(q, p):
			// p > q, so we have a decision.
			return false
		}
		// p == q; try the next comparison.
	}
	// All comparisons to here said "equal", so just return whatever
	// the final comparison reports.
	return ms.less[k](p, q)
}

func NewTaskManager() *TaskManager {
	log.Debug("Creating Taskmanager")
	tm := new(TaskManager)
	tm.NextCommand = make(chan *Command)
	go tm.commandProvider()
	go tm.taskEventReciever()
	go tm.jobAllProvider()
	go tm.taskAllProvider()
	go tm.commandAllProvider()
	return tm
}

func (tm *TaskManager) jobAllProvider() {
	for {
		tm.RLock()
		jobs := []*Job{}
		for _, job := range tm.Jobs {
			jobs = append(jobs, job)
		}
		tm.RUnlock()
		select {
		case AllJobs <- jobs:
		case <-time.After(time.Second * 3):
		}
	}
}

func (tm *TaskManager) taskAllProvider() {
	for {
		tm.RLock()
		tasks := []*Task{}
		for _, job := range tm.Jobs {
			for _, task := range job.getTasks() {
				tasks = append(tasks, task)
			}
		}
		tm.RUnlock()
		AllTasks <- tasks
	}
}

func (tm *TaskManager) commandAllProvider() {
	for {
		tm.Lock()
		commands := []*Command{}
		OrderedBy(created).Sort(tm.Jobs)
		for _, job := range tm.Jobs {
			for _, task := range job.getTasks() {
				for _, command := range task.getCommands() {
					commands = append(commands, command)
				}
			}
		}
		tm.Unlock()
		AllCommands <- commands
	}
}

func (tm *TaskManager) taskEventReciever() {
	for {
		message := <-CommandEvent
		command := tm.getCommandsFromId(message.GetId())
		if command != nil {
			err := command.FSM.Event("finish")
			if err != nil {
				log.Fatal(err)
			}
		}
	}

}

func (tm *TaskManager) taskSorter() {

}

func (tm *TaskManager) getCommandsFromId(id string) *Command {
	for _, job := range tm.Jobs {
		if job.State != STOPPED {
			for _, task := range job.getTasks() {
				if task.State != STOPPED {
					for _, command := range task.getCommands() {
						if command.ID.String() == id {
							return command
						}
					}
				}
			}
		}
	}
	return nil
}

func (tm *TaskManager) getReadyCommands() []*Command {
	tm.Lock()
	defer tm.Unlock()
	readyCommands := []*Command{}
	OrderedBy(prio, created).Sort(tm.Jobs)
	for _, job := range tm.Jobs {
		if job.State != STOPPED {
			for _, task := range job.getTasks() {
				if task.State != STOPPED {
					for _, command := range task.getCommands() {
						if command.State == READY {
							readyCommands = append(readyCommands, command)
						}
					}
				}
			}
		}
	}
	return readyCommands
}

func (tm *TaskManager) commandProvider() {
	for {
		readyCommands := tm.getReadyCommands()
		// Sort the commands
		if len(readyCommands) == 0 {
			time.Sleep(time.Millisecond * 100)
			continue
		}
		command := readyCommands[0]
		err := command.FSM.Event("assign")
		if err != nil {
			log.WithField("module", "taskmanager").Error(err)
		}
		tm.NextCommand <- command
	}
}

func (tm *TaskManager) AddJob(job *Job) {
	tm.Lock()
	defer tm.Unlock()
	tm.Jobs = append(tm.Jobs, job)
	err := job.FSM.Event("ready")
	if err != nil {
		log.WithField("module", "taskmanager").Error(err)
	}
}
