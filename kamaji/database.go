package kamaji

import (
	"github.com/garyburd/redigo/redis"
	"sync"
)

//interface Database
type Database struct {
	Client redis.Conn
}

var _initCtx sync.Once
var _instance *Database

func NewDatabase() *Database {
	_initCtx.Do(func() { _instance = new(Database) })
	return _instance
}

func (db *Database) Connect(Addr string) *Database {
	c, err := redis.Dial("tcp", Addr)
	if err == nil {
		db.Client = c
	}

	return db
}
