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
}

var stringToStatus = map[string]Status{
	"UNKNOWN":   UNKNOWN,
	"CREATING":  CREATING,
	"IDLE":      IDLE,
	"RUNNING":   RUNNING,
	"STOPPING":  STOPPING,
	"STOPPED":   STOPPED,
	"PAUSED":    PAUSED,
	"DONE":      DONE,
	"ERROR":     ERROR,
	"ARCHIVING": ARCHIVING,
}

func (js Status) String() string {
	return statusStrings[js]
}

func StatusFromString(status string) Status {
	return stringToStatus[status]
}
