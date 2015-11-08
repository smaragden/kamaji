package kamaji_test

import (
	"github.com/smaragden/kamaji/kamaji"
	"math/rand"
	"net"
	"testing"
)

var tests = []string{
	RandStringBytes(10),
	RandStringBytes(625000), // 5Mb
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func RandStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func TestClientMessage(t *testing.T) {
	for i, message_string := range tests {
		message := []byte(message_string)
		go func() {
			conn, err := net.Dial("tcp", ":3000")
			if err != nil {
				t.Fatal(err)
			}
			defer conn.Close()
			client := kamaji.NewClient(conn)
			n, err := client.SendMessage(message)
			t.Logf("Bytes Written: %d\n", n)
		}()

		l, err := net.Listen("tcp", ":3000")
		if err != nil {
			t.Fatal(err)
		}

		for {
			conn, err := l.Accept()
			if err != nil {
				return
			}
			defer conn.Close()
			client := kamaji.NewClient(conn)
			recv_message, err := client.ReadMessage()
			recv_string := string(recv_message[:])
			if err != nil {
				t.Error(err)
			}
			if recv_string != message_string {
				t.Errorf("recv_message = %s, want %s", recv_string[:10], message_string[:10])
				for i, ch := range message {
					if ch != recv_message[i] {
						t.Logf("%d, %d==%d\n", i-1, message[i-1], recv_message[i-1])
						t.Logf("Diff occured at index: %d, %d!=%d\n", i, message[i], recv_message[i])
						t.Logf("%d, %d!=%d\n", i+1, message[i+1], recv_message[i+1])
						break
					}
				}
			} else {
				t.Logf("Test %d passed.", i+1)
			}
			break
		}
		l.Close()
	}
}
