package kamaji

type State int

// "Enums" for job states
const (
	// General
	UNKNOWN State = 0

	// Tasks Intermidates
	CREATING  State = 20
	ASSIGNING State = 21
	STOPPING  State = 22
	DELETING  State = 23
	// Tasks
	PAUSED  State = 25
	STOPPED State = 26
	BLOCKED State = 27
	DONE    State = 28

	// Node Intermediates
	DISCONNECTING State = 30
	// Node
	OFFLINE State = 35
	ONLINE  State = 36
	SERVICE State = 37

	READY   State = 80
	WORKING State = 90
	ERROR   State = 100
)

var StateToStrings = map[State]string{
	UNKNOWN:       "UNKNOWN",
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
	ERROR:         "ERROR",
}

var stringToState = map[string]State{
	"UNKNOWN":       UNKNOWN,
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
	"ERROR":         ERROR,
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
