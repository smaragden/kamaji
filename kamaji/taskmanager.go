package kamaji

import (
	//"code.google.com/p/go-uuid/uuid"
	log "github.com/Sirupsen/logrus"
	"sync"
	"time"
)

var CommandEvent chan *KamajiMessage
var AllCommands chan []*Command

func init() {
	log.SetLevel(Config.LOG_LEVEL_TASKMANAGER)
	CommandEvent = make(chan *KamajiMessage, 5)
	AllCommands = make(chan []*Command)
}

type TaskManager struct {
	sync.RWMutex
	Jobs        []*Job
	NextCommand chan *Command
}

func NewTaskManager() *TaskManager {
	log.Debug("Creating Taskmanager")
	tm := new(TaskManager)
	tm.NextCommand = make(chan *Command)
	go tm.commandProvider()
	go tm.taskEventReciever()
	go tm.taskAllProvider()
	return tm
}

func (tm *TaskManager) taskAllProvider() {
	for {
		tm.Lock()
		commands := []*Command{}
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
		command.FSM.Event("assign")
		tm.NextCommand <- command
	}
}

func (tm *TaskManager) AddJob(job *Job) {
	tm.Lock()
	defer tm.Unlock()
	tm.Jobs = append(tm.Jobs, job)
	job.FSM.Event("ready")
}
