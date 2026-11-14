package macos

import (
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/abhisheksinha1506/lightsout/pkg/wait"
)

// End-to-end check of the real NSWorkspace code: open TextEdit, find it in
// Applications, quit it with Quit, and check every TextEdit process exited.
// It quits TextEdit, so it only runs when LIGHTSOUT_MACOS_E2E=1 (as in CI).
func TestRealTextEditIsListedAndQuits(t *testing.T) {
	if os.Getenv("LIGHTSOUT_MACOS_E2E") != "1" {
		t.Skip("set LIGHTSOUT_MACOS_E2E=1 to run (opens and quits TextEdit)")
	}

	if err := exec.Command("open", "-a", "TextEdit").Run(); err != nil {
		t.Fatal(err)
	}

	var pids []int
	found := wait.Until(30*time.Second, 200*time.Millisecond, func() bool {
		apps, err := Applications()
		if err != nil {
			return false
		}
		pids = apps["TextEdit"]
		return len(pids) > 0
	})
	if !found {
		apps, err := Applications()
		t.Fatalf("TextEdit not listed after 30s; listed apps: %v (err %v)", apps, err)
	}

	if err := Quit(pids, 15*time.Second); err != nil {
		t.Fatalf("Quit(%v) = %v", pids, err)
	}
	for _, pid := range pids {
		if isRunning(pid) {
			t.Errorf("TextEdit process %d still running after Quit", pid)
		}
	}
}
