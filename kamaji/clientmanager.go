package kamaji

import (
	"bufio"
	"code.google.com/p/go-uuid/uuid"
	"encoding/gob"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/golang/protobuf/proto"
	"github.com/looplab/fsm"
	//"github.com/shirou/gopsutil/host"
	"io"
	"net"
	//"os"
	"sync"
	//"text/tabwriter"
)

func init() {
	log.SetLevel(Config.LOG_LEVEL_CLIENTMANAGER)
}

type ClientManager struct {
	*PoolManager
	Addr         string
	Port         int
	poolsUpdated chan int
}

func NewClientManager(name string, address string, port int) *ClientManager {
	fmt.Println("Creating ClientManager")
	cm := new(ClientManager)
	cm.PoolManager = NewPoolManager(name)
	cm.Addr = address
	cm.Port = port
	cm.CreatePools([]string{OFFLINE.S(), ONLINE.S(), READY.S(), WORKING.S(), SERVICE.S()})
	cm.poolsUpdated = make(chan int)
	//go cm.poolsReporter()
	return cm
}

type Client struct {
	net.Conn
	ID              uuid.UUID
	State           State
	reader          *bufio.Reader
	writer          *bufio.Writer
	encoder         *gob.Encoder
	decoder         *gob.Decoder
	FSM             *fsm.FSM
	currentCommands []*Command
	statusUpdate    chan State
	sender          chan *KamajiMessage
}

func NewClient(conn net.Conn) *Client {
	c := new(Client)
	c.Conn = conn
	c.ID = uuid.NewRandom()
	c.State = UNKNOWN
	c.reader = bufio.NewReader(conn)
	c.writer = bufio.NewWriterSize(conn, 4096)
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
			ONLINE.S():        func(e *fsm.Event) { c.onlineClient(e) },
			OFFLINE.S():       func(e *fsm.Event) { c.offlineClient(e) },
			READY.S():         func(e *fsm.Event) { c.readyClient(e) },
			WORKING.S():       func(e *fsm.Event) { c.workClient(e) },
			SERVICE.S():       func(e *fsm.Event) { c.serviceClient(e) },
			DISCONNECTING.S(): func(e *fsm.Event) { c.disconnectClient(e) },
		},
	)

	c.statusUpdate = make(chan State)
	c.sender = make(chan *KamajiMessage)
	return c
}

func (c *Client) isEqual(other *Client) bool {
	return c.ID.String() == other.ID.String()
}

func (c *Client) enterState(e *fsm.Event) {
	c.State = StateFromString(e.Dst)
	//fmt.Println("Sending status change over channel")
	c.statusUpdate <- c.State
	log.WithFields(log.Fields{
		"module": "clientmanager",
		"client": c.ID,
		"from":   e.Src,
		"to":     e.Dst,
	}).Debug("Changing Client State")
}

func (c *Client) onlineClient(e *fsm.Event) {
	//fmt.Printf("Starting Client: %s\n", c.ID)
}

func (c *Client) offlineClient(e *fsm.Event) {
	c.Conn.Close()
	//fmt.Printf("Stopping Client: %s\n", c.ID)
}

func (c *Client) readyClient(e *fsm.Event) {
	//fmt.Printf("Setting Client %s to %s.\n", c.ID, e.Dst)
}

func (c *Client) workClient(e *fsm.Event) {
	//fmt.Printf("Setting Client %s to %s.\n", c.ID, e.Dst)
}

func (c *Client) serviceClient(e *fsm.Event) {
	//fmt.Printf("Setting Client %s to %s.\n", c.ID, e.Dst)
}

func (c *Client) disconnectClient(e *fsm.Event) {
	//fmt.Printf("Client %s Disconnecting.\n", c.ID)
}

type ClientPool struct {
	sync.RWMutex
	Name        string
	clientQueue chan *Client
	clients     []*Client
}

func NewClientPool(name string) *ClientPool {
	cp := new(ClientPool)
	cp.Name = name
	cp.clientQueue = make(chan *Client, 1024)
	return cp
}

