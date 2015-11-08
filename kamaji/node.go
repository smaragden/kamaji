package kamaji

import (
	"code.google.com/p/go-uuid/uuid"
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/golang/protobuf/proto"
	"github.com/looplab/fsm"
	"net"
	"sync"
)

func init() {
	log.SetLevel(Config.LOG_LEVEL_CLIENTMANAGER)
}

type Node struct {
	*Client
	ID uuid.UUID
	sync.RWMutex
	State           State
	FSM             *fsm.FSM
	currentCommands []*Command
	statusUpdate    chan State
	sender          chan *KamajiMessage
}

func NewNode(conn net.Conn) *Node {
	n := new(Node)
	n.Client = NewClient(conn)
	n.ID = uuid.NewRandom()
	n.State = UNKNOWN
	n.FSM = fsm.NewFSM(
		n.State.S(),
		fsm.Events{
			{Name: "offline", Src: []string{ONLINE.S(), READY.S(), DISCONNECTING.S()}, Dst: OFFLINE.S()},
			{Name: "online", Src: []string{UNKNOWN.S(), READY.S(), OFFLINE.S()}, Dst: ONLINE.S()},
			{Name: "ready", Src: []string{ONLINE.S(), WORKING.S()}, Dst: READY.S()},
			{Name: "assign", Src: []string{READY.S()}, Dst: ASSIGNING.S()},
			{Name: "work", Src: []string{ASSIGNING.S()}, Dst: WORKING.S()},
			{Name: "service", Src: []string{UNKNOWN.S(), ONLINE.S(), OFFLINE.S()}, Dst: SERVICE.S()},
		},
		fsm.Callbacks{
			"after_event": func(e *fsm.Event) { n.afterEvent(e) },
			OFFLINE.S():   func(e *fsm.Event) { n.offlineNode(e) },
		},
	)

	n.sender = make(chan *KamajiMessage)
	return n
}

func (n *Node) ChangeState(state string) {
	n.Lock()
	defer n.Unlock()
	err := n.FSM.Event(state)
	if err != nil {
		log.WithFields(log.Fields{"module": "nodemanager", "fuction": "stateChanger", "node": n.ID}).Fatal(err)
	}
}

func (c *Node) isEqual(other *Node) bool {
	return c.ID.String() == other.ID.String()
}

func (c *Node) afterEvent(e *fsm.Event) {
	c.State = StateFromString(e.Dst)
	log.WithFields(log.Fields{
		"module": "node",
		"node":   c.ID,
		"from":   e.Src,
		"to":     e.Dst,
	}).Debug("Changing Node State")
	//c.statusUpdate <- c.State
}

func (c *Node) offlineNode(e *fsm.Event) {
	c.Conn.Close()
	// TODO: Handle stray jobs
	//fmt.Printf("Stopping Node: %s\n", c.ID)
}

func (n *Node) messageSender() {
	for {
		message := <-n.sender
		message_data, err := proto.Marshal(message)
		if err != nil {
			fmt.Println(err)
			continue
		}
		_, err = n.SendMessage(message_data)
		if err != nil {
			log.WithFields(log.Fields{"module": "nodemanager", "function": "messageSender"}).Error(err)
			continue
		}
	}
}

func (n *Node) assignCommand(command *Command) error {
	message := &KamajiMessage{
		Action: KamajiMessage_ASSIGN.Enum(),
		Entity: KamajiMessage_COMMAND.Enum(),
		Id:     proto.String(command.ID.String()),
	}
	n.currentCommands = append(n.currentCommands, command)
	log.WithFields(log.Fields{"module": "nodemanager", "client": n.ID, "command": command.Name}).Info("Assigning command to client.")
	n.sender <- message
	return nil
}

func (n *Node) removeCommand(command_id string) error {
	log.WithFields(log.Fields{"module": "nodemanager", "client": n.ID, "command": command_id}).Debug("Clean up command.")
	for i, node_commands := range n.currentCommands {
		if node_commands.ID.String() == command_id {
			n.currentCommands = append(n.currentCommands[:i], n.currentCommands[i+1:]...)
			return nil
		}
	}
	return errors.New("Couldn't find command on client.")
}

func (n *Node) requestStatusUpdate(state State) error {
	log.WithFields(log.Fields{"module": "nodemanager", "client": n.ID, "state": n.State, "new_state": state}).Debug("State Update Request")
	message := &KamajiMessage{
		Action: KamajiMessage_STATUS_UPDATE.Enum(),
		Entity: KamajiMessage_NODE.Enum(),
		Id:     proto.String(n.ID.String()),
		Statusupdate: &KamajiMessage_StatusUpdate{
			Destination: proto.Int32(int32(state)),
		},
	}
	n.sender <- message
	return nil
}

func (n *Node) recieveMessage() (*KamajiMessage, error) {
	tmp, err := n.ReadMessage()
	if err != nil {
		fmt.Println("Error reading message: ", err)
		return nil, err
	}
	message := &KamajiMessage{}
	err = proto.Unmarshal(tmp, message)
	if err != nil {
		fmt.Println("error unmarshall message.", err)
		return nil, err
	}
	return message, nil
}
