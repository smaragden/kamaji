package kamaji_test

import (
	"github.com/smaragden/kamaji/kamaji"
	"testing"
)

func TestDispatcherLoop(t *testing.T) {
	dispatcher := kamaji.NewDispatcher()
	t.Logf("Dispatcher %+v", dispatcher)
}
