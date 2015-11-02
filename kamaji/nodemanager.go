package kamaji

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/golang/protobuf/proto"
	"io"
	"net"
)

type NodeManager struct {
	*PoolManager
	Addr         string
	Port         int
	poolsUpdated chan int
}

func NewNodeManager(name string, address string, port int) *NodeManager {
	fmt.Println("Creating NodeManager")
	cm := new(NodeManager)
	cm.PoolManager = NewPoolManager(name)
	cm.Addr = address
	cm.Port = port
	cm.CreatePools([]string{OFFLINE.S(), ONLINE.S(), READY.S(), WORKING.S(), SERVICE.S()})
	cm.poolsUpdated = make(chan int)
	//go cm.poolsReporter()
	return cm
}

func (nm *NodeManager) GetAddrStr() string {
	return fmt.Sprintf("%s:%d", nm.Addr, nm.Port)
}

func (nm *NodeManager) Start() {
	ln, err := net.Listen("tcp", nm.GetAddrStr())
	if err != nil {
		log.Fatalf("Failed to listen: %s", err)
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("Failed to accept: %s", err)
			continue
		}
		go nm.handleConnection(conn)
	}
}

func (nm *NodeManager) AddNode(conn net.Conn) *Node {
	c := NewNode(conn)
	go nm.NodeStatusChanged(c)
	go c.messageSender()
	return c
}

func (nm *NodeManager) getAvailableNode() *Node {
	for _, v := range nm.Pools[READY.S()].Items {
		node := v.(*Node)
		if node.State == READY {
			return node
		}
	}
	return nil
}

func (nm *NodeManager) NodeStatusChanged(n *Node) error {
	for {
		status := <-n.statusUpdate
		switch status {
		case ONLINE:
			nm.MoveItemToPool(n, "ONLINE")
		case OFFLINE:
			nm.MoveItemToPool(n, "OFFLINE")
		case READY:
			nm.MoveItemToPool(n, "READY")
		case WORKING:
			nm.MoveItemToPool(n, "WORKING")
		case SERVICE:
			nm.MoveItemToPool(n, "SERVICE")
		case DONE:
			nm.MoveItemToPool(n, "DONE")
		}
	}
	return nil
}

func (nm *NodeManager) handleNodeMessage(n *Node, message *KamajiMessage) {
	switch message.GetAction() {
	case KamajiMessage_STATUS_UPDATE:
		{
			status := message.GetStatusupdate()
			if State(status.GetDestination()) == READY {
				n.FSM.Event("ready")
			}
			if State(status.GetDestination()) == SERVICE {
				n.FSM.Event("service")
			}
		}
	case KamajiMessage_QUERY:
		{
			message := &KamajiMessage{
				Action: KamajiMessage_QUERY.Enum(),
				Entity: KamajiMessage_NODE.Enum(),
			}
			for p, _ := range nm.itemToPool {
				client := p.(*Node)
				clientItem := &KamajiMessage_NodeItem{
					Id:    proto.String(client.ID.String()),
					State: proto.Int32(int32(client.State)),
				}
				message.Messageitems = append(message.Messageitems, clientItem)
			}
			n.sender <- message
		}
	}
}

func (nm *NodeManager) handleCommandMessage(n *Node, message *KamajiMessage) {
	switch message.GetAction() {
	case KamajiMessage_STATUS_UPDATE:
		{
			status := message.GetStatusupdate()
			if State(status.GetDestination()) == DONE {
				CommandEvent <- message
				n.FSM.Event("ready")
			}
		}
	case KamajiMessage_QUERY:
		{
			fmt.Println("Got command query")
			message := &KamajiMessage{
				Action: KamajiMessage_QUERY.Enum(),
				Entity: KamajiMessage_COMMAND.Enum(),
			}
			allCommands := <-AllCommands
			for _, command := range allCommands {
				commandItem := &KamajiMessage_CommandItem{
					Name:  proto.String(fmt.Sprintf("%s | %s | %s", command.Task.Job.Name, command.Task.Name, command.Name)),
					Id:    proto.String(command.ID.String()),
					State: proto.Int32(int32(command.State)),
				}
				message.Commanditems = append(message.Commanditems, commandItem)
			}
			n.sender <- message
		}
	}
}

func (nm *NodeManager) handleMessage(n *Node, message *KamajiMessage) error {
	switch message.GetEntity() {
	case KamajiMessage_NODE:
		nm.handleNodeMessage(n, message)
	case KamajiMessage_COMMAND:
		nm.handleCommandMessage(n, message)
	}
	return nil
}

func (nm *NodeManager) handleConnection(conn net.Conn) {
	node := nm.AddNode(conn)
	node.FSM.Event("online")
	defer node.FSM.Event("offline")
	for {
		if node.State == ONLINE {
			err := node.requestStatusUpdate(READY)
			if err != nil {
				log.WithFields(log.Fields{"module": "clientmanager", "client": node.ID}).Error(err)
			}
		}
		message, err := node.recieveMessage()
		if err != nil {
			if err == io.EOF {
				log.WithFields(log.Fields{"module": "clientmanager", "client": node.ID}).Error(err)
				node.FSM.Event("disconnect")
			}
			break
		}
		err = nm.handleMessage(node, message)
		if err != nil {
			log.WithFields(log.Fields{"module": "clientmanager", "client": node.ID}).Error(err)
		}
	}
}
