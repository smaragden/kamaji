package kamaji

import (
	"bytes"
	"fmt"
	"sync"
)

type Application struct {
	sync.RWMutex
	name      string
	count     int
	available int
}

func NewApplication(name string, count int) *Application {
	a := new(Application)
	a.name = name
	a.count = count
	a.available = count
	return a
}

func (a *Application) Borrow() (int, bool) {
	if a.available > 0 {
		a.available--
		return 1, true
	}
	return 0, false
}

func (a *Application) Return() (int, bool) {
	if a.available < a.count {
		a.available++
		return 1, true
	}
	return 0, false
}

type LicenseManager struct {
	Applications map[string]*Application
}

func NewLicenseManager() *LicenseManager {
	lm := new(LicenseManager)
	lm.Applications = make(map[string]*Application)
	return lm
}

func (lm LicenseManager) lkey() string {
	return "licenses"
}

func (lm LicenseManager) akey(app *Application) string {
	var buffer bytes.Buffer
	buffer.WriteString(lm.lkey())
	buffer.WriteString(":")
	buffer.WriteString(app.name)
	return buffer.String()
}

func (lm LicenseManager) AddApplication(name string, count int) int {
	_, ok := lm.Applications[name]
	if ok {
		_ = fmt.Errorf("Application: %q already exists!", name)
		return 0
	}
	lm.Applications[name] = NewApplication(name, count)
	return count
}

func (lm LicenseManager) Borrow(name string) (int, bool) {
	app, ok := lm.Applications[name]
	if ok {
		app.Lock()
		n, err := app.Borrow()
		app.Unlock()
		return n, err
	}
	return 0, false
}

func (lm LicenseManager) Return(name string) (int, bool) {
	app, ok := lm.Applications[name]
	if ok {
		app.Lock()
		n, err := app.Return()
		app.Unlock()
		return n, err
	}
	return 0, false
}

func (lm LicenseManager) Status(name string) Application {
	app, ok := lm.Applications[name]
	if ok {
		return *app
	}
	return *app
}

func (lm LicenseManager) Store() bool {
	db := NewDatabase()
	for _, app := range lm.Applications {
		_, err := db.Client.Do("HSET", lm.akey(app), "count", app.count)
		if err != nil {
			panic(err)
		}
	}
	return true
}

/*
func (lm LicenseManager) Store() bool {
	db := NewDatabase()
	for _, app := range lm.Applications {
		_, err := db.Client.Do("SADD", lm.akey(app), app.count)
		if err != nil {
			panic(err)
		}
	}
	return true
}
*/
