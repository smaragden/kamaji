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
	go HttpServe(d.nm)
	for {
		node := d.nm.getAvailableNode() // Create a blocking channel
		if node == nil {
			time.Sleep(time.Millisecond * 5)
			continue
		}
		command := <-d.tm.NextCommand
		log.WithFields(log.Fields{
			"module":  "dispatcher",
			"job":     command.Task.Job.Name,
			"task":    command.Task.Name,
			"command": command.Name,
			"client":  node.ID,
		}).Debug("Dispatch Task")
		node.FSM.Event("work")
		e := command.FSM.Event("start")
		if e != nil {
			log.Fatal(e)
		}
		//fmt.Printf("%s|%s|%s, %s, %s\n", command.Task.Job.Name, command.Task.Name, command.Name, command.Status, client.ID)
		err := node.assignCommand(command)
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
