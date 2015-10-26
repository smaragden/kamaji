package kamaji

import (
	"container/list"
	//"fmt"
)

type Dispatcher struct {
	Jobs *list.List
}

func NewDispatcher() *Dispatcher {
	d := new(Dispatcher)
	return d
}
