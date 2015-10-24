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

var statuses = [...]string{
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

func (js Status) String() string {
	return statuses[js]
}
