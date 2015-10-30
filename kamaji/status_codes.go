package kamaji

type Status int

// "Enums" for job states
const (
	UNKNOWN Status = iota
	CREATING
	IDLE
	RUNNING
	STOPPING
	STOPPED
	PAUSED
	DONE
	ERROR
	ARCHIVING
	ONLINE
	OFFLINE
	AVAILABLE
	WORKING
	DISCONNECTED
	READY
	ASSIGNING
)

var statusStrings = [...]string{
	"UNKNOWN",
	"CREATING",
	"IDLE",
	"RUNNING",
	"STOPPING",
	"STOPPED",
	"PAUSED",
	"DONE",
	"ERROR",
	"ARCHIVING",
	"ONLINE",
	"OFFLINE",
	"AVAILABLE",
	"WORKING",
	"DISCONNECTED",
	"READY",
	"ASSIGNING",
}

var stringToStatus = map[string]Status{
	"UNKNOWN":      UNKNOWN,
	"CREATING":     CREATING,
	"IDLE":         IDLE,
	"RUNNING":      RUNNING,
	"STOPPING":     STOPPING,
	"STOPPED":      STOPPED,
	"PAUSED":       PAUSED,
	"DONE":         DONE,
	"ERROR":        ERROR,
	"ARCHIVING":    ARCHIVING,
	"ONLINE":       ONLINE,
	"OFFLINE":      OFFLINE,
	"AVAILABLE":    AVAILABLE,
	"WORKING":      WORKING,
	"DISCONNECTED": DISCONNECTED,
	"READY":        READY,
	"ASSIGNING":    ASSIGNING,
}

func (js Status) String() string {
	return statusStrings[js]
}

func StatusFromString(status string) Status {
	return stringToStatus[status]
}
