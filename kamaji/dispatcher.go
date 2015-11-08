package kamaji

import (
	log "github.com/Sirupsen/logrus"
	"time"
)

func init() {
	log.SetLevel(Config.LOG_LEVEL_DISPATCHER)
}

type Dispatcher struct {
	lm      *LicenseManager
	nm      *NodeManager
	tm      *TaskManager
	running bool
	logger  *log.Entry
}

func NewDispatcher(lm *LicenseManager, nm *NodeManager, tm *TaskManager) *Dispatcher {
	log.Debug("Create Dispatcher")
	d := new(Dispatcher)
	d.lm = lm
	d.nm = nm
	d.tm = tm
	d.running = false
	d.logger = log.WithField("module", "Dispatcher")
	return d
}

func (d *Dispatcher) Start() {
	for {
		log.WithField("module", "dispatcher").Debug("Waiting for node")
		node := <-d.nm.NextNode
		if node == nil {
			time.Sleep(time.Millisecond * 1000)
			continue
		}
		log.WithField("module", "dispatcher").Debug("Waiting for command")
		command := <-d.tm.NextCommand
		log.WithFields(log.Fields{
			"module":  "dispatcher",
			"job":     command.Task.Job.Name,
			"task":    command.Task.Name,
			"command": command.Name,
			"client":  node.ID,
		}).Debug("Dispatch Task")
		//node.changeState <- "work"
		err := command.FSM.Event("start")
		if err != nil {
			log.Fatal(err)
		}
		//fmt.Printf("%s|%s|%s, %s, %s\n", command.Task.Job.Name, command.Task.Name, command.Name, command.Status, client.ID)
		err = node.assignCommand(command)
		if err != nil {
			log.WithFields(log.Fields{
				"module":  "dispatcher",
				"command": command.Name,
				"client":  node.ID,
				"action":  "assign",
			}).Error(err)
		}
	}
}
