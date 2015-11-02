package kamaji_test

import (
	"fmt"
	"github.com/smaragden/kamaji/kamaji"
	"net"
	"testing"
)

type ClientManager struct {
	*kamaji.PoolManager
	Addr         string
	Port         int
	poolsUpdated chan int
}

func NewClientManager(name string, address string, port int) *ClientManager {
	cm := new(ClientManager)
	cm.PoolManager = kamaji.NewPoolManager(name)
	cm.Addr = address
	cm.Port = port
	cm.CreatePools([]string{kamaji.OFFLINE.String(), kamaji.ONLINE.String(), kamaji.AVAILABLE.String(), kamaji.WORKING.String()})
	return cm
}

type Test struct {
	Name        string
	currentPool string
}

func (t *Test) SetCurrentPool(pool string) {
	t.currentPool = pool
}

func (t *Test) IsEqual(other *Test) bool {
	return t.Name == other.Name
}

////////////////////////////////////
func TestPoolManager(t *testing.T) {
	pool_manager := kamaji.NewClientManager("TestPoolManager", "test", 1234)
	fmt.Printf("%s %+v %T\n%s %+v %T\n", pool_manager.Name, pool_manager, pool_manager, pool_manager.PoolManager.Name, pool_manager.PoolManager, pool_manager)
	fmt.Printf("Pools: %+v\n", pool_manager.Pools)
	conn, err := net.Dial("tcp", ":3000")
	if err != nil {
		fmt.Println(err)
	}
	test := kamaji.NewClient(conn)
	pool_manager.MoveClientToPool(test, "AVAILABLE")
	for i, pool := range pool_manager.Pools {
		fmt.Printf("%q, %+v\n", i, pool)
		for j, item := range pool.Items {
			fmt.Printf("\t%d, %+v\n", j, item)
		}
	}

	pool_manager.MoveClientToPool(test, "WORKING")
	for i, pool := range pool_manager.Pools {
		fmt.Printf("%q, %+v\n", i, pool)
		for j, item := range pool.Items {
			fmt.Printf("\t%d, %+v\n", j, item)
		}
	}
	//eq := test.isEqual(pool_manager.Pools["AVAILABLE"].Items[0])
	//fmt.Printf("is equal %q\n", eq)
}
