//go:build linux

package linux

import (
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/abhisheksinha1506/lightsout/pkg/wait"
)

// End-to-end check of the real wmctrl/xprop code: open an xterm, find it in
// Applications, close it with Close, and check the process exits. It needs an
// X display with a window manager, wmctrl, xprop and xterm, so it only runs
// when LIGHTSOUT_LINUX_E2E=1 (as in CI).
func TestRealXtermIsListedAndCloses(t *testing.T) {
	if os.Getenv("LIGHTSOUT_LINUX_E2E") != "1" {
		t.Skip("set LIGHTSOUT_LINUX_E2E=1 to run (opens and closes an xterm)")
	}

	xterm := exec.Command("xterm", "-title", "lightsout e2e")
	if err := xterm.Start(); err != nil {
		t.Fatal(err)
	}
	exited := make(chan error, 1)
	go func() { exited <- xterm.Wait() }()
	// Only matters if the test fails before the xterm was closed.
	t.Cleanup(func() { _ = xterm.Process.Kill() })

	var windows []string
	found := wait.Until(15*time.Second, 200*time.Millisecond, func() bool {
		apps, err := Applications()
		if err != nil {
			return false
		}
		windows = apps["XTerm"]
		return len(windows) > 0
	})
	if !found {
		apps, err := Applications()
		t.Fatalf("XTerm not listed after 15s; listed apps: %v (err %v)", apps, err)
	}

	if err := Close(windows); err != nil {
		t.Fatalf("Close(%v) = %v", windows, err)
	}
	select {
	case err := <-exited:
		t.Logf("xterm exited after its window closed (%v)", err)
	case <-time.After(10 * time.Second):
		t.Fatal("xterm still running 10s after its window closed")
	}
}
