package kamaji

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"time"
)

func init() {
	log.SetLevel(log.DebugLevel)
}

type Dispatcher struct {
	lm      *LicenseManager
	cm      *ClientManager
	tm      *TaskManager
	running bool
	logger  *log.Entry
}

func NewDispatcher(lm *LicenseManager, cm *ClientManager, tm *TaskManager) *Dispatcher {
	d := new(Dispatcher)
	d.lm = lm
	d.cm = cm
	d.tm = tm
	d.running = false
	d.logger = log.WithField("module", "Dispatcher")
	return d
}

func (d *Dispatcher) Start() {
	for {
		d.logger.Debug("Dispatcher")
		client := d.cm.getAvailableClient() // Create a blocking channel
		if client == nil {
			time.Sleep(time.Second)
			continue
		}
		command := <-d.tm.NextCommand
		fmt.Printf("%s|%s|%s, %s, %s\n", command.Task.Job.Name, command.Task.Name, command.Name, command.Status, client.ID)
		client.FSM.Event("work")
		command.FSM.Event("start")
	}
}
