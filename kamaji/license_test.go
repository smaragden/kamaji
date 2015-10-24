package kamaji_test

import (
	"github.com/smaragden/kamaji/kamaji"
	//"math/rand"
	"sync"
	"testing"
	"time"
)

func TestNewLicenceManager(t *testing.T) {
	want := 0
	lm := kamaji.NewLicenseManager()
	got := len(lm.Applications)
	if got != want {
		t.Errorf("Applications = %q, want %q", got, want)
	}

	lm.AddApplication("maya", 5)
	want = 1
	got = len(lm.Applications)
	if got != want {
		t.Errorf("Applications = %d, want %d", got, want)
	}

	want = 0
	got = lm.AddApplication("maya", 0)
	if got != want {
		t.Errorf("Applications = %d, want %d", got, want)
	}

	want = 1
	got = len(lm.Applications)
	if got != want {
		t.Errorf("Applications = %d, want %d", got, want)
	}
	t.Logf("%+v", lm.Status("maya"))
	want = 1
	got, _ = lm.Borrow("maya")
	if got != want {
		t.Errorf("Applications = %d, want %d", got, want)
	}
	t.Logf("%+v", lm.Status("maya"))
	want = 1
	got, _ = lm.Borrow("maya")
	if got != want {
		t.Errorf("Applications = %d, want %d", got, want)
	}
	t.Logf("%+v", lm.Status("maya"))
	want = 1
	got, _ = lm.Return("maya")
	if got != want {
		t.Errorf("Applications = %d, want %d", got, want)
	}
	t.Logf("%+v", lm.Status("maya"))
}

func f(n int, lm *kamaji.LicenseManager, t *testing.T, wg *sync.WaitGroup) {
	for {
		_, ok := lm.Borrow("arnold")
		if ok {
			_, _ = lm.Return("arnold")
			break
		} else {
			time.Sleep(time.Millisecond)
		}
	}
	wg.Done()
}

func TestConcurrency(t *testing.T) {
	licenses := 1000
	gouroutines := 1000000
	lm := kamaji.NewLicenseManager()
	_ = lm.AddApplication("arnold", licenses)
	var wg sync.WaitGroup
	wg.Add(gouroutines)
	for i := 0; i < gouroutines; i++ {
		go f(i, lm, t, &wg)
	}
	wg.Wait()
	t.Logf("FINISHED! | %d goroutines | %d licenses | %+v", gouroutines, licenses, lm.Status("arnold"))
}
