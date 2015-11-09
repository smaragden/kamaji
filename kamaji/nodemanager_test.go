package kamaji_test

import (
    "github.com/smaragden/kamaji/kamaji"
    "testing"
    "time"
    "net"
)

func client(t *testing.T) {
    conn, err := net.Dial("tcp", ":3000")
    if err != nil {
        t.Fatal(err)
    }
    defer conn.Close()
    time.Sleep(time.Second * 10)
}

func TestNodeManager(t *testing.T) {
    nm := kamaji.NewNodeManager("", 3000)
    go nm.Start()
    time.Sleep(time.Millisecond * 10)
    go client(t)
    time.Sleep(time.Millisecond * 10)
    nm.Stop()
}

