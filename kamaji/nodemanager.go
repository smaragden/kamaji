package kamaji

import (
    "fmt"
    log "github.com/Sirupsen/logrus"
    "github.com/golang/protobuf/proto"
    "net"
    "time"
    "github.com/smaragden/kamaji/kamaji/proto"
    "sync"
    "errors"
)

type NodeList []*Node

type NodeManager struct {
    Addr         string
    Port         int
    poolsUpdated chan int
    NextNode     chan *Node
    Nodes        NodeList
    done         chan struct{}
    waitGroup    *sync.WaitGroup
}

func NewNodeManager(address string, port int) *NodeManager {
    cm := new(NodeManager)
    cm.Addr = address
    cm.Port = port
    cm.NextNode = make(chan *Node)
    cm.done = make(chan struct{})
    cm.waitGroup = &sync.WaitGroup{}
    go cm.nodeProvider()
    return cm
}

func (nm *NodeManager) nodeProvider() {
    nm.waitGroup.Add(1)
    defer nm.waitGroup.Done()
    for {
        select {
        case <-nm.done:
            return
        case <-time.After(time.Millisecond * 10):
            for _, node := range nm.Nodes {
                if node.State == READY {
                    select {
                    case <-nm.done:
                        return
                    case nm.NextNode <- node:
                    }
                }
            }
        }
    }
}

func (nm *NodeManager) GetAddrStr() string {
    return fmt.Sprintf("%s:%d", nm.Addr, nm.Port)
}

