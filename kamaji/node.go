package kamaji

import (
    "github.com/pborman/uuid"
    "errors"
    "fmt"
    log "github.com/Sirupsen/logrus"
    "github.com/golang/protobuf/proto"
    "github.com/looplab/fsm"
    "net"
    "sync"
    "github.com/smaragden/kamaji/kamaji/proto"
    "time"
    "io"
    "syscall"
)

// Node represents a instance of a worker node. Most likely a separate machine.
type Node struct {
    sync.RWMutex
    *Client
    ID              uuid.UUID
    Name            string
    State           State
    fsm             *fsm.FSM
    currentCommands []*Command
    statusUpdate    chan State
    Send            chan *proto_msg.KamajiMessage
    Receive         chan *proto_msg.KamajiMessage
    done            chan bool
    waitGroup       *sync.WaitGroup
}

// Create and return a new Node. channels for sending and receiving messages are setup.
// And two goroutines are spawned Send and Receive to handle incoming and outgoing messages.
func NewNode(conn net.Conn) *Node {
    n := new(Node)
    n.Client = NewClient(conn)
    n.ID = uuid.NewRandom()
    n.Name = n.ID.String()
    n.State = UNKNOWN
    n.fsm = fsm.NewFSM(
        n.State.S(),
        fsm.Events{
            {Name: "offline", Src: StateList(ONLINE, READY, ASSIGNING, WORKING), Dst: OFFLINE.S()},
            {Name: "online", Src: StateList(UNKNOWN, READY, OFFLINE), Dst: ONLINE.S()},
            {Name: "ready", Src: StateList(ONLINE, WORKING, ASSIGNING), Dst: READY.S()},
            {Name: "assign", Src: StateList(READY), Dst: ASSIGNING.S()},
            {Name: "work", Src: StateList(ASSIGNING), Dst: WORKING.S()},
            {Name: "service", Src: StateList(UNKNOWN, ONLINE, OFFLINE), Dst: SERVICE.S()},
        },
        fsm.Callbacks{
            "after_event": func(e *fsm.Event) { n.afterEvent(e) },
            "before_offline": func(e *fsm.Event) { n.beforeOffline(e) },
            OFFLINE.S():   func(e *fsm.Event) { n.offlineNode(e) },
        },
    )

    n.Send = make(chan *proto_msg.KamajiMessage)
    n.Receive = make(chan *proto_msg.KamajiMessage)
    n.done = make(chan bool)
    n.waitGroup = &sync.WaitGroup{}
    if n.Conn != nil {
        go n.messageTransmitter()
        go n.messageReciever()
    }
    return n
}

// Synchronous state changer. This method should almost always be called when you want to change state.
func (n *Node) ChangeState(state string) {
    n.Lock()
    defer n.Unlock()
    err := n.fsm.Event(state)
    if err != nil {
        log.WithFields(log.Fields{"module": "nodemanager", "fuction": "stateChanger", "node": n.Name}).Fatal(err)
    }
}

func (n *Node) afterEvent(e *fsm.Event) {
    n.State = StateFromString(e.Dst)
    log.WithFields(log.Fields{
        "module": "node",
        "node":   n.Name,
        "from":   e.Src,
        "to":     e.Dst,
    }).Debug("Changing Node State")
}

// We close the message handlers (messageTransmitter, messageReciever) before entering the offline state.
func (n *Node) beforeOffline(e *fsm.Event) {
    close(n.done)
    n.waitGroup.Wait()
}

// Close the nodes connection.
func (n *Node) offlineNode(e *fsm.Event) {
    n.Conn.Close()
    // TODO: Handle stray tasks
}

// Assign a command to this node and report to the client.
// we add a reference to the command on this node to be able to track it later.
func (n *Node) assignCommand(command *Command) error {
    message := &proto_msg.KamajiMessage{
        Action: proto_msg.KamajiMessage_ASSIGN.Enum(),
        Entity: proto_msg.KamajiMessage_COMMAND.Enum(),
        Id:     proto.String(command.ID.String()),
    }
    n.currentCommands = append(n.currentCommands, command)
    log.WithFields(log.Fields{"module": "nodemanager", "client": n.Name, "job": command.Task.Job.Name, "task": command.Task.Name, "command": command.Name}).Info("Assigning command to client.")
    n.Send <- message
    return nil
}

// Remove the command that was added in assignCommand.
func (n *Node) removeCommand(command_id string) (*Command, error) {
    log.WithFields(log.Fields{"module": "nodemanager", "client": n.Name, "command": command_id}).Debug("Clean up command.")
    for i, node_command := range n.currentCommands {
        if node_command.ID.String() == command_id {
            n.currentCommands = append(n.currentCommands[:i], n.currentCommands[i + 1:]...)
            return node_command, nil
        }
    }
    return nil, errors.New("Couldn't find command on client.")
}

// Send a request to the client that we want to update our state. The client will respond with the status it agrees with.
func (n *Node) requestStatusUpdate(state State) error {
    log.WithFields(log.Fields{"module": "nodemanager", "client": n.Name, "state": n.State, "new_state": state}).Debug("State Update Request")
    message := &proto_msg.KamajiMessage{
        Action: proto_msg.KamajiMessage_STATUS_UPDATE.Enum(),
        Entity: proto_msg.KamajiMessage_NODE.Enum(),
        Id:     proto.String(n.ID.String()),
        Statusupdate: &proto_msg.KamajiMessage_StatusUpdate{
            Destination: proto.Int32(int32(state)),
        },
    }
    n.Send <- message
    return nil
}

// Send a message to the client.
func (n *Node) messageTransmitter() {
    n.waitGroup.Add(1)
    defer n.waitGroup.Done()
    for {
        select {
        case <-n.done:
            return
        case message := <-n.Send:
            message_data, err := proto.Marshal(message)
            if err != nil {
                fmt.Println(err)
                continue
            }
            n.SetDeadline(time.Now().Add(6e6)) //TODO: Find proper value, this affects the time to offline the node
            _, err = n.SendMessage(message_data)
            if err != nil {
                if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
                    continue
                }
                log.WithFields(log.Fields{"module": "nodemanager", "function": "messageSender"}).Error(err)
                return
            }
        case <-time.After(time.Millisecond * 10):
            continue
        }
    }
}

// Receive a message from the client
func (n *Node) messageReciever() {
    n.waitGroup.Add(1)
    defer n.waitGroup.Done()
    for {
        n.SetDeadline(time.Now().Add(6e8)) //TODO: Find proper value, this affects the time to offline the node
        tmp, err := n.ReadMessage()
        if err != nil {
            select {
            case <-n.done:
                return
            default:
                if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
                    continue
                }
                if err != io.EOF {
                    if opErr, _ := err.(*net.OpError); opErr.Err.Error() == syscall.ECONNRESET.Error() {
                        log.WithFields(log.Fields{"module": "nodemanager", "function": "messageSender"}).Error("Suicide!")
                    }
                }
                return

            }
        }
        message := &proto_msg.KamajiMessage{}
        err = proto.Unmarshal(tmp, message)
        if err != nil {
            fmt.Println("[messageReciever] error unmarshall message.", err)
            return
        }
        n.Receive <- message
    }
}
