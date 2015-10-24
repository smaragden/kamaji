package kamaji

import (
	"fmt"
)

type Application struct {
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
		return app.Borrow()
	}
	return 0, false
}

func (lm LicenseManager) Return(name string) (int, bool) {
	app, ok := lm.Applications[name]
	if ok {
		return app.Return()
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
	for name, app := range lm.Applications {
		err := db.Client.Set(name, app.available, 0).Err()
		if err != nil {
			panic(err)
		}

	}
	return true
}
