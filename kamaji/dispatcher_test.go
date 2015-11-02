package kamaji_test

import (
	"encoding/gob"
	"fmt"
	"github.com/smaragden/kamaji/kamaji"
	"io"
	"log"
	"math/rand"
	"net"
	"testing"
	"time"
)

type Message struct {
	Action string
}

type ClientConn struct {
	net.Conn
	encoder *gob.Encoder
	decoder *gob.Decoder
}

func TestDispatcher(t *testing.T) {
	fmt.Println("Starting")
	lm := kamaji.NewLicenseManager()
	ch := kamaji.NewClientManager("ClientManager", "", 1314)
	go ch.Start()
	tm := kamaji.NewTaskManager()
	d := kamaji.NewDispatcher(lm, ch, tm)
	job_count := 10
	task_count := 5
	command_count := 20
	for i := 0; i < job_count; i++ {
		job := kamaji.NewJob(fmt.Sprintf("Job %d", i))
		for j := 0; j < task_count; j++ {
			task := kamaji.NewTask(fmt.Sprintf("Task %d", j), job)
			for k := 0; k < command_count; k++ {
				_ = kamaji.NewCommand(fmt.Sprintf("Command %d", k), task)
			}
		}
		tm.AddJob(job)
	}
	go d.Start()

	client_count := 1
	fmt.Println("Starting: ", client_count, " clients.")
	//var wg sync.WaitGroup
	//wg.Add(client_count)
	quit := make(chan bool, client_count)
	for i := 0; i < client_count; i++ {
		go cli(i, quit)
		time.Sleep(time.Millisecond * 2)
	}
	//wg.Wait()
	time.Sleep(time.Second * 10)
	for i := 0; i < client_count; i++ {
		quit <- true
	}
	fmt.Println("Exiting")
}

func cli(cn int, quit chan bool) {
	time.Sleep(time.Millisecond * time.Duration(rand.Intn(1000)))
	conn, err := net.Dial("tcp", "127.0.0.1:1314")
	if err != nil {
		fmt.Println("Error!")
		return
	}
	defer conn.Close()
	clientConn := new(ClientConn)
	clientConn.Conn = conn
	clientConn.encoder = gob.NewEncoder(conn)
	clientConn.decoder = gob.NewDecoder(conn)
	for {
		select {
		case <-quit:
			break
		default:
			{
				message := &Message{}
				err = clientConn.decoder.Decode(message)
				if err != nil {
					if err == io.EOF {
						//log.Printf("Connection Lost: %s", err)
					}
					break
				}
				if message.Action == "report" {
					//fmt.Println("Server asked me to report.")
					err := clientConn.encoder.Encode(Message{Action: "report"})
					if err != nil {
						log.Fatal("encode error:", err)
					}
				}
			}
		}
	}
}
