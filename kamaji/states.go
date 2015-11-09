package kamaji

type State int

// "Enums" for states
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

// String representation of a State
func (js State) String() string {
    return StateToStrings[js]
}

// Alias for String()
func (js State) S() string {
    return StateToStrings[js]
}

// Return a State from a string
func StateFromString(State string) State {
    return stringToState[State]
}

// Takes arbitrary number of state and returns a string list of state names
func StateList(states ...State) []string {
    var stateList []string
    for _, state := range states {
        stateList = append(stateList, state.String())
    }
    return stateList
}
