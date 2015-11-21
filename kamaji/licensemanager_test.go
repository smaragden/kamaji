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