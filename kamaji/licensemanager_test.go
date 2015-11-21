package kamaji_test
import (
	"testing"
	"github.com/smaragden/kamaji/kamaji"
	"sync"
	"time"
)

func TestLicenseManageAsync(t *testing.T) {
	lm := kamaji.NewLicenseManager()
	lm.AddLicense("maya", 20)
	for {
		if lm.Borrow("maya"){
			t.Logf("Got a license. Licenses left: %d", lm.Available("maya"))
		}else{
			t.Logf("No more licenses available!")
			break
		}
	}
	lm.Return("maya")
	lm.Return("maya")
	lm.Return("maya")
	for {
		if lm.Borrow("maya"){
			t.Logf("Got a license. Licenses left: %d", lm.Available("maya"))
		}else{
			t.Logf("No more licenses available!")
			break
		}
	}
}

func licenseCheckout(n int, lm *kamaji.LicenseManager, t *testing.T, wg *sync.WaitGroup) {
	for {
		ok := lm.Borrow("maya")
		if ok {
			time.Sleep(time.Nanosecond)
			_ = lm.Return("maya")
			break
		}else{
			time.Sleep(time.Nanosecond*2)
		}
	}
	wg.Done()
}

func TestLicenseManageSync(t *testing.T) {
	var licenses uint16 = 1000
	goroutines := 10000
	lm := kamaji.NewLicenseManager()
	lm.AddLicense("maya", licenses)
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go licenseCheckout(i, lm, t, &wg)
	}
	wg.Wait()
	t.Logf("Licenses after %d goroutines fought for %d licenses: %d", goroutines, licenses, lm.Available("maya"))
	if lm.Available("maya") != licenses {
		t.Errorf("Expected: %d, got: %d", licenses, lm.Available("maya"))
	}
}

func TestLicenseManageMulti(t *testing.T) {
	lm := kamaji.NewLicenseManager()
	lm.AddLicense("maya", 2)
	lm.AddLicense("arnold", 2)
	lm.AddLicense("nuke", 2)
	lics, err := lm.MatchRequirements([]string{"maya","arnold","nuke"})
	if err != nil {
		t.Error(err)
	}
	if lm.Available("maya") != 1 {
		t.Errorf("Expected: 1, got: %d", lm.Available("maya"))
	}
	if lm.Available("arnold") != 1 {
		t.Errorf("Expected: 1, got: %d", lm.Available("arnold"))
	}
	if lm.Available("nuke") != 1 {
		t.Errorf("Expected: 1, got: %d", lm.Available("nuke"))
	}
	if lics[0].Name != "maya"{
		t.Errorf("Expected: maya, got: %s", lics[0].Name)
	}
	if lics[1].Name != "arnold"{
		t.Errorf("Expected: arnold, got: %s", lics[1].Name)
	}
	if lics[2].Name != "nuke"{
		t.Errorf("Expected: nuke, got: %s", lics[2].Name)
	}
	lics, _ = lm.MatchRequirements([]string{"maya","arnold","nuke"})
	lics, err = lm.MatchRequirements([]string{"maya","arnold","nuke"})
	if err == nil {
		t.Logf("I expected an error here.")
	}
}
