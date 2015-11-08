package kamaji

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/golang/protobuf/proto"
	"io"
	"net"
	"time"
)

type NodeList []*Node

type NodeManager struct {
	Addr         string
	Port         int
	poolsUpdated chan int
	NextNode     chan *Node
	Nodes        NodeList
}

func NewNodeManager(name string, address string, port int) *NodeManager {
	fmt.Println("Creating NodeManager")
	cm := new(NodeManager)
	cm.Addr = address
	cm.Port = port
	cm.NextNode = make(chan *Node)
	go cm.nodeProvider()
	return cm
}

func (nm *NodeManager) nodeProvider() {
	for {
		for _, node := range nm.Nodes {
			if node.State == READY {
				node.ChangeState("assign")
				nm.NextNode <- node
			}
		}
		time.Sleep(time.Millisecond * 100)
	}
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
	n := NewNode(conn)
	go n.messageSender()
	nm.Nodes = append(nm.Nodes, n)
	return n
}

func (nm *NodeManager) handleNodeMessage(n *Node, message *KamajiMessage) {
	switch message.GetAction() {
	case KamajiMessage_STATUS_UPDATE:
		{
			status := message.GetStatusupdate()
			switch State(status.GetDestination()) {
			case READY:
				n.ChangeState("ready")
			case WORKING:
				n.ChangeState("work")
			case SERVICE:
				n.ChangeState("service")
			}
		}
	case KamajiMessage_QUERY:
		{
			message := &KamajiMessage{
				Action: KamajiMessage_QUERY.Enum(),
				Entity: KamajiMessage_NODE.Enum(),
			}
			for _, node := range nm.Nodes {
				nodeItem := &KamajiMessage_NodeItem{
					Id:    proto.String(node.ID.String()),
					State: proto.Int32(int32(node.State)),
				}
				message.Messageitems = append(message.Messageitems, nodeItem)
			}
			n.sender <- message
		}
	}
}

func (nm *NodeManager) handleJobMessage(n *Node, message *KamajiMessage) {
	switch message.GetAction() {
	case KamajiMessage_STATUS_UPDATE:
		{
			status := message.GetStatusupdate()
			switch State(status.GetDestination()) {
			case DONE:
				CommandEvent <- message
				n.ChangeState("ready")
			}
		}
	case KamajiMessage_QUERY:
		{
			fmt.Println("Got job query")
			message := &KamajiMessage{
				Action: KamajiMessage_QUERY.Enum(),
				Entity: KamajiMessage_JOB.Enum(),
			}
			allJobs := <-AllJobs
			fmt.Printf("Jobs: %+v\n", allJobs)
			for _, job := range allJobs {
				jobItem := &KamajiMessage_JobItem{
					Name:  proto.String(fmt.Sprintf("%s", job.Name)),
					Id:    proto.String(job.ID.String()),
					State: proto.Int32(int32(job.State)),
				}
				message.Jobitems = append(message.Jobitems, jobItem)
			}
			n.sender <- message
		}
	}
}

func (nm *NodeManager) handleTaskMessage(n *Node, message *KamajiMessage) {
	switch message.GetAction() {
	case KamajiMessage_STATUS_UPDATE:
		{
			status := message.GetStatusupdate()
			switch State(status.GetDestination()) {
			case DONE:
				CommandEvent <- message
				n.ChangeState("ready")
			}
		}
	case KamajiMessage_QUERY:
		{
			fmt.Println("Got task query")
			message := &KamajiMessage{
				Action: KamajiMessage_QUERY.Enum(),
				Entity: KamajiMessage_TASK.Enum(),
			}
			allTasks := <-AllTasks
			for _, task := range allTasks {
				taskItem := &KamajiMessage_TaskItem{
					Name:  proto.String(fmt.Sprintf("%s", task.Name)),
					Id:    proto.String(task.ID.String()),
					State: proto.Int32(int32(task.State)),
				}
				message.Taskitems = append(message.Taskitems, taskItem)
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
				err := n.removeCommand(message.GetId())
				if err != nil {
					log.WithFields(log.Fields{"module": "nodemanager", "node": n.ID}).Error(err)
				}
				log.WithFields(log.Fields{"module": "nodemanager", "client": n.ID, "command": message.GetId()}).Info("Command Done.")
				n.ChangeState("ready")
				CommandEvent <- message
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
	case KamajiMessage_JOB:
		nm.handleJobMessage(n, message)
	case KamajiMessage_TASK:
		nm.handleTaskMessage(n, message)
	case KamajiMessage_COMMAND:
		nm.handleCommandMessage(n, message)
	}
	return nil
}

func (nm *NodeManager) handleConnection(conn net.Conn) {
	node := nm.AddNode(conn)
	node.ChangeState("online")
	defer node.ChangeState("offline")
	for {
		if node.State == ONLINE {
			err := node.requestStatusUpdate(READY)
			if err != nil {
				log.WithFields(log.Fields{"module": "nodemanager", "node": node.ID}).Error(err)
			}
		}
		message, err := node.recieveMessage()
		if err != nil {
			if err == io.EOF {
				log.WithFields(log.Fields{"module": "nodemanager", "node": node.ID}).Error(err)
			}
			fmt.Println(err)
			break
		}
		err = nm.handleMessage(node, message)
		if err != nil {
			log.WithFields(log.Fields{"module": "nodemanager", "node": node.ID}).Error(err)
		}
	}
}
