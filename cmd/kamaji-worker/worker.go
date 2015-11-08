package main

import (
	"github.com/smaragden/kamaji/kamaji"
	//"bufio"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/docopt/docopt-go"
	"github.com/golang/protobuf/proto"
	"github.com/shirou/gopsutil/host"
	"math/rand"
	"net"
	"strconv"
	"sync"
	"time"
)

type Stats struct {
	Host *host.HostInfoStat
}

type Message struct {
	Action string
	Stats  Stats
}

type ClientConn struct {
	*kamaji.Client
	sender chan *kamaji.KamajiMessage
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

func doWork(client *ClientConn, message *kamaji.KamajiMessage) {
	time.Sleep(time.Duration(rand.Int31n(5000)) * time.Millisecond)
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

func handleClientMessage(client *ClientConn, message *kamaji.KamajiMessage) {
	switch message.GetAction() {
	case kamaji.KamajiMessage_STATUS_UPDATE:
		status := message.GetStatusupdate()
		fmt.Println(kamaji.State(status.GetDestination()).S())

		response := &kamaji.KamajiMessage{
			Action: kamaji.KamajiMessage_STATUS_UPDATE.Enum(),
			Entity: kamaji.KamajiMessage_NODE.Enum(),
			Statusupdate: &kamaji.KamajiMessage_StatusUpdate{
				Destination: proto.Int32(int32(kamaji.READY)),
			},
		}
		client.sender <- response
	}

}

func handleCommandMessage(client *ClientConn, message *kamaji.KamajiMessage) {
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

func handleMessage(client *ClientConn, message *kamaji.KamajiMessage) {
	switch message.GetEntity() {
	case kamaji.KamajiMessage_NODE:
		handleClientMessage(client, message)
	case kamaji.KamajiMessage_COMMAND:
		handleCommandMessage(client, message)
	}
}

func cli(cn int, wg *sync.WaitGroup) {
	defer wg.Done()
	time.Sleep(time.Millisecond * time.Duration(rand.Intn(1000)))
	conn, err := net.Dial("tcp", "127.0.0.1:1314")
	if err != nil {
		fmt.Println("Error!")
		return
	}
	defer conn.Close()
	clientConn := new(ClientConn)
	clientConn.Client = kamaji.NewClient(conn)
	clientConn.sender = make(chan *kamaji.KamajiMessage)
	go clientConn.messageSender()
	//go reportStats(clientConn)
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
		time.Sleep(time.Second * 10)
	}
	fmt.Println("Exiting")
}
