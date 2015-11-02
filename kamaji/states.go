package kamaji

type State int

// "Enums" for job states
const (
	// General
	UNKNOWN State = 0
	ERROR   State = 10
	READY   State = 11
	WORKING State = 12

	// Tasks
	PAUSED  State = 20
	STOPPED State = 21
	BLOCKED State = 22
	DONE    State = 23
	// Tasks Intermidates
	CREATING  State = 25
	ASSIGNING State = 26
	STOPPING  State = 27
	DELETING  State = 28

	// Client
	OFFLINE State = 30
	ONLINE  State = 31
	SERVICE State = 32
	// Client Intermediates
	DISCONNECTING State = 35
)

var StateToStrings = map[State]string{
	UNKNOWN:       "UNKNOWN",
	ERROR:         "ERROR",
	READY:         "READY",
	WORKING:       "WORKING",
	PAUSED:        "PAUSED",
	STOPPED:       "STOPPED",
	BLOCKED:       "BLOCKED",
	DONE:          "DONE",
	CREATING:      "CREATING",
	ASSIGNING:     "ASSIGNING",
	STOPPING:      "STOPPING",
	DELETING:      "DELETING",
	OFFLINE:       "OFFLINE",
	ONLINE:        "ONLINE",
	SERVICE:       "SERVICE",
	DISCONNECTING: "DISCONNECTING",
}

var stringToState = map[string]State{
	"UNKNOWN":       UNKNOWN,
	"ERROR":         ERROR,
	"READY":         READY,
	"WORKING":       WORKING,
	"PAUSED":        PAUSED,
	"STOPPED":       STOPPED,
	"BLOCKED":       BLOCKED,
	"DONE":          DONE,
	"CREATING":      CREATING,
	"ASSIGNING":     ASSIGNING,
	"STOPPING":      STOPPING,
	"DELETING":      DELETING,
	"OFFLINE":       OFFLINE,
	"ONLINE":        ONLINE,
	"SERVICE":       SERVICE,
	"DISCONNECTING": DISCONNECTING,
}

func (js State) String() string {
	return StateToStrings[js]
}

func (js State) S() string {
	return StateToStrings[js]
}

func StateFromString(State string) State {
	return stringToState[State]
}

/*

// General
UNKNOWN  =   0
ERROR     =  10
READY      = 11
WORKING     =12

// Job, Task, Command
PAUSED      20
STOPPED     21
BLOCKED     22
// Intermidates
CREATING    25
ASSIGNING   26
STOPPING    27
DELETING    28

// Client
OFFLINE     30
ONLINE      31
// Intermediates
DISCONNECTING 35
*/
