package kamaji

import (
	log "github.com/Sirupsen/logrus"
)

type Configuration struct {
	// Logging
	LOG_LEVEL_TASK          log.Level
	LOG_LEVEL_DISPATCHER    log.Level
	LOG_LEVEL_CLIENTMANAGER log.Level
	LOG_LEVEL_TASKMANAGER   log.Level
}

var Config = Configuration{
	LOG_LEVEL_TASK:          log.InfoLevel,
	LOG_LEVEL_DISPATCHER:    log.InfoLevel,
	LOG_LEVEL_CLIENTMANAGER: log.InfoLevel,
	LOG_LEVEL_TASKMANAGER:   log.InfoLevel,
}
