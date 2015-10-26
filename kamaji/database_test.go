package kamaji_test

import (
	"github.com/smaragden/kamaji/kamaji"
	"testing"
)

func TestDatabase(t *testing.T) {
	db := kamaji.NewDatabase()
	db.Connect("localhost:6379")
	lm := kamaji.NewLicenseManager()
	lm.AddApplication("maya", 12)
	lm.AddApplication("arnold", 20)
	lm.AddApplication("nuke", 6)
	lm.Store()
}
