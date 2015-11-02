package kamaji

import (
	"reflect"
	"sync"
)

// Pool is a list of objects
type Pool struct {
	sync.RWMutex
	Name  string
	Items []interface{}
}

// Find an item in the pool and return it's index
// If the item is not found return -1
func (p *Pool) itemIndex(item interface{}) int {
	var value reflect.Value
	value = reflect.ValueOf(item)
	method := value.MethodByName("IsEqual")
	index := -1
	if method.IsValid() {
		for p, i := range p.Items {
			equal := method.Call([]reflect.Value{reflect.ValueOf(i)})[0]
			if equal.Bool() {
				index = p
				break
			}
		}
	}
	return index
}

// Push objects to the end of the pool
func (p *Pool) Push(item interface{}) interface{} {
	p.Lock()
	defer p.Unlock()
	p.Items = append(p.Items, item)
	return item
}

// Remove an item from the queue and retur it
func (p *Pool) Pop(item interface{}) interface{} {
	index := p.itemIndex(item)
	if index != -1 {
		p.Lock()
		defer p.Unlock()
		item := p.Items[index]
		p.Items = append(p.Items[:index], p.Items[index+1:]...)
		return item
	}
	return nil
}

// Remove an item from the pool
func (p *Pool) Remove(other interface{}) int {
	index := p.itemIndex(other)
	if index != -1 {
		p.Items = append(p.Items[:index], p.Items[index+1:]...)
	}
	return index
}

type PoolManager struct {
	Name       string
	Pools      map[string]*Pool
	itemToPool map[interface{}]*Pool
}

// The PoolManager controls the pools
func NewPoolManager(name string) *PoolManager {
	pm := new(PoolManager)
	pm.Name = name
	pm.Pools = make(map[string]*Pool)
	pm.itemToPool = make(map[interface{}]*Pool)
	return pm
}

// Create a new named pool.
func (pm *PoolManager) CreatePool(name string) *Pool {
	p := new(Pool)
	p.Name = name
	pm.Pools[name] = p
	return p
}

// Create a new named pools from slice of strings.
func (pm *PoolManager) CreatePools(pools []string) []*Pool {
	result := make([]*Pool, len(pools))
	for i, name := range pools {
		p := new(Pool)
		p.Name = name
		pm.Pools[name] = p
		result[i] = p
	}
	return result
}

// Move an item from one pool to another.
func (pm *PoolManager) MoveItemToPool(item interface{}, pool string) error {
	current_pool := pm.getCurrentPool(item)
	if current_pool != nil {
		current_pool.Remove(item)
	}
	to_pool := pm.Pools[pool]
	to_pool.Push(item)
	pm.itemToPool[item] = pm.Pools[pool]
	return nil
}

// Get the pool a specific item lives in
func (pm *PoolManager) getCurrentPool(item interface{}) *Pool {
	pool, ok := pm.itemToPool[item]
	if ok == true {
		return pool
	}
	return nil
}
