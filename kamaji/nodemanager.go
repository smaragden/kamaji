package kamaji

import (
    "fmt"
    log "github.com/Sirupsen/logrus"
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

    nodeLock     sync.RWMutex
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
                node.Lock()
                if node.State == READY {
                    select {
                    case <-nm.done:
                        return
                    case nm.NextNode <- node:
                    }
                }
                node.Unlock()
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
    defer ln.Close()
    for {
        conn, err := ln.Accept()
        if err != nil {
            log.Printf("Failed to accept: %s", err)
            continue
        }
        go nm.handleConnection(conn)
    }
}

// Promptly closing the connections.
// Let the node do it own cleanup.
func (nm *NodeManager) Stop() {
    log.WithFields(log.Fields{
        "module":  "nodemanager",
        "action":  "stop",
    }).Info("Stopping Node Manager")
    close(nm.done)
    for _, node := range nm.Nodes {
        node.Conn.Close()
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
            }
        }
    }
}

func (nm *NodeManager) handleCommandMessage(n *Node, message *proto_msg.KamajiMessage) {
    switch message.GetAction() {
    case proto_msg.KamajiMessage_STATUS_UPDATE:
        {
            status := message.GetStatusupdate()
            if State(status.GetDestination()) == DONE {
                command, err := n.removeCommand(message.GetId())
                if err != nil {
                    log.WithFields(log.Fields{"module": "nodemanager", "node": n.Name}).Error(err)
                }
                log.WithFields(log.Fields{"module": "nodemanager", "node": n.Name, "command": message.GetId()}).Info("Command Done.")
                n.ChangeState("ready")
                command.ChangeState("finish")
            }
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
    defer node.ChangeState("offline")
    for {
        if node.State == ONLINE {
            err := node.requestStatusUpdate(READY)
            if err != nil {
                log.WithFields(log.Fields{"module": "nodemanager", "node": node.Name}).Error(err)
            }
        }
        select {
        case <-nm.done:
            return
        case message := <-node.Receive:
            err, pass := nm.handleMessage(node, message)
            if err != nil {
                if !pass {
                    log.WithFields(log.Fields{"module": "nodemanager", "node": node.Name}).Error(err)
                }
            }
        }
    }
}