func (ch *ClientManager) poolsReporter() {
	for {
		num := <-ch.poolsUpdated
		if num == 0 {
			ch.poolStatus()
		}

	}
}

func (ch *ClientManager) GetAddrStr() string {
	return fmt.Sprintf("%s:%d", ch.Addr, ch.Port)
}

func (ch *ClientManager) Start() {
	ln, err := net.Listen("tcp", ch.GetAddrStr())
	if err != nil {
		log.Fatalf("Failed to listen: %s", err)
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("Failed to accept: %s", err)
			continue
		}
		go ch.handleConnection(conn)
	}
}

func (ch *ClientManager) AddClient(conn net.Conn) *Client {
	c := NewClient(conn)
	go ch.ClientStatusChanged(c)
	go c.messageSender()
	return c
}

func (ch *ClientManager) getAvailableClient() *Client {
	for _, v := range ch.Pools["READY"].Items {
		client := v.(*Client)
		if client.State == READY {
			return client
		}
	}
	return nil
}

func (ch *ClientManager) poolStatus() {
	for p, v := range ch.Pools {
		fmt.Printf("Pool: %s, numClients: %d\n", p, len(v.Items))
	}
}
func (ch *ClientManager) ClientStatusChanged(c *Client) error {
	for {
		status := <-c.statusUpdate
		fmt.Printf("Got status update. %s\n", status)
		switch status {
		case ONLINE:
			ch.MoveClientToPool(c, "ONLINE")
		case OFFLINE:
			ch.MoveClientToPool(c, "OFFLINE")
		case READY:
			ch.MoveClientToPool(c, "READY")
		case WORKING:
			ch.MoveClientToPool(c, "WORKING")
		case SERVICE:
			ch.MoveClientToPool(c, "SERVICE")
		case DONE:
			ch.MoveClientToPool(c, "DONE")
		}
	}
	return nil
}

func (c *Client) messageSender() {
	for {
		message := <-c.sender
		data, err := proto.Marshal(message)
		if err != nil {
			log.WithFields(log.Fields{"module": "clientmanager"}).Error(err)
			c.writer.Flush()
			continue
		}
		fmt.Printf("About to send %d bytes\n", len(data))
		n, err := c.writer.Write(data)
		fmt.Printf("%d bytes written.\n", n)
		if err != nil {
			log.WithFields(log.Fields{"module": "clientmanager"}).Error(err)
			c.writer.Flush()
			continue

		}
		c.writer.Flush()
	}
}

func (cm *ClientManager) handleClientMessage(c *Client, message *KamajiMessage) {
	switch message.GetAction() {
	case KamajiMessage_STATUS_UPDATE:
		{
			status := message.GetStatusupdate()
			if State(status.GetDestination()) == READY {
				c.FSM.Event("ready")
			}
			if State(status.GetDestination()) == SERVICE {
				c.FSM.Event("service")
			}
		}
	case KamajiMessage_QUERY:
		{
			message := &KamajiMessage{
				Action: KamajiMessage_QUERY.Enum(),
				Entity: KamajiMessage_CLIENT.Enum(),
			}
			for p, _ := range cm.itemToPool {
				client := p.(*Client)
				clientItem := &KamajiMessage_ClientItem{
					Id:    proto.String(client.ID.String()),
					State: proto.Int32(int32(client.State)),
				}
				message.Messageitems = append(message.Messageitems, clientItem)
			}
			c.sender <- message
		}
	}
}

