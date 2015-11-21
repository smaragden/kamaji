package kamaji_test

import (
    "github.com/smaragden/kamaji/kamaji"
    "testing"
    "time"
    "net"
    "strconv"
)

func client(t *testing.T, fall bool, port string) {
    conn, err := net.Dial("tcp", ":"+port)
    if err != nil {
        t.Fatal(err)
    }
    defer conn.Close()
    if !fall {
        for{
            buf := make([]byte, 0, 4096)
            _, err :=conn.Read(buf)
            if err != nil {
                break
            }
        }
    }
}

// Test Bringing up 100 nodes and then stop the nodemanager.
func TestNodeManager(t *testing.T) {
    port := 1414
    numClients := 100
    nm := kamaji.NewNodeManager("", port)
    go nm.Start()
    time.Sleep(time.Millisecond * 10)
    for i := 0; i < numClients; i++{
        go client(t, false, strconv.Itoa(port))
        time.Sleep(time.Millisecond * 5)
    }
    for {
        if len(nm.Nodes) == numClients{
            break
        }
    }
    nm.Stop()
}

/*
// Test Bringing up 100 nodes and then stop the nodemanager.
func TestNodeManagerStop(t *testing.T) {
    port := 1415
    numClients := 100
    nm := kamaji.NewNodeManager("", port)
    go nm.Start()
    time.Sleep(time.Millisecond * 10)
    for i := 0; i < numClients; i++{
        go client(t, false, strconv.Itoa(port))
    }
    for {
        if len(nm.Nodes) > 50 {
            break
        }
        time.Sleep(time.Millisecond * 5)
    }
    nm.Stop()
}


// Test Bringing up 100 nodes and let them exit, then stop the nodemanager.
func TestNodeManagerFallthrough(t *testing.T) {
    numClients := 100
    port := 1416
    nm := kamaji.NewNodeManager("", port)
    go nm.Start()
    time.Sleep(time.Millisecond * 10)
    defer nm.Stop()
    for i := 0; i < numClients; i++{
        go client(t, true, strconv.Itoa(port))
    }
    for {
        if len(nm.Nodes) == numClients{
            return
        }
        time.Sleep(time.Millisecond * 5)
    }
}
*/
