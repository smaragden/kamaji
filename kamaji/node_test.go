package kamaji_test

import (
	"github.com/smaragden/kamaji/kamaji"
	"sync"
	"testing"
	"time"
)

func changeNodeState(node *kamaji.Node, state string, t *testing.T, wg *sync.WaitGroup) {
	node.ChangeState <- state
	wg.Done()
}
func TestNodeState(t *testing.T) {
	node_count := 10
	var nodes []*kamaji.Node
	for i := 1; i < node_count+1; i++ {
		node := kamaji.NewNode(nil)
		nodes = append(nodes, node)
	}
	stateSequence := []string{"online", "ready"}
	var wg sync.WaitGroup
	for _, state := range stateSequence {
		for _, node := range nodes {
			wg.Add(1)
			go changeNodeState(node, state, t, &wg)
		}
		time.Sleep(time.Millisecond * 10)
	}
	wg.Wait()
	time.Sleep(time.Millisecond * 1000)
}