func (nm *NodeManager) Start() {
    log.WithFields(log.Fields{
        "module":  "nodemanager",
        "action":  "start",
    }).Info("Starting Node Manager.")
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

func (nm *NodeManager) Stop() {
    log.WithFields(log.Fields{
        "module":  "nodemanager",
        "action":  "stop",
    }).Info("Stopping Node Manager.")
    close(nm.done)
    for _, node := range nm.Nodes {
        node.ChangeState("offline")
    }
}


func (nm *NodeManager) AddNode(conn net.Conn) *Node {
    n := NewNode(conn)
    nm.Nodes = append(nm.Nodes, n)
    return n
}

func (nm *NodeManager) handleNodeMessage(n *Node, message *proto_msg.KamajiMessage) {
    switch message.GetAction() {
    case proto_msg.KamajiMessage_STATUS_UPDATE:
        {
            status := message.GetStatusupdate()
            name := status.GetName()
            if name != "" {
                log.WithFields(log.Fields{"module": "nodemanager", "old_name": n.Name, "new_name": name}).Debug("Rename Node.")
                n.Name = name
            }
            switch State(status.GetDestination()) {
            case READY:
                n.ChangeState("ready")
            case WORKING:
                n.ChangeState("work")
            case SERVICE:
                n.ChangeState("service")
            }
        }
    case proto_msg.KamajiMessage_QUERY:
        {
            message := &proto_msg.KamajiMessage{
                Action: proto_msg.KamajiMessage_QUERY.Enum(),
                Entity: proto_msg.KamajiMessage_NODE.Enum(),
            }
            for _, node := range nm.Nodes {
                nodeItem := &proto_msg.KamajiMessage_NodeItem{
                    Id:    proto.String(node.ID.String()),
                    State: proto.Int32(int32(node.State)),
                }
                message.Messageitems = append(message.Messageitems, nodeItem)
            }
            n.Send <- message
        }
    }
}

func (nm *NodeManager) handleJobMessage(n *Node, message *proto_msg.KamajiMessage) {
    switch message.GetAction() {
    case proto_msg.KamajiMessage_STATUS_UPDATE:
        {
            status := message.GetStatusupdate()
            switch State(status.GetDestination()) {
            case DONE:
                CommandEvent <- message
                n.ChangeState("ready")
            }
        }
    case proto_msg.KamajiMessage_QUERY:
        {
            /*
            fmt.Println("Got job query")
            message := &proto_msg.KamajiMessage{
                Action: proto_msg.KamajiMessage_QUERY.Enum(),
                Entity: proto_msg.KamajiMessage_JOB.Enum(),
            }
            allJobs := <-AllJobs
            fmt.Printf("Jobs: %+v\n", allJobs)
            for _, job := range allJobs {
                jobItem := &proto_msg.KamajiMessage_JobItem{
                    Name:  proto.String(fmt.Sprintf("%s", job.Name)),
                    Id:    proto.String(job.ID.String()),
                    State: proto.Int32(int32(job.State)),
                }
                message.Jobitems = append(message.Jobitems, jobItem)
            }
            n.Send <- message
            */
        }
    }
}

func (nm *NodeManager) handleTaskMessage(n *Node, message *proto_msg.KamajiMessage) {
    switch message.GetAction() {
    case proto_msg.KamajiMessage_STATUS_UPDATE:
        {
            status := message.GetStatusupdate()
            switch State(status.GetDestination()) {
            case DONE:
                CommandEvent <- message
                n.ChangeState("ready")
            }
        }
    case proto_msg.KamajiMessage_QUERY:
        {
            /*
            fmt.Println("Got task query")
            message := &proto_msg.KamajiMessage{
                Action: proto_msg.KamajiMessage_QUERY.Enum(),
                Entity: proto_msg.KamajiMessage_TASK.Enum(),
            }
            allTasks := <-AllTasks
            for _, task := range allTasks {
                taskItem := &proto_msg.KamajiMessage_TaskItem{
                    Name:  proto.String(fmt.Sprintf("%s", task.Name)),
                    Id:    proto.String(task.ID.String()),
                    State: proto.Int32(int32(task.State)),
                }
                message.Taskitems = append(message.Taskitems, taskItem)
            }
            n.Send <- message
            */
        }
    }
}

func (nm *NodeManager) handleCommandMessage(n *Node, message *proto_msg.KamajiMessage) {
    switch message.GetAction() {
    case proto_msg.KamajiMessage_STATUS_UPDATE:
        {
            status := message.GetStatusupdate()
            if State(status.GetDestination()) == DONE {
                err := n.removeCommand(message.GetId())
                if err != nil {
                    log.WithFields(log.Fields{"module": "nodemanager", "node": n.Name}).Error(err)
                }
                log.WithFields(log.Fields{"module": "nodemanager", "node": n.Name, "command": message.GetId()}).Info("Command Done.")
                n.ChangeState("ready")
                CommandEvent <- message
            }
        }
    case proto_msg.KamajiMessage_QUERY:
        {
            /*
            fmt.Println("Got command query")
            message := &proto_msg.KamajiMessage{
                Action: proto_msg.KamajiMessage_QUERY.Enum(),
                Entity: proto_msg.KamajiMessage_COMMAND.Enum(),
            }
            allCommands := <-AllCommands
            for _, command := range allCommands {
                commandItem := &proto_msg.KamajiMessage_CommandItem{
                    Name:  proto.String(fmt.Sprintf("%s | %s | %s", command.Task.Job.Name, command.Task.Name, command.Name)),
                    Id:    proto.String(command.ID.String()),
                    State: proto.Int32(int32(command.State)),
                }
                message.Commanditems = append(message.Commanditems, commandItem)
            }
            n.Send <- message
            */
        }
    }
}

func (nm *NodeManager) handleMessage(n *Node, message *proto_msg.KamajiMessage) (error, bool) {
    select {
    case <-nm.done:
        return errors.New("Node Manager shutting down. Message Rejected."), true
    default:
        {
            switch message.GetEntity() {
            case proto_msg.KamajiMessage_NODE:
                nm.handleNodeMessage(n, message)
            case proto_msg.KamajiMessage_JOB:
                nm.handleJobMessage(n, message)
            case proto_msg.KamajiMessage_TASK:
                nm.handleTaskMessage(n, message)
            case proto_msg.KamajiMessage_COMMAND:
                nm.handleCommandMessage(n, message)
            }
        }
    }
    return nil, true
}

func (nm *NodeManager) handleConnection(conn net.Conn) {
    node := nm.AddNode(conn)
    node.ChangeState("online")
    for {
        if node.State == ONLINE {
            err := node.requestStatusUpdate(READY)
            if err != nil {
                log.WithFields(log.Fields{"module": "nodemanager", "node": node.Name}).Error(err)
            }
        }
        message := <-node.Receive
        err, pass := nm.handleMessage(node, message)
        if err != nil {
            if !pass {
                log.WithFields(log.Fields{"module": "nodemanager", "node": node.Name}).Error(err)
            }
        }
    }
}

