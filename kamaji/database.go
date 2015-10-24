package kamaji

import (
	"gopkg.in/redis.v3"
	"sync"
)

type Database struct {
	Client *redis.Client
}

var _initCtx sync.Once
var _instance *Database

func NewDatabase() *Database {
	_initCtx.Do(func() { _instance = new(Database) })
	return _instance
}

func (db *Database) Connect(Addr string) *Database {
	db.Client = redis.NewClient(&redis.Options{
		Addr:     Addr,
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	return db
}
