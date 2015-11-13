package main

import (
    "code.google.com/p/go-uuid/uuid"
    "github.com/smaragden/kamaji/kamaji"
//"bufio"
    "fmt"
    log "github.com/Sirupsen/logrus"
    "github.com/docopt/docopt-go"
    "github.com/golang/protobuf/proto"
    "net"
    "strconv"
    "sync"
    "time"
    "github.com/smaragden/kamaji/kamaji/proto"
)

type ClientConn struct {
    *kamaji.Client
    ID     uuid.UUID
    Name   string
    sender chan *proto_msg.KamajiMessage
}

func (c *ClientConn) messageSender() {
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

func doWork(client *ClientConn, message *proto_msg.KamajiMessage) {
    fmt.Println("Doing Work Started: ", message.GetId())
    time.Sleep(time.Millisecond * 1000)
    fmt.Println("Doing Work Done: ", message.GetId())
    response := &proto_msg.KamajiMessage{
        Action: proto_msg.KamajiMessage_STATUS_UPDATE.Enum(),
        Entity: proto_msg.KamajiMessage_COMMAND.Enum(),
        Id:     message.Id,
        Statusupdate: &proto_msg.KamajiMessage_StatusUpdate{
            Destination: proto.Int32(int32(kamaji.DONE)),
        },
    }
    client.sender <- response
}

func handleClientMessage(client *ClientConn, message *proto_msg.KamajiMessage) {
    switch message.GetAction() {
    case proto_msg.KamajiMessage_STATUS_UPDATE:
        status := message.GetStatusupdate()
        fmt.Println("Got Node status update request: ", kamaji.State(status.GetDestination()).S())

        response := &proto_msg.KamajiMessage{
            Action: proto_msg.KamajiMessage_STATUS_UPDATE.Enum(),
            Entity: proto_msg.KamajiMessage_NODE.Enum(),
            Statusupdate: &proto_msg.KamajiMessage_StatusUpdate{
                Destination: proto.Int32(int32(kamaji.State(status.GetDestination()))),
                Name: proto.String(client.Name),
            },
        }
        client.sender <- response
    }

}

func handleCommandMessage(client *ClientConn, message *proto_msg.KamajiMessage) {
    switch message.GetAction() {
    case proto_msg.KamajiMessage_ASSIGN:
        response := &proto_msg.KamajiMessage{
            Action: proto_msg.KamajiMessage_STATUS_UPDATE.Enum(),
            Entity: proto_msg.KamajiMessage_NODE.Enum(),
            Statusupdate: &proto_msg.KamajiMessage_StatusUpdate{
                Destination: proto.Int32(int32(kamaji.WORKING)),
            },
        }
        fmt.Println("Send Node status update request: ", kamaji.WORKING.S(), ", ", client.ID)
        client.sender <- response
        go doWork(client, message)
    }
}

func handleMessage(client *ClientConn, message *proto_msg.KamajiMessage) {
    switch message.GetEntity() {
    case proto_msg.KamajiMessage_NODE:
        handleClientMessage(client, message)
    case proto_msg.KamajiMessage_COMMAND:
        handleCommandMessage(client, message)
    }
}

func cli(cn int, wg *sync.WaitGroup) {
    defer wg.Done()
    //time.Sleep(time.Millisecond * time.Duration(rand.Intn(1000)))
    conn, err := net.Dial("tcp", "127.0.0.1:1314")
    if err != nil {
        fmt.Println("Error!")
        return
    }
    defer conn.Close()
    clientConn := new(ClientConn)
    clientConn.Client = kamaji.NewClient(conn)
    clientConn.ID = uuid.NewRandom()
    clientConn.Name = clientConn.ID.String()
    clientConn.Name = fmt.Sprintf("node%03d.test.now", cn)
    clientConn.sender = make(chan *proto_msg.KamajiMessage)
    go clientConn.messageSender()
    //go reportStats(clientConn)
    for {
        tmp, err := clientConn.ReadMessage()
        if err != nil {
            break
        }
        message := &proto_msg.KamajiMessage{}
        err = proto.Unmarshal(tmp, message)
        if err != nil {
            fmt.Println(err)
            break
        }
        handleMessage(clientConn, message)
    }
    fmt.Println("Exiting Client Loop.")
}

func main() {
    usage := `Kamaji Client Spawner.

Usage:
  client_spawner [-n=5]

Options:
  -n --num_clients=N  Number of clients. [default: 5]
  -h --help                 Show this screen.`
    arguments, err := docopt.Parse(usage, nil, true, "Kamaji Client Spawner", false)
    if err != nil {
        fmt.Println(err)
    }
    client_count, _ := strconv.Atoi(arguments["--num_clients"].(string))
    for {
        fmt.Println("Starting: ", client_count, " clients.")
        var wg sync.WaitGroup
        wg.Add(client_count)
        for i := 0; i < client_count; i++ {
            go cli(i, &wg)
            time.Sleep(time.Millisecond * 2)
        }
        wg.Wait()
        time.Sleep(time.Second * 3)
    }
    fmt.Println("Exiting")
}
