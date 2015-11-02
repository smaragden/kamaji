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
	cm      *ClientManager
	tm      *TaskManager
	running bool
	logger  *log.Entry
}

func NewDispatcher(lm *LicenseManager, cm *ClientManager, tm *TaskManager) *Dispatcher {
	log.Debug("Create Dispatcher")
	d := new(Dispatcher)
	d.lm = lm
	d.cm = cm
	d.tm = tm
	d.running = false
	d.logger = log.WithField("module", "Dispatcher")
	return d
}

func (d *Dispatcher) Start() {
	go HttpServe(d.cm)
	for {
		client := d.cm.getAvailableClient() // Create a blocking channel
		if client == nil {
			time.Sleep(time.Millisecond * 5)
			continue
		}
		command := <-d.tm.NextCommand
		log.WithFields(log.Fields{
			"module":  "dispatcher",
			"job":     command.Task.Job.Name,
			"task":    command.Task.Name,
			"command": command.Name,
			"client":  client.ID,
		}).Debug("Dispatch Task")
		client.FSM.Event("work")
		e := command.FSM.Event("start")
		if e != nil {
			log.Fatal(e)
		}
		//fmt.Printf("%s|%s|%s, %s, %s\n", command.Task.Job.Name, command.Task.Name, command.Name, command.Status, client.ID)
		err := client.assignCommand(command)
		if err != nil {
			log.WithFields(log.Fields{
				"module":  "dispatcher",
				"command": command.Name,
				"client":  client.ID,
				"action":  "assign",
			}).Error(err)
		}
	}
}
