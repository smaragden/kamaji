package main

import (
    "github.com/smaragden/kamaji/kamaji"
    "fmt"
    log "github.com/Sirupsen/logrus"
    "github.com/docopt/docopt-go"
    "github.com/golang/protobuf/proto"
    "net"
    "os"
)

type Node struct {
    *kamaji.Client
    Name   string
    sender chan *kamaji.KamajiMessage
}

func (c *Node) messageSender() {
    for {
        for {
            message := <-c.sender
            message_data, err := proto.Marshal(message)
            if err != nil {
                fmt.Println(err)
                continue
            }
            _, err = c.SendMessage(message_data)
            if err != nil {
                log.WithFields(log.Fields{"module": "clientmanager"}).Error(err)
                continue
            }
        }
    }
}

func doWork(client *Node, message *kamaji.KamajiMessage) {
    response := &kamaji.KamajiMessage{
        Action: kamaji.KamajiMessage_STATUS_UPDATE.Enum(),
        Entity: kamaji.KamajiMessage_COMMAND.Enum(),
        Id:     message.Id,
        Statusupdate: &kamaji.KamajiMessage_StatusUpdate{
            Destination: proto.Int32(int32(kamaji.DONE)),
        },
    }
    client.sender <- response
}

func handleClientMessage(client *Node, message *kamaji.KamajiMessage) {
    switch message.GetAction() {
    case kamaji.KamajiMessage_STATUS_UPDATE:
        status := message.GetStatusupdate()
        fmt.Println(kamaji.State(status.GetDestination()).S())

        response := &kamaji.KamajiMessage{
            Action: kamaji.KamajiMessage_STATUS_UPDATE.Enum(),
            Entity: kamaji.KamajiMessage_NODE.Enum(),
            Statusupdate: &kamaji.KamajiMessage_StatusUpdate{
                Destination: proto.Int32(int32(kamaji.READY)),
                Name: proto.String(client.Name),
            },
        }
        client.sender <- response
    }

}

func handleCommandMessage(client *Node, message *kamaji.KamajiMessage) {
    switch message.GetAction() {
    case kamaji.KamajiMessage_ASSIGN:
        response := &kamaji.KamajiMessage{
            Action: kamaji.KamajiMessage_STATUS_UPDATE.Enum(),
            Entity: kamaji.KamajiMessage_NODE.Enum(),
            Statusupdate: &kamaji.KamajiMessage_StatusUpdate{
                Destination: proto.Int32(int32(kamaji.WORKING)),
            },
        }
        client.sender <- response
        go doWork(client, message)
    }
}

func handleMessage(client *Node, message *kamaji.KamajiMessage) {
    switch message.GetEntity() {
    case kamaji.KamajiMessage_NODE:
        handleClientMessage(client, message)
    case kamaji.KamajiMessage_COMMAND:
        handleCommandMessage(client, message)
    }
}

func main() {
    usage := `Kamaji Client Spawner.

Usage:
  kamaji-node <server> [-p <port>] [-n <name>]
  kamaji-node (-h | --help | --version)

Options:
  -p, --port=<port>  	Port. [default: 1314]
  -n, --name=<name>     Alternative name to use for this node.
  -h, --help            Show this screen.
  -v, --verbose			Verbose output.`
    arguments, err := docopt.Parse(usage, nil, true, "Kamaji Client Spawner", false)
    if err != nil {
        fmt.Println(err)
    }
    fmt.Printf("%+v\n", arguments)
    conn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", arguments["<server>"], arguments["--port"]))
    if err != nil {
        fmt.Println("Error!")
        return
    }
    defer conn.Close()
    clientConn := new(Node)
    clientConn.Client = kamaji.NewClient(conn)
    name, ok := arguments["name"]
    if ok == true && name != nil {
        clientConn.Name = name.(string)
    }else {
        hostname, err := os.Hostname()
        if err == nil {
            clientConn.Name = hostname
        }
    }
    fmt.Printf("Name: %s\n", clientConn.Name)
    clientConn.sender = make(chan *kamaji.KamajiMessage)
    go clientConn.messageSender()
    for {
        tmp, err := clientConn.ReadMessage()
        if err != nil {
            break
        }
        message := &kamaji.KamajiMessage{}
        err = proto.Unmarshal(tmp, message)
        if err != nil {
            fmt.Println(err)
            break
        }
        handleMessage(clientConn, message)
    }
    fmt.Println("Exiting")
}
