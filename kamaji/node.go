package kamaji

import (
	"code.google.com/p/go-uuid/uuid"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/golang/protobuf/proto"
	"github.com/looplab/fsm"
	"net"
)

func init() {
	log.SetLevel(Config.LOG_LEVEL_CLIENTMANAGER)
}

type Node struct {
	*Client
	ID              uuid.UUID
	State           State
	FSM             *fsm.FSM
	currentCommands []*Command
	statusUpdate    chan State
	sender          chan *KamajiMessage
}

func NewNode(conn net.Conn) *Node {
	c := new(Node)
	c.Client = NewClient(conn)
	c.ID = uuid.NewRandom()
	c.State = UNKNOWN
	c.FSM = fsm.NewFSM(
		c.State.S(),
		fsm.Events{
			{Name: "offline", Src: []string{ONLINE.S(), READY.S(), DISCONNECTING.S()}, Dst: OFFLINE.S()},
			{Name: "online", Src: []string{UNKNOWN.S(), READY.S(), OFFLINE.S()}, Dst: ONLINE.S()},
			{Name: "ready", Src: []string{ONLINE.S(), WORKING.S()}, Dst: READY.S()},
			{Name: "work", Src: []string{READY.S()}, Dst: WORKING.S()},
			{Name: "service", Src: []string{UNKNOWN.S(), ONLINE.S(), OFFLINE.S()}, Dst: SERVICE.S()},
			{Name: "disconnect", Src: []string{ONLINE.S(), READY.S(), WORKING.S(), SERVICE.S()}, Dst: DISCONNECTING.S()},
		},
		fsm.Callbacks{
			"enter_state":     func(e *fsm.Event) { c.enterState(e) },
			ONLINE.S():        func(e *fsm.Event) { c.onlineNode(e) },
			OFFLINE.S():       func(e *fsm.Event) { c.offlineNode(e) },
			READY.S():         func(e *fsm.Event) { c.readyNode(e) },
			WORKING.S():       func(e *fsm.Event) { c.workNode(e) },
			SERVICE.S():       func(e *fsm.Event) { c.serviceNode(e) },
			DISCONNECTING.S(): func(e *fsm.Event) { c.disconnectNode(e) },
		},
	)

	c.statusUpdate = make(chan State)
	c.sender = make(chan *KamajiMessage)
	return c
}

func (c *Node) isEqual(other *Node) bool {
	return c.ID.String() == other.ID.String()
}

func (c *Node) enterState(e *fsm.Event) {
	c.State = StateFromString(e.Dst)
	//fmt.Println("Sending status change over channel")
	c.statusUpdate <- c.State
	log.WithFields(log.Fields{
		"module": "clientmanager",
		"client": c.ID,
		"from":   e.Src,
		"to":     e.Dst,
	}).Debug("Changing Node State")
}

func (c *Node) onlineNode(e *fsm.Event) {
	//fmt.Printf("Starting Node: %s\n", c.ID)
}

func (c *Node) offlineNode(e *fsm.Event) {
	c.Conn.Close()
	//fmt.Printf("Stopping Node: %s\n", c.ID)
}

func (c *Node) readyNode(e *fsm.Event) {
	//fmt.Printf("Setting Node %s to %s.\n", c.ID, e.Dst)
}

func (c *Node) workNode(e *fsm.Event) {
	//fmt.Printf("Setting Node %s to %s.\n", c.ID, e.Dst)
}

func (c *Node) serviceNode(e *fsm.Event) {
	//fmt.Printf("Setting Node %s to %s.\n", c.ID, e.Dst)
}

func (c *Node) disconnectNode(e *fsm.Event) {
	//fmt.Printf("Node %s Disconnecting.\n", c.ID)
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
			log.WithFields(log.Fields{"module": "clientmanager"}).Error(err)
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
	log.WithFields(log.Fields{"module": "clientmanager", "client": n.ID, "command": command.Name}).Debug("Assigning command to client.")
	n.sender <- message
	return nil
}

func (n *Node) requestStatusUpdate(state State) error {
	log.WithFields(log.Fields{"module": "clientmanager", "client": n.ID, "status": n.State}).Debug("State online, request status update.")
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
		return nil, err
	}
	message := &KamajiMessage{}
	err = proto.Unmarshal(tmp, message)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return message, nil
}