func (cm *ClientManager) handleCommandMessage(c *Client, message *KamajiMessage) {
	switch message.GetAction() {
	case KamajiMessage_STATUS_UPDATE:
		{
			status := message.GetStatusupdate()
			if State(status.GetDestination()) == DONE {
				CommandEvent <- message
				c.FSM.Event("ready")
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
			c.sender <- message
		}
	}
}

func (cm *ClientManager) handleMessage(c *Client, message *KamajiMessage) error {
	switch message.GetEntity() {
	case KamajiMessage_CLIENT:
		cm.handleClientMessage(c, message)
	case KamajiMessage_COMMAND:
		cm.handleCommandMessage(c, message)
	}
	return nil
}

func (c *Client) assignCommand(command *Command) error {
	message := &KamajiMessage{
		Action: KamajiMessage_ASSIGN.Enum(),
		Entity: KamajiMessage_COMMAND.Enum(),
		Id:     proto.String(command.ID.String()),
	}
	log.WithFields(log.Fields{"module": "clientmanager", "client": c.ID, "command": command.Name}).Debug("Assigning command to client.")
	c.sender <- message
	return nil
}

func (c *Client) requestStatusUpdate(state State) error {
	log.WithFields(log.Fields{"module": "clientmanager", "client": c.ID, "status": c.State}).Debug("State online, request status update.")
	message := &KamajiMessage{
		Action: KamajiMessage_STATUS_UPDATE.Enum(),
		Entity: KamajiMessage_CLIENT.Enum(),
		Id:     proto.String(c.ID.String()),
		Statusupdate: &KamajiMessage_StatusUpdate{
			Destination: proto.Int32(int32(state)),
		},
	}
	c.sender <- message
	return nil
}

func (c *Client) recieveMessage() (*KamajiMessage, error) {
	message := &KamajiMessage{}
	tmp := make([]byte, 4096)
	n, err := c.reader.Read(tmp)
	if err != nil {
		if err == io.EOF {
			return nil, err
		}
		fmt.Println(err)
		return nil, err
	}
	err = proto.Unmarshal(tmp[0:n], message)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return message, nil
}

func (cm *ClientManager) handleConnection(conn net.Conn) {
	client := cm.AddClient(conn)
	client.FSM.Event("online")
	defer client.FSM.Event("offline")
	for {
		if client.State == ONLINE {
			err := client.requestStatusUpdate(READY)
			if err != nil {
				log.WithFields(log.Fields{"module": "clientmanager", "client": client.ID}).Error(err)
			}
		}
		message, err := client.recieveMessage()
		if err != nil {
			if err == io.EOF {
				log.WithFields(log.Fields{"module": "clientmanager", "client": client.ID}).Error(err)
				client.FSM.Event("disconnect")
			}
			break
		}
		err = cm.handleMessage(client, message)
		if err != nil {
			log.WithFields(log.Fields{"module": "clientmanager", "client": client.ID}).Error(err)
		}
		/*
			header := &HeaderMsg{}
			err := client.decoder.Decode(header)
			if err != nil {
				if err == io.EOF {
					//log.WithFields(log.Fields{"module": "clientmanager"}).Error(err)
					client.FSM.Event("disconnect")
				}
				break
			}
			if header.Msg == CLIENT_ANNOUNCE {
				msg := &ClientAnnounceMsg{}
				err := client.decoder.Decode(msg)
				if err != nil {
					log.WithFields(log.Fields{"module": "clientmanager"}).Error(err)
				}
				if msg.State != client.State {
					log.WithFields(log.Fields{"module": "clientmanager", "client": client.ID, "from": client.State, "to": msg.State}).Debug("Client Reported for duty.")
					switch msg.State {
					case READY:
						client.FSM.Event("ready")
					}

				}
			}

			if header.Msg == TASK_STATUS {
				msg := &TaskMsg{}
				err := client.decoder.Decode(msg)
				if err != nil {
					log.WithFields(log.Fields{"module": "clientmanager"}).Error(err)
				}
				if msg.State == DONE {
					CommandEvent <- msg
					client.FSM.Event("ready")
				}
			}
		*/
		/*
			if message.Action == "stats" {
				w := new(tabwriter.Writer)
				w.Init(os.Stdout, 0, 8, 0, '\t', 0)
				fmt.Fprintln(w)
				fmt.Fprintln(w, "hostname\tuptime\tos\tplatform\t.")
				fmt.Fprintf(w, "%s\t%d\t%s\t%s\t.", message.Stats.Host.Hostname, message.Stats.Host.Uptime, message.Stats.Host.OS, message.Stats.Host.Platform)
				fmt.Fprintln(w)
				fmt.Fprintln(w)
				w.Flush()
			}
		*/
	}
}

type ClientPoolHandler struct {
}
