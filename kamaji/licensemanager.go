package kamaji

import (
    "bytes"
    "fmt"
    "sync"
    log "github.com/Sirupsen/logrus"
)

func init() {
    level, err := log.ParseLevel(Config.Logging.Licensemanager)
    if err == nil {
        log.SetLevel(level)
    }
}
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

func (nm LicenseManager) Start() {
    log.WithFields(log.Fields{
        "module":  "licensemanager",
        "action":  "start",
    }).Info("Starting License Manager.")
}

func (lm LicenseManager) Stop() {
    log.WithFields(log.Fields{
        "module":  "licencemanager",
        "action":  "stop",
    }).Info("Stopping License Manager.")
}