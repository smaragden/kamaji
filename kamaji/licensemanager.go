package kamaji

import (
	"errors"
	"fmt"
)

type License struct {
	Name string
	Count uint16
	Queue chan bool
}

func NewLicense(name string, count uint16) *License {
	l := new(License)
	l.Name = name
	l.Count = count
	l.Queue = make(chan bool, count)
	for i := uint16(0); i < count; i++ {
		l.Queue <- true
	}
	return l
}

// Borrow a license and return true if succeeded and false if failed.
func (l *License) Borrow() bool {
	select {
	case res := <-l.Queue:
		return res
	//case <-time.After(time.Duration(1e6 * int64(Config.LicenseManager.Interval))):
	default:
		return false
	}
}

// Return a license and return true if succeeded and false if failed.
func (l *License) Return() bool {
	select {
	case l.Queue <- true:
		return true
	default:
		return false
	}
}

// Get available licenses
func (l *License) Available() uint16 {
	return uint16(len(l.Queue))
}

type LicenseManager struct {
	Licenses map[string]*License
}

func NewLicenseManager() *LicenseManager {
	lm := new(LicenseManager)
	lm.Licenses = make(map[string]*License)
	return lm
}

func (lm *LicenseManager) AddLicense(name string, count uint16) error {
	_, ok := lm.Licenses[name]
    if ok {
         return errors.New(fmt.Sprintf("License already exists: %s", name))
    }
    lm.Licenses[name] = NewLicense(name, count)
	return nil
}

// Borrow a license and return true if succeeded and false if failed.
func (lm *LicenseManager) Borrow(name string) bool {
	return lm.Licenses[name].Borrow()
}

// Borrow multiple licenses and return true if succeeded and false if failed.
// We also return the licenses we already aquired if we fail
func (lm *LicenseManager) borrowMultiple(names []string) bool {
	var aquired_licenses []string
	for _, name := range names {
		ok := lm.Licenses[name].Borrow()
		if !ok{
			for _, lic := range aquired_licenses {
				lm.Return(lic)
			}
			return false
		}
		aquired_licenses = append(aquired_licenses, name)
	}
	return true
}

// Return a license and return true if succeeded and false if failed.
func (lm *LicenseManager) Return(name string) bool {
	return lm.Licenses[name].Return()
}

func (lm *LicenseManager) Available(name string) uint16 {
	return lm.Licenses[name].Available()
}

func (lm *LicenseManager) MatchRequirements(licenses []string) ([]*License, error){
	// First check if we got available licenses
	aquired_licenses := make([]*License, len(licenses))
	for i, lic := range licenses {
		if lm.Available(lic) == 0{
			return nil, errors.New("Couldn't match requirements.")
		}
		aquired_licenses[i] = lm.Licenses[lic]
	}
	// Aquire the licenses
	ok := lm.borrowMultiple(licenses)
	if !ok {
		return nil, errors.New("Couldn't aquire all licenses.")
	}
	return aquired_licenses, nil
